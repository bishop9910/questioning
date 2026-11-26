// Package server 组装 kratos HTTP 服务：路由注册、错误编码、静态页兜底。
package server

import (
	"encoding/json"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"path"
	"strings"
	"time"

	khttp "github.com/go-kratos/kratos/v3/transport/http"

	"questioning/internal/biz"
	"questioning/internal/conf"
	"questioning/internal/service"
	"questioning/web"
)

// NewHTTPServer 按配置建服务并把 /api/v1 下的路由挂上。
func NewHTTPServer(c *conf.Config, svc *service.Service, logger *slog.Logger) *khttp.Server {
	srv := khttp.NewServer(
		khttp.Address(c.Server.Addr),
		khttp.Timeout(time.Duration(c.Server.TimeoutSec)*time.Second),
		khttp.Filter(recoverFilter(logger), accessLog(logger)),
		khttp.ErrorEncoder(encodeError),
		khttp.NotFoundHandler(staticHandler(logger)),
	)

	r := srv.Route("/api/v1")

	r.POST("/auth/register", svc.Register)
	r.POST("/auth/login", svc.Login)
	r.GET("/auth/me", svc.Me)

	r.GET("/surveys", svc.ListSurveys)
	r.POST("/surveys", svc.CreateSurvey)
	r.GET("/surveys/{id}", svc.SurveyDetail)
	r.PUT("/surveys/{id}", svc.UpdateSurvey)
	r.DELETE("/surveys/{id}", svc.DeleteSurvey)
	r.PUT("/surveys/{id}/status", svc.SetStatus)
	r.PUT("/surveys/{id}/questions", svc.SaveQuestions)
	r.GET("/surveys/{id}/stats", svc.Stats)
	r.GET("/surveys/{id}/responses", svc.Responses)
	r.GET("/surveys/{id}/responses/{rid}", svc.ResponseDetail)
	r.GET("/surveys/{id}/export.csv", svc.ExportCSV)

	r.GET("/public/surveys/{id}", svc.PublicSurvey)
	r.POST("/public/surveys/{id}/submit", svc.Submit)

	return srv
}

// encodeError 统一错误格式：{"message": "..."}。
func encodeError(w http.ResponseWriter, r *http.Request, err error) {
	code, msg := 500, "服务器内部错误"
	var be *biz.Error
	if errors.As(err, &be) {
		code, msg = be.Status, be.Message
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": msg})
}

// recoverFilter 兜住 handler 里的 panic，避免整个进程被一个请求带崩。
func recoverFilter(logger *slog.Logger) khttp.FilterFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.Error("请求 panic", "path", r.URL.Path, "panic", rec)
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte(`{"message":"服务器内部错误"}`))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// accessLog 打印每个请求的方法、路径与耗时。
func accessLog(logger *slog.Logger) khttp.FilterFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			logger.Info("http", "method", r.Method, "path", r.URL.Path, "cost", time.Since(start).String())
		})
	}
}

// staticHandler 给前端产物做兜底：命中文件就发文件，否则回 index.html 交给前端路由。
func staticHandler(logger *slog.Logger) http.HandlerFunc {
	sub, err := web.FS()
	if err != nil {
		logger.Error("前端资源加载失败", "err", err)
		return func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "前端资源未打包，先执行 pnpm build", http.StatusInternalServerError)
		}
	}
	index, readErr := fs.ReadFile(sub, "index.html")
	if readErr != nil {
		logger.Warn("web/dist 里没有 index.html", "err", readErr)
	}
	files := http.FileServer(http.FS(sub))

	return func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"接口不存在"}`))
			return
		}
		clean := path.Clean("/" + r.URL.Path)
		if clean != "/" {
			if f, err := sub.Open(strings.TrimPrefix(clean, "/")); err == nil {
				f.Close()
				files.ServeHTTP(w, r)
				return
			}
		}
		if readErr != nil {
			http.Error(w, "前端资源未打包，先执行 pnpm build", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(index)
	}
}
