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
	activeDatabasesCounter = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_db_count",
			Help: "Active databases created on the Server",
		},
	)
	savedPlaygroundSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "saved_playground_size",
			Help:    "Histogram of saved playground size in byte",
			Buckets: []float64{1000, 5000, 10000, 100000, 300000},
		},
		[]string{"type"},
	)
	cleanupDuration = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "cleanup_duration_seconds",
			Help: "Database cleanup in second",
		},
	)
	badgerBackupSize = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "badger_backup_size_bytes",
			Help: "Size of last badger backup in bytes",
		},
	)
	gzipCounter = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "request_compressed_with_gzip",
			Help: "Nb of request compressed with gzip",
		},
	)
)

func initPrometheusCounter(storage *badger.DB) {
	prometheus.MustRegister(requestDurations)
	prometheus.MustRegister(activeDatabasesCounter)
	prometheus.MustRegister(savedPlaygroundSize)
	prometheus.MustRegister(cleanupDuration)
	prometheus.MustRegister(badgerBackupSize)
	prometheus.MustRegister(gzipCounter)

	computeSavedPlaygroundStats(storage)
}

func computeSavedPlaygroundStats(storage *badger.DB) {

	storage.View(func(txn *badger.Txn) error {

		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			item.Value(func(val []byte) error {
				p := &page{}
				p.decode(val)
				savedPlaygroundSize.WithLabelValues(p.label()).Observe(float64(len(val)))
				return nil
			})
		}
		return nil
	})
}
