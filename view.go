package main

import (
	"net/http"
	"strings"

	"github.com/dgraph-io/badger"
)

// view a saved playground page identified by its ID
func (s *server) viewHandler(w http.ResponseWriter, r *http.Request) {

	id := strings.TrimPrefix(r.URL.Path, viewEndpoint)
	p, err := s.loadPage([]byte(id))
	if err != nil {
		s.logger.Printf("fail to load page with id %s : %v", id, err)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("this playground doesn't exist"))
		return
	}
	err = templates.Execute(w, p)
	if err != nil {
		s.logger.Printf("fail to execute template with page %s: %v", p.String(), err)
		return
	}
}

func (s *server) loadPage(id []byte) (*page, error) {
	p := &page{
		MongoVersion: s.mongodbVersion,
	}
	err := s.storage.View(func(txn *badger.Txn) error {
		item, err := txn.Get(id)
		if err != nil {
			return err
		}
		val, err := item.Value()
		if err != nil {
			return err
		}
		p.decode(val)
		return nil
	})
	return p, err
}
