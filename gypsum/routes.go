package gypsum

import (
	"fmt"
	"log"
	"strconv"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

func serve() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(static.Serve("/", static.LocalFile("./public", true)))

	authorized := r.Group("/api/v1", gin.BasicAuth(gin.Accounts{
		Config.Username: Config.Password,
	}))
	authorized.GET("/rules", func(c *gin.Context) {
		c.JSON(200, rules)
	})
	authorized.GET("/rules/:rid", func(c *gin.Context) {
		ruleIdStr := c.Param("rid")
		ruleId, err := strconv.ParseUint(ruleIdStr, 10, 64)
		if err != nil {
			c.JSON(404, gin.H{
				"code":    100,
				"message": "no such rule",
			})
		} else {
			r, ok := rules[ruleId]
			if ok {
				c.JSON(200, r)
			} else {
				c.JSON(404, gin.H{
					"code":    100,
					"message": "no such rule",
				})
			}
		}
	})
	authorized.POST("/rules", func(c *gin.Context) {
		var rule Rule
		if err := c.BindJSON(&rule); err != nil {
			c.JSON(400, gin.H{
				"code":    200,
				"message": fmt.Sprintf("converting error: %s", err),
			})
			return
		}
		cursor++
		_ = db.Put([]byte("gypsum-$meta-cursor"), ToBytes(cursor), nil)
		v, err := rule.ToBytes()
		if err != nil {
			c.JSON(400, gin.H{
				"code":    200,
				"message": fmt.Sprintf("converting error: %s", err),
			})
			return
		} else {
			_ = db.Put(append([]byte("gypsum-rules-"), ToBytes(cursor)...), v, nil)
			_ = rule.Register(cursor)
			rules[cursor] = rule
			c.JSON(201, gin.H{
				"code":    0,
				"message": "ok",
			})
			return
		}
	})
	authorized.DELETE("/rules/:rid", func(c *gin.Context) {
		ruleIdStr := c.Param("rid")
		ruleId, err := strconv.ParseUint(ruleIdStr, 10, 64)
		if err != nil {
			c.JSON(404, gin.H{
				"code":    100,
				"message": "no such rule",
			})
		} else {
			_, ok := rules[ruleId]
			if ok {
				delete(rules, ruleId)
				_ = db.Delete(append([]byte("gypsum-rules-"), ToBytes(ruleId)...), nil)
				zeroMatcher[ruleId].Delete()
				c.JSON(200, gin.H{
					"code":    0,
					"message": "deleted",
				})
			} else {
				c.JSON(404, gin.H{
					"code":    100,
					"message": "no such rule",
				})
			}
		}
	})
	authorized.PUT("/rules/:rid", func(c *gin.Context) {
		ruleIdStr := c.Param("rid")
		ruleId, err := strconv.ParseUint(ruleIdStr, 10, 64)
		if err != nil {
			c.JSON(404, gin.H{
				"code":    100,
				"message": "no such rule",
			})
			return
		}
		_, ok := rules[ruleId]
		if !ok {
			c.JSON(404, gin.H{
				"code":    100,
				"message": "no such rule",
			})
			return
		}
		var rule Rule
		if err := c.BindJSON(&rule); err != nil {
			c.JSON(400, gin.H{
				"code":    200,
				"message": fmt.Sprintf("converting error: %s", err),
			})
			return
		}
		v, err := rule.ToBytes()
		if err != nil {
			c.JSON(400, gin.H{
				"code":    200,
				"message": fmt.Sprintf("converting error: %s", err),
			})
			return
		}
		db.Put(append([]byte("gypsum-rules-"), ToBytes(ruleId)...), v, nil)
		zeroMatcher[ruleId].Delete()
		rule.Register(ruleId)
		rules[ruleId] = rule
		c.JSON(200, gin.H{
			"code":    0,
			"message": "ok",
		})
		return
	})
	err := r.Run(Config.Listen)
	if err != nil {
		log.Printf("binding address error: %s", err)
		//panic(err)
	}
}
