// Package model 放各层共用的数据结构，字段与 docs/api.md 对齐。
package model

// User 是后台用户，role 目前只有 admin 与 user。
type User struct {
	ID           int64
	Username     string
	Nickname     string
	Role         string
	PasswordHash string
	CreatedAt    int64
}

// Survey 是一份问卷，status 为 draft / published / closed。
type Survey struct {
	ID          int64
	OwnerID     int64
	Title       string
	Description string
	Status      string
	Anonymous   bool
	CreatedAt   int64
	UpdatedAt   int64
	PublishedAt int64
	Questions   []Question
	// ResponseCount 只用于列表页展示。
	ResponseCount int
}

// Question 一道题，type 为 single / multiple / text / rating。
// MinSelect / MaxSelect 只对多选题有意义，0 表示不限。
type Question struct {
	ID          int64
	SurveyID    int64
	Type        string
	Title       string
	Description string
	Required    bool
	MinSelect   int
	MaxSelect   int
	OrderNo     int
	Options     []Option
}

// Option 是选择题的一个选项，NextQuestionID 为 0 表示不跳转。
type Option struct {
	ID             int64
	QuestionID     int64
	Label          string
	NextQuestionID int64
	OrderNo        int
}

// Response 是一份答卷的元信息。
type Response struct {
	ID          int64
	SurveyID    int64
	SubmittedAt int64
	DurationSec int
	Device      string
}

// Answer 是一份答卷里某道题的作答，OptionIDs 用逗号拼接存储。
type Answer struct {
	ID         int64
	ResponseID int64
	QuestionID int64
	OptionIDs  []int64
	Text       string
}
