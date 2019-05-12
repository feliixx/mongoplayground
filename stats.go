package main

import (
	"github.com/dgraph-io/badger"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	requestDurations = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "request_durations_seconds",
			Help:       "Request latency distributions",
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
)

func init() {
	prometheus.MustRegister(requestDurations)
	prometheus.MustRegister(activeDatabases)
	prometheus.MustRegister(savedPlayground)
}

func (s *server) computeSavedPlaygroundStats() error {

	return s.storage.View(func(txn *badger.Txn) error {

		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		p := &page{}
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			value, err := item.Value()
			if err != nil {
				return err
			}
			p.decode(value)
			savedPlayground.WithLabelValues(p.label()).Inc()
		}
		return nil
	})
}
