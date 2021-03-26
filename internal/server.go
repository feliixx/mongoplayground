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
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
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

// Server is struct implementing http.Handler and holding
// mongodb and badger connection
type Server struct {
	mux            *http.ServeMux
	session        *mongo.Client
	mongodbVersion []byte
	storage        *badger.DB
	logger         *log.Logger

	// activeDB holds info of the database created / used during
	// the last cleanupInterval. Its access is garded by activeDbLock
	activeDbLock sync.RWMutex
	activeDB     map[string]dbMetaInfo

	// map storing static content compressed with brotli
	staticContent map[string][]byte
	// local dir to store badger backups
	backupDir string
}

// NewServer returns a new instance of Server
func NewServer(logger *log.Logger, badgerDir, backupDir string) (*Server, error) {

	session, mongodbVersion, err := createMongodbSession()
	if err != nil {
		return nil, err
	}

	badgerDB, err := badger.Open(badger.DefaultOptions(badgerDir))
	if err != nil {
		return nil, err
	}

	staticContent, err := compressStaticResources(mongodbVersion)
	if err != nil {
		return nil, fmt.Errorf("fail to compress static resources: %v", err)
	}

	s := &Server{
		mux:            http.DefaultServeMux,
		session:        session,
		mongodbVersion: mongodbVersion,
		storage:        badgerDB,
		activeDB:       map[string]dbMetaInfo{},
		logger:         logger,
		backupDir:      backupDir,
		staticContent:  staticContent,
	}

	initPrometheusCounter(s.storage)

	go func(s *Server) {
		for range time.Tick(cleanupInterval) {
			s.removeExpiredDB()
		}
	}(s)

	go func(s *Server) {
		for range time.Tick(backupInterval) {
			s.backup()
		}
	}(s)

	s.mux.HandleFunc(homeEndpoint, s.homeHandler)
	s.mux.HandleFunc(viewEndpoint, s.viewHandler)
	s.mux.HandleFunc(runEndpoint, s.runHandler)
	s.mux.HandleFunc(saveEndpoint, s.saveHandler)
	s.mux.HandleFunc(staticEndpoint, s.staticHandler)
	s.mux.HandleFunc(healthEndpoint, s.healthHandler)
	s.mux.Handle(metricsEndpoint, promhttp.Handler())

	return s, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	start := time.Now()
	s.mux.ServeHTTP(w, r)

	label := r.URL.Path
	if strings.HasPrefix(label, viewEndpoint) {
		label = viewEndpoint
	}
	if strings.HasPrefix(label, staticEndpoint) {
		label = staticEndpoint
	}
	requestDurations.WithLabelValues(label).Observe(float64(time.Since(start)) / float64(time.Second))
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
