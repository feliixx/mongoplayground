package main

import (
	"fmt"
	"net/http"

	"github.com/dgraph-io/badger"
)

// save the playground and return the playground url, which looks
// like:
//
//   https://mongoplayground.net/p/nJhd-dhf3Ea
func (s *server) saveHandler(w http.ResponseWriter, r *http.Request) {

	p, err := newPage(
		r.FormValue("mode"),
		r.FormValue("config"),
		r.FormValue("query"),
	)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	id, err := s.save(p)
	if err != nil {
		s.logger.Printf("fail to save playground %s with id %s: %v", p.String(), id, err)
		w.Write([]byte("fail to save playground"))
		return
	}
	fmt.Fprintf(w, "%sp/%s", r.Referer(), id)
}

func (s *server) save(p *page) ([]byte, error) {

	id, val := p.ID(), p.encode()

	// before saving, check if the playground is not already
	// saved
	alreadySaved := false
	s.storage.View(func(txn *badger.Txn) error {

		_, err := txn.Get(id)
		// if the key is not found, an 'ErrKeyNotFound' is returned.
		// hence if the error is nil, the playground is already saved
		if err == nil {
			alreadySaved = true
		}
		return nil
	})

	if !alreadySaved {
		err := s.storage.Update(func(txn *badger.Txn) error {
			return txn.Set(id, val)
		})
		if err != nil {
			return nil, err
		}
		// At this point, we know for sure that a new playground
		// has been saved, so update the stats
		savedPlayground.WithLabelValues(p.label()).Inc()
	}
	return id, nil
}
