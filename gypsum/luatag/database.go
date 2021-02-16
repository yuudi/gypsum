package luatag

import (
	"bytes"
	"encoding/gob"

	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
	lua "github.com/yuin/gopher-lua"
)

func init() {
	gob.Register(lua.LNumber(0))
	gob.Register(lua.LString(""))
	gob.Register(lua.LBool(false))
	gob.Register(lua.LNil)
	gob.Register(lua.LTable{})
}

var db *leveldb.DB

func SetDB(newDB *leveldb.DB) {
	db = newDB
}

func dbLoader(L *lua.LState) int {
	mod := L.NewTable()
	L.SetFuncs(mod, map[string]lua.LGFunction{
		"get": dbGet,
		"put": dbPut,
	})
	L.Push(mod)
	return 1
}

func dbGet(L *lua.LState) int {
	key := L.ToString(1)
	defaultValue := L.Get(2)
	bytesKey := []byte(key)
	bytesData, err := db.Get(append([]byte("gypsum-userDB-lua-"), bytesKey...), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			L.Push(defaultValue)
			return 1
		}
		log.Error(err)
		L.Push(lua.LNil)
		L.Push(lua.LString("database error: " + err.Error()))
		return 2
	}
	var data *lua.LValue
	buffer := bytes.Buffer{}
	buffer.Write(bytesData)
	decoder := gob.NewDecoder(&buffer)
	if err := decoder.Decode(&data); err != nil {
		log.Errorf("error when reading data from database: %s", err)
		L.Push(lua.LNil)
		L.Push(lua.LString("error when reading data from database: " + err.Error()))
		return 2
	}
	L.Push(*data)
	return 1
}

func dbPut(L *lua.LState) int {
	key := L.ToString(1)
	value := L.Get(2)
	bytesKey := []byte(key)
	buffer := bytes.Buffer{}
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(&value); err != nil {
		log.Errorf("error when encode valueStore as bytes: %s", err)
		L.Push(lua.LString("error when encode valueStore as bytes: " + err.Error()))
		return 1
	}
	if err := db.Put(append([]byte("gypsum-userDB-lua-"), bytesKey...), buffer.Bytes(), nil); err != nil {
		log.Errorf("error when put value to database: %s", err)
		L.Push(lua.LString("error when put value to database: " + err.Error()))
		return 1
	}
	return 0
}
