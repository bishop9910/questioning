// Package biz 业务层：校验、跳题路径模拟、统计与索引。
package biz

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"questioning/internal/data"
	"questioning/internal/model"
	"questioning/pkg/ds"
	"questioning/pkg/jwt"
)

// Error 是带 HTTP 状态码的业务错误，service 层直接翻译成 JSON。
type Error struct {
	Status  int
	Message string
}

func (e *Error) Error() string { return e.Message }

func failf(status int, format string, args ...any) error {
	return &Error{Status: status, Message: fmt.Sprintf(format, args...)}
}

// Usecase 串起存储层与数据结构。
type Usecase struct {
	store      *data.Store
	secret     []byte
	ttl        time.Duration
	indexCache *ds.LRU[int64, *responseIndex]
}

// New 建业务对象；indexCache 缓存每份问卷的答卷时间索引，最多留 16 份。
func New(store *data.Store, secret []byte, ttl time.Duration) *Usecase {
	return &Usecase{
		store:      store,
		secret:     secret,
		ttl:        ttl,
		indexCache: ds.NewLRU[int64, *responseIndex](16),
	}
}

// ---- 账号 ----

// Register 注册普通用户并直接返回 token。
func (u *Usecase) Register(username, password, nickname string) (model.User, string, error) {
	username = strings.TrimSpace(username)
	nickname = strings.TrimSpace(nickname)
	if l := len([]rune(username)); l < 3 || l > 32 {
		return model.User{}, "", failf(400, "用户名长度需要 3-32 个字符")
	}
	if len(password) < 6 {
		return model.User{}, "", failf(400, "密码至少 6 位")
	}
	if nickname == "" {
		nickname = username
	}
	if _, err := u.store.UserByUsername(username); err == nil {
		return model.User{}, "", failf(409, "用户名已被注册")
	} else if !errors.Is(err, data.ErrNotFound) {
		return model.User{}, "", err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return model.User{}, "", err
	}
	user, err := u.store.CreateUser(username, string(hash), nickname, "user")
	if err != nil {
		return model.User{}, "", err
	}
	token, err := jwt.Sign(u.secret, user.ID, u.ttl)
	return user, token, err
}

// Login 校验密码并签发 token。
func (u *Usecase) Login(username, password string) (model.User, string, error) {
	user, err := u.store.UserByUsername(strings.TrimSpace(username))
	if errors.Is(err, data.ErrNotFound) {
		return model.User{}, "", failf(401, "用户名或密码错误")
	}
	if err != nil {
		return model.User{}, "", err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return model.User{}, "", failf(401, "用户名或密码错误")
	}
	token, err := jwt.Sign(u.secret, user.ID, u.ttl)
	return user, token, err
}

// User 取当前用户。
func (u *Usecase) User(id int64) (model.User, error) {
	user, err := u.store.UserByID(id)
	if errors.Is(err, data.ErrNotFound) {
		return model.User{}, failf(401, "登录状态已失效")
	}
	return user, err
}

// ---- 问卷 ----

// ListSurveys 当前用户的问卷列表。
func (u *Usecase) ListSurveys(ownerID int64) ([]model.Survey, error) {
	return u.store.ListSurveys(ownerID)
}

// CreateSurvey 新建草稿。
func (u *Usecase) CreateSurvey(ownerID int64, title, description string, anonymous bool) (model.Survey, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return model.Survey{}, failf(400, "问卷标题不能为空")
	}
	if len([]rune(title)) > 100 {
		return model.Survey{}, failf(400, "问卷标题最多 100 个字")
	}
	return u.store.CreateSurvey(ownerID, title, strings.TrimSpace(description), anonymous)
}

// SurveyDetail 取问卷 + 题目，并校验归属。
func (u *Usecase) SurveyDetail(id, ownerID int64) (model.Survey, error) {
	sv, err := u.owned(id, ownerID)
	if err != nil {
		return sv, err
	}
	if sv.Questions, err = u.store.Questions(id); err != nil {
		return sv, err
	}
	return sv, nil
}

func (u *Usecase) owned(id, ownerID int64) (model.Survey, error) {
	sv, err := u.store.SurveyByID(id)
	if errors.Is(err, data.ErrNotFound) {
		return sv, failf(404, "问卷不存在")
	}
	if err != nil {
		return sv, err
	}
	if sv.OwnerID != ownerID {
		return sv, failf(403, "无权操作这份问卷")
	}
	return sv, nil
}

// UpdateSurvey 改标题、说明、匿名开关。
func (u *Usecase) UpdateSurvey(id, ownerID int64, title, description string, anonymous bool) error {
	if _, err := u.owned(id, ownerID); err != nil {
		return err
	}
	title = strings.TrimSpace(title)
	if title == "" {
		return failf(400, "问卷标题不能为空")
	}
	if len([]rune(title)) > 100 {
		return failf(400, "问卷标题最多 100 个字")
	}
	return u.store.UpdateSurvey(id, title, strings.TrimSpace(description), anonymous)
}

// DeleteSurvey 删问卷。
func (u *Usecase) DeleteSurvey(id, ownerID int64) error {
	if _, err := u.owned(id, ownerID); err != nil {
		return err
	}
	u.indexCache.Invalidate(id)
	return u.store.DeleteSurvey(id)
}

// SetStatus 切换 draft / published / closed。
func (u *Usecase) SetStatus(id, ownerID int64, status string) error {
	sv, err := u.owned(id, ownerID)
	if err != nil {
		return err
	}
	switch status {
	case "draft", "published", "closed":
	default:
		return failf(400, "状态只能是 draft、published 或 closed")
	}
	if status == "published" && sv.Status == "published" {
		return nil
	}
	return u.store.SetStatus(id, status)
}

var validTypes = map[string]bool{"single": true, "multiple": true, "text": true, "rating": true}

// SaveQuestions 整卷覆盖保存，写库前校验题型、选项与跳题环。
func (u *Usecase) SaveQuestions(id, ownerID int64, inputs []data.SaveQuestionInput) ([]model.Question, error) {
	if _, err := u.owned(id, ownerID); err != nil {
		return nil, err
	}
	if len(inputs) == 0 {
		return nil, failf(400, "至少需要一道题")
	}
	if len(inputs) > 200 {
		return nil, failf(400, "一道问卷最多 200 题")
	}

	// 前端新增题还没有 id，约定用负数表示「第 n 题」（-1 就是第一题，-2 第二题…），
	// 存储层保存时再换成真实 id。
	realIDs := map[int64]bool{}
	for _, in := range inputs {
		if in.ID > 0 {
			realIDs[in.ID] = true
		}
	}
	effID := func(i int, id int64) int64 {
		if id > 0 {
			return id
		}
		return -int64(i + 1)
	}
	validNext := func(next int64) bool {
		if next > 0 {
			return realIDs[next]
		}
		n := -next
		return n >= 1 && n <= int64(len(inputs))
	}

	for i, in := range inputs {
		if !validTypes[in.Type] {
			return nil, failf(400, "第 %d 题题型不合法", i+1)
		}
		if strings.TrimSpace(in.Title) == "" {
			return nil, failf(400, "第 %d 题题干不能为空", i+1)
		}
		if n := len([]rune(in.Title)); n > 200 {
			return nil, failf(400, "第 %d 题题干最多 200 个字", i+1)
		}
		switch in.Type {
		case "single", "multiple":
			if len(in.Options) < 2 {
				return nil, failf(400, "第 %d 题至少需要 2 个选项", i+1)
			}
			if len(in.Options) > 50 {
				return nil, failf(400, "第 %d 题最多 50 个选项", i+1)
			}
			for j, opt := range in.Options {
				if strings.TrimSpace(opt.Label) == "" {
					return nil, failf(400, "第 %d 题第 %d 个选项为空", i+1, j+1)
				}
			}
		}
	}

	// 跳题有向图：题 → 选项指向的题，有环就拒绝保存
	g := ds.NewGraph()
	for i, in := range inputs {
		for _, opt := range in.Options {
			next := opt.NextQuestionID
			if next == 0 {
				continue
			}
			if !validNext(next) {
				return nil, failf(400, "第 %d 题的选项指向了不存在的题目", i+1)
			}
			g.AddEdge(effID(i, in.ID), next)
		}
	}
	if cycle := g.Cycle(); cycle != nil {
		return nil, failf(400, "跳转逻辑存在环，请检查题目跳转设置")
	}
	return u.store.SaveQuestions(id, inputs)
}

// ---- 公开填写 ----

// PublicSurvey 取已发布问卷给填写页。
func (u *Usecase) PublicSurvey(id int64) (model.Survey, error) {
	sv, err := u.store.SurveyByID(id)
	if errors.Is(err, data.ErrNotFound) {
		return sv, failf(404, "问卷不存在")
	}
	if err != nil {
		return sv, err
	}
	if sv.Status != "published" {
		return sv, failf(403, "问卷未发布或已关闭")
	}
	if sv.Questions, err = u.store.Questions(id); err != nil {
		return sv, err
	}
	return sv, nil
}

// Submit 校验必答与选项归属后落库，同时让答卷索引缓存失效。
func (u *Usecase) Submit(id int64, durationSec int, device string, answers []data.NewAnswer) (int64, error) {
	sv, err := u.PublicSurvey(id)
	if err != nil {
		return 0, err
	}
	if len(sv.Questions) == 0 {
		return 0, failf(400, "问卷还没有题目")
	}
	if durationSec < 0 {
		durationSec = 0
	}
	if durationSec > 86400 {
		durationSec = 86400
	}

	// 题目与选项归属校验
	optionOwner := map[int64]int64{} // optionID → questionID
	optionNext := map[int64]int64{}  // optionID → nextQuestionID
	answerMap := map[int64]data.NewAnswer{}
	for _, q := range sv.Questions {
		for _, o := range q.Options {
			optionOwner[o.ID] = q.ID
			optionNext[o.ID] = o.NextQuestionID
		}
	}
	for _, a := range answers {
		if _, ok := optionOwnerByQuestion(sv, a.QuestionID); !ok {
			return 0, failf(400, "答案里出现了不属于本问卷的题目")
		}
		for _, oid := range a.OptionIDs {
			if optionOwner[oid] != a.QuestionID {
				return 0, failf(400, "选项与题目不匹配")
			}
		}
		answerMap[a.QuestionID] = a
	}

	// 按跳题规则模拟一遍作答路径，检查必答
	path := walkPath(sv.Questions, answerMap, optionNext)
	for _, q := range path {
		a, answered := answerMap[q.ID]
		if !answered && q.Required {
			return 0, failf(400, "第 %d 题「%s」为必答题", q.OrderNo+1, q.Title)
		}
		if !answered {
			continue
		}
		switch q.Type {
		case "text":
			if q.Required && strings.TrimSpace(a.Text) == "" {
				return 0, failf(400, "第 %d 题「%s」为必答题", q.OrderNo+1, q.Title)
			}
		default:
			if q.Required && len(a.OptionIDs) == 0 {
				return 0, failf(400, "第 %d 题「%s」为必答题", q.OrderNo+1, q.Title)
			}
		}
	}

	// 没走到的题目不落库
	kept := make([]data.NewAnswer, 0, len(answers))
	onPath := map[int64]bool{}
	for _, q := range path {
		onPath[q.ID] = true
	}
	for _, a := range answers {
		if !onPath[a.QuestionID] {
			continue
		}
		if strings.TrimSpace(a.Text) == "" && len(a.OptionIDs) == 0 {
			continue
		}
		kept = append(kept, a)
	}

	responseID, err := u.store.CreateResponse(id, durationSec, device, kept)
	if err != nil {
		return 0, err
	}
	u.indexCache.Invalidate(id)
	return responseID, nil
}

// walkPath 从第一题开始按跳转规则走一遍，返回实际经过的题目序列。
// 单选按所选项的跳转走；多选取选中项里最后一个配置了跳转的选项。
func walkPath(questions []model.Question, answers map[int64]data.NewAnswer, optionNext map[int64]int64) []model.Question {
	if len(questions) == 0 {
		return nil
	}
	byID := map[int64]int{}
	for i, q := range questions {
		byID[q.ID] = i
	}
	var path []model.Question
	idx := 0
	visited := map[int64]bool{}
	for idx >= 0 && idx < len(questions) {
		q := questions[idx]
		if visited[q.ID] {
			break // 理论上保存时已拦住环，这里兜底
		}
		visited[q.ID] = true
		path = append(path, q)

		next := int64(0)
		if a, ok := answers[q.ID]; ok && q.Type != "text" {
			for _, o := range q.Options {
				for _, picked := range a.OptionIDs {
					if picked == o.ID {
						if n := optionNext[o.ID]; n != 0 {
							next = n
						}
					}
				}
			}
		}
		if next != 0 {
			if i, ok := byID[next]; ok {
				idx = i
			} else {
				break
			}
			continue
		}
		idx++
	}
	return path
}

func optionOwnerByQuestion(sv model.Survey, questionID int64) (model.Question, bool) {
	for _, q := range sv.Questions {
		if q.ID == questionID {
			return q, true
		}
	}
	return model.Question{}, false
}
