package server

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed static/dist
var embeddedStatic embed.FS

func (s *Server) registerStatic() {
	staticFiles, err := fs.Sub(embeddedStatic, "static/dist")
	if err != nil {
		panic(err)
	}
	indexHTML, err := fs.ReadFile(staticFiles, "index.html")
	if err != nil {
		panic(err)
	}
	fileServer := http.FileServer(http.FS(staticFiles))

	s.r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			fail(c, http.StatusNotFound, "接口不存在")
			return
		}

		requestPath := strings.TrimPrefix(c.Request.URL.Path, "/")
		if requestPath == "" {
			c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
			return
		}

		if file, err := staticFiles.Open(requestPath); err == nil {
			_ = file.Close()
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}

		c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
	})
}
