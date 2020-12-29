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

	"github.com/yuudi/gypsum/gypsum/luatag"
)

// const (
// 	kindInteger = reflect.Int | reflect.Int8 | reflect.Int16 | reflect.Int32 | reflect.Int64 | reflect.Uint | reflect.Uint8 | reflect.Uint16 | reflect.Uint32 | reflect.Uint64
// 	kindQQID    = reflect.String | reflect.Int | reflect.Int8 | reflect.Int16 | reflect.Int32 | reflect.Int64 | reflect.Uint | reflect.Uint8 | reflect.Uint16 | reflect.Uint32 | reflect.Uint64
// )

func initTemplating() error {
	// register filters
	if err := pongo2.RegisterFilter("escq", filterEscapeCQCode); err != nil {
		return err
	}

	// disable HTML auto-escape
	pongo2.SetAutoescape(false)

	// register functions
	pongo2.Globals["at"] = at
	pongo2.Globals["res"] = resourcePath
	pongo2.Globals["image"] = image
	pongo2.Globals["dynamic_image"] = dynamicImage
	pongo2.Globals["sleep"] = sleep
	pongo2.Globals["db_get"] = dbGet
	pongo2.Globals["db_put"] = dbPut

	// register lua
	if err := pongo2.RegisterTag("lua", luatag.TagLuaParser); err != nil {
		return err
	}
	return nil
}

var cqEscape = strings.NewReplacer("&", "&amp;","[", "&#91;","]", "&#93;",",", "&#44;")

func filterEscapeCQCode(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	return pongo2.AsValue(cqEscape.Replace(in.String())), nil
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

func image(src string) string {
	// onenot can handle it well :)
	return fmt.Sprintf("[CQ:image,file=%v] ", src)
}

func dynamicImage(src string) string {
	return fmt.Sprintf("[CQ:image,cache=0,file=%v] ", src)
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

func randomInt(min, max interface{}) int {
	rand.Intn(100)
	return 0
}

func dbGet(key interface{}) interface{} {
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
	bytesData, err := db.Get(append([]byte("gypsum-userDB-"), bytesKey...), nil)
	if err == leveldb.ErrNotFound {
		return nil
	}
	var data interface{}
	buffer := bytes.Buffer{}
	buffer.Write(bytesData)
	decoder := gob.NewDecoder(&buffer)
	if err := decoder.Decode(data); err != nil {
		log.Errorf("error when reading data from database: %s", err)
		return nil
	}
	return data
}

func dbPut(key interface{}, value interface{}) {
	var bytesKey []byte
	switch key.(type) {
	case string:
		bytesKey = []byte(key.(string))
	case int:
		bytesKey = U64ToBytes(uint64(key.(int)))
	default:
		log.Errorf("cannot use %#v (%T) as database key", key, key)
	}
	buffer := bytes.Buffer{}
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(value); err != nil {
		log.Errorf("error when encode %#v (%T) as bytes: %s", value, value, err)
		return
	}
	if err := db.Put(append([]byte("gypsum-userDB-"), bytesKey...), buffer.Bytes(), nil); err != nil {
		log.Errorf("error when put value to database %s", err)
		return
	}
}
