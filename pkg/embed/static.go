// Package embed 提供静态资源嵌入功能
package embed

import (
	"embed"
	"io/fs"
	"net/http"
	"path"

	"kube-tide/internal/utils/logger"
)

//go:embed web/dist/assets
//go:embed web/dist/index.html
var webFS embed.FS

// GetFileSystem 返回嵌入的文件系统
func GetFileSystem() http.FileSystem {
	log := logger.GetLogger()

	// 创建子文件系统，仅包含web/dist目录
	sub, err := fs.Sub(webFS, "web/dist")
	if err != nil {
		log.Error("创建子文件系统失败", logger.Error(err))
		return nil
	}

	return http.FS(sub)
}

// StaticHandler 创建静态资源处理器
func StaticHandler(prefix string) http.Handler {
	fileServer := http.FileServer(GetFileSystem())

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 删除前缀并清理路径
		p := path.Clean(r.URL.Path)

		// 如果请求路径是目录，尝试提供index.html
		if p == "/" || p == "" {
			r.URL.Path = "/"
		} else {
			r.URL.Path = p[len(prefix):]
		}

		fileServer.ServeHTTP(w, r)
	})
}
