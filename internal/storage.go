// mongoplayground: a sandbox to test and share MongoDB queries
// Copyright (C) 2017 Adrien Petel
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package internal

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// interval between two MongoDB cleanup
	cleanupInterval = 4 * time.Hour
	// interval between two Badger backup
	backupInterval = 24 * time.Hour
)

type storage struct {
	mongoSession *mongo.Client
	mongoVersion []byte

	kvStore *badger.DB
	// local dir to store badger backups
	backupDir string

	// activeDB holds info of the database created / used during
	// the last cleanupInterval. Its access is garded by activeDbLock
	activeDbLock sync.RWMutex
	activeDB     map[string]dbMetaInfo
}

func newStorage(badgerDir, backupDir string) (*storage, error) {

	session, mongodbVersion, err := createMongodbSession()
	if err != nil {
		return nil, err
	}

	kvStore, err := badger.Open(badger.DefaultOptions(badgerDir))
	if err != nil {
		return nil, err
	}

	s := &storage{
		mongoSession: session,
		mongoVersion: mongodbVersion,
		kvStore:      kvStore,
		activeDB:     map[string]dbMetaInfo{},
		backupDir:    backupDir,
	}

	initPrometheusCounter(s.kvStore)

	go func(s *storage) {
		for range time.Tick(cleanupInterval) {
			s.removeExpiredDB()
		}
	}(s)

	go func(s *storage) {
		for range time.Tick(backupInterval) {
			s.backup()
		}
	}(s)

	return s, nil
}

// remove database not used since the previous cleanup in MongoDB
func (s *storage) removeExpiredDB() {

	now := time.Now()

	s.activeDbLock.Lock()
	for name, infos := range s.activeDB {
		if now.Sub(time.Unix(infos.lastUsed, 0)) > cleanupInterval {
			err := s.mongoSession.Database(name).Drop(context.Background())
			if err != nil {
				log.Printf("fail to drop database %v: %v", name, err)
			}
			delete(s.activeDB, name)
		}
	}
	s.activeDbLock.Unlock()

	cleanupDuration.Set(time.Since(now).Seconds())
	activeDatabasesCounter.Set(float64(len(s.activeDB)))
}

// create a backup from the badger db, and store it in backupDir.
// keep a backup of last seven days only. Older backups are
// overwritten
// upload the last backup to google drive. Previous backup is moved to trash
// and automatically removed after 30 days
func (s *storage) backup() {

	if _, err := os.Stat(s.backupDir); os.IsNotExist(err) {
		os.Mkdir(s.backupDir, os.ModePerm)
	}

	fileName := fmt.Sprintf("%s/badger_%d.bak", s.backupDir, time.Now().Weekday())

	localBackup(s.kvStore, fileName)
	saveBackupToGoogleDrive(fileName)
}

type dbMetaInfo struct {
	// list of collections in the database
	collections sort.StringSlice
	// last usage of this database, stored as Unix time
	lastUsed int64
	// true if all collections of the database are empty
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

func createMongodbSession() (session *mongo.Client, version []byte, err error) {
	session, err = mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		return nil, nil, fmt.Errorf("fail to create mongodb client: %v", err)
	}
	err = session.Connect(context.Background())
	if err != nil {
		return nil, nil, fmt.Errorf("fail to connect to mongodb: %v", err)
	}
	return session, getMongodVersion(session), nil
}
