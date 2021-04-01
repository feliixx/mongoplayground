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
