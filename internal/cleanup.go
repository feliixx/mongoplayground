package internal

import (
	"sort"
	"time"

	"golang.org/x/net/context"
)

type dbMetaInfo struct {
	// list of collections in the database
	collections sort.StringSlice
	// last usage of this database, stored as Unix time
	lastUsed int64
	// wether all collections of the databse are empty
	emptyDatabase bool
}

func (d *dbMetaInfo) hasCollection(collectionName string) bool {
	for _, name := range d.collections {
		if name == collectionName {
			return true
		}
	}
	return false
}

// remove database not used since the previous cleanup in MongoDB
func (s *Server) removeExpiredDB() {

	now := time.Now()

	s.activeDbLock.Lock()
	for name, infos := range s.activeDB {
		if now.Sub(time.Unix(infos.lastUsed, 0)) > cleanupInterval {
			err := s.session.Database(name).Drop(context.Background())
			if err != nil {
				s.logger.Printf("fail to drop database %v: %v", name, err)
			}
			delete(s.activeDB, name)
		}
	}
	s.activeDbLock.Unlock()

	cleanupDuration.Set(time.Since(now).Seconds())
	activeDatabases.Set(float64(len(s.activeDB)))
}
