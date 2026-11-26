// Package data 是 SQLite 存储层：建表、种子数据与各表读写。
package data

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"

	"questioning/internal/model"
)

// ErrNotFound 表示按主键查不到记录。
var ErrNotFound = errors.New("记录不存在")

// Store 包住 *sql.DB，业务层只跟它打交道。
type Store struct {
	db *sql.DB
}

const schema = `
CREATE TABLE IF NOT EXISTS users (
	id            INTEGER PRIMARY KEY AUTOINCREMENT,
	username      TEXT    NOT NULL UNIQUE,
	password_hash TEXT    NOT NULL,
	nickname      TEXT    NOT NULL DEFAULT '',
	role          TEXT    NOT NULL DEFAULT 'user',
	created_at    INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS surveys (
	id           INTEGER PRIMARY KEY AUTOINCREMENT,
	owner_id     INTEGER NOT NULL,
	title        TEXT    NOT NULL,
	description  TEXT    NOT NULL DEFAULT '',
	status       TEXT    NOT NULL DEFAULT 'draft',
	anonymous    INTEGER NOT NULL DEFAULT 0,
	created_at   INTEGER NOT NULL,
	updated_at   INTEGER NOT NULL,
	published_at INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS questions (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	survey_id   INTEGER NOT NULL,
	type        TEXT    NOT NULL,
	title       TEXT    NOT NULL,
	description TEXT    NOT NULL DEFAULT '',
	required    INTEGER NOT NULL DEFAULT 0,
	order_no    INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_questions_survey ON questions(survey_id, order_no);

CREATE TABLE IF NOT EXISTS options (
	id               INTEGER PRIMARY KEY AUTOINCREMENT,
	question_id      INTEGER NOT NULL,
	label            TEXT    NOT NULL,
	next_question_id INTEGER NOT NULL DEFAULT 0,
	order_no         INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_options_question ON options(question_id, order_no);

CREATE TABLE IF NOT EXISTS responses (
	id           INTEGER PRIMARY KEY AUTOINCREMENT,
	survey_id    INTEGER NOT NULL,
	submitted_at INTEGER NOT NULL,
	duration_sec INTEGER NOT NULL DEFAULT 0,
	device       TEXT    NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_responses_survey ON responses(survey_id, submitted_at);

CREATE TABLE IF NOT EXISTS answers (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	response_id INTEGER NOT NULL,
	question_id INTEGER NOT NULL,
	option_ids  TEXT    NOT NULL DEFAULT '',
	text        TEXT    NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_answers_response ON answers(response_id);
CREATE INDEX IF NOT EXISTS idx_answers_question ON answers(question_id);
`

func nowUnix() int64 { return time.Now().Unix() }

// Open 打开（不存在则创建）数据库并建表。
func Open(path string) (*Store, error) {
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("创建数据目录失败: %w", err)
		}
	}
	dsn := path + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}
	// SQLite 写入串行，单连接足够且避免 database is locked
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("建表失败: %w", err)
	}
	return &Store{db: db}, nil
}

// Close 关库。
func (s *Store) Close() error { return s.db.Close() }

// SeedUser 是演示账号描述。
type SeedUser struct {
	Username string
	Password string
	Nickname string
	Role     string
}

// Seed 在用户表为空时写入演示账号，返回是否写入了数据。
func (s *Store) Seed(users []SeedUser) (bool, error) {
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&n); err != nil {
		return false, err
	}
	if n > 0 {
		return false, nil
	}
	for _, u := range users {
		hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return false, err
		}
		if _, err := s.CreateUser(u.Username, string(hash), u.Nickname, u.Role); err != nil {
			return false, err
		}
	}
	return true, nil
}

// ---- 用户 ----

// CreateUser 插入用户。
func (s *Store) CreateUser(username, passwordHash, nickname, role string) (model.User, error) {
	u := model.User{Username: username, PasswordHash: passwordHash, Nickname: nickname, Role: role, CreatedAt: nowUnix()}
	res, err := s.db.Exec(`INSERT INTO users(username, password_hash, nickname, role, created_at) VALUES(?,?,?,?,?)`,
		u.Username, u.PasswordHash, u.Nickname, u.Role, u.CreatedAt)
	if err != nil {
		return model.User{}, err
	}
	u.ID, err = res.LastInsertId()
	return u, err
}

const userCols = `id, username, password_hash, nickname, role, created_at`

func scanUser(row interface{ Scan(...any) error }) (model.User, error) {
	var u model.User
	err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Nickname, &u.Role, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return u, ErrNotFound
	}
	return u, err
}

// UserByUsername 按用户名查（登录用）。
func (s *Store) UserByUsername(username string) (model.User, error) {
	return scanUser(s.db.QueryRow(`SELECT `+userCols+` FROM users WHERE username=?`, username))
}

// UserByID 按主键查。
func (s *Store) UserByID(id int64) (model.User, error) {
	return scanUser(s.db.QueryRow(`SELECT `+userCols+` FROM users WHERE id=?`, id))
}

// ---- 问卷 ----

const surveyCols = `id, owner_id, title, description, status, anonymous, created_at, updated_at, published_at`

func scanSurvey(row interface{ Scan(...any) error }, withCount bool) (model.Survey, error) {
	var (
		sv      model.Survey
		anony   int
		count   int
		targets = []any{&sv.ID, &sv.OwnerID, &sv.Title, &sv.Description, &sv.Status, &anony, &sv.CreatedAt, &sv.UpdatedAt, &sv.PublishedAt}
	)
	if withCount {
		targets = append(targets, &count)
	}
	err := row.Scan(targets...)
	if errors.Is(err, sql.ErrNoRows) {
		return sv, ErrNotFound
	}
	sv.Anonymous = anony == 1
	sv.ResponseCount = count
	return sv, err
}

// CreateSurvey 建一份草稿问卷。
func (s *Store) CreateSurvey(ownerID int64, title, description string, anonymous bool) (model.Survey, error) {
	now := nowUnix()
	anony := 0
	if anonymous {
		anony = 1
	}
	res, err := s.db.Exec(`INSERT INTO surveys(owner_id,title,description,status,anonymous,created_at,updated_at) VALUES(?,?,?,'draft',?,?,?)`,
		ownerID, title, description, anony, now, now)
	if err != nil {
		return model.Survey{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return model.Survey{}, err
	}
	return s.SurveyByID(id)
}

// UpdateSurvey 改标题/说明/匿名设置。
func (s *Store) UpdateSurvey(id int64, title, description string, anonymous bool) error {
	anony := 0
	if anonymous {
		anony = 1
	}
	_, err := s.db.Exec(`UPDATE surveys SET title=?, description=?, anonymous=?, updated_at=? WHERE id=?`,
		title, description, anony, nowUnix(), id)
	return err
}

// SurveyByID 查单份问卷（不含题目）。
func (s *Store) SurveyByID(id int64) (model.Survey, error) {
	return scanSurvey(s.db.QueryRow(`SELECT `+surveyCols+` FROM surveys WHERE id=?`, id), false)
}

// ListSurveys 列出某用户的全部问卷，带答卷数。
func (s *Store) ListSurveys(ownerID int64) ([]model.Survey, error) {
	rows, err := s.db.Query(`SELECT `+surveyCols+`,
		(SELECT COUNT(*) FROM responses r WHERE r.survey_id = s.id)
		FROM surveys s WHERE owner_id=? ORDER BY updated_at DESC`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []model.Survey{}
	for rows.Next() {
		sv, err := scanSurvey(rows, true)
		if err != nil {
			return nil, err
		}
		out = append(out, sv)
	}
	return out, rows.Err()
}

// SetStatus 切状态，发布时补 published_at；关闭不清空它。
func (s *Store) SetStatus(id int64, status string) error {
	if status == "published" {
		_, err := s.db.Exec(`UPDATE surveys SET status=?, published_at=?, updated_at=? WHERE id=?`,
			status, nowUnix(), nowUnix(), id)
		return err
	}
	_, err := s.db.Exec(`UPDATE surveys SET status=?, updated_at=? WHERE id=?`, status, nowUnix(), id)
	return err
}

// DeleteSurvey 删问卷，连带题目、选项、答卷与答案。
func (s *Store) DeleteSurvey(id int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmts := []string{
		`DELETE FROM answers WHERE response_id IN (SELECT id FROM responses WHERE survey_id=?)`,
		`DELETE FROM answers WHERE question_id IN (SELECT id FROM questions WHERE survey_id=?)`,
		`DELETE FROM responses WHERE survey_id=?`,
		`DELETE FROM options WHERE question_id IN (SELECT id FROM questions WHERE survey_id=?)`,
		`DELETE FROM questions WHERE survey_id=?`,
		`DELETE FROM surveys WHERE id=?`,
	}
	for _, q := range stmts {
		if _, err := tx.Exec(q, id); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// Questions 取一份问卷的题目与选项，按 order_no 排好。
func (s *Store) Questions(surveyID int64) ([]model.Question, error) {
	rows, err := s.db.Query(`SELECT id, survey_id, type, title, description, required, order_no
		FROM questions WHERE survey_id=? ORDER BY order_no, id`, surveyID)
	if err != nil {
		return nil, err
	}
	var out []model.Question
	index := map[int64]int{}
	for rows.Next() {
		var (
			q        model.Question
			required int
		)
		if err := rows.Scan(&q.ID, &q.SurveyID, &q.Type, &q.Title, &q.Description, &required, &q.OrderNo); err != nil {
			rows.Close()
			return nil, err
		}
		q.Required = required == 1
		index[q.ID] = len(out)
		out = append(out, q)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return out, nil
	}

	optRows, err := s.db.Query(`SELECT o.id, o.question_id, o.label, o.next_question_id, o.order_no
		FROM options o JOIN questions q ON q.id = o.question_id
		WHERE q.survey_id=? ORDER BY o.order_no, o.id`, surveyID)
	if err != nil {
		return nil, err
	}
	defer optRows.Close()
	for optRows.Next() {
		var o model.Option
		if err := optRows.Scan(&o.ID, &o.QuestionID, &o.Label, &o.NextQuestionID, &o.OrderNo); err != nil {
			return nil, err
		}
		if i, ok := index[o.QuestionID]; ok {
			out[i].Options = append(out[i].Options, o)
		}
	}
	return out, optRows.Err()
}

// ResponseIDs 返回某问卷全部答卷 ID（导出/统计时按需取）。
func (s *Store) ResponseIDs(surveyID int64) ([]int64, error) {
	rows, err := s.db.Query(`SELECT id FROM responses WHERE survey_id=? ORDER BY submitted_at DESC`, surveyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// placeholders 生成 "?,?,?"。
func placeholders(n int) string {
	if n <= 0 {
		return ""
	}
	return strings.TrimSuffix(strings.Repeat("?,", n), ",")
}
