package luatag

import (
	"time"

	zeroMessage "github.com/wdvxdr1123/ZeroBot/message"
	lua "github.com/yuin/gopher-lua"
)

func Writer(w interface{ WriteString(string) (int, error) }, safe bool) func(*lua.LState) int {
	return func(L *lua.LState) int {
		top := L.GetTop()
		for i := 1; i <= top; i++ {
			// write all arguments
			data := L.ToStringMeta(L.Get(i)).String()
			if !safe {
				data = zeroMessage.EscapeCQCodeText(data)
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

func luaSleep(L *lua.LState) int {
	arg := L.ToNumber(1)
	duration := time.Duration(float64(arg) * float64(time.Second))
	time.Sleep(duration)
	return 0
}
