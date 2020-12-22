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
	db, err = leveldb.OpenFile("gypsum_data/data", nil)
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

func loadData() error {
	loadRules()
	loadTriggers()
	return nil
}

func loadRules() {
	rules = make(map[uint64]Rule)
	zeroMatcher = make(map[uint64]*zero.Matcher)
	iter := db.NewIterator(util.BytesPrefix([]byte("gypsum-rules-")), nil)
	defer func() {
		iter.Release()
		if err := iter.Error(); err != nil {
			log.Printf("载入数据错误：%s", err)
		}
	}()
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
}

func loadTriggers() {
	triggers = make(map[uint64]Trigger)
	zeroTrigger = make(map[uint64]*zero.Matcher)
	iter := db.NewIterator(util.BytesPrefix([]byte("gypsum-triggers-")), nil)
	defer func() {
		iter.Release()
		if err := iter.Error(); err != nil {
			log.Printf("载入数据错误：%s", err)
		}
	}()
	for iter.Next() {
		key := ToUint(iter.Key()[16:])
		value := iter.Value()
		t, e := TriggerFromByte(value)
		if e != nil {
			log.Printf("无法加载规则%d：%s", key, e)
			continue
		}
		triggers[key] = *t
		if e := t.Register(key); e != nil {
			log.Printf("无法注册规则%d：%s", key, e)
			continue
		}
	}
}
