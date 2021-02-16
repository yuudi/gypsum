package luatag

import (
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	lua "github.com/yuin/gopher-lua"
	luaJson "layeh.com/gopher-json"
)

func botModLoader(L *lua.LState) int {
	mod := L.NewTable()
	L.SetFuncs(mod, map[string]lua.LGFunction{
		"api": botApi,
	})
	L.Push(mod)
	return 1
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
