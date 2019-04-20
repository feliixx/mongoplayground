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

	id, val := p.ID(), p.encode()
	err := s.storage.Update(func(txn *badger.Txn) error {
		return txn.Set(id, val)
	})
	if err != nil {
		s.logger.Printf("fail to save playground %s with id %s", p.String(), id)
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "%sp/%s", r.Referer(), id)
}
