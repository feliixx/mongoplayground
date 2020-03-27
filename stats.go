package main

import (
	"time"

	"github.com/dgraph-io/badger/v2"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	requestDurations = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "request_durations_seconds",
			Help:       "Request latency distributions",
			MaxAge:     1 * time.Hour, // compute quantiles on last hour
			Objectives: map[float64]float64{0.5: 0.05, 0.75: 0.01, 0.95: 0.001},
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
