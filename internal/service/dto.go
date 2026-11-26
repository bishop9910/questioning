package service

import (
	"time"

	"questioning/internal/model"
)

// ts 把 unix 秒转成 RFC3339，0 返回空串（未发布时 publishedAt 就是空）。
func ts(unix int64) string {
	if unix <= 0 {
		return ""
	}
	return time.Unix(unix, 0).Format(time.RFC3339)
}

// UserDTO 是返回给前端的用户信息（不含密码）。
type UserDTO struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Role     string `json:"role"`
}

func toUserDTO(u model.User) UserDTO {
	return UserDTO{ID: u.ID, Username: u.Username, Nickname: u.Nickname, Role: u.Role}
}

// AuthResp 是登录/注册的返回。
type AuthResp struct {
	Token string  `json:"token"`
	User  UserDTO `json:"user"`
}

// OptionDTO 是选项。
type OptionDTO struct {
	ID             int64  `json:"id"`
	Label          string `json:"label"`
	NextQuestionID int64  `json:"nextQuestionId"`
}

// QuestionDTO 是题目。
type QuestionDTO struct {
	ID          int64       `json:"id"`
	Type        string      `json:"type"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	Options     []OptionDTO `json:"options"`
}

// SurveyDTO 是问卷（列表与详情共用，详情才带 questions）。
type SurveyDTO struct {
	ID            int64         `json:"id"`
	Title         string        `json:"title"`
	Description   string        `json:"description"`
	Status        string        `json:"status"`
	Anonymous     bool          `json:"anonymous"`
	CreatedAt     string        `json:"createdAt"`
	UpdatedAt     string        `json:"updatedAt"`
	PublishedAt   string        `json:"publishedAt"`
	ResponseCount int           `json:"responseCount"`
	Questions     []QuestionDTO `json:"questions"`
}

func toSurveyDTO(sv model.Survey, withQuestions bool) SurveyDTO {
	out := SurveyDTO{
		ID:            sv.ID,
		Title:         sv.Title,
		Description:   sv.Description,
		Status:        sv.Status,
		Anonymous:     sv.Anonymous,
		CreatedAt:     ts(sv.CreatedAt),
		UpdatedAt:     ts(sv.UpdatedAt),
		PublishedAt:   ts(sv.PublishedAt),
		ResponseCount: sv.ResponseCount,
		Questions:     []QuestionDTO{},
	}
	if withQuestions {
		for _, q := range sv.Questions {
			qd := QuestionDTO{
				ID:          q.ID,
				Type:        q.Type,
				Title:       q.Title,
				Description: q.Description,
				Required:    q.Required,
				Options:     []OptionDTO{},
			}
			for _, o := range q.Options {
				qd.Options = append(qd.Options, OptionDTO{ID: o.ID, Label: o.Label, NextQuestionID: o.NextQuestionID})
			}
			out.Questions = append(out.Questions, qd)
		}
	}
	return out
}

// ---- 请求体 ----

// RegisterReq 注册请求。
type RegisterReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
}

// LoginReq 登录请求。
type LoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// SurveySaveReq 新建/修改问卷。
type SurveySaveReq struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Anonymous   bool   `json:"anonymous"`
}

// StatusReq 发布/关闭问卷。
type StatusReq struct {
	Status string `json:"status"`
}

// OptionSaveDTO 保存时的选项：nextQuestionId 可以是负数（引用同批的新题）。
type OptionSaveDTO struct {
	Label          string `json:"label"`
	NextQuestionID int64  `json:"nextQuestionId"`
}

// QuestionSaveDTO 保存时的题目：id > 0 更新，id <= 0 新增。
type QuestionSaveDTO struct {
	ID          int64           `json:"id"`
	Type        string          `json:"type"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Required    bool            `json:"required"`
	Options     []OptionSaveDTO `json:"options"`
}

// QuestionsSaveReq 整卷保存。
type QuestionsSaveReq struct {
	Questions []QuestionSaveDTO `json:"questions"`
}

// AnswerReq 提交时的单题作答。
type AnswerReq struct {
	QuestionID int64   `json:"questionId"`
	OptionIDs  []int64 `json:"optionIds"`
	Text       string  `json:"text"`
}

// SubmitReq 提交答卷。
type SubmitReq struct {
	Answers     []AnswerReq `json:"answers"`
	DurationSec int         `json:"durationSec"`
}

// AnswerDTO 是答卷详情里的单题作答。
type AnswerDTO struct {
	QuestionID int64   `json:"questionId"`
	OptionIDs  []int64 `json:"optionIds"`
	Text       string  `json:"text"`
}

// ResponseDTO 是答卷元信息。
type ResponseDTO struct {
	ID          int64  `json:"id"`
	SubmittedAt string `json:"submittedAt"`
	DurationSec int    `json:"durationSec"`
	Device      string `json:"device"`
}

// ResponseListDTO 是分页结果。
type ResponseListDTO struct {
	Total int           `json:"total"`
	List  []ResponseDTO `json:"list"`
}

// ResponseDetailDTO 是单份答卷。
type ResponseDetailDTO struct {
	ResponseDTO
	Answers []AnswerDTO `json:"answers"`
}
