package luatag

import (
	"log"

	lua "github.com/yuin/gopher-lua"
)

func Writer(w interface{ WriteString(string) (int, error) }) func(*lua.LState) int {
	return func(L *lua.LState) int {
		lv := L.ToString(1)
		if _, err := w.WriteString(lv); err != nil {
			log.Printf("write buffer error: %s", err)
		}
		return 0
	}
}
