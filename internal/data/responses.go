package data

import (
	"database/sql"
	"errors"

	"questioning/internal/model"
)

// NewAnswer 是提交答卷时的一道题作答。
type NewAnswer struct {
	QuestionID int64
	OptionIDs  []int64
	Text       string
}

// CreateResponse 写入一份答卷和它的全部答案，返回答卷 ID。
func (s *Store) CreateResponse(surveyID int64, durationSec int, device string, answers []NewAnswer) (int64, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`INSERT INTO responses(survey_id, submitted_at, duration_sec, device) VALUES(?,?,?,?)`,
		surveyID, nowUnix(), durationSec, device)
	if err != nil {
		return 0, err
	}
	responseID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	for _, a := range answers {
		if _, err := tx.Exec(`INSERT INTO answers(response_id, question_id, option_ids, text) VALUES(?,?,?,?)`,
			responseID, a.QuestionID, joinIDs(a.OptionIDs), a.Text); err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return responseID, nil
}

// ListResponses 取一份问卷的全部答卷元信息（统计与索引用）。
func (s *Store) ListResponses(surveyID int64) ([]model.Response, error) {
	rows, err := s.db.Query(`SELECT id, survey_id, submitted_at, duration_sec, device
		FROM responses WHERE survey_id=? ORDER BY submitted_at DESC, id DESC`, surveyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Response{}
	for rows.Next() {
		var r model.Response
		if err := rows.Scan(&r.ID, &r.SurveyID, &r.SubmittedAt, &r.DurationSec, &r.Device); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// ResponseByID 查单份答卷。
func (s *Store) ResponseByID(id int64) (model.Response, error) {
	var r model.Response
	err := s.db.QueryRow(`SELECT id, survey_id, submitted_at, duration_sec, device FROM responses WHERE id=?`, id).
		Scan(&r.ID, &r.SurveyID, &r.SubmittedAt, &r.DurationSec, &r.Device)
	if errors.Is(err, sql.ErrNoRows) {
		return r, ErrNotFound
	}
	return r, err
}

// AllAnswers 取一份问卷的全部答案，按答卷 ID 分组。
func (s *Store) AllAnswers(surveyID int64) (map[int64][]model.Answer, error) {
	rows, err := s.db.Query(`SELECT a.id, a.response_id, a.question_id, a.option_ids, a.text
		FROM answers a JOIN responses r ON r.id = a.response_id
		WHERE r.survey_id=? ORDER BY a.response_id, a.question_id`, surveyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[int64][]model.Answer{}
	for rows.Next() {
		var (
			a    model.Answer
			join string
		)
		if err := rows.Scan(&a.ID, &a.ResponseID, &a.QuestionID, &join, &a.Text); err != nil {
			return nil, err
		}
		a.OptionIDs = parseIDs(join)
		out[a.ResponseID] = append(out[a.ResponseID], a)
	}
	return out, rows.Err()
}

// AnswersByQuestion 取某道题的全部答案（统计文本题用）。
func (s *Store) AnswersByQuestion(questionID int64) ([]model.Answer, error) {
	rows, err := s.db.Query(`SELECT id, response_id, question_id, option_ids, text FROM answers WHERE question_id=?`, questionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Answer
	for rows.Next() {
		var (
			a    model.Answer
			join string
		)
		if err := rows.Scan(&a.ID, &a.ResponseID, &a.QuestionID, &join, &a.Text); err != nil {
			return nil, err
		}
		a.OptionIDs = parseIDs(join)
		out = append(out, a)
	}
	return out, rows.Err()
}
