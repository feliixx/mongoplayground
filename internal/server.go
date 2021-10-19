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
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	homeEndpoint    = "/"
	viewEndpoint    = "/p/"
	runEndpoint     = "/run"
	saveEndpoint    = "/save"
	staticEndpoint  = "/static/"
	metricsEndpoint = "/metrics"
	healthEndpoint  = "/health"

	readTimeout  = 10 * time.Second
	writeTimeout = 30 * time.Second
	idleTimeout  = 3 * time.Minute

	errInternalServerError = "Internal server error.\n  Please file an issue here:\n\n  https://github.com/feliixx/mongoplayground/issues"
)

// NewServer initialize a badger and a mongodb connection,
// and return an http server
func NewServer(mongoUri string, dropFirst bool, badgerDir, backupDir string, mailInfo *MailInfo) (*http.Server, error) {

	storage, err := newStorage(mongoUri, dropFirst, badgerDir, backupDir, mailInfo)
	if err != nil {
		return nil, err
	}
	return newHttpServerWithStorage(storage)
}

func newHttpServerWithStorage(storage *storage) (*http.Server, error) {

	staticContent, err := compressStaticResources(storage.mongoVersion)
	if err != nil {
		return nil, fmt.Errorf("fail to compress static resources: %v", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc(homeEndpoint, staticContent.homeHandler)
	mux.HandleFunc(viewEndpoint, storage.viewHandler)
	mux.HandleFunc(runEndpoint, storage.runHandler)
	mux.HandleFunc(saveEndpoint, storage.saveHandler)
	mux.HandleFunc(staticEndpoint, staticContent.staticHandler)
	mux.HandleFunc(healthEndpoint, storage.healthHandler)
	mux.Handle(metricsEndpoint, promhttp.Handler())

	return &http.Server{
		Addr:         ":8080",
		Handler:      latencyAndPanicObserver(mux, storage.mailInfo),
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}, nil
}

// Wrapper for each http request.
// Monitors latency of each endpoint, and explicitely
// logs panic stack trace so they can be send to loki.
// If smtp info are configured, also send the stack trace
// by email
func latencyAndPanicObserver(handler http.Handler, mailInfo *MailInfo) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer handleAnyPanic(w, r, mailInfo)

		start := time.Now()
		handler.ServeHTTP(w, r)

		label := r.URL.Path
		if strings.HasPrefix(label, viewEndpoint) {
			label = viewEndpoint
		}
		if strings.HasPrefix(label, staticEndpoint) {
			label = staticEndpoint
		}

		if label != homeEndpoint &&
			label != viewEndpoint &&
			label != runEndpoint &&
			label != saveEndpoint &&
			label != staticEndpoint &&
			label != healthEndpoint &&
			label != metricsEndpoint {
			label = "invalid"
		}
		requestDurations.WithLabelValues(label).Observe(float64(time.Since(start)) / float64(time.Second))
	})
}

func handleAnyPanic(w http.ResponseWriter, r *http.Request, mailInfo *MailInfo) {

	if panic := recover(); panic != nil {

		stackTrace := string(debug.Stack())
		log.Print(stackTrace)

		if mailInfo != nil {
			go mailInfo.sendRequestAndStackTraceByEmail(r, stackTrace)
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte(errInternalServerError))
	}
}
