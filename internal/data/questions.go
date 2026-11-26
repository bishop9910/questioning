package data

import (
	"strconv"

	"questioning/internal/model"
	"questioning/pkg/ds"
)

// SaveOptionInput 保存时的选项输入。
type SaveOptionInput struct {
	Label          string
	NextQuestionID int64
}

// SaveQuestionInput 保存时的题目输入：ID > 0 表示已有题目，ID <= 0 表示新题，
// 负数是前端给的临时 id，选项的 NextQuestionID 可以引用它。
type SaveQuestionInput struct {
	ID          int64
	Type        string
	Title       string
	Description string
	Required    bool
	Options     []SaveOptionInput
}

// SaveQuestions 整卷覆盖保存。
//
// 题目顺序用双向链表维护：先把库里的题目按 order_no 串成链，再按前端数组顺序
// 逐个 MoveAfter 到目标位置（O(1) 重连指针），不在新数组里的节点摘掉即可。
func (s *Store) SaveQuestions(surveyID int64, inputs []SaveQuestionInput) ([]model.Question, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 1. 捞出已有题目，串成链表
	rows, err := tx.Query(`SELECT id FROM questions WHERE survey_id=? ORDER BY order_no, id`, surveyID)
	if err != nil {
		return nil, err
	}
	existing := map[int64]bool{}
	list := ds.NewList()
	nodes := map[int64]*ds.Node{}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		existing[id] = true
		nodes[id] = list.Push(id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// 2. 归一化 id：id 为 0 的新题自动分配负数临时 id
	type norm struct {
		key int64
		in  SaveQuestionInput
	}
	normed := make([]norm, 0, len(inputs))
	byKey := map[int64]SaveQuestionInput{}
	for i, in := range inputs {
		key := in.ID
		if key == 0 {
			// 与前端约定一致：负数 -n 表示第 n 题
			key = -int64(i + 1)
		}
		normed = append(normed, norm{key: key, in: in})
		byKey[key] = in
	}

	// 3. 按前端顺序重排链表
	keep := map[int64]bool{}
	var prev *ds.Node
	for _, n := range normed {
		keep[n.key] = true
		node, ok := nodes[n.key]
		if !ok {
			node = list.Push(n.key)
			nodes[n.key] = node
		}
		list.MoveAfter(node, prev)
		prev = node
	}
	for id := range existing {
		if !keep[id] {
			if node, ok := nodes[id]; ok {
				list.Remove(node)
			}
		}
	}
	ordered := list.Values()

	// 4. 删除被移出的题目及其答案、选项
	for id := range existing {
		if keep[id] {
			continue
		}
		stmts := []string{
			`DELETE FROM answers WHERE question_id=?`,
			`DELETE FROM options WHERE question_id=?`,
			`UPDATE options SET next_question_id=0 WHERE next_question_id=?`,
			`DELETE FROM questions WHERE id=?`,
		}
		for _, q := range stmts {
			if _, err := tx.Exec(q, id); err != nil {
				return nil, err
			}
		}
	}

	// 5. 按新顺序写题目，记录临时 id → 真实 id
	realIDs := map[int64]int64{}
	for pos, key := range ordered {
		in := byKey[key]
		required := 0
		if in.Required {
			required = 1
		}
		if key > 0 && existing[key] {
			if _, err := tx.Exec(`UPDATE questions SET type=?, title=?, description=?, required=?, order_no=?
				WHERE id=? AND survey_id=?`,
				in.Type, in.Title, in.Description, required, pos, key, surveyID); err != nil {
				return nil, err
			}
			realIDs[key] = key
			continue
		}
		res, err := tx.Exec(`INSERT INTO questions(survey_id, type, title, description, required, order_no)
			VALUES(?,?,?,?,?,?)`, surveyID, in.Type, in.Title, in.Description, required, pos)
		if err != nil {
			return nil, err
		}
		id, err := res.LastInsertId()
		if err != nil {
			return nil, err
		}
		realIDs[key] = id
	}

	// 6. 重建选项（评分题固定 1-5，忽略前端传的选项）
	for _, key := range ordered {
		qid := realIDs[key]
		if _, err := tx.Exec(`DELETE FROM options WHERE question_id=?`, qid); err != nil {
			return nil, err
		}
		in := byKey[key]
		labels := in.Options
		if in.Type == "rating" {
			labels = make([]SaveOptionInput, 0, 5)
			for i := 1; i <= 5; i++ {
				labels = append(labels, SaveOptionInput{Label: strconv.Itoa(i)})
			}
		}
		for i, opt := range labels {
			next := opt.NextQuestionID
			if next < 0 {
				// 负数 -n 指第 n 题（前端新增题还没 id），换成真实 id
				rid, ok := realIDs[next]
				if !ok {
					rid = 0
				}
				next = rid
			}
			if _, err := tx.Exec(`INSERT INTO options(question_id, label, next_question_id, order_no)
				VALUES(?,?,?,?)`, qid, opt.Label, next, i); err != nil {
				return nil, err
			}
		}
	}

	if _, err := tx.Exec(`UPDATE surveys SET updated_at=? WHERE id=?`, nowUnix(), surveyID); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return s.Questions(surveyID)
}

// OptionIDsFor 读一份答卷的全部答案。
func (s *Store) AnswerByResponse(responseID int64) ([]model.Answer, error) {
	rows, err := s.db.Query(`SELECT id, response_id, question_id, option_ids, text FROM answers WHERE response_id=?`, responseID)
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

func parseIDs(s string) []int64 {
	if s == "" {
		return nil
	}
	var out []int64
	for _, part := range splitComma(s) {
		if v, err := strconv.ParseInt(part, 10, 64); err == nil {
			out = append(out, v)
		}
	}
	return out
}

func splitComma(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	return append(out, s[start:])
}

func joinIDs(ids []int64) string {
	if len(ids) == 0 {
		return ""
	}
	s := ""
	for i, id := range ids {
		if i > 0 {
			s += ","
		}
		s += strconv.FormatInt(id, 10)
	}
	return s
}
