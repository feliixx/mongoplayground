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
	"compress/gzip"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/andybalholm/brotli"
)

const (
	templateConfig = `[{"key":1},{"key":2}]`
	templateQuery  = "db.collection.find()"
)

var homeTemplate = template.Must(template.ParseFS(assets, homeTemplateFile))

// return a playground with the default configuration
func (s *staticContent) homeHandler(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != homeEndpoint {
		log.Printf("file not found: %s", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
		w.Write(nil)
		return
	}

	page := &page{
		Mode:         bsonMode,
		Config:       []byte(templateConfig),
		Query:        []byte(templateQuery),
		MongoVersion: s.mongodbVersion,
	}

	serveHomeTemplate(
		w,
		page,
		strings.Contains(r.Header.Get("Accept-Encoding"), brotliEncoding),
	)
}

func serveHomeTemplate(w http.ResponseWriter, page *page, useBrotli bool) {

	var writer io.WriteCloser

	if useBrotli {
		w.Header().Set("Content-Encoding", brotliEncoding)
		writer = brotli.NewWriter(w)
	} else {
		w.Header().Set("Content-Encoding", gzipEncoding)
		writer = gzip.NewWriter(w)
	}

	w.Header().Set("Cache-Control", "no-transform")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	homeTemplate.Execute(writer, page)
	writer.Close()
}
