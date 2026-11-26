// Package web 用 embed 把前端构建产物打进二进制，部署时只拷一个可执行文件。
package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var dist embed.FS

// FS 返回 web/dist 的只读文件系统。
func FS() (fs.FS, error) {
	return fs.Sub(dist, "dist")
}
