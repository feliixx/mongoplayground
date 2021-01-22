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

package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var templates = template.Must(template.ParseFiles("web/playground.html"))

const (
	staticDir = "web/static"
	badgerDir = "storage"
	backupDir = "backups"

	homeEndpoint    = "/"
	viewEndpoint    = "/p/"
	runEndpoint     = "/run"
	saveEndpoint    = "/save"
	staticEndpoint  = "/static/"
	metricsEndpoint = "/metrics"
	healthEndpoint  = "/health"

	// interval between two MongoDB cleanup
	cleanupInterval = 4 * time.Hour
	// interval between two Badger backup
	backupInterval = 24 * time.Hour
)

type server struct {
	mux     *http.ServeMux
	session *mongo.Client
	storage *badger.DB
	logger  *log.Logger

	// mutex guards the activeDB map
	mutex    sync.RWMutex
	activeDB map[string]dbMetaInfo

	staticContent  map[string][]byte
	mongodbVersion []byte
}

func newServer(logger *log.Logger) (*server, error) {

	session, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		return nil, fmt.Errorf("fail to create mongodb client: %v", err)
	}
	err = session.Connect(context.Background())
	if err != nil {
		return nil, fmt.Errorf("fail to connect to mongodb: %v", err)
	}

	db, err := badger.Open(badger.DefaultOptions(badgerDir))
	if err != nil {
		return nil, err
	}

	s := &server{
		mux:            http.DefaultServeMux,
		session:        session,
		storage:        db,
		activeDB:       map[string]dbMetaInfo{},
		logger:         logger,
		mongodbVersion: getMongodVersion(session),
	}

	err = s.compressStaticResources()
	if err != nil {
		return nil, err
	}

	err = s.computeSavedPlaygroundStats()
	if err != nil {
		return nil, fmt.Errorf("fail to read data from badger: %v", err)
	}

	go func(s *server) {
		for range time.Tick(cleanupInterval) {
			s.removeExpiredDB()
		}
	}(s)

	go func(s *server) {
		for range time.Tick(backupInterval) {
			s.backup()
		}
	}(s)

	s.mux.HandleFunc(homeEndpoint, s.newPageHandler)
	s.mux.HandleFunc(viewEndpoint, s.viewHandler)
	s.mux.HandleFunc(runEndpoint, s.runHandler)
	s.mux.HandleFunc(saveEndpoint, s.saveHandler)
	s.mux.HandleFunc(staticEndpoint, s.staticHandler)
	s.mux.HandleFunc(healthEndpoint, s.healthHandler)
	s.mux.Handle(metricsEndpoint, promhttp.Handler())

	return s, nil
}

func getMongodVersion(client *mongo.Client) []byte {

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

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	start := time.Now()
	s.mux.ServeHTTP(w, r)

	label := r.URL.Path
	if strings.HasPrefix(label, viewEndpoint) {
		label = viewEndpoint
	}

	if label == runEndpoint || label == viewEndpoint || label == homeEndpoint || label == saveEndpoint {
		requestDurations.WithLabelValues(label).Observe(float64(time.Since(start)) / float64(time.Second))
	}
}

// return a playground with the default configuration
func (s *server) newPageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Encoding", "gzip")
	w.Write(s.staticContent[homeEndpoint])
}

// remove database not used since the previous cleanup in MongoDB
func (s *server) removeExpiredDB() {

	now := time.Now()

	s.mutex.Lock()
	for name, infos := range s.activeDB {
		if now.Sub(time.Unix(infos.lastUsed, 0)) > cleanupInterval {
			err := s.session.Database(name).Drop(context.Background())
			if err != nil {
				s.logger.Printf("fail to drop database %v: %v", name, err)
			}
			delete(s.activeDB, name)
		}
	}
	s.mutex.Unlock()

	cleanupDuration.Set(time.Since(now).Seconds())
	activeDatabases.Set(float64(len(s.activeDB)))
}

// create a backup from the badger db, and store it in backupDir.
// keep a backup of last seven days only. Older backups are
// overwritten
func (s *server) backup() {

	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		os.Mkdir(backupDir, os.ModePerm)
	}

	fileName := fmt.Sprintf("%s/badger_%d.bak", backupDir, time.Now().Weekday())
	f, err := os.Create(fileName)
	if err != nil {
		s.logger.Printf("fail to create file %s: %v", fileName, err)
	}
	defer f.Close()

	_, err = s.storage.Backup(f, 1)
	if err != nil {
		s.logger.Printf("backup failed: %v", err)
	}
	fileInfo, err := f.Stat()
	if err != nil {
		s.logger.Printf("fail to get backup stats")
	}
	badgerBackup.Set(float64(fileInfo.Size()))
}

type dbMetaInfo struct {
	collections   sort.StringSlice
	lastUsed      int64
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
