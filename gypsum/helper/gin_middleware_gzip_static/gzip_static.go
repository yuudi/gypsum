package gin_middleware_gzip_static

import (
	"embed"
	"mime"
	"net/http"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// ServeEmbedGzipStatic Serve Gzip Static
//
// *warning* only used by gypsum
// make sure all files are pre-compressed
func ServeEmbedGzipStatic(fs embed.FS, prefix, buildTag string) gin.HandlerFunc {
	prefix = strings.TrimPrefix(prefix, "/")
	return func(c *gin.Context) {
		if c.Request.Header.Get("Range") != "" {
			c.Status(416) // 懒得写了
			return
		}
		gzipped := false
		reqPath := c.Params.ByName("filepath")
		filePath := path.Join(prefix, reqPath)
		fileReq, err := fs.Open(filePath)
		if err != nil {
			c.Status(404)
			return
		}
		contentType := mime.TypeByExtension(filepath.Ext(reqPath))
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		if ShouldCompress(c.Request) {
			gzFile, err := fs.Open(filePath + ".gz")
			if err == nil {
				fileReq = gzFile
				gzipped = true
			}
		}
		stat, err := fileReq.Stat()
		if err != nil {
			c.Status(404)
			return
		}
		if stat.IsDir() {
			c.Status(404)
			return
		}
		if c.Request.Header.Get("If-None-Match") == buildTag {
			c.Status(304)
			return
		}
		contentLength := stat.Size()
		data := make([]byte, contentLength)
		_, err = fileReq.Read(data)
		if err != nil {
			c.Status(500)
			return
		}
		c.Header("Vary", "Accept-Encoding")
		c.Header("ETag", buildTag)
		c.Header("Content-Length", strconv.FormatInt(contentLength, 10))
		if gzipped {
			c.Header("Content-Encoding", "gzip")
		}
		c.Data(200, contentType, data)
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
