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
	"github.com/dgraph-io/badger/v2"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	requestDurations = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "request_durations_seconds",
			Help:    "Histogram of latencies for HTTP request",
			Buckets: []float64{0.001, 0.01, 0.1, 0.25, 0.5, 1, 2.5, 10, 60},
		},
		[]string{"handler"},
	)
	activeDatabases = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_db_count",
			Help: "Active databases created on the server",
		},
	)
	savedPlayground = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "saved_playground_count",
			Help: "Saved playground",
		},
		[]string{"type"},
	)
	cleanupDuration = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "cleanup_duration_seconds",
			Help: "Database cleanup in second",
		},
	)
)

func init() {
	prometheus.MustRegister(requestDurations)
	prometheus.MustRegister(activeDatabases)
	prometheus.MustRegister(savedPlayground)
	prometheus.MustRegister(cleanupDuration)
}

func (s *server) computeSavedPlaygroundStats() error {

	return s.storage.View(func(txn *badger.Txn) error {

		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			item.Value(func(val []byte) error {
				p := &page{}
				p.decode(val)
				savedPlayground.WithLabelValues(p.label()).Inc()
				return nil
			})
		}
		return nil
	})
}
