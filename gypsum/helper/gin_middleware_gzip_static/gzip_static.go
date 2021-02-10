package gin_middleware_gzip_static

import (
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// ServeGzipStatic Serve Gzip Static
//
// *warning* only used by gypsum
// make sure all files are pre-compressed
func ServeGzipStatic(fs http.FileSystem) gin.HandlerFunc {
	return func(c *gin.Context) {
		filePath := c.Params.ByName("filepath")
		_, err := fs.Open(filePath)
		if err != nil {
			c.Data(404, "text/plain", []byte("404 Not Found"))
			return
		}
		contentType := mime.TypeByExtension(filepath.Ext(filePath))
		c.Header("Content-Type", contentType)
		var ext string
		if ShouldCompress(c.Request) {
			c.Header("Content-Encoding", "gzip")
			ext = ".gz"
		} else {
			ext = ""
		}
		c.FileFromFS(filePath+ext, fs)
	}
}

func ShouldCompress(req *http.Request) bool {
	if !strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") ||
		strings.Contains(req.Header.Get("Connection"), "Upgrade") ||
		strings.Contains(req.Header.Get("Content-Type"), "text/event-stream") {

		return false
	}
	return true
}
