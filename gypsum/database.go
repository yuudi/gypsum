package gypsum

import (
	"github.com/syndtr/goleveldb/leveldb"
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
	loadJobs()
	return nil
}
