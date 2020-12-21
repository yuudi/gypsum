package gypsum

import (
	"fmt"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

func serve() {
	if _, err := os.Stat("./public"); os.IsNotExist(err) {
		if err := os.Mkdir("./public", 0644); err != nil {
			log.Printf("%s", err)
		}
		if _, err := os.Stat("./public/index.html"); os.IsNotExist(err) {
			if err := ioutil.WriteFile("./public/index.html", []byte(simpleHtmlPage), 0644); err != nil {
				log.Printf("%s", err)
			}
		}
	}
	r := gin.Default()
	r.Use(static.Serve("/", static.LocalFile("./public", true)))
	r.GET("/api/v1/rules", func(c *gin.Context) {
		c.JSON(200, rules)
	})
	r.GET("/api/v1/rules/:rid", func(c *gin.Context) {
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
	r.POST("/api/v1/rules", func(c *gin.Context) {
		var rule Rule
		if err := c.BindJSON(&rule); err != nil {
			c.JSON(400, gin.H{
				"code":    200,
				"message": fmt.Sprintf("converting error: %s", err),
			})
			return
		}
		cursor += 1
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
			_ = rule.Register("rule" + strconv.FormatUint(cursor, 10))
			rules[cursor] = rule
			c.JSON(201, gin.H{
				"code":    0,
				"message": "ok",
			})
			return
		}
	})
	err := r.Run(Listen)
	if err != nil {
		log.Printf("binding address error: %s", err)
		//panic(err)
	}
}
