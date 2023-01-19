package internal

import (
	"sort"
	"sync"
)

type cache struct {
	sync.Mutex
	list map[string]dbMetaInfo
}

type dbMetaInfo struct {
	// list of collections in the database
	collections sort.StringSlice
	// last usage of this database, stored as Unix time
	lastUsed int64
	// true if database is ready to use: 
	// either is created on the server, or has a config error
	ready bool
	// any error that occured while creating the db
	err error
}

func (d *dbMetaInfo) hasCollection(collectionName string) bool {
	for _, name := range d.collections {
		if name == collectionName {
			return true
		}
	}
	return false
}
