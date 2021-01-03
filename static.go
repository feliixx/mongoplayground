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

package main

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const (
	templateConfigOld = `[
  {
    "collection": "collection",
    "count": 10,
    "content": {
		"k": {
		  "type": "int",
		  "minInt": 0, 
		  "maxInt": 10
		}
	}
  }
]`
	templateConfigNew = `[{"key":1},{"key":2}]`
	templateQuery     = "db.collection.find()"
)

// serve static ressources (css/js/html)
func (s *server) staticHandler(w http.ResponseWriter, r *http.Request) {

	name := strings.TrimPrefix(r.URL.Path, staticEndpoint)

	content, ok := s.staticContent[name]
	if !ok {
		s.logger.Printf("static resource %s doesn't exist", name)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", contentTypeFromName(name))
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Cache-Control", "public, max-age=31536000")

	w.Write(content)
}

func contentTypeFromName(name string) string {

	if strings.HasSuffix(name, ".css") {
		return "text/css; charset=utf-8"
	}
	if strings.HasSuffix(name, ".js") {
		return "application/javascript; charset=utf-8"
	}
	if strings.HasSuffix(name, ".png") {
		return "image/png"
	}
	return "text/html; charset=utf-8"
}

// load static resources (javascript, css, docs and default page)
// and compress them in order to serve them faster
func (s *server) compressStaticResources() error {

	var buf bytes.Buffer
	zw, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	zw.Name, zw.ModTime = homeEndpoint, time.Now()

	p, _ := newPage(bsonLabel, templateConfigNew, templateQuery)
	p.MongoVersion = s.mongodbVersion
	if err := templates.Execute(zw, p); err != nil {
		return err
	}
	if err := s.add(zw, &buf); err != nil {
		return err
	}

	files, err := ioutil.ReadDir(staticDir)
	if err != nil {
		return err
	}
	for _, f := range files {
		buf.Reset()
		zw.Reset(&buf)

		zw.Name, zw.ModTime = f.Name(), time.Now()
		b, err := ioutil.ReadFile(staticDir + "/" + f.Name())
		if err != nil {
			return err
		}
		if _, err = zw.Write(b); err != nil {
			return err
		}
		if err := s.add(zw, &buf); err != nil {
			return err
		}
	}
	return nil
}

func (s *server) add(zw *gzip.Writer, buf *bytes.Buffer) error {
	if s.staticContent == nil {
		s.staticContent = map[string][]byte{}
	}
	if err := zw.Close(); err != nil {
		return err
	}
	c := make([]byte, buf.Len())
	copy(c, buf.Bytes())
	s.staticContent[zw.Name] = c
	return nil
}
