package gypsum

import (
	"github.com/flosch/pongo2"

	"github.com/yuudi/gypsum/gypsum/luatag"
)

func initTemplating() error {
	if err := pongo2.RegisterTag("lua", luatag.TagLuaParser); err != nil {
		return err
	}
	return nil
}
