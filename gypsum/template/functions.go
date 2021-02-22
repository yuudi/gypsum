package template

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/flosch/pongo2"
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"

	"github.com/yuudi/gypsum/gypsum/helper"
)

func At(qq ...interface{}) *pongo2.Value {
	ats := make([]string, len(qq))
	for i, qqID := range qq {
		ats[i] = atqq(qqID)
	}
	return pongo2.AsSafeValue(strings.Join(ats, ""))
}

func atqq(qq interface{}) string {
	switch qq.(type) {
	case int, int32, int64, uint, uint32, uint64, string:
		return fmt.Sprintf("[CQ:at,qq=%v] ", qq)
	default:
		log.Warnf("error: cannot accept %#v as qqid", qq)
		return "ERROR"
	}
}

func Image(src string, args ...int) *pongo2.Value {
	// onenot can handle it well :)
	var cache int
	switch len(args) {
	case 0:
		cache = 1
	case 1:
		cache = args[0]
	default:
		log.Warn("function image: too many arguments")
		cache = args[0]
	}
	return pongo2.AsSafeValue(fmt.Sprintf("[CQ:image,cache=%d,file=%s] ", cache, src))
}

func Record(src string, args ...int) *pongo2.Value {
	// same as image :)
	var cache int
	switch len(args) {
	case 0:
		cache = 1
	case 1:
		cache = args[0]
	default:
		log.Warn("function image: too many arguments")
		cache = args[0]
	}
	return pongo2.AsSafeValue(fmt.Sprintf("[CQ:record,cache=%d,file=%s] ", cache, src))
}

func Sleep(duration interface{}) string {
	seconds, err := helper.AnyToFloat(duration)
	if err != nil {
		log.Warnf("error: cannot accept %#v as interger", duration)
		return "ERROR"
	}
	time.Sleep(time.Duration(seconds * float64(time.Second)))
	return ""
}

func RandomInt(input ...interface{}) (int, error) {
	var min, max int
	var err error
	switch len(input) {
	case 0:
		min = 0
		max = 99
	case 1:
		min = 0
		max, err = helper.AnyToInt(input[0])
		if err != nil {
			return 0, err
		}
	case 2:
		min, err = helper.AnyToInt(input[0])
		max, err = helper.AnyToInt(input[1])
		if err != nil {
			return 0, err
		}
	default:
		log.Warn("too many argument for random")
		min, err = helper.AnyToInt(input[0])
		max, err = helper.AnyToInt(input[1])
		if err != nil {
			return 0, err
		}
	}
	return rand.Intn(max) + min, nil
}

func FileGetContents(filename string) string {
	if strings.HasPrefix(filename, "http://") || strings.HasPrefix(filename, "https://") {
		res, err := http.Get(filename)
		if err != nil {
			log.Error("模板解析错误", err)
			return ""
		}
		if res.StatusCode >= 400 {
			log.Error("模板解析错误，网络请求返回值", res.StatusCode)
			return ""
		}
		content, err := io.ReadAll(res.Body)
		if err != nil {
			log.Error("模板解析错误，读取网络资源", err)
			return ""
		}
		return string(content)
	}
	content, err := os.ReadFile(filename)
	if err != nil {
		log.Error("模板解析错误，读取文件", err)
		return ""
	}
	return string(content)
}

func ParseJson(jsonBody string) interface{} {
	var r interface{}
	err := jsoniter.UnmarshalFromString(jsonBody, &r)
	if err != nil {
		log.Error("模板解析错误，解析json", err)
		return nil
	}
	return r
}

func RandomLine(c string) string {
	lines := strings.Split(c, "\n")
	choice := rand.Intn(len(lines))
	return lines[choice]
}

func RandomFile(dirPath string) string {
	dir, err := os.ReadDir(dirPath)
	if err != nil {
		log.Error("模板解析错误，读取目录", err)
		return ""
	}
	choice := rand.Intn(len(dir))
	return path.Join(dirPath, dir[choice].Name())
}

func Sequence(input interface{}) ([]int, error) {
	length, err := helper.AnyToInt(input)
	if err != nil {
		return nil, err
	}
	if length < 0 {
		return nil, errors.New("length is negative")
	}
	s := make([]int, length)
	for i := range s {
		s[i] = i
	}
	return s, nil
}
