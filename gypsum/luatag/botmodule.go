package luatag

import (
	"time"

	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	lua "github.com/yuin/gopher-lua"
	luaJson "layeh.com/gopher-json"

	"github.com/yuudi/gypsum/gypsum/helper/cqcode"
)

func botModLoaderFunc(event *zero.Event) lua.LGFunction {
	return func(L *lua.LState) int {
		mod := L.NewTable()
		L.SetFuncs(mod, map[string]lua.LGFunction{
			"api":          botApi,
			"send_private": sendPrivateMessage,
			"send_group":   sendGroupMessage,
			"send":         sendToEvent(event),
			"get":          getNextMessage(event),
			"approve":      approveToEvent(event),
			"withdraw":     withdrawEventMessage(event),
			"set_title":    setTitleToEvent(event),
			"group_ban":    setGroupBanToEvent(event),
		})
		L.Push(mod)
		return 1
	}
}

func botApi(L *lua.LState) int {
	action := L.ToString(1)
	luaParams := L.ToTable(2)
	params := make(map[string]interface{})
	if luaParams == nil {
		params = nil
	} else {
		luaParams.ForEach(func(k lua.LValue, v lua.LValue) {
			key := k.String()
			switch v.Type() {
			case lua.LTString:
				params[key] = v.String()
			case lua.LTNumber:
				params[key] = float64(v.(lua.LNumber))
			case lua.LTBool:
				params[key] = bool(v.(lua.LBool))
			default:
				log.Errorf("error when calling api from lua: cannot use type %s", v.Type().String())
			}
		})
	}
	result := zero.CallAction(action, params)
	luaResult, _ := luaJson.Decode(L, []byte(result.Raw))
	L.Push(luaResult)
	return 1
}

func sendToEvent(event *zero.Event) lua.LGFunction {
	return func(L *lua.LState) int {
		if event == nil {
			log.Warn("cannot send without event")
			L.Push(lua.LNil)
			L.Push(lua.LString("cannot send without event"))
			return 2
		}
		msg := L.ToString(1)
		if msg == "" {
			L.Push(lua.LNil)
			L.Push(lua.LString("cannot send empty message"))
			return 2
		}
		safe := L.ToBool(2)
		if !safe {
			msg = cqcode.Escape(msg)
		}
		messageID := zero.Send(*event, msg)
		L.Push(lua.LNumber(messageID))
		return 1
	}
}

func sendPrivateMessage(L *lua.LState) int {
	userID := int64(L.ToNumber(1))
	if userID == 0 {
		L.Push(lua.LNil)
		L.Push(lua.LString("cannot send without user_id"))
		return 2
	}
	message := L.ToString(2)
	if len(message) == 0 {
		L.Push(lua.LNil)
		L.Push(lua.LString("cannot send, message is empty"))
		return 2
	}
	safe := L.ToBool(3)
	if !safe {
		message = cqcode.Escape(message)
	}
	messageID := zero.SendPrivateMessage(userID, message)
	L.Push(lua.LNumber(messageID))
	return 1
}

func sendGroupMessage(L *lua.LState) int {
	groupID := int64(L.ToNumber(1))
	if groupID == 0 {
		L.Push(lua.LNil)
		L.Push(lua.LString("cannot send without group_id"))
		return 2
	}
	message := L.ToString(2)
	if len(message) == 0 {
		L.Push(lua.LNil)
		L.Push(lua.LString("cannot send, message is empty"))
		return 2
	}
	safe := L.ToBool(3)
	if !safe {
		message = cqcode.Escape(message)
	}
	messageID := zero.SendGroupMessage(groupID, message)
	L.Push(lua.LNumber(messageID))
	return 1
}

func getNextMessage(event *zero.Event) lua.LGFunction {
	return func(L *lua.LState) int {
		var rules []zero.Rule
		var userID, groupID int64
		argUserID := L.Get(1)
		if argUserID == lua.LNil {
			if event == nil {
				L.Push(lua.LNil)
				L.Push(lua.LString("cannot infer user, there is no event"))
				return 2
			}
			userID = event.UserID
		} else {
			userID = int64(lua.LVAsNumber(argUserID))
		}
		if userID != 0 {
			rules = append(rules, zero.CheckUser(userID))
		}

		argGroupID := L.Get(2)
		if argGroupID == lua.LNil {
			if event == nil {
				L.Push(lua.LNil)
				L.Push(lua.LString("cannot infer group, there is no event"))
				return 2
			}
			groupID = event.GroupID
		} else {
			groupID = int64(lua.LVAsNumber(argGroupID))
		}
		if groupID == 0 {
			// 0 表示私聊
			rules = append(rules, func(event *zero.Event, _ zero.State) bool {
				return event.MessageType == "private"
			})
		} else {
			rules = append(rules, func(ev *zero.Event, _ zero.State) bool {
				return ev.GroupID == groupID
			})
		}
		timeout := L.ToNumber(3)
		if timeout == 0 {
			timeout = 30
		}
		timeoutDuration := time.Duration(float64(timeout) * float64(time.Second))

		filterFunc := L.ToFunction(4)
		if filterFunc != nil {
			cp := lua.P{
				Fn:      filterFunc,
				NRet:    1,
				Protect: true, // return error instead of panic
				Handler: nil,
			}
			userDefinedRule := func(event *zero.Event, state zero.State) bool {
				luaEvent, err := luaJson.Decode(L, []byte(event.RawEvent.Raw))
				if err != nil {
					panic(err)
				}
				err = L.CallByParam(cp, luaEvent)
				if err != nil {
					log.Error("lua filter function execution error: " + err.Error())
					return false
				}
				defer L.Pop(1)
				return L.ToBool(-1)
			}
			rules = append(rules, userDefinedRule)
		}

		message := make(chan string)
		tempMather := zero.Matcher{
			Temp:     true,
			Block:    true,
			Priority: 1,
			State:    map[string]interface{}{},
			Type:     zero.Type("message"),
			Rules:    rules,
			Handler: func(_ *zero.Matcher, ev zero.Event, _ zero.State) zero.Response {
				message <- ev.RawMessage
				return zero.SuccessResponse
			},
		}
		zero.StoreTempMatcher(&tempMather)
		select {
		case reply := <-message:
			L.Push(lua.LString(reply))
			return 1
		case <-time.After(timeoutDuration):
			tempMather.Delete()
			L.Push(lua.LNil)
			L.Push(lua.LNil)
			return 2
		}
	}
}

func approveToEvent(event *zero.Event) lua.LGFunction {
	return func(L *lua.LState) int {
		if event == nil {
			log.Warn("cannot approve without event")
			L.Push(lua.LString("cannot approve without event"))
			return 1
		}
		if event.PostType != "request" {
			L.Push(lua.LString("cannot approve on event: " + event.PostType))
			return 1
		}
		switch event.RequestType {
		case "friend":
			go zero.SetFriendAddRequest(event.Flag, true, "")
		case "group":
			go zero.SetGroupAddRequest(event.Flag, event.SubType, true, "")
		}
		return 0
	}
}

func withdrawEventMessage(event *zero.Event) lua.LGFunction {
	return func(L *lua.LState) int {
		if event == nil {
			log.Warn("cannot withdraw without event")
			L.Push(lua.LString("cannot withdraw without event"))
			return 1
		}
		if event.PostType != "message" {
			L.Push(lua.LString("cannot withdraw on event: " + event.PostType))
			return 1
		}
		if event.GroupID == 0 {
			L.Push(lua.LString("cannot withdraw on message " + event.MessageType))
			return 1
		}
		go zero.DeleteMessage(event.MessageID)
		return 0
	}
}

func setTitleToEvent(event *zero.Event) lua.LGFunction {
	return func(L *lua.LState) int {
		if event == nil {
			log.Warn("cannot set title without event")
			L.Push(lua.LString("cannot set title without event"))
			return 1
		}
		if event.GroupID == 0 {
			L.Push(lua.LString("cannot set title without group"))
			return 1
		}
		title := L.ToString(1)
		targetID := int64(L.ToNumber(2))
		if targetID == 0 {
			targetID = event.UserID
		}
		go zero.SetGroupSpecialTitle(event.GroupID, targetID, title)
		return 0
	}
}

func setGroupBanToEvent(event *zero.Event) lua.LGFunction {
	return func(L *lua.LState) int {
		if event == nil {
			log.Warn("cannot ban without event")
			L.Push(lua.LString("cannot ban without event"))
			return 1
		}
		if event.GroupID == 0 {
			L.Push(lua.LString("cannot ban without group"))
			return 1
		}
		seconds := L.ToNumber(1)
		duration := int64(float64(seconds) * float64(time.Second))
		targetID := int64(L.ToNumber(2))
		if targetID == 0 {
			targetID = event.UserID
		}
		go zero.SetGroupBan(event.GroupID, targetID, duration)
		return 0
	}
}
