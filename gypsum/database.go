package gypsum

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	zero "github.com/wdvxdr1123/ZeroBot"
	"log"
)

var db *leveldb.DB
var cursor uint64

func initDb() (err error) {
	db, err = leveldb.OpenFile("gypsum/data", nil)
	if err != nil {
		return err
	}
	data, e := db.Get([]byte("gypsum-$meta-cursor"), nil)
	if e != nil {
		if e == leveldb.ErrNotFound {
			cursor = 0
		} else {
			return e
		}
	} else {
		cursor = ToUint(data)
	}
	return
}

func loadData() (err error) {
	iter := db.NewIterator(util.BytesPrefix([]byte("gypsum-rules-")), nil)
	rules = make(map[uint64]Rule)
	zeroMatcher = make(map[uint64]*zero.Matcher)
	for iter.Next() {
		key := ToUint(iter.Key()[13:])
		value := iter.Value()
		r, e := RuleFromBytes(value)
		if e != nil {
			log.Printf("无法加载规则%d：%s", key, e)
			continue
		}
		rules[key] = *r
		if e := r.Register(key); e != nil {
			log.Printf("无法注册规则%d：%s", key, e)
			continue
		}
	}
	iter.Release()
	err = iter.Error()
	return
}
