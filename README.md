# Questioning · 在线调研问卷管理系统

> 数据结构课程设计 · Vue 3 + go-kratos v3 + SQLite · 前端编进二进制，一个文件跑起来

自己就能用的问卷系统：后台建卷、拖拽排版、给选项配跳题，填写端手机 / 平板 / 桌面自适应，回收后实时出图，一键导出 CSV。

课程设计的重点在 `pkg/ds/`：题目排序用双向链表、选项计数用开放寻址哈希表、答卷分页用跳表、跳题逻辑用图 + 环检测、高频词用堆、答卷索引用 LRU，全部手写并带单元测试，不引第三方容器库。

---

## 功能

1. **账号**：注册 / 登录，bcrypt 存口令，自实现 HS256 JWT；每个用户只能看到、操作自己的问卷。
2. **问卷管理**：新建、改标题和说明、匿名开关、发布、关闭、删除；列表支持关键词搜索（前端过滤）。
3. **拖拽式编辑器**
   - 4 种题型：单选、多选、填空、评分（1–5）
   - 题目卡片可直接拖拽重排，也可以用 ↑ ↓ 按钮微调；支持删除
   - 题目属性：题型、题干、补充说明、是否必答、选项增删
   - **跳题**：单选 / 多选题的每个选项都能配置「选了它之后跳到第几题」，0 表示按顺序进入下一题；保存时做环检测，A→B→A 这种死循环直接报 400
   - 撤销 / 重做（按钮，或 Ctrl+Z / Ctrl+Shift+Z）
4. **填写端**：公开链接 `/s/:id`，手机单列、桌面居中卡片；进度条、必答校验、上一题 / 下一题；严格按跳题逻辑只走相关题目。
5. **实时统计**：回收量、平均耗时、近 7 天回收趋势；每题按类型出图——单选 / 多选环形图 + 明细、评分分布柱状图、填空高频词排行。
6. **答卷浏览与导出**：分页浏览答卷、点开看单份详情；一键导出 CSV（带 BOM，Excel 打开不乱码；多选答案用 `|` 连接）。

---

## 技术栈

| 层次 | 选型 | 说明 |
| --- | --- | --- |
| 后端框架 | `github.com/go-kratos/kratos/v3` **v3.0.0** | HTTP transport + filter（日志 / panic 恢复），配置用 `gopkg.in/yaml.v3` 直读 |
| 语言 | Go 1.25 | kratos v3 要求 ≥ 1.25；Linux / Windows 都能直接编 |
| 数据库 | SQLite（`modernc.org/sqlite`） | 纯 Go 实现，**不需要 cgo / gcc**；库文件 `data/questioning.db` |
| 前端 | Vue 3 + Vite + Vue Router | `<script setup>` 单文件组件（纯 JavaScript，不用 TS，编译快）；组件用 Naive UI，图表用 ECharts |
| 前后端结合 | `//go:embed` | `web/dist` 编进二进制，后端顺带托管静态页，**没有第二个服务、没有代理、没有 CORS** |

---

## 目录结构

```
Questioning/
├── cmd/server/main.go       # 入口：装配 conf → data → biz → service → server
├── configs/config.yaml      # 端口、库路径、JWT 密钥、种子账号
├── internal/
│   ├── conf/                # 配置结构体（yaml.v3）
│   ├── data/                # 建表、仓储（SQLite）
│   ├── biz/                 # 业务：问卷 / 答题 / 跳题路径 / 统计 / 导出
│   ├── model/               # 领域模型
│   ├── service/             # HTTP 接口层：DTO、参数校验、错误信息、CSV
│   └── server/              # kratos http.Server、路由、静态页兜底
├── pkg/
│   ├── ds/                  # ★ 手写数据结构 + 单元测试
│   │   ├── list.go          # 双向链表
│   │   ├── stack.go         # 链式栈（DFS 用）
│   │   ├── hashmap.go       # 开放寻址哈希表
│   │   ├── skiplist.go      # 跳表
│   │   ├── graph.go         # 邻接表 + 环检测 + 可达
│   │   ├── heap.go          # 二叉堆（Top-K）
│   │   ├── lru.go           # LRU 缓存
│   │   └── ds_test.go       # 七个数据结构的单元测试
│   └── jwt/                 # HS256 签发 / 校验
├── web/                     # Vue 3 前端（独立 package.json）
│   ├── src/pages/           # 登录、问卷列表、编辑器、统计、填写端
│   ├── src/api.js           # fetch 封装
│   ├── src/router.js        # 路由 + 登录守卫
│   └── dist/                # 构建产物，随仓库提交（见快速开始）
├── docs/api.md              # 接口约定
└── README.md
```

---

## 数据结构设计与算法说明

`pkg/ds/` 全部手写，`go test ./pkg/ds/...` 跑单测。

| 业务场景 | 数据结构 | 关键操作与复杂度 |
| --- | --- | --- |
| 题目拖拽排序、整卷保存时重排 `order_no` | 双向链表 `DoublyList` | 已有节点换位 O(1)，按顺序输出 O(n) |
| 跳题路径求解、图的迭代 DFS | 链式栈 `Stack`（泛型） | push / pop / peek O(1) |
| 选项票数、填空词频 | 开放寻址哈希表 `HashMap` | 均摊 O(1)，负载因子 0.75 时扩容 2 倍并再散列 |
| 答卷按时间分页、区间筛选 | 跳表 `SkipList` | 插入 O(log n)，区间遍历 O(log n + k)，第 0 层链表支持倒序翻页 |
| 跳题逻辑校验与可达题集合 | 邻接表图 `Graph` | 环检测 O(V+E)（三色 DFS），可达集合 O(V+E) |
| 高频词 Top-K | 二叉堆 `MinHeap` | 建堆 O(n)，取前 k 个 O(n log k) |
| 答卷索引缓存 | `LRU`（哈希表 + 双向链表） | 命中 O(1)，容量 16 份问卷 |

几个算法点：

- **环检测**：整卷保存时把「题目 → 跳转目标」建成有向图（新题用负数占位 `-n` 表示第 n 题，保存时换成真实 id），DFS 三色标记找环，有环返回 400。
- **跳题路径**：交卷时从第 1 题出发，按「所选项的 `nextQuestionId`」逐步前进，得到本次实际作答路径。
- **必答校验**：只校验路径上的题目；路径外的题不参与校验，也不会写进答卷。
- **高频词**：填空答案按 unicode 标点 / 空白切分，丢弃单字，再用哈希表计数、小顶堆取前 10。
- **分页**：每份问卷的答卷建一棵跳表做索引（`submittedAt` 为 score，response id 为 value），分页和区间查询走跳表取 id 再回库，新答卷提交时让 LRU 失效并重建。

---

## 快速开始

前置：Go ≥ 1.25。`web/dist` 已经随仓库提交，克隆下来直接编，不需要装 Node：

```bash
go build -o questioning ./cmd/server
./questioning              # 默认 http://localhost:8000，配置见 configs/config.yaml
```

第一次编译要下依赖（kratos + 纯 Go 的 SQLite 驱动，几十 MB）；国内网络慢就加一句 `export GOPROXY=https://goproxy.cn,direct`。

Windows 就是 `go build -o questioning.exe ./cmd/server`，然后命令行或双击运行。

想改前端（需要 Node ≥ 20）：

```bash
cd web
npm install
npm run build              # 产物写到 web/dist
cd ..
go build -o questioning ./cmd/server   # 重新编译，前端被打进二进制
```

**种子账号**（首次启动自动建表并写入）：

| 账号 | 口令 | 角色 |
| --- | --- | --- |
| `admin` | `admin123` | 管理员 |
| `demo` | `demo123` | 普通用户 |

演示流程：`admin` 登录 → 新建问卷 → 加 3 道题（单选 / 填空 / 评分）→ 给单选选项配跳题 → 保存 → 发布 → 复制 `/s/1` 链接在浏览器手机尺寸下填写 → 回统计页看图表 → 导出 CSV。

**一个二进制就够了**：静态页、SQLite 都在本地；删掉 `data/questioning.db` 就是全新环境（下次启动自动重建并写种子账号）。

---

## 接口一览

统一前缀 `/api/v1`，JSON 收发，错误返回 HTTP 状态码 + `{"message":"中文"}`；除 `/auth/*` 和 `/public/*` 外都要 `Authorization: Bearer <token>`。详细字段见 `docs/api.md`。

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/auth/register` | 注册，返回 token + user |
| POST | `/auth/login` | 登录，返回 token + user |
| GET | `/auth/me` | 当前用户 |
| GET | `/surveys` | 当前用户的问卷列表 |
| POST | `/surveys` | 新建问卷 |
| GET | `/surveys/{id}` | 问卷详情（含题目、选项、跳题） |
| PUT | `/surveys/{id}` | 改标题 / 说明 / 匿名 |
| DELETE | `/surveys/{id}` | 删除问卷（连带题目、选项、答卷） |
| PUT | `/surveys/{id}/status` | 发布 / 关闭（`{"status":"published"}`） |
| PUT | `/surveys/{id}/questions` | 整卷题目覆盖保存（顺序、选项、跳题），保存前环检测 |
| GET | `/public/surveys/{id}` | **公开**取卷（只返回已发布的） |
| POST | `/public/surveys/{id}/submit` | **公开**提交答卷 |
| GET | `/surveys/{id}/stats` | 统计概览 + 每题图表数据 |
| GET | `/surveys/{id}/responses` | 答卷列表（`page` / `size` / `from` / `to`） |
| GET | `/surveys/{id}/responses/{rid}` | 单份答卷详情 |
| GET | `/surveys/{id}/export.csv` | 导出 CSV |

表：`users`、`surveys`、`questions`、`options`、`responses`、`answers`。

---

## 几个约定（保持简单）

- **单体**：不拆服务、不引 etcd / consul / Redis / MQ，kratos 只用 HTTP transport，进程内 LRU 做缓存。
- **不用 protoc，也不用 wire**：手写路由 + DTO，`cmd/server/main.go` 里手动装配依赖，克隆下来 `go run ./cmd/server` 就能跑。
- **一个进程一个端口**：前端静态页由后端直接托管，浏览器只连 `:8000`；非 `/api` 的未知路径回 `index.html`，刷新不会 404。
- **不加额外依赖**：不引 ORM、不引中文分词库，词频按标点切分。

---

## 开发进度

- [x] M1 骨架：kratos 启动、配置加载、SQLite 建表 + 种子账号、JWT 鉴权
- [x] M2 `pkg/ds/` 七个数据结构 + 单元测试
- [x] M3 问卷 / 题目接口，整卷保存 + 环检测
- [x] M4 前端：登录、问卷列表、拖拽编辑器
- [x] M5 填写端：响应式、跳题、必答校验
- [x] M6 统计图表 + 答卷列表 + CSV 导出
- [x] M7 联调自测、`web/dist` 构建提交

## 许可

课程设计作品，MIT License。
