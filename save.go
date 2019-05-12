package main

import (
	"fmt"
	"net/http"

	"github.com/dgraph-io/badger"
)

// save the playground and return the playground ID
func (s *server) saveHandler(w http.ResponseWriter, r *http.Request) {

	p := &page{
		Mode:   modeByte(r.FormValue("mode")),
		Config: []byte(r.FormValue("config")),
		Query:  []byte(r.FormValue("query")),
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

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
	err := s.storage.Update(func(txn *badger.Txn) error {
		return txn.Set(id, val)
	})
	if err != nil {
		return nil, err
	}

	savedPlayground.WithLabelValues(p.label()).Inc()

	return id, nil
}
