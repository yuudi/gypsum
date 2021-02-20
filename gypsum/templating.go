package gypsum

import (
	"fmt"
	"net/url"

	"github.com/flosch/pongo2"
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	lua "github.com/yuin/gopher-lua"

	"github.com/yuudi/gypsum/gypsum/helper"
	"github.com/yuudi/gypsum/gypsum/helper/cqcode"
	"github.com/yuudi/gypsum/gypsum/luatag"
	"github.com/yuudi/gypsum/gypsum/template"
)

func initTemplating() error {
	// replace default HTML filter to CQ filter
	if err := pongo2.ReplaceFilter("escape", filterEscapeCQCode); err != nil {
		return err
	}

	// enable auto-escape
	pongo2.SetAutoescape(true)

	if err := pongo2.RegisterFilter("silence", filterSilence); err != nil {
		return err
	}

	// register functions
	pongo2.Globals["at"] = template.At
	pongo2.Globals["res"] = resourcePathFunc(Config.ResourceShare)
	pongo2.Globals["image"] = template.Image
	pongo2.Globals["record"] = template.Record
	pongo2.Globals["sleep"] = template.Sleep
	pongo2.Globals["url_encode"] = url.QueryEscape
	pongo2.Globals["random_int"] = template.RandomInt
	pongo2.Globals["random_line"] = template.RandomLine
	pongo2.Globals["random_file"] = template.RandomFile
	pongo2.Globals["file_get_contents"] = template.FileGetContents
	pongo2.Globals["parse_json"] = template.ParseJson
	pongo2.Globals["db_get"] = template.DatabaseGet
	pongo2.Globals["db_put"] = template.DatabasePut

	// register tags
	if err := pongo2.RegisterTag("lua", luatag.TagLuaParser); err != nil {
		return err
	}
	//if err := pongo2.RegisterTag("approve", template.TagApproveParser); err != nil {
	//	return err
	//}
	//if err := pongo2.RegisterTag("at_sender", template.TagAtSenderParser); err != nil {
	//	return err
	//}
	//if err := pongo2.RegisterTag("withdraw", template.TagWithdrawParser); err != nil {
	//	return err
	//}
	if err := pongo2.RegisterTag("random_choice", template.TagRandomChoiceParser); err != nil {
		return err
	}
	if err := pongo2.RegisterTag("send_private", template.TagSendParser(template.PrivateMessageType)); err != nil {
		return err
	}
	if err := pongo2.RegisterTag("send_group", template.TagSendParser(template.GroupMessageType)); err != nil {
		return err
	}

	// set lua `res` func
	luatag.SetResFunc(resourcePathFunc(Config.ResourceShare))

	return nil
}

func filterEscapeCQCode(in *pongo2.Value, _ *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	return pongo2.AsValue(cqcode.Escape(in.String())), nil
}

func filterSilence(_ *pongo2.Value, _ *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	return pongo2.AsValue(nil), nil
}

func buildExecutionContext(matcher *zero.Matcher, event zero.Event, state zero.State, luaState *lua.LState) pongo2.Context {
	return pongo2.Context{
		"matcher": matcher,
		"state":   state,
		"event": func() interface{} {
			e := make(map[string]interface{})
			if err := jsoniter.UnmarshalFromString(event.RawEvent.Raw, &e); err != nil {
				log.Errorf("error when decode event json: %s", err)
			}
			return e
		},
		"json_event": &event.RawEvent.Raw,
		"at_sender": func() string {
			if event.GroupID == 0 {
				log.Warnf("cannot at sender in event %s/%s", event.PostType, event.SubType)
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
		"withdraw": func() {
			if event.MessageType != "group" {
				log.Warnf("cannot withdraw: message is not a group message: %#v", event)
			}
			zero.DeleteMessage(event.MessageID)
		},
		"set_title": func(title string, qqid ...int64) {
			if event.MessageType != "group" {
				log.Warnf("cannot set title: message is not a group message: %#v", event)
			}
			if len(qqid) == 0 {
				zero.SetGroupSpecialTitle(event.GroupID, event.UserID, title)
			} else {
				for _, user := range qqid {
					zero.SetGroupSpecialTitle(event.GroupID, user, title)
				}
			}
		},
		"group_ban": func(duration interface{}, qqid ...int64) {
			if event.GroupID == 0 {
				log.Warnf("cannot ban sender in event %s/%s", event.PostType, event.SubType)
				return
			}
			d, err := helper.AnyToInt64(duration)
			if err != nil {
				log.Warnf("cannot convert %#v to int64", duration)
				return
			}
			if len(qqid) == 0 {
				zero.SetGroupBan(event.GroupID, event.UserID, d)
			} else {
				for _, user := range qqid {
					zero.SetGroupBan(event.GroupID, user, d)
				}
			}
		},
		"_event": &event,
		"_lua":   luaState,
	}
}
