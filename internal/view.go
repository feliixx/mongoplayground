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
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/dgraph-io/badger/v2"
)

const errNoMatchingPlayground = "this playground doesn't exist"

// view a saved playground page identified by its ID
func (s *storage) viewHandler(w http.ResponseWriter, r *http.Request) {

	id := extractPageIDFromURL(r.URL.Path)

	page, err := s.loadPage(id)
	if err != nil {
		log.Printf("fail to load page with id %s : %v", id, err)
		serveNoMatchingPlayground(w)
		return
	}

	serveHomeTemplate(w, page)
}

func extractPageIDFromURL(url string) []byte {

	id := strings.TrimPrefix(url, viewEndpoint)
	if len(id) > pageIDLength {
		id = id[:pageIDLength]
	}
	return []byte(id)
}

func (s *storage) loadPage(id []byte) (*page, error) {

	if len(id) != pageIDLength {
		return nil, errors.New("invalid page id length")
	}

	p := &page{
		MongoVersion: s.mongoVersion,
	}
	err := s.kvStore.View(func(txn *badger.Txn) error {
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

func serveNoMatchingPlayground(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(errNoMatchingPlayground))
}
