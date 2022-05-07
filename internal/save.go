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
	"fmt"
	"net/http"

	"github.com/dgraph-io/badger/v2"
)

// save the playground and return the playground url, which looks
// like:
//
//   https://mongoplayground.net/p/nJhd-dhf3Ea
func (s *storage) saveHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Cache-control", "no-transform")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	p, err := newPage(
		r.FormValue("mode"),
		r.FormValue("config"),
		r.FormValue("query"),
	)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	id := s.save(p)

	fmt.Fprintf(w, "%sp/%s", r.Referer(), id)
}

func (s *storage) save(p *page) []byte {

	id, val := p.ID(), p.encode()

	// before saving, check if the playground is not already
	// saved
	alreadySaved := false
	s.kvStore.View(func(txn *badger.Txn) error {

		_, err := txn.Get(id)
		// if the key is not found, an 'ErrKeyNotFound' is returned.
		// hence if the error is nil, the playground is already saved
		if err == nil {
			alreadySaved = true
		}
		return nil
	})

	if !alreadySaved {
		s.kvStore.Update(func(txn *badger.Txn) error {
			return txn.Set(id, val)
		})
		// At this point, we know for sure that a new playground
		// has been saved, so update the stats
		savedPlaygroundSize.WithLabelValues(p.label()).Observe(float64(len(val)))
	}
	return id
}
