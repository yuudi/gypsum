package cqcode

import "strings"

var cqEscape = strings.NewReplacer("&", "&amp;", "[", "&#91;", "]", "&#93;", ",", "&#44;")
var cqParse = strings.NewReplacer("&amp;", "&", "&#91;", "[", "&#93;", "]", "&#44;", ",")

func Escape(q string) string {
	return cqEscape.Replace(q)
}

func Parse(q string) string {
	return cqParse.Replace(q)
}
