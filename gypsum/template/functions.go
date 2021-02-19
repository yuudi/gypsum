package template

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"

	"github.com/yuudi/gypsum/gypsum/helper"
)

func At(qq ...interface{}) string {
	ats := make([]string, len(qq))
	for i, qqID := range qq {
		ats[i] = atqq(qqID)
	}
	return strings.Join(ats, "")
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

func Image(src string, args ...int) string {
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
	return fmt.Sprintf("[CQ:image,cache=%d,file=%s] ", cache, src)
}

func Record(src string, args ...int) string {
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
	return fmt.Sprintf("[CQ:record,cache=%d,file=%s] ", cache, src)
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

func RandomInt(input ...int) int {
	var min, max int
	switch len(input) {
	case 0:
		min = 0
		max = 99
	case 1:
		min = 0
		max = input[0]
	case 2:
		min = input[0]
		max = input[1]
	default:
		log.Warn("too many argument for random")
		min = input[0]
		max = input[1]
	}
	return rand.Intn(max) + min
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
