// Package service 是 HTTP 接入层：解析参数、调业务层、拼 DTO。
package service

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	khttp "github.com/go-kratos/kratos/v3/transport/http"

	"questioning/internal/biz"
	"questioning/internal/data"
	"questioning/internal/model"
	"questioning/pkg/jwt"
)

// Service 持有业务层与 JWT 密钥。
type Service struct {
	uc     *biz.Usecase
	secret []byte
}

// New 建接入层。
func New(uc *biz.Usecase, secret []byte) *Service {
	return &Service{uc: uc, secret: secret}
}

// ---- 错误与参数处理 ----

func badBody() error {
	return &biz.Error{Status: 400, Message: "请求体格式不正确"}
}

func badPath() error {
	return &biz.Error{Status: 400, Message: "路径参数不合法"}
}

// render 统一写 JSON 并声明 charset=utf-8（kratos 的 ctx.JSON 不带 charset，
// 有些客户端会按 Latin-1 解码导致中文乱码）。
func (s *Service) render(ctx khttp.Context, code int, v any) error {
	w := ctx.Response()
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	return json.NewEncoder(w).Encode(v)
}

// fail 把业务错误写成 {"message": "..."}；非业务错误记日志后回 500。
func (s *Service) fail(ctx khttp.Context, err error) error {
	var be *biz.Error
	if errors.As(err, &be) {
		return s.render(ctx, be.Status, map[string]string{"message": be.Message})
	}
	return s.render(ctx, 500, map[string]string{"message": "服务器内部错误"})
}

func pathID(ctx khttp.Context, name string) (int64, error) {
	id, err := strconv.ParseInt(ctx.Vars().Get(name), 10, 64)
	if err != nil || id <= 0 {
		return 0, badPath()
	}
	return id, nil
}

// currentUserID 解析 Authorization: Bearer <token>。
func (s *Service) currentUserID(ctx khttp.Context) (int64, error) {
	h := ctx.Request().Header.Get("Authorization")
	if !strings.HasPrefix(h, "Bearer ") {
		return 0, &biz.Error{Status: 401, Message: "请先登录"}
	}
	id, err := jwt.Parse(s.secret, strings.TrimSpace(strings.TrimPrefix(h, "Bearer ")))
	if err != nil {
		return 0, &biz.Error{Status: 401, Message: "登录状态已失效，请重新登录"}
	}
	return id, nil
}

// parseTimeParam 支持 unix 秒或 2006-01-02 两种写法，0/空返回 0。
func parseTimeParam(raw string) int64 {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	if v, err := strconv.ParseInt(raw, 10, 64); err == nil {
		return v
	}
	if t, err := time.ParseInLocation("2006-01-02", raw, time.Local); err == nil {
		return t.Unix()
	}
	return 0
}

// deviceOf 从 UA 里猜一个设备类型，统计里展示用。
func deviceOf(ua string) string {
	l := strings.ToLower(ua)
	switch {
	case strings.Contains(l, "iphone"):
		return "iPhone"
	case strings.Contains(l, "ipad"):
		return "iPad"
	case strings.Contains(l, "android"):
		return "Android"
	case strings.Contains(l, "mobile"):
		return "手机"
	default:
		return "桌面端"
	}
}

// ---- 账号 ----

// Register 注册。
func (s *Service) Register(ctx khttp.Context) error {
	var req RegisterReq
	if err := ctx.Bind(&req); err != nil {
		return s.fail(ctx, badBody())
	}
	user, token, err := s.uc.Register(req.Username, req.Password, req.Nickname)
	if err != nil {
		return s.fail(ctx, err)
	}
	return s.render(ctx, 200, AuthResp{Token: token, User: toUserDTO(user)})
}

// Login 登录。
func (s *Service) Login(ctx khttp.Context) error {
	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return s.fail(ctx, badBody())
	}
	user, token, err := s.uc.Login(req.Username, req.Password)
	if err != nil {
		return s.fail(ctx, err)
	}
	return s.render(ctx, 200, AuthResp{Token: token, User: toUserDTO(user)})
}

// Me 当前用户。
func (s *Service) Me(ctx khttp.Context) error {
	id, err := s.currentUserID(ctx)
	if err != nil {
		return s.fail(ctx, err)
	}
	user, err := s.uc.User(id)
	if err != nil {
		return s.fail(ctx, err)
	}
	return s.render(ctx, 200, toUserDTO(user))
}

// ---- 问卷管理 ----

// ListSurveys 我的问卷列表。
func (s *Service) ListSurveys(ctx khttp.Context) error {
	uid, err := s.currentUserID(ctx)
	if err != nil {
		return s.fail(ctx, err)
	}
	list, err := s.uc.ListSurveys(uid)
	if err != nil {
		return s.fail(ctx, err)
	}
	out := make([]SurveyDTO, 0, len(list))
	for _, sv := range list {
		out = append(out, toSurveyDTO(sv, false))
	}
	return s.render(ctx, 200, map[string]any{"list": out})
}

// CreateSurvey 新建问卷。
func (s *Service) CreateSurvey(ctx khttp.Context) error {
	uid, err := s.currentUserID(ctx)
	if err != nil {
		return s.fail(ctx, err)
	}
	var req SurveySaveReq
	if err := ctx.Bind(&req); err != nil {
		return s.fail(ctx, badBody())
	}
	sv, err := s.uc.CreateSurvey(uid, req.Title, req.Description, req.Anonymous)
	if err != nil {
		return s.fail(ctx, err)
	}
	return s.render(ctx, 200, toSurveyDTO(sv, true))
}

// SurveyDetail 问卷详情（带题目）。
func (s *Service) SurveyDetail(ctx khttp.Context) error {
	uid, err := s.currentUserID(ctx)
	if err != nil {
		return s.fail(ctx, err)
	}
	id, err := pathID(ctx, "id")
	if err != nil {
		return s.fail(ctx, err)
	}
	sv, err := s.uc.SurveyDetail(id, uid)
	if err != nil {
		return s.fail(ctx, err)
	}
	return s.render(ctx, 200, toSurveyDTO(sv, true))
}

// UpdateSurvey 改标题等基本信息。
func (s *Service) UpdateSurvey(ctx khttp.Context) error {
	uid, err := s.currentUserID(ctx)
	if err != nil {
		return s.fail(ctx, err)
	}
	id, err := pathID(ctx, "id")
	if err != nil {
		return s.fail(ctx, err)
	}
	var req SurveySaveReq
	if err := ctx.Bind(&req); err != nil {
		return s.fail(ctx, badBody())
	}
	if err := s.uc.UpdateSurvey(id, uid, req.Title, req.Description, req.Anonymous); err != nil {
		return s.fail(ctx, err)
	}
	sv, err := s.uc.SurveyDetail(id, uid)
	if err != nil {
		return s.fail(ctx, err)
	}
	return s.render(ctx, 200, toSurveyDTO(sv, true))
}

// DeleteSurvey 删问卷。
func (s *Service) DeleteSurvey(ctx khttp.Context) error {
	uid, err := s.currentUserID(ctx)
	if err != nil {
		return s.fail(ctx, err)
	}
	id, err := pathID(ctx, "id")
	if err != nil {
		return s.fail(ctx, err)
	}
	if err := s.uc.DeleteSurvey(id, uid); err != nil {
		return s.fail(ctx, err)
	}
	return s.render(ctx, 200, map[string]string{"message": "已删除"})
}

// SetStatus 发布/关闭。
func (s *Service) SetStatus(ctx khttp.Context) error {
	uid, err := s.currentUserID(ctx)
	if err != nil {
		return s.fail(ctx, err)
	}
	id, err := pathID(ctx, "id")
	if err != nil {
		return s.fail(ctx, err)
	}
	var req StatusReq
	if err := ctx.Bind(&req); err != nil {
		return s.fail(ctx, badBody())
	}
	if err := s.uc.SetStatus(id, uid, req.Status); err != nil {
		return s.fail(ctx, err)
	}
	sv, err := s.uc.SurveyDetail(id, uid)
	if err != nil {
		return s.fail(ctx, err)
	}
	return s.render(ctx, 200, toSurveyDTO(sv, true))
}

// SaveQuestions 整卷保存题目与跳转。
func (s *Service) SaveQuestions(ctx khttp.Context) error {
	uid, err := s.currentUserID(ctx)
	if err != nil {
		return s.fail(ctx, err)
	}
	id, err := pathID(ctx, "id")
	if err != nil {
		return s.fail(ctx, err)
	}
	var req QuestionsSaveReq
	if err := ctx.Bind(&req); err != nil {
		return s.fail(ctx, badBody())
	}
	inputs := make([]data.SaveQuestionInput, 0, len(req.Questions))
	for _, q := range req.Questions {
		in := data.SaveQuestionInput{
			ID:          q.ID,
			Type:        q.Type,
			Title:       q.Title,
			Description: q.Description,
			Required:    q.Required,
		}
		for _, o := range q.Options {
			in.Options = append(in.Options, data.SaveOptionInput{Label: o.Label, NextQuestionID: o.NextQuestionID})
		}
		inputs = append(inputs, in)
	}
	questions, err := s.uc.SaveQuestions(id, uid, inputs)
	if err != nil {
		return s.fail(ctx, err)
	}
	sv, err := s.uc.SurveyDetail(id, uid)
	if err != nil {
		return s.fail(ctx, err)
	}
	sv.Questions = questions
	return s.render(ctx, 200, toSurveyDTO(sv, true))
}

// ---- 公开填写 ----

// PublicSurvey 填写页拿问卷。
func (s *Service) PublicSurvey(ctx khttp.Context) error {
	id, err := pathID(ctx, "id")
	if err != nil {
		return s.fail(ctx, err)
	}
	sv, err := s.uc.PublicSurvey(id)
	if err != nil {
		return s.fail(ctx, err)
	}
	return s.render(ctx, 200, toSurveyDTO(sv, true))
}

// Submit 交卷。
func (s *Service) Submit(ctx khttp.Context) error {
	id, err := pathID(ctx, "id")
	if err != nil {
		return s.fail(ctx, err)
	}
	var req SubmitReq
	if err := ctx.Bind(&req); err != nil {
		return s.fail(ctx, badBody())
	}
	answers := make([]data.NewAnswer, 0, len(req.Answers))
	for _, a := range req.Answers {
		answers = append(answers, data.NewAnswer{QuestionID: a.QuestionID, OptionIDs: a.OptionIDs, Text: strings.TrimSpace(a.Text)})
	}
	device := deviceOf(ctx.Request().UserAgent())
	responseID, err := s.uc.Submit(id, req.DurationSec, device, answers)
	if err != nil {
		return s.fail(ctx, err)
	}
	return s.render(ctx, 200, map[string]int64{"id": responseID})
}

// ---- 统计与导出 ----

// Stats 统计页数据。
func (s *Service) Stats(ctx khttp.Context) error {
	uid, err := s.currentUserID(ctx)
	if err != nil {
		return s.fail(ctx, err)
	}
	id, err := pathID(ctx, "id")
	if err != nil {
		return s.fail(ctx, err)
	}
	st, err := s.uc.Stats(id, uid)
	if err != nil {
		return s.fail(ctx, err)
	}
	return s.render(ctx, 200, st)
}

// Responses 答卷分页列表。
func (s *Service) Responses(ctx khttp.Context) error {
	uid, err := s.currentUserID(ctx)
	if err != nil {
		return s.fail(ctx, err)
	}
	id, err := pathID(ctx, "id")
	if err != nil {
		return s.fail(ctx, err)
	}
	q := ctx.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	size, _ := strconv.Atoi(q.Get("size"))
	list, total, err := s.uc.Responses(id, uid, page, size, parseTimeParam(q.Get("from")), parseTimeParam(q.Get("to")))
	if err != nil {
		return s.fail(ctx, err)
	}
	out := make([]ResponseDTO, 0, len(list))
	for _, r := range list {
		out = append(out, ResponseDTO{ID: r.ID, SubmittedAt: ts(r.SubmittedAt), DurationSec: r.DurationSec, Device: r.Device})
	}
	return s.render(ctx, 200, ResponseListDTO{Total: total, List: out})
}

// ResponseDetail 单份答卷。
func (s *Service) ResponseDetail(ctx khttp.Context) error {
	uid, err := s.currentUserID(ctx)
	if err != nil {
		return s.fail(ctx, err)
	}
	id, err := pathID(ctx, "id")
	if err != nil {
		return s.fail(ctx, err)
	}
	rid, err := pathID(ctx, "rid")
	if err != nil {
		return s.fail(ctx, err)
	}
	r, answers, err := s.uc.ResponseDetail(id, uid, rid)
	if err != nil {
		return s.fail(ctx, err)
	}
	detail := ResponseDetailDTO{
		ResponseDTO: ResponseDTO{ID: r.ID, SubmittedAt: ts(r.SubmittedAt), DurationSec: r.DurationSec, Device: r.Device},
		Answers:     make([]AnswerDTO, 0, len(answers)),
	}
	for _, a := range answers {
		detail.Answers = append(detail.Answers, AnswerDTO{QuestionID: a.QuestionID, OptionIDs: a.OptionIDs, Text: a.Text})
	}
	return s.render(ctx, 200, detail)
}

// ExportCSV 导出答卷，带 BOM 方便 Excel 直接打开。
func (s *Service) ExportCSV(ctx khttp.Context) error {
	uid, err := s.currentUserID(ctx)
	if err != nil {
		return s.fail(ctx, err)
	}
	id, err := pathID(ctx, "id")
	if err != nil {
		return s.fail(ctx, err)
	}
	raw, err := s.uc.Export(id, uid)
	if err != nil {
		return s.fail(ctx, err)
	}

	var buf bytes.Buffer
	buf.WriteString("\xEF\xBB\xBF")
	w := csv.NewWriter(&buf)

	header := []string{"答卷ID", "提交时间", "耗时(秒)", "设备"}
	for _, q := range raw.Questions {
		header = append(header, q.Title)
	}
	if err := w.Write(header); err != nil {
		return s.fail(ctx, err)
	}

	for _, r := range raw.Responses {
		byQuestion := make(map[int64]model.Answer, len(raw.Answers[r.ID]))
		for _, a := range raw.Answers[r.ID] {
			byQuestion[a.QuestionID] = a
		}
		row := []string{strconv.FormatInt(r.ID, 10), ts(r.SubmittedAt), strconv.Itoa(r.DurationSec), r.Device}
		for _, q := range raw.Questions {
			a, ok := byQuestion[q.ID]
			if !ok {
				row = append(row, "")
				continue
			}
			if q.Type == "text" {
				row = append(row, a.Text)
				continue
			}
			labels := make([]string, 0, len(a.OptionIDs))
			for _, oid := range a.OptionIDs {
				for _, o := range q.Options {
					if o.ID == oid {
						labels = append(labels, o.Label)
					}
				}
			}
			row = append(row, strings.Join(labels, "|"))
		}
		if err := w.Write(row); err != nil {
			return s.fail(ctx, err)
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return s.fail(ctx, err)
	}

	name := fmt.Sprintf("survey-%d-%s.csv", id, time.Now().Format("20060102"))
	res := ctx.Response()
	res.Header().Set("Content-Type", "text/csv; charset=utf-8")
	res.Header().Set("Content-Disposition", `attachment; filename="`+name+`"`)
	_, err = res.Write(buf.Bytes())
	return err
}
