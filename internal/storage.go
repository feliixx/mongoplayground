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
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/dgraph-io/badger/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// interval between two MongoDB cleanup
	cleanupInterval = 10 * time.Minute
	// how long a database is kept after its last use 
	maxUnusedDuration = 6 * time.Hour
	// interval between two Badger backup
	backupInterval = 24 * time.Hour
)

type storage struct {
	mongoSession *mongo.Client
	mongoVersion []byte

	kvStore *badger.DB
	// local dir to store badger backups
	backupDir           string
	backupServiceStatus serviceInfo

	activeDB *cache

	mailInfo *MailInfo

	cloudflareInfo *CloudflareInfo

	googleDriveInfo *GoogleDriveInfo
}

func newStorage(mongoUri string, dropFirst bool, cloudflareInfo *CloudflareInfo, mailInfo *MailInfo, googleDriveInfo *GoogleDriveInfo) (*storage, error) {

	session, err := createMongodbSession(mongoUri)
	if err != nil {
		return nil, err
	}

	kvStore, err := badger.Open(badger.DefaultOptions("storage"))
	if err != nil {
		return nil, err
	}

	s := &storage{
		mongoSession: session,
		mongoVersion: getMongoVersion(session),
		kvStore:      kvStore,
		activeDB: &cache{
			list: map[string]dbMetaInfo{},
		},
		backupDir: "backups",
		backupServiceStatus: serviceInfo{
			Name:   "backup",
			Status: statusUp,
		},
		mailInfo:        mailInfo,
		cloudflareInfo:  cloudflareInfo,
		googleDriveInfo: googleDriveInfo,
	}

	if dropFirst {
		s.deleteExistingDB()
	}

	initPrometheusCounter(s.kvStore)

	go func(s *storage) {
		for range time.Tick(cleanupInterval) {
			s.removeUnusedDB()
		}
	}(s)

	go func(s *storage) {
		for range time.Tick(backupInterval) {
			s.backup()
		}
	}(s)

	return s, nil
}

// delete all database having a name with 32 char
func (s *storage) deleteExistingDB() error {

	dbNames, err := s.mongoSession.ListDatabaseNames(context.Background(), bson.D{})
	if err != nil {
		return err
	}

	for _, name := range dbNames {
		if len(name) == 32 {
			log.Printf("Deleting db '%s'", name)
			err = s.mongoSession.Database(name).Drop(context.Background())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *storage) removeUnusedDB() {

	now := time.Now()

	s.activeDB.Lock()
	for name, info := range s.activeDB.list {

		// if database creation failed, always remove it from the cache 
		// as soon as possible, to prevent temporary error due to MongoDB 
		// from being kept for too long.
		if info.err != nil {
			delete(s.activeDB.list, name)
		}

		if now.Sub(time.Unix(info.lastUsed, 0)) > maxUnusedDuration {
			if info.err == nil {
				err := s.mongoSession.Database(name).Drop(context.Background())
				if err != nil {
					log.Printf("fail to drop database %v: %v", name, err)
				}
			}
			delete(s.activeDB.list, name)
		}
	}
	s.activeDB.Unlock()

	cleanupDuration.Set(time.Since(now).Seconds())
	activeDatabasesCounter.Set(float64(len(s.activeDB.list)))
}

// create a backup from the badger db, and store it in backupDir.
// keep a backup of last seven days only. Older backups are
// overwritten
// upload the last backup to google drive. Previous backup is moved to trash
// and automatically removed after 30 days
func (s *storage) backup() {

	log.Print("starting backup...")

	if _, err := os.Stat(s.backupDir); os.IsNotExist(err) {
		os.Mkdir(s.backupDir, os.ModePerm)
	}

	fileName := fmt.Sprintf("%s/badger_%d.bak", s.backupDir, time.Now().Weekday())

	err := localBackup(s.kvStore, fileName)
	if err != nil {
		s.handleBackupError("error in local backup", err)
		return
	}

	if s.googleDriveInfo != nil {
		err = s.googleDriveInfo.saveBackupToGoogleDrive(fileName)
		if err != nil {
			s.handleBackupError("error while uploading backup", err)
			return
		}
	}

	s.backupServiceStatus.Status = statusUp
	s.backupServiceStatus.Cause = ""

	// as backup() run once a day, also update the mongodb
	// server version ( in case the cluster has automatically
	// been upgraded )
	currentMongoVersion := getMongoVersion(s.mongoSession)
	if !bytes.Equal(currentMongoVersion, s.mongoVersion) && s.cloudflareInfo != nil {
		s.mongoVersion = currentMongoVersion
		s.cloudflareInfo.clearCloudflareCache()
	}
}

func (s *storage) handleBackupError(message string, err error) {

	errorMsg := fmt.Sprintf("%s: %v", message, err)
	log.Print(errorMsg)

	s.backupServiceStatus.Status = statusDegrade
	s.backupServiceStatus.Cause = errorMsg
	if s.mailInfo != nil {
		s.mailInfo.sendErrorByEmail(errorMsg)
	}
}

func createMongodbSession(mongoUri string) (*mongo.Client, error) {

	session, err := mongo.NewClient(options.Client().ApplyURI(mongoUri))
	if err != nil {
		return nil, fmt.Errorf("fail to create mongodb client: %v", err)
	}
	err = session.Connect(context.Background())
	if err != nil {
		return nil, fmt.Errorf("fail to connect to mongodb: %v", err)
	}
	return session, nil
}

func getMongoVersion(client *mongo.Client) []byte {

	result := client.Database("admin").RunCommand(context.Background(), bson.M{"buildInfo": 1})

	var buildInfo struct {
		Version []byte
	}
	err := result.Decode(&buildInfo)
	if err != nil {
		return []byte("unknown")
	}
	return buildInfo.Version
}
