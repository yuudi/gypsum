package gypsum

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

//go:embed public/*
var publicAssets embed.FS

func serveWeb() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	authorized := r.Group("/api/v1", gin.BasicAuth(gin.Accounts{
		Config.Username: Config.Password,
	}))
	authorized.GET("/groups", getGroups)
	authorized.GET("/groups/:gid", getGroupByID)
	authorized.POST("/groups", createGroup)
	authorized.POST("/groups/:gid/groups", createGroup)
	authorized.PUT("/groups/:gid/items/:type/:iid", addGroupItem)
	authorized.GET("/groups/:gid/archive", exportGroup)
	//authorized.POST("/groups", importGroup)
	//authorized.POST("/groups/:gid/groups", importGroup)
	authorized.DELETE("/groups/:gid", deleteGroup)
	authorized.PATCH("/groups/:gid", renameGroup)
	authorized.GET("/rules", getRules)
	authorized.GET("/rules/:rid", getRuleByID)
	authorized.POST("/rules", createRule)
	authorized.POST("/groups/:gid/rules", createRule)
	authorized.DELETE("/rules/:rid", deleteRule)
	authorized.PUT("/rules/:rid", modifyRule)
	authorized.GET("/triggers", getTriggers)
	authorized.GET("/triggers/:tid", getTriggerByID)
	authorized.POST("/triggers", createTrigger)
	authorized.POST("/groups/:gid/triggers", createTrigger)
	authorized.DELETE("/triggers/:tid", deleteTrigger)
	authorized.PUT("/triggers/:tid", modifyTrigger)
	authorized.GET("/jobs", getJobs)
	authorized.GET("/jobs/:jid", getJobByID)
	authorized.POST("/jobs", createJob)
	authorized.POST("/groups/:gid/jobs", createJob)
	authorized.DELETE("/jobs/:jid", deleteJob)
	authorized.PUT("/jobs/:jid", modifyJob)
	authorized.GET("/resources", getResources)
	authorized.GET("/resources/:rid", getResourceByID)
	authorized.GET("/resources/:rid/content", downloadResource)
	authorized.POST("/resources/:name", uploadResource)
	authorized.POST("/groups/:gid/resources/:name", uploadResource)
	authorized.DELETE("/resources/:rid", deleteResource)
	authorized.PATCH("/resources/:rid", renameResource)

	assets, err := fs.Sub(publicAssets, "public")
	if err != nil {
		log.Fatal("directory `public` not compiled")
	}
	publicHttpFs := http.FS(assets)

	r.NoRoute(func(c *gin.Context) {
		c.FileFromFS(c.Request.URL.Path, publicHttpFs)
	}, func(c *gin.Context) {
		// wildcard for history router
		c.FileFromFS("index.html", publicHttpFs)
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
