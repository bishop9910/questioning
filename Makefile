# 在线调研问卷管理系统 —— 常用命令
#
#   make           等于 make build（先打包前端，再编译成单个二进制）
#   make build-go  只编译 Go（用仓库里已有的 web/dist）
#   make run       编译后直接跑起来，浏览器开 http://127.0.0.1:8000
#   make dev-web   单独起 Vite 开发服务器（改前端热更新）
#   make dev-api   单独起后端（配合 dev-web 用）
#   make check     格式化 + vet + 单测
#   make clean     删掉二进制、dist、本地数据库
#
# 需要：Go >= 1.25、Node >= 20 + pnpm（只构建前端时才需要）

APP    := questioning
PORT   ?= 8000
GO     ?= go
PNPM   ?= pnpm
CONFIG ?= configs/config.yaml

ifeq ($(OS),Windows_NT)
BIN := $(APP).exe
else
BIN := $(APP)
endif

.DEFAULT_GOAL := build
.PHONY: help build build-go web run dev-web dev-api test fmt vet tidy check clean

help:
	@echo "make build     - 构建前端并编译 $(BIN)"
	@echo "make build-go  - 只编译 Go（跳过前端）"
	@echo "make run       - 编译并启动，端口 $(PORT)"
	@echo "make dev-web   - 前端开发服务器"
	@echo "make dev-api   - 后端开发服务器"
	@echo "make check     - fmt + vet + test"
	@echo "make clean     - 清理产物"

# 前端：装依赖 + 打包到 web/dist（会被 go:embed 编进二进制）
web:
	cd web && $(PNPM) install
	cd web && $(PNPM) run build

build: web build-go

build-go:
	$(GO) build -o $(BIN) ./cmd/server

run: build-go
	./$(BIN) -conf $(CONFIG)

dev-web:
	cd web && $(PNPM) install && $(PNPM) run dev

dev-api:
	$(GO) run ./cmd/server -conf $(CONFIG)

test:
	$(GO) test ./...

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

tidy:
	$(GO) mod tidy

check: fmt vet test

clean:
	rm -f $(BIN) $(APP) $(APP).exe
	rm -rf web/dist web/node_modules data
