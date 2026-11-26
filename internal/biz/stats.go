package biz

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
	"unicode"

	"questioning/internal/data"
	"questioning/internal/model"
	"questioning/pkg/ds"
)

// responseIndex 是一份问卷的答卷时间索引：跳表按提交时间排序，map 存元信息。
type responseIndex struct {
	list *ds.SkipList
	meta map[int64]model.Response
}

// indexFor 取（或重建）某问卷的索引，建好后放进 LRU 缓存。
func (u *Usecase) indexFor(surveyID int64) (*responseIndex, error) {
	if idx, ok := u.indexCache.Get(surveyID); ok {
		return idx, nil
	}
	rows, err := u.store.ListResponses(surveyID)
	if err != nil {
		return nil, err
	}
	idx := &responseIndex{list: ds.NewSkipList(), meta: make(map[int64]model.Response, len(rows))}
	for _, r := range rows {
		idx.list.Insert(r.SubmittedAt, r.ID)
		idx.meta[r.ID] = r
	}
	u.indexCache.Put(surveyID, idx)
	return idx, nil
}

// Responses 分页查答卷：无时间筛选走跳表倒序翻页，有时间筛选走区间查询。
func (u *Usecase) Responses(surveyID, ownerID int64, page, size int, from, to int64) ([]model.Response, int, error) {
	if _, err := u.owned(surveyID, ownerID); err != nil {
		return nil, 0, err
	}
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 200 {
		size = 20
	}
	idx, err := u.indexFor(surveyID)
	if err != nil {
		return nil, 0, err
	}

	var ids []int64
	total := 0
	if from > 0 || to > 0 {
		if from <= 0 {
			from = 0
		}
		if to <= 0 {
			to = time.Now().Add(24 * time.Hour).Unix()
		}
		asc := idx.list.RangeAsc(from, to)
		total = len(asc)
		// 区间查询结果按时间升序，翻页要按新到旧，先整体反转
		desc := make([]int64, 0, len(asc))
		for i := len(asc) - 1; i >= 0; i-- {
			desc = append(desc, asc[i])
		}
		start := (page - 1) * size
		if start > len(desc) {
			start = len(desc)
		}
		end := start + size
		if end > len(desc) {
			end = len(desc)
		}
		ids = desc[start:end]
	} else {
		total = idx.list.Len()
		ids = idx.list.Page((page-1)*size, size)
	}

	out := make([]model.Response, 0, len(ids))
	for _, id := range ids {
		out = append(out, idx.meta[id])
	}
	return out, total, nil
}

// ResponseDetail 取单份答卷和它的答案。
func (u *Usecase) ResponseDetail(surveyID, ownerID, responseID int64) (model.Response, []model.Answer, error) {
	if _, err := u.owned(surveyID, ownerID); err != nil {
		return model.Response{}, nil, err
	}
	r, err := u.store.ResponseByID(responseID)
	if errors.Is(err, data.ErrNotFound) {
		return r, nil, failf(404, "答卷不存在")
	}
	if err != nil {
		return r, nil, err
	}
	if r.SurveyID != surveyID {
		return r, nil, failf(404, "答卷不属于这份问卷")
	}
	answers, err := u.store.AnswerByResponse(responseID)
	return r, answers, err
}

// ---- 统计 ----

// DailyPoint 是近 7 天里某天的答卷数。
type DailyPoint struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// OptionStat 是一个选项的票数。
type OptionStat struct {
	ID      int64   `json:"id"`
	Label   string  `json:"label"`
	Count   int     `json:"count"`
	Percent float64 `json:"percent"`
}

// WordStat 是一条词频。
type WordStat struct {
	Word  string `json:"word"`
	Count int    `json:"count"`
}

// QuestionStat 是一道题的统计结果。
type QuestionStat struct {
	ID       int64        `json:"id"`
	Type     string       `json:"type"`
	Title    string       `json:"title"`
	Answered int          `json:"answered"`
	Options  []OptionStat `json:"options"`
	Words    []WordStat   `json:"words"`
}

// Stats 是统计页的完整数据。
type Stats struct {
	SurveyID       int64          `json:"surveyId"`
	Title          string         `json:"title"`
	Total          int            `json:"total"`
	AvgDurationSec int            `json:"avgDurationSec"`
	Daily          []DailyPoint   `json:"daily"`
	Questions      []QuestionStat `json:"questions"`
}

// Stats 汇总一份问卷的答卷。
func (u *Usecase) Stats(surveyID, ownerID int64) (Stats, error) {
	sv, err := u.owned(surveyID, ownerID)
	if err != nil {
		return Stats{}, err
	}
	if sv.Questions, err = u.store.Questions(surveyID); err != nil {
		return Stats{}, err
	}
	responses, err := u.store.ListResponses(surveyID)
	if err != nil {
		return Stats{}, err
	}
	answers, err := u.store.AllAnswers(surveyID)
	if err != nil {
		return Stats{}, err
	}

	st := Stats{
		SurveyID:  sv.ID,
		Title:     sv.Title,
		Total:     len(responses),
		Daily:     dailyPoints(responses),
		Questions: make([]QuestionStat, 0, len(sv.Questions)),
	}
	if len(responses) > 0 {
		sum := 0
		for _, r := range responses {
			sum += r.DurationSec
		}
		st.AvgDurationSec = int(math.Round(float64(sum) / float64(len(responses))))
	}

	for _, q := range sv.Questions {
		stat := QuestionStat{ID: q.ID, Type: q.Type, Title: q.Title, Options: []OptionStat{}, Words: []WordStat{}}

		// 选项票数走哈希表，文本词频走哈希表 + 小顶堆 Top-K
		optionCount := ds.NewHashMap()
		wordCount := ds.NewHashMap()
		for _, list := range answers {
			for _, a := range list {
				if a.QuestionID != q.ID {
					continue
				}
				if len(a.OptionIDs) > 0 || strings.TrimSpace(a.Text) != "" {
					stat.Answered++
				}
				for _, oid := range a.OptionIDs {
					optionCount.Add(fmt.Sprint(oid), 1)
				}
				for _, w := range tokenize(a.Text) {
					wordCount.Add(w, 1)
				}
			}
		}
		for _, o := range q.Options {
			c := optionCount.Get(fmt.Sprint(o.ID))
			stat.Options = append(stat.Options, OptionStat{ID: o.ID, Label: o.Label, Count: c, Percent: percent(c, stat.Answered)})
		}
		items := make([]ds.WordCount, 0, wordCount.Len())
		for _, kv := range wordCount.Items() {
			items = append(items, ds.WordCount{Word: kv.Key, Count: kv.Val})
		}
		for _, w := range ds.TopK(items, 10) {
			stat.Words = append(stat.Words, WordStat{Word: w.Word, Count: w.Count})
		}
		st.Questions = append(st.Questions, stat)
	}
	return st, nil
}

func percent(part, total int) float64 {
	if total <= 0 {
		return 0
	}
	return math.Round(float64(part)/float64(total)*1000) / 10
}

// dailyPoints 统计最近 7 天（含今天）的答卷数，缺失的日期补 0。
func dailyPoints(responses []model.Response) []DailyPoint {
	counts := ds.NewHashMap()
	for _, r := range responses {
		counts.Add(time.Unix(r.SubmittedAt, 0).Format("2006-01-02"), 1)
	}
	out := make([]DailyPoint, 0, 7)
	now := time.Now()
	for i := 6; i >= 0; i-- {
		day := now.AddDate(0, 0, -i).Format("2006-01-02")
		out = append(out, DailyPoint{Date: day, Count: counts.Get(day)})
	}
	return out
}

// tokenize 做简易分词：按空白、标点和符号切开，丢掉单字。
// 课程设计不引中文分词库，够画词频图就行。
func tokenize(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	raw := strings.FieldsFunc(s, func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsPunct(r) || unicode.IsSymbol(r)
	})
	out := make([]string, 0, len(raw))
	for _, w := range raw {
		if len([]rune(w)) >= 2 {
			out = append(out, w)
		}
	}
	return out
}

// ExportData 是导出 CSV 需要的原始数据。
type ExportData struct {
	Title     string
	Questions []model.Question
	Responses []model.Response
	Answers   map[int64][]model.Answer
}

// Export 取导出所需数据。
func (u *Usecase) Export(surveyID, ownerID int64) (ExportData, error) {
	sv, err := u.owned(surveyID, ownerID)
	if err != nil {
		return ExportData{}, err
	}
	if sv.Questions, err = u.store.Questions(surveyID); err != nil {
		return ExportData{}, err
	}
	responses, err := u.store.ListResponses(surveyID)
	if err != nil {
		return ExportData{}, err
	}
	answers, err := u.store.AllAnswers(surveyID)
	if err != nil {
		return ExportData{}, err
	}
	return ExportData{Title: sv.Title, Questions: sv.Questions, Responses: responses, Answers: answers}, nil
}
