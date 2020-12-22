package gypsum

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"strconv"
	"strings"

	"github.com/flosch/pongo2"
	zero "github.com/wdvxdr1123/ZeroBot"
)

type TriggerCategory int

type Trigger struct {
	GroupID     int64  `json:"group_id"`
	UserID      int64  `json:"user_id"`
	TriggerType string `json:"trigger_type"`
	Response    string `json:"response"`
	Priority    int    `json:"priority"`
	Block       bool   `json:"block"`
}

var (
	triggers    map[uint64]Trigger
	zeroTrigger map[uint64]*zero.Matcher
)

func (t *Trigger) ToBytes() ([]byte, error) {
	buffer := bytes.Buffer{}
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(t); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func TriggerFromByte(b []byte) (*Trigger, error) {
	t := &Trigger{}
	buffer := bytes.Buffer{}
	buffer.Write(b)
	decoder := gob.NewDecoder(&buffer)
	err := decoder.Decode(t)
	return t, err
}

func noticeRule(noticeType string) zero.Rule {
	ntype := strings.SplitN(noticeType, "/", 2)
	if len(ntype) == 1 {
		return func(event *zero.Event, _ zero.State) bool {
			return event.NoticeType == ntype[0]
		}
	}
	return func(event *zero.Event, _ zero.State) bool {
		return event.NoticeType == ntype[0] && event.SubType == ntype[1]
	}
}

func (t *Trigger) Register(id uint64) error {
	tmpl, err := pongo2.FromString(t.Response)
	if err != nil {
		log.Printf("模板预处理出错：%s", err)
		return err
	}
	zeroTrigger[id] = zero.OnNotice(noticeRule(t.TriggerType), groupRule(t.GroupID), userRule(t.UserID)).SetPriority(t.Priority).SetBlock(t.Block).Handle(templateTriggerHandler(*tmpl))
	return nil
}

func templateTriggerHandler(tmpl pongo2.Template) zero.Handler {
	return func(matcher *zero.Matcher, event zero.Event, state zero.State) zero.Response {
		reply, err := tmpl.Execute(pongo2.Context{
			"matcher": matcher,
			"event":   event,
			"state":   state,
		})
		if err != nil {
			log.Printf("渲染模板出错：%s", err)
			return zero.FinishResponse
		}
		reply = strings.TrimSpace(reply)
		if reply != "" {
			zero.Send(event, reply)
		}
		return zero.FinishResponse
	}
}

func getTriggers(c *gin.Context) {
	c.JSON(200, triggers)
}

func getTriggerById(c *gin.Context) {
	triggerIdStr := c.Param("tid")
	triggerId, err := strconv.ParseUint(triggerIdStr, 10, 64)
	if err != nil {
		c.JSON(404, gin.H{
			"code":    1000,
			"message": "no such trigger",
		})
	} else {
		t, ok := triggers[triggerId]
		if ok {
			c.JSON(200, t)
		} else {
			c.JSON(404, gin.H{
				"code":    1000,
				"message": "no such trigger",
			})
		}
	}
}

func createTrigger(c *gin.Context) {
	var trigger Trigger
	if err := c.BindJSON(&trigger); err != nil {
		c.JSON(400, gin.H{
			"code":    2000,
			"message": fmt.Sprintf("converting error: %s", err),
		})
		return
	}
	cursor++
	if err := db.Put([]byte("gypsum-$meta-cursor"), ToBytes(cursor), nil); err != nil {
		c.JSON(500, gin.H{
			"code":    3000,
			"message": fmt.Sprintf("Server got itself into trouble: %s", err),
		})
		return
	}
	v, err := trigger.ToBytes()
	if err != nil {
		c.JSON(400, gin.H{
			"code":    2000,
			"message": fmt.Sprintf("converting error: %s", err),
		})
		return
	}
	if err := trigger.Register(cursor); err != nil {
		c.JSON(400, gin.H{
			"code":    2001,
			"message": fmt.Sprintf("trigger error: %s", err),
		})
		return
	}
	if err := db.Put(append([]byte("gypsum-triggers-"), ToBytes(cursor)...), v, nil); err != nil {
		c.JSON(500, gin.H{
			"code":    3000,
			"message": fmt.Sprintf("Server got itself into trouble: %s", err),
		})
		return
	}
	triggers[cursor] = trigger
	c.JSON(201, gin.H{
		"code":    0,
		"message": "ok",
	})
	return
}

func deleteTrigger(c *gin.Context) {
	triggerIdStr := c.Param("tid")
	triggerId, err := strconv.ParseUint(triggerIdStr, 10, 64)
	if err != nil {
		c.JSON(404, gin.H{
			"code":    1000,
			"message": "no such trigger",
		})
	} else {
		_, ok := triggers[triggerId]
		if ok {
			delete(triggers, triggerId)
			if err := db.Delete(append([]byte("gypsum-triggers-"), ToBytes(triggerId)...), nil); err != nil {
				c.JSON(500, gin.H{
					"code":    3001,
					"message": fmt.Sprintf("Server got itself into trouble: %s", err),
				})
				return
			}
			zeroTrigger[triggerId].Delete()
			c.JSON(200, gin.H{
				"code":    0,
				"message": "deleted",
			})
		} else {
			c.JSON(404, gin.H{
				"code":    1000,
				"message": "no such trigger",
			})
		}
	}
}

func modifyTrigger(c *gin.Context) {
	triggerIdStr := c.Param("tid")
	triggerId, err := strconv.ParseUint(triggerIdStr, 10, 64)
	if err != nil {
		c.JSON(404, gin.H{
			"code":    1000,
			"message": "no such trigger",
		})
		return
	}
	_, ok := triggers[triggerId]
	if !ok {
		c.JSON(404, gin.H{
			"code":    100,
			"message": "no such trigger",
		})
		return
	}
	var trigger Trigger
	if err := c.BindJSON(&trigger); err != nil {
		c.JSON(400, gin.H{
			"code":    2000,
			"message": fmt.Sprintf("converting error: %s", err),
		})
		return
	}
	v, err := trigger.ToBytes()
	if err != nil {
		c.JSON(400, gin.H{
			"code":    2000,
			"message": fmt.Sprintf("converting error: %s", err),
		})
		return
	}
	matcher := zeroTrigger[triggerId]
	if err := trigger.Register(triggerId); err != nil {
		c.JSON(400, gin.H{
			"code":    2001,
			"message": fmt.Sprintf("trigger error: %s", err),
		})
		return
	}
	matcher.Delete()
	if err := db.Put(append([]byte("gypsum-triggers-"), ToBytes(triggerId)...), v, nil); err != nil {
		c.JSON(500, gin.H{
			"code":    3002,
			"message": fmt.Sprintf("Server got itself into trouble: %s", err),
		})
		return
	}
	triggers[triggerId] = trigger
	c.JSON(200, gin.H{
		"code":    0,
		"message": "ok",
	})
	return
}
