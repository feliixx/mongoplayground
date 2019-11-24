package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/globalsign/mgo"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	templates = template.Must(template.ParseFiles("playground.html"))
)

const (
	staticDir = "static"
	badgerDir = "storage"
	backupDir = "backups"

	homeEndpoint    = "/"
	viewEndpoint    = "/p/"
	runEndpoint     = "/run"
	saveEndpoint    = "/save"
	staticEndpoint  = "/static/"
	metricsEndpoint = "/metrics"

	// interval between two MongoDB cleanup
	cleanupInterval = 4 * time.Hour
	// interval between two Badger backup
	backupInterval = 24 * time.Hour
)

type dbMetaInfo struct {
	collections []string
	lastUsed    int64
}

type server struct {
	mux     *http.ServeMux
	session *mgo.Session
	storage *badger.DB
	logger  *log.Logger

	// mutex guards the activeDB map
	mutex    sync.RWMutex
	activeDB map[string]dbMetaInfo

	staticContent  map[string][]byte
	mongodbVersion []byte
}

func newServer(logger *log.Logger) (*server, error) {

	session, err := mgo.Dial("mongodb://")
	if err != nil {
		return nil, fmt.Errorf("fail to connect to mongodb: %v", err)
	}
	info, _ := session.BuildInfo()
	version := []byte(info.Version)

	opts := badger.DefaultOptions
	opts.Dir = badgerDir
	opts.ValueDir = badgerDir
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	s := &server{
		mux:            http.DefaultServeMux,
		session:        session,
		storage:        db,
		activeDB:       map[string]dbMetaInfo{},
		logger:         logger,
		mongodbVersion: version,
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
	s.mux.Handle(metricsEndpoint, promhttp.Handler())

	return s, nil
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

	session := s.session.Copy()
	defer session.Close()

	now := time.Now()
	defer cleanupDuration.Set(float64(time.Since(now)) / float64(time.Millisecond))

	s.mutex.Lock()
	defer s.mutex.Unlock()

	for name, infos := range s.activeDB {
		if now.Sub(time.Unix(infos.lastUsed, 0)) > cleanupInterval {
			err := session.DB(name).DropDatabase()
			if err != nil {
				s.logger.Printf("fail to drop database %v: %v", name, err)
			}
			delete(s.activeDB, name)
			activeDatabases.Dec()
		}
	}
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
}
