// 问卷系统入口：读配置 → 开库 → 装配业务/接入/服务 → 跑 kratos。
package main

import (
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/go-kratos/kratos/v3"

	"questioning/internal/biz"
	"questioning/internal/conf"
	"questioning/internal/data"
	"questioning/internal/server"
	"questioning/internal/service"
)

func main() {
	configPath := flag.String("conf", "configs/config.yaml", "配置文件路径")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := conf.Load(*configPath)
	if err != nil {
		logger.Error("加载配置失败", "err", err)
		os.Exit(1)
	}

	store, err := data.Open(cfg.Database.Path)
	if err != nil {
		logger.Error("打开数据库失败", "err", err)
		os.Exit(1)
	}
	defer store.Close()

	if cfg.Seed.Enabled {
		seeded, err := store.Seed([]data.SeedUser{
			{Username: "admin", Password: "admin123", Nickname: "管理员", Role: "admin"},
			{Username: "demo", Password: "demo123", Nickname: "演示用户", Role: "user"},
		})
		if err != nil {
			logger.Error("写入演示账号失败", "err", err)
			os.Exit(1)
		}
		if seeded {
			logger.Info("已写入演示账号", "accounts", "admin/admin123, demo/demo123")
		}
	}

	secret := []byte(cfg.Server.JWTSecret)
	uc := biz.New(store, secret, time.Duration(cfg.Server.TokenTTLHours)*time.Hour)
	svc := service.New(uc, secret)
	srv := server.NewHTTPServer(cfg, svc, logger)

	app := kratos.New(
		kratos.Name("questioning"),
		kratos.Version("v1.0.0"),
		kratos.Logger(logger),
		kratos.Server(srv),
	)
	if err := app.Run(); err != nil {
		logger.Error("服务退出", "err", err)
		os.Exit(1)
	}
}
