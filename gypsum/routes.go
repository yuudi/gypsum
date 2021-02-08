package gypsum

import (
	"embed"
	"io/fs"
	"net/http"

	gzipForGin "github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	gzipStatic "github.com/yuudi/gypsum/helper/gin_middleware_gzip_static"
)

//go:generate gzip -rk9 web

//go:embed web/assets
var publicAssets embed.FS

//go:embed web/index.html
var publicIndex []byte

//go:embed web/index.html.gz
var publicIndexGz []byte

func serveWeb() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	api := r.Group("/api/v1")
	api.Use(gzipForGin.Gzip(gzipForGin.BestSpeed))
	api.Use(gin.BasicAuth(gin.Accounts{
		Config.Username: Config.Password,
	}))

	api.GET("/groups", getGroups)
	api.GET("/groups/:gid", getGroupByID)
	api.POST("/groups", createGroup)
	api.POST("/groups/:gid/groups", createGroup)
	api.PUT("/groups/:gid/items/:type/:iid", addGroupItem)
	api.GET("/groups/:gid/archive", exportGroup)
	api.DELETE("/groups/:gid", deleteGroup)
	api.PATCH("/groups/:gid", renameGroup)
	api.GET("/rules", getRules)
	api.GET("/rules/:rid", getRuleByID)
	api.POST("/rules", createRule)
	api.POST("/groups/:gid/rules", createRule)
	api.DELETE("/rules/:rid", deleteRule)
	api.PUT("/rules/:rid", modifyRule)
	api.GET("/triggers", getTriggers)
	api.GET("/triggers/:tid", getTriggerByID)
	api.POST("/triggers", createTrigger)
	api.POST("/groups/:gid/triggers", createTrigger)
	api.DELETE("/triggers/:tid", deleteTrigger)
	api.PUT("/triggers/:tid", modifyTrigger)
	api.GET("/jobs", getJobs)
	api.GET("/jobs/:jid", getJobByID)
	api.POST("/jobs", createJob)
	api.POST("/groups/:gid/jobs", createJob)
	api.DELETE("/jobs/:jid", deleteJob)
	api.PUT("/jobs/:jid", modifyJob)
	api.GET("/resources", getResources)
	api.GET("/resources/:rid", getResourceByID)
	api.GET("/resources/:rid/content", downloadResource)
	api.POST("/resources/:name", uploadResource)
	api.POST("/groups/:gid/resources/:name", uploadResource)
	api.DELETE("/resources/:rid", deleteResource)
	api.PATCH("/resources/:rid", renameResource)

	webAssets, err := fs.Sub(publicAssets, "web/assets")
	if err != nil {
		log.Fatal("directory `web` not compiled")
	}
	//assetsGroup := r.Group("/assets")
	//assetsGroup.Use(gzipStatic.ServeGzipStatic)
	//assetsGroup.StaticFS("/", http.FS(webAssets))

	r.GET("/assets/*filepath", gzipStatic.ServeGzipStatic(http.FS(webAssets)))

	homePage := func(c *gin.Context) {
		if gzipStatic.ShouldCompress(c.Request) {
			c.Header("Vary", "Accept-Encoding")
			c.Header("Content-Encoding", "gzip")
			c.Data(200, "text/html; charset=utf-8", publicIndexGz)
		} else {
			c.Data(200, "text/html", publicIndex)
		}
	}

	r.GET("/index.html", homePage)
	r.GET("/", homePage)
	//// wildcard for history router
	//r.NoRoute(homePage)
	r.NoRoute(func(c *gin.Context) {
		c.Data(404, "text/plain", []byte("404 Not Found"))
	})

	err = r.Run(Config.Listen)
	if err != nil {
		log.Errorf("binding address error: %s", err)
		// panic(err)
	}
}

type RestError struct {
	Status  int
	Code    int
	Message string
}

func (r RestError) Error() string {
	return r.Message
}
