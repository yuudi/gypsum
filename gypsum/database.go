package gypsum

import (
	"bytes"
	"encoding/gob"
	"math/rand"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"

	"github.com/yuudi/gypsum/gypsum/helper"
	"github.com/yuudi/gypsum/gypsum/luatag"
	"github.com/yuudi/gypsum/gypsum/template"
)

var db *leveldb.DB
var itemCursor cursorType
var botAdmins []int64

type cursorType struct {
	id   uint64
	lock sync.Mutex
}

func (c *cursorType) readFromDB() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	data, err := db.Get([]byte("gypsum-$meta-cursor"), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			c.id = 0
		} else {
			return err
		}
	} else {
		c.id = helper.ToUint(data)
	}
	return nil
}

func (c *cursorType) Require() uint64 {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.id += 1
	if err := db.Put([]byte("gypsum-$meta-cursor"), helper.U64ToBytes(c.id), nil); err != nil {
		// 不太可能失败吧。。
		c.id -= 1
		panic(err)
	}
	return c.id
}

func initDb() error {
	var err error
	db, err = leveldb.OpenFile("gypsum_data/data", nil)
	if err != nil {
		return err
	}
	// read cursor
	if err = itemCursor.readFromDB(); err != nil {
		return err
	}
	// read coldsalt
	coldSalt, err = db.Get([]byte("gypsum-$meta-coldsalt"), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			rand.Read(coldSalt)
			err = db.Put([]byte("gypsum-$meta-coldsalt"), coldSalt, nil)
			if err != nil {
				log.Warnf("error when write database: %s", err)
			}
		} else {
			return err
		}
	}
	// read bot owners
	botOwnersBytes, err := db.Get([]byte("gypsum-$meta-botowners"), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			botAdmins = nil
		} else {
			return err
		}
	} else {
		if err = gob.NewDecoder(bytes.NewReader(botOwnersBytes)).Decode(&botAdmins); err != nil {
			return err
		}
	}

	luatag.SetDB(db)
	template.SetDB(db)
	return nil
}

func loadData() error {
	loadGroups()
	loadRules()
	loadTriggers()
	loadJobs()
	loadResources()
	return nil
}
