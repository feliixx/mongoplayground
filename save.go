package main

import (
	"fmt"
	"net/http"

	"github.com/dgraph-io/badger/v2"
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

	id := s.save(p)

	fmt.Fprintf(w, "%sp/%s", r.Referer(), id)
}

func (s *server) save(p *page) []byte {

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
		s.storage.Update(func(txn *badger.Txn) error {
			return txn.Set(id, val)
		})
		// At this point, we know for sure that a new playground
		// has been saved, so update the stats
		savedPlayground.WithLabelValues(p.label()).Inc()
	}
	return id
}
