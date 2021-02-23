package template

import (
	"bytes"
	"encoding/gob"

	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"

	"github.com/yuudi/gypsum/gypsum/helper"
)

var db *leveldb.DB

func SetDB(newDB *leveldb.DB) {
	db = newDB
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

func DatabaseGet(key interface{}, defaultValue ...interface{}) interface{} {
	if len(defaultValue) > 1 {
		log.Warn("too many arguments for calling db_get")
	}
	var bytesKey []byte
	switch k := key.(type) {
	case string:
		bytesKey = []byte(k)
	case int:
		bytesKey = helper.U64ToBytes(uint64(k))
	default:
		log.Errorf("cannot use %#v (%T) as database key", key, key)
		return nil
	}
	bytesData, err := db.Get(append([]byte("gypsum-userDB-p-"), bytesKey...), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			if len(defaultValue) == 0 {
				log.Warnf("cannot find key in database: %v", key)
				return nil
			} else {
				return defaultValue[0]
			}
		}
		log.Error(err)
		return nil
	}
	var data StoredValue
	decoder := gob.NewDecoder(bytes.NewReader(bytesData))
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

func DatabasePut(key, value interface{}) *int {
	var bytesKey []byte
	switch k := key.(type) {
	case string:
		bytesKey = []byte(k)
	case int:
		bytesKey = helper.U64ToBytes(uint64(k))
	default:
		log.Errorf("cannot use %#v (%T) as database key", key, key)
		return nil
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
