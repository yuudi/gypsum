package luatag

import (
	"time"

	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	lua "github.com/yuin/gopher-lua"

	"github.com/yuudi/gypsum/gypsum/helper/cqcode"
)

func Writer(w interface{ WriteString(string) (int, error) }, safe bool) func(*lua.LState) int {
	return func(L *lua.LState) int {
		top := L.GetTop()
		for i := 1; i <= top; i++ {
			// write all arguments
			data := L.ToStringMeta(L.Get(i)).String()
			if !safe {
				data = cqcode.Escape(data)
			}
			_, _ = w.WriteString(data)
			// write space between arguments
			if i != top {
				_, _ = w.WriteString(" ")
			}
		}
		//// write end of line
		//_, _ = w.WriteString("\n")
		return 0
	}
}

func Sender(event *zero.Event, safe bool) func(*lua.LState) int {
	if event == nil {
		log.Warn("cannot send without event")
		return func(_ *lua.LState) int {
			return 0
		}
	}
	return func(L *lua.LState) int {
		msg := L.ToString(1)
		if !safe {
			msg = cqcode.Escape(msg)
		}
		if event.GroupID == 0 {
			go zero.SendPrivateMessage(event.UserID, msg)
		} else {
			go zero.SendGroupMessage(event.GroupID, msg)
		}
		return 0
	}
}

func luaSleep(L *lua.LState) int {
	arg := L.ToNumber(1)
	duration := time.Duration(float64(arg) * float64(time.Second))
	time.Sleep(duration)
	return 0
}
