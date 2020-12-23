package gypsum

import (
	"log"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

func serveWeb() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.Use(static.Serve("/", static.LocalFile("./public", true)))

	authorized := r.Group("/api/v1", gin.BasicAuth(gin.Accounts{
		Config.Username: Config.Password,
	}))
	authorized.GET("/rules", getRules)
	authorized.GET("/rules/:rid", getRuleByID)
	authorized.POST("/rules", createRule)
	authorized.DELETE("/rules/:rid", deleteRule)
	authorized.PUT("/rules/:rid", modifyRule)
	authorized.GET("/triggers", getTriggers)
	authorized.GET("/triggers/:tid", getTriggerByID)
	authorized.POST("/triggers", createTrigger)
	authorized.DELETE("/triggers/:tid", deleteTrigger)
	authorized.PUT("/triggers/:tid", modifyTrigger)
	authorized.GET("/jobs", getJobs)
	authorized.GET("/jobs/:jid", getJobByID)
	authorized.POST("/jobs", createJob)
	authorized.DELETE("/jobs/:jid", deleteJob)
	authorized.PUT("/jobs/:jid", modifyJob)

	err := r.Run(Config.Listen)
	if err != nil {
		log.Printf("binding address error: %s", err)
		//panic(err)
	}
}
