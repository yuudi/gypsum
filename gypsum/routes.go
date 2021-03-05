package gypsum

import (
	"embed"
	"encoding/base64"
	"math/rand"
	"path"
	"strings"

	gzipForGin "github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	gzipStatic "github.com/yuudi/gypsum/gypsum/helper/gin_middleware_gzip_static"
	"github.com/yuudi/gypsum/gypsum/helper/selfsign"
)

//go:generate gzip -rk9 web

//go:embed web
var publicAssets embed.FS

//go:embed web/index.html
var publicIndex []byte

//go:embed web/index.html.gz
var publicIndexGz []byte

func initWeb() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	api := r.Group("/api/v1")
	api.Use(gzipForGin.Gzip(gzipForGin.BestSpeed))

	initialLoginAuth()
	api.Use(authMiddleware)

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
	api.GET("/schedulers", getJobs)
	api.GET("/jobs/:jid", getJobByID)
	api.GET("/schedulers/:jid", getJobByID)
	api.POST("/jobs", createJob)
	api.POST("/schedulers", createJob)
	api.POST("/groups/:gid/jobs", createJob)
	api.POST("/groups/:gid/schedulers", createJob)
	api.DELETE("/jobs/:jid", deleteJob)
	api.DELETE("/schedulers/:jid", deleteJob)
	api.PUT("/jobs/:jid", modifyJob)
	api.PUT("/schedulers/:jid", modifyJob)
	api.GET("/resources", getResources)
	api.GET("/resources/:rid", getResourceByID)
	api.GET("/resources/:rid/content", downloadResource)
	api.POST("/resources/:name", uploadResource)
	api.POST("/groups/:gid/resources/:name", uploadResource)
	api.DELETE("/resources/:rid", deleteResource)
	api.PATCH("/resources/:rid", renameResource)

	// debug
	api.POST("/debug", userTest)

	// admin
	api.GET("/gypsum/update", getUpdateStatus)
	api.PUT("/gypsum/update", requestUpdateGypsum)
	// admin (non-auth)
	r.GET("/api/v1/gypsum/information", getGypsumInformation)
	r.PUT("/api/v1/gypsum/login", loginHandler)

	// resource backref
	r.GET("/contents/resources/:filename", serveResource)

	if Config.ExternalAssets == "" {
		// internal assets
		tagByte := make([]byte, 12)
		rand.Read(tagByte)
		r.GET("/assets/*filepath", gzipStatic.ServeEmbedGzipStatic(publicAssets, "/web/assets", base64.URLEncoding.EncodeToString(tagByte)))

		// home page
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
	} else {
		// external assets
		r.Static("/assets/", path.Join(Config.ExternalAssets, "assets"))
		homePage := func(c *gin.Context) {
			c.File(path.Join(Config.ExternalAssets, "index.html"))
		}
		r.GET("/index.html", homePage)
		r.GET("/", homePage)
	}

	//// wildcard for history router
	//r.NoRoute(homePage)
	r.NoRoute(func(c *gin.Context) {
		c.Data(404, "text/plain", []byte("404 Not Found"))
	})

	go serveWeb(r, Config.Listen)
}

func serveWeb(r *gin.Engine, listen string) {
	if strings.HasPrefix(listen, "http://") {
		err := r.Run(listen[len("http://"):])
		if err != nil {
			log.Fatalf("binding address error: %s", err)
		}
	} else if strings.HasPrefix(listen, "https://") {
		pub, priv, err := selfsign.GetTlsKeys()
		if err != nil {
			log.Fatalf("get tls key error: %s", err)
		}
		err = r.RunTLS(listen[len("https://"):], pub, priv)
		if err != nil {
			log.Fatalf("binding address error: %s", err)
		}
	} else if strings.HasPrefix(listen, "unix://") {
		err := r.RunUnix(listen[len("unix:/"):])
		if err != nil {
			log.Fatalf("binding address error: %s", err)
		}
	} else {
		err := r.Run(listen)
		if err != nil {
			log.Fatalf("binding address error: %s", err)
		}
	}
}
