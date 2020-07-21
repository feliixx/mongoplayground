package main

import (
	"errors"
	"net/http"
	"strings"

	"github.com/dgraph-io/badger/v2"
)

const errNoMatchingPlayground = "this playground doesn't exist"

// view a saved playground page identified by its ID
func (s *server) viewHandler(w http.ResponseWriter, r *http.Request) {

	id := extractPageIDFromURL(r.URL.Path)

	p, err := s.loadPage(id)
	if err != nil {
		s.logger.Printf("fail to load page with id %s : %v", id, err)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(errNoMatchingPlayground))
		return
	}
	err = templates.Execute(w, p)
	if err != nil {
		s.logger.Printf("fail to execute template for playground %s: %v", id, err)
		return
	}
}

func extractPageIDFromURL(url string) []byte {

	id := strings.TrimPrefix(url, viewEndpoint)
	if len(id) > pageIDLength {
		id = id[:pageIDLength]
	}
	return []byte(id)
}

func (s *server) loadPage(id []byte) (*page, error) {

	if len(id) != pageIDLength {
		return nil, errors.New("invalid page id length")
	}

	p := &page{
		MongoVersion: s.mongodbVersion,
	}
	err := s.storage.View(func(txn *badger.Txn) error {
		item, err := txn.Get(id)
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			p.decode(val)
			return nil
		})
	})
	return p, err
}
