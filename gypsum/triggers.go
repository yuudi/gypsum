package gypsum

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"strconv"
	"strings"

	"github.com/flosch/pongo2"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb/util"
	zero "github.com/wdvxdr1123/ZeroBot"
	lua "github.com/yuin/gopher-lua"
)

type TriggerCategory int

type Trigger struct {
	DisplayName string  `json:"display_name"`
	Active      bool    `json:"active"`
	GroupsID    []int64 `json:"groups_id"`
	UsersID     []int64 `json:"users_id"`
	TriggerType string  `json:"trigger_type"`
	Response    string  `json:"response"`
	Priority    int     `json:"priority"`
	Block       bool    `json:"block"`
	ParentGroup uint64  `json:"-"`
}

var (
	triggers    map[uint64]*Trigger
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
		DisplayName: "",
		Active:      true,
		GroupsID:    []int64{},
		UsersID:     []int64{},
		TriggerType: "",
		Response:    "",
		Priority:    50,
		Block:       true,
		ParentGroup: 0,
	}
	buffer := bytes.Buffer{}
	buffer.Write(b)
	decoder := gob.NewDecoder(&buffer)
	err := decoder.Decode(t)
	return t, err
}

func noticeRule(noticeTypeStr string) zero.Rule {
	noticeType := strings.SplitN(noticeTypeStr, "/", 2)
	if len(noticeType) == 1 {
		return func(event *zero.Event, _ zero.State) bool {
			return event.DetailType == noticeType[0]
		}
	}
	return func(event *zero.Event, _ zero.State) bool {
		return event.DetailType == noticeType[0] && event.SubType == noticeType[1]
	}
}

func (t *Trigger) Register(id uint64) error {
	if !t.Active {
		return nil
	}
	tmpl, err := pongo2.FromString(t.Response)
	if err != nil {
		log.Errorf("模板预处理出错：%s", err)
		return err
	}
	zeroTrigger[id] = zero.OnNotice(noticeRule(t.TriggerType), groupsRule(t.GroupsID), usersRule(t.UsersID)).SetPriority(t.Priority).SetBlock(t.Block).Handle(templateTriggerHandler(*tmpl))
	return nil
}

func templateTriggerHandler(tmpl pongo2.Template) zero.Handler {
	return func(matcher *zero.Matcher, event zero.Event, state zero.State) zero.Response {
		var luaState *lua.LState
		defer func() {
			if luaState != nil {
				luaState.Close()
			}
		}()
		reply, err := tmpl.Execute(pongo2.Context{
			"matcher": matcher,
			"state":   state,
			"event": func() interface{} {
				e := make(map[string]interface{})
				if err := jsoniter.UnmarshalFromString(event.RawEvent.Raw, &e); err != nil {
					log.Errorf("error when decode event json: %s", err)
				}
				return e
			},
			"at_sender": func() string {
				if event.GroupID == 0 {
					log.Errorf("cannot at sender in event %s/%s", event.PostType, event.SubType)
					return ""
				}
				return fmt.Sprintf("[CQ:at,qq=%d]", event.UserID)
			},
			"approve": func() {
				if event.PostType != "request" {
					log.Warnf("cannot approve: event is not a request: %#v", event)
				}
				switch event.RequestType {
				case "friend":
					zero.SetFriendAddRequest(event.Flag, true, "")
				case "group":
					zero.SetGroupAddRequest(event.Flag, event.SubType, true, "")
				}
			},
			"_lua": luaState,
		})
		if err != nil {
			log.Errorf("渲染模板出错：%s", err)
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
	triggers = make(map[uint64]*Trigger)
	zeroTrigger = make(map[uint64]*zero.Matcher)
	iter := db.NewIterator(util.BytesPrefix([]byte("gypsum-triggers-")), nil)
	defer func() {
		iter.Release()
		if err := iter.Error(); err != nil {
			log.Errorf("载入数据错误：%s", err)
		}
	}()
	for iter.Next() {
		key := ToUint(iter.Key()[16:])
		value := iter.Value()
		t, e := TriggerFromByte(value)
		if e != nil {
			log.Errorf("无法加载规则%d：%s", key, e)
			continue
		}
		triggers[key] = t
		if e := t.Register(key); e != nil {
			log.Errorf("无法注册规则%d：%s", key, e)
			continue
		}
	}
}

func (t *Trigger) GetParentID() uint64 {
	return t.ParentGroup
}

func (t *Trigger) GetDisplayName() string {
	return t.DisplayName
}

func (t *Trigger) SaveToDB(idx uint64) error {
	v, err := t.ToBytes()
	if err != nil {
		return err
	}
	return db.Put(append([]byte("gypsum-triggers-"), U64ToBytes(idx)...), v, nil)
}

func (t *Trigger) NewParent(selfID, parentID uint64) error {
	v, err := t.ToBytes()
	if err != nil {
		return err
	}
	t.ParentGroup = parentID
	err = db.Put(append([]byte("gypsum-triggers-"), U64ToBytes(selfID)...), v, nil)
	return err
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
	parentStr := c.Param("gid")
	var parentID uint64
	if len(parentStr) == 0 {
		parentID = 0
	} else {
		var err error
		parentID, err = strconv.ParseUint(parentStr, 10, 64)
		if err != nil {
			c.JSON(404, gin.H{
				"code":    1000,
				"message": "no such group",
			})
			return
		}
	}
	parentGroup, ok := groups[parentID]
	if !ok {
		c.JSON(404, gin.H{
			"code":    1000,
			"message": "group not found",
		})
		return
	}

	trigger.ParentGroup = parentID
	// syntax check
	if err := checkTemplate(trigger.Response); err != nil {
		c.JSON(422, gin.H{
			"code":    2041,
			"message": fmt.Sprintf("template error: %s", err),
		})
		return
	}
	//save
	cursor++
	parentGroup.Items = append(parentGroup.Items, Item{
		ItemType:    TriggerItem,
		DisplayName: trigger.DisplayName,
		ItemID:      cursor,
	})
	if err := parentGroup.SaveToDB(parentID); err != nil {
		log.Error(err)
		c.JSON(500, gin.H{
			"code":    3000,
			"message": fmt.Sprintf("Server got itself into trouble: %s", err),
		})
		return
	}
	if err := db.Put([]byte("gypsum-$meta-cursor"), U64ToBytes(cursor), nil); err != nil {
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
	if err := db.Put(append([]byte("gypsum-triggers-"), U64ToBytes(cursor)...), v, nil); err != nil {
		c.JSON(500, gin.H{
			"code":    3000,
			"message": fmt.Sprintf("Server got itself into trouble: %s", err),
		})
		return
	}
	triggers[cursor] = &trigger
	c.JSON(201, gin.H{
		"code":       0,
		"message":    "ok",
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
	if !ok {
		c.JSON(404, gin.H{
			"code":    1000,
			"message": "no such trigger",
		})
		return
	}

	// remove self from parent
	if err := DeleteFromParent(oldTrigger.ParentGroup, triggerID); err != nil {
		log.Errorf("error when delete group %d from parent group %d: %s", triggerID, oldTrigger.ParentGroup, err)
	}

	// remove self from database
	delete(triggers, triggerID)
	if err := db.Delete(append([]byte("gypsum-triggers-"), U64ToBytes(triggerID)...), nil); err != nil {
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
	// check syntax
	if err := checkTemplate(newTrigger.Response); err != nil {
		c.JSON(422, gin.H{
			"code":    2041,
			"message": fmt.Sprintf("template error: %s", err),
		})
		return
	}
	oldMatcher, ok := zeroTrigger[triggerID]
	newTrigger.ParentGroup = oldTrigger.ParentGroup
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
	if err := newTrigger.Register(triggerID); err != nil {
		c.JSON(400, gin.H{
			"code":    2001,
			"message": fmt.Sprintf("trigger error: %s", err),
		})
		return
	}
	if err := newTrigger.SaveToDB(triggerID); err != nil {
		c.JSON(500, gin.H{
			"code":    3002,
			"message": fmt.Sprintf("Server got itself into trouble: %s", err),
		})
		return
	}
	triggers[triggerID] = &newTrigger
	if newTrigger.DisplayName != oldTrigger.DisplayName {
		if err = ChangeNameForParent(newTrigger.ParentGroup, triggerID, newTrigger.DisplayName); err != nil {
			log.Errorf("error when change trigger %d from parent group %d: %s", triggerID, newTrigger.ParentGroup, err)
		}
	}
	c.JSON(200, gin.H{
		"code":    0,
		"message": "ok",
	})
	return
}
