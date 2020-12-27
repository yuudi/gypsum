package luatag

import (
	"log"

	zero "github.com/wdvxdr1123/ZeroBot"
	lua "github.com/yuin/gopher-lua"
	luajson "layeh.com/gopher-json"
)

func Writer(w interface{ WriteString(string) (int, error) }) func(*lua.LState) int {
	return func(L *lua.LState) int {
		top := L.GetTop()
		for i := 1; i <= top; i++{
			_, _ = w.WriteString(L.ToStringMeta(L.Get(i)).String())
			if i != top{
				_, _ = w.WriteString(" ")
			}
		}
		_, _ = w.WriteString("\n")
		return 0
	}
}

func botApi(L *lua.LState) int {
	action := L.ToString(1)
	lparams := L.ToTable(2)
	params := make(map[string]interface{})
	if lparams == nil {
		params = nil
	} else {
		lparams.ForEach(func(k lua.LValue, v lua.LValue) {
			key := k.String()
			switch v.Type() {
			case lua.LTString:
				params[key] = v.String()
			case lua.LTNumber:
				params[key] = float64(v.(lua.LNumber))
			case lua.LTBool:
				params[key] = bool(v.(lua.LBool))
			default:
				log.Printf("error when calling api from lua: cannot use type %s", v.Type().String())
			}
		})
	}
	result := zero.CallAction(action, params)
	lresult, _ := luajson.Decode(L, []byte(result.Raw))
	L.Push(lresult)
	return 1
}
