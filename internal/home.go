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
	"log"
	"net/http"
	"strconv"
	"strings"
)

// return a playground with the default configuration
func (s *staticContent) homeHandler(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != homeEndpoint {
		log.Printf("file not found: %s", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
		w.Write(nil)
		return
	}

	acceptedEncoding := gzipEncoding
	if strings.Contains(r.Header.Get("Accept-Encoding"), brotliEncoding) {
		acceptedEncoding = brotliEncoding
	}

	resource,_ := s.getResource(homeEndpoint, acceptedEncoding)

	w.Header().Set("Content-Encoding", resource.contentEncoding)
	w.Header().Set("Content-Type", resource.contentType)
	w.Header().Set("Content-Length", strconv.Itoa(len(resource.content)))
	w.Write(resource.content)
}
