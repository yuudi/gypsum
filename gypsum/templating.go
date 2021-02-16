package gypsum

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/flosch/pongo2"
	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"

	"github.com/yuudi/gypsum/gypsum/helper/cqcode"
	"github.com/yuudi/gypsum/gypsum/luatag"
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
	pongo2.Globals["at"] = at
	pongo2.Globals["res"] = resourcePath
	pongo2.Globals["image"] = image
	pongo2.Globals["sleep"] = sleep
	pongo2.Globals["random_int"] = randomInt
	pongo2.Globals["db_get"] = dbGet
	pongo2.Globals["db_put"] = dbPut

	// register lua
	if err := pongo2.RegisterTag("lua", luatag.TagLuaParser); err != nil {
		log.Errorf("lua引擎初始化错误：%s", err)
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

func at(qq ...interface{}) string {
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

func image(src string, args ...int) string {
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

func sleep(duration interface{}) string {
	seconds, err := ToFloat(duration)
	if err != nil {
		log.Warnf("error: cannot accept %#v as interger", duration)
		return "ERROR"
	}
	time.Sleep(time.Duration(seconds * float64(time.Second)))
	return ""
}

func randomInt(input ...int) int {
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

type ValueType int

const (
	IntValueType ValueType = iota
	StrValueType
)

type StoredValue struct {
	ValueType ValueType
	IntValue  int
	StrValue  string
}

func dbGet(key interface{}, defaultValue ...interface{}) interface{} {
	if len(defaultValue) > 1 {
		log.Warn("too many arguments for calling db_get")
	}
	var bytesKey []byte
	switch key.(type) {
	case string:
		bytesKey = []byte(key.(string))
	case int:
		bytesKey = U64ToBytes(uint64(key.(int)))
	default:
		log.Errorf("cannot use %#v (%T) as database key", key, key)
		return nil
	}
	bytesData, err := db.Get(append([]byte("gypsum-userDB-p-"), bytesKey...), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			log.Warnf("cannot find key in database: %v", key)
			if len(defaultValue) == 0 {
				return nil
			} else {
				return defaultValue[0]
			}
		}
		log.Error(err)
		return nil
	}
	var data StoredValue
	buffer := bytes.Buffer{}
	buffer.Write(bytesData)
	decoder := gob.NewDecoder(&buffer)
	if err := decoder.Decode(&data); err != nil {
		log.Errorf("error when reading data from database: %s", err)
		return nil
	}
	switch data.ValueType {
	case IntValueType:
		return data.IntValue
	case StrValueType:
		return data.StrValue
	default:
		log.Errorf("Unknown value type from StoredValue: %v", data.ValueType)
		return nil
	}
}

func dbPut(key, value interface{}) *int {
	var bytesKey []byte
	switch key.(type) {
	case string:
		bytesKey = []byte(key.(string))
	case int:
		bytesKey = U64ToBytes(uint64(key.(int)))
	default:
		log.Errorf("cannot use %#v (%T) as database key", key, key)
	}
	var valueStore *StoredValue
	switch value.(type) {
	case string:
		valueStore = &StoredValue{
			ValueType: StrValueType,
			StrValue:  value.(string),
		}
	case int:
		valueStore = &StoredValue{
			ValueType: IntValueType,
			IntValue:  value.(int),
		}
	default:
		log.Errorf("cannot store %#v (%T) to database", value, value)
	}
	buffer := bytes.Buffer{}
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(valueStore); err != nil {
		log.Errorf("error when encode valueStore as bytes: %s", err)
		return nil
	}
	if err := db.Put(append([]byte("gypsum-userDB-p-"), bytesKey...), buffer.Bytes(), nil); err != nil {
		log.Errorf("error when put value to database %s", err)
		return nil
	}
	return nil
}
