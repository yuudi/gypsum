package gypsum

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/flosch/pongo2"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/syndtr/goleveldb/leveldb/util"
	zero "github.com/wdvxdr1123/ZeroBot"
)

type TriggerCategory int

type Trigger struct {
	Active      bool   `json:"active"`
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
	t := &Trigger{
		Active:      true,
		GroupID:     0,
		UserID:      0,
		TriggerType: "",
		Response:    "",
		Priority:    50,
		Block:       true,
	}
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
	if !t.Active {
		return nil
	}
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
			"state":   state,
			"event": func() interface{} {
				e := make(map[string]interface{})
				if err := jsoniter.Unmarshal(event.RawEvent, &e); err != nil {
					log.Printf("error when decode event json: %s", err)
				}
				return e
			},
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

func loadTriggers() {
	triggers = make(map[uint64]Trigger)
	zeroTrigger = make(map[uint64]*zero.Matcher)
	iter := db.NewIterator(util.BytesPrefix([]byte("gypsum-triggers-")), nil)
	defer func() {
		iter.Release()
		if err := iter.Error(); err != nil {
			log.Printf("载入数据错误：%s", err)
		}
	}()
	for iter.Next() {
		key := ToUint(iter.Key()[16:])
		value := iter.Value()
		t, e := TriggerFromByte(value)
		if e != nil {
			log.Printf("无法加载规则%d：%s", key, e)
			continue
		}
		triggers[key] = *t
		if e := t.Register(key); e != nil {
			log.Printf("无法注册规则%d：%s", key, e)
			continue
		}
	}
}

func getTriggers(c *gin.Context) {
	c.JSON(200, triggers)
}

func getTriggerByID(c *gin.Context) {
	triggerIDStr := c.Param("tid")
	triggerID, err := strconv.ParseUint(triggerIDStr, 10, 64)
	if err != nil {
		c.JSON(404, gin.H{
			"code":    1000,
			"message": "no such trigger",
		})
	} else {
		t, ok := triggers[triggerID]
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
		"trigger_id": cursor,
	})
	return
}

func deleteTrigger(c *gin.Context) {
	triggerIDStr := c.Param("tid")
	triggerID, err := strconv.ParseUint(triggerIDStr, 10, 64)
	if err != nil {
		c.JSON(404, gin.H{
			"code":    1000,
			"message": "no such trigger",
		})
		return
	}
	oldTrigger, ok := triggers[triggerID]
	if ok {
		delete(triggers, triggerID)
		if err := db.Delete(append([]byte("gypsum-triggers-"), ToBytes(triggerID)...), nil); err != nil {
			c.JSON(500, gin.H{
				"code":    3001,
				"message": fmt.Sprintf("Server got itself into trouble: %s", err),
			})
			return
		}
		if oldTrigger.Active {
			zeroTrigger[triggerID].Delete()
		}
		c.JSON(200, gin.H{
			"code":    0,
			"message": "deleted",
		})
		return
	}
	c.JSON(404, gin.H{
		"code":    1000,
		"message": "no such trigger",
	})
}

func modifyTrigger(c *gin.Context) {
	triggerIDStr := c.Param("tid")
	triggerID, err := strconv.ParseUint(triggerIDStr, 10, 64)
	if err != nil {
		c.JSON(404, gin.H{
			"code":    1000,
			"message": "no such trigger",
		})
		return
	}
	oldTrigger, ok := triggers[triggerID]
	if !ok {
		c.JSON(404, gin.H{
			"code":    100,
			"message": "no such trigger",
		})
		return
	}
	var newTrigger Trigger
	if err := c.BindJSON(&newTrigger); err != nil {
		c.JSON(400, gin.H{
			"code":    2000,
			"message": fmt.Sprintf("converting error: %s", err),
		})
		return
	}
	v, err := newTrigger.ToBytes()
	if err != nil {
		c.JSON(400, gin.H{
			"code":    2000,
			"message": fmt.Sprintf("converting error: %s", err),
		})
		return
	}
	oldMatcher, ok := zeroTrigger[triggerID]
	if err := newTrigger.Register(triggerID); err != nil {
		c.JSON(400, gin.H{
			"code":    2001,
			"message": fmt.Sprintf("trigger error: %s", err),
		})
		return
	}
	if oldTrigger.Active {
		if !ok {
			c.JSON(500, gin.H{
				"code":    7012,
				"message": "error when delete old rule: matcher not found",
			})
			return
		}
		oldMatcher.Delete()
	}
	if err := db.Put(append([]byte("gypsum-triggers-"), ToBytes(triggerID)...), v, nil); err != nil {
		c.JSON(500, gin.H{
			"code":    3002,
			"message": fmt.Sprintf("Server got itself into trouble: %s", err),
		})
		return
	}
	triggers[triggerID] = newTrigger
	c.JSON(200, gin.H{
		"code":    0,
		"message": "ok",
	})
	return
}
