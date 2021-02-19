package gypsum

import (
	"net/url"

	"github.com/flosch/pongo2"

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
	pongo2.Globals["res"] = resourcePath
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
	if err := pongo2.RegisterTag("random_choice", template.TagRandomChoiceParser); err != nil {
		return err
	}
	if err := pongo2.RegisterTag("send_private", template.TagSendParser(template.PrivateMessageType)); err != nil {
		return err
	}
	if err := pongo2.RegisterTag("send_group", template.TagSendParser(template.GroupMessageType)); err != nil {
		return err
	}
	return nil
}

func filterEscapeCQCode(in *pongo2.Value, _ *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	return pongo2.AsValue(cqcode.Escape(in.String())), nil
}

func filterSilence(_ *pongo2.Value, _ *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	return pongo2.AsValue(nil), nil
}
