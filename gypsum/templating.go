package gypsum

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/flosch/pongo2"

	"github.com/yuudi/gypsum/gypsum/luatag"
)

// const (
// 	kindInteger = reflect.Int | reflect.Int8 | reflect.Int16 | reflect.Int32 | reflect.Int64 | reflect.Uint | reflect.Uint8 | reflect.Uint16 | reflect.Uint32 | reflect.Uint64
// 	kindQQID    = reflect.String | reflect.Int | reflect.Int8 | reflect.Int16 | reflect.Int32 | reflect.Int64 | reflect.Uint | reflect.Uint8 | reflect.Uint16 | reflect.Uint32 | reflect.Uint64
// )

func initTemplating() error {
	// register filters
	if err := pongo2.RegisterFilter("escCQ", filterEscapeCQCode); err != nil {
		return err
	}

	// disable HTML auto-escape
	pongo2.SetAutoescape(false)

	// register functions
	pongo2.Globals["at"] = at
	pongo2.Globals["sleep"] = sleep

	// register lua
	if err := pongo2.RegisterTag("lua", luatag.TagLuaParser); err != nil {
		return err
	}
	return nil
}

func filterEscapeCQCode(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	output := strings.Replace(in.String(), "&", "&amp;", -1)
	output = strings.Replace(output, "[", "&#91;", -1)
	output = strings.Replace(output, "]", "&#93;", -1)
	output = strings.Replace(output, ",", "&#44;", -1)
	return pongo2.AsValue(output), nil
}

func at(qq ...interface{}) string {
	if len(qq) == 0 {
		return "" // TODO: get qqid from context
	}
	ats := make([]string, len(qq))
	for i, qqID := range qq {
		ats[i] = atqq(qqID)
	}
	return strings.Join(ats, "")
}

func atqq(qq interface{}) string {
	switch qq.(type) {
	case int, int32, int64, uint, uint32, uint64:
		// doing nothing
	case string:
		if qq != "all" {
			log.Printf("error: cannot accept %#v as qqid", qq)
			return "ERROR"
		}
	default:
		log.Printf("error: cannot accept %#v as qqid", qq)
		return "ERROR"
	}
	return fmt.Sprintf("[CQ:at,qq=%v] ", qq)
}

func sleep(duration interface{}) string {
	seconds, err := ToFloat(duration)
	if err != nil {
		log.Printf("error: cannot accept %#v as interger", duration)
		return "ERROR"
	}
	time.Sleep(time.Duration(seconds * float64(time.Second)))
	return ""
}
