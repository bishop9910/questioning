# 接口约定

- 前缀：`/api/v1`
- 请求/响应：`application/json`（CSV 导出除外）
- 鉴权：`Authorization: Bearer <token>`；`/auth/*` 与 `/public/*` 不需要
- 时间：RFC3339 字符串，如 `2026-11-26T16:00:00+08:00`
- 出错：HTTP 4xx/5xx + `{"message":"中文提示"}`

## 数据对象

```jsonc
// User
{ "id": 1, "username": "admin", "nickname": "管理员", "role": "admin" }

// SurveyListItem
{ "id": 1, "title": "课程满意度调查", "description": "", "status": "draft",
  "questionCount": 3, "responseCount": 12,
  "createdAt": "...", "updatedAt": "...", "publishedAt": "" }

// Option   nextQuestionId: 0 表示不跳转（顺序下一题 / 交卷）
{ "id": 11, "label": "满意", "nextQuestionId": 0 }

// Question
{ "id": 1, "type": "single", "title": "你的年级", "description": "",
  "required": true, "options": [ ... ] }

// SurveyDetail
{ "id": 1, "title": "...", "description": "", "status": "draft", "anonymous": false,
  "createdAt": "...", "updatedAt": "...", "publishedAt": "",
  "questions": [ ... ] }
```

`type`：`single` 单选 / `multiple` 多选 / `text` 填空 / `rating` 评分（1–5，选项自动生成）。
`status`：`draft` / `published` / `closed`。

## 账号

| 方法 | 路径 | 请求 | 响应 |
| --- | --- | --- | --- |
| POST | `/auth/register` | `{"username","password","nickname"}` | `{"token","user"}` |
| POST | `/auth/login` | `{"username","password"}` | `{"token","user"}` |
| GET | `/auth/me` | - | `{"user"}` |

## 问卷

| 方法 | 路径 | 请求 | 响应 |
| --- | --- | --- | --- |
| GET | `/surveys` | query: `keyword` `status` `page`(1) `size`(10) | `{"total","list":[SurveyListItem]}` |
| POST | `/surveys` | `{"title","description"}` | `SurveyDetail` |
| GET | `/surveys/{id}` | - | `SurveyDetail` |
| PUT | `/surveys/{id}` | `{"title","description","anonymous"}` | `SurveyDetail` |
| DELETE | `/surveys/{id}` | - | `{}` |
| PUT | `/surveys/{id}/status` | `{"status":"published"}` | `SurveyDetail` |
| PUT | `/surveys/{id}/questions` | `{"questions":[Question]}` | `{"questions":[Question]}` |

`PUT /surveys/{id}/questions` 是整卷覆盖保存：

- 带 `id > 0` 的题目原地更新，`id == 0` 的插入；本次没出现的题目被删除（连带其答案）。
- 顺序按数组下标写 `order_no`（内部用双向链表重排）。
- 跳题目标必须指向本卷中的题目或 0；有环返回 400，`message` 里指出环上题目。

## 公开填写

| 方法 | 路径 | 请求 | 响应 |
| --- | --- | --- | --- |
| GET | `/public/surveys/{id}` | - | `SurveyDetail`（含 `questions[].options[].nextQuestionId`，填写页据此跳题） |
| POST | `/public/surveys/{id}/submit` | `{"answers":[{"questionId","optionIds":[],"text":""}],"durationSec":90}` | `{"id":5}` |

公开取卷只返回 `published` 的问卷（草稿或已关闭返回 403）。`nextQuestionId` 会一并返回，填写页用它实现选项跳题；0 表示顺序进入下一题。

## 统计与答卷

| 方法 | 路径 | 请求 | 响应 |
| --- | --- | --- | --- |
| GET | `/surveys/{id}/stats` | - | 见下 |
| GET | `/surveys/{id}/responses` | query: `page` `size` `from` `to`(unix 秒) | `{"total","list":[{"id","submittedAt","durationSec","device"}]}` |
| GET | `/surveys/{id}/responses/{rid}` | - | `{"id","submittedAt","durationSec","device","answers":[{"questionId","optionIds":[11],"text":""}]}` |
| GET | `/surveys/{id}/export.csv` | - | `text/csv` 附件 |

统计响应：

```jsonc
{
  "surveyId": 1, "title": "...", "total": 12, "avgDurationSec": 95,
  "daily": [ { "date": "2026-11-20", "count": 3 }, ... ],   // 近 7 天
  "questions": [
    { "id": 1, "type": "single", "title": "你的年级", "answered": 12,
      "options": [ { "id": 11, "label": "大一", "count": 7, "percent": 58.3 } ],
      "words": [] },
    { "id": 3, "type": "text", "title": "建议", "answered": 8,
      "options": [], "words": [ { "word": "食堂", "count": 5 } ] }
  ]
}
```

`percent` 保留一位小数；`words` 只有 `text` 题有。

## CSV 导出

表头：`答卷ID,提交时间,耗时(秒),设备` + 每题一列（题目标题）。多选答案用 `|` 连接，评分输出数字，填空输出原文。文件带 UTF-8 BOM。
