package gypsum

import (
	"math/rand"

	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
)

var db *leveldb.DB
var itemCursor uint64

func initDb() error {
	var err error
	db, err = leveldb.OpenFile("gypsum_data/data", nil)
	if err != nil {
		return err
	}
	data, err := db.Get([]byte("gypsum-$meta-cursor"), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			itemCursor = 0
		} else {
			return err
		}
	} else {
		itemCursor = ToUint(data)
	}
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
