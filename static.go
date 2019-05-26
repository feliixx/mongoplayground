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
	templateConfig = `[
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
	templateQuery = "db.collection.find()"
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
	return "text/html; charset=utf-8"
}

// load static ressources (javascript, css, docs and default page)
// and compress them in order to serve them faster
func (s *server) compressStaticResources() error {

	var buf bytes.Buffer
	zw, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	zw.Name, zw.ModTime = homeEndpoint, time.Now()

	p := newPage(bsonLabel, templateConfig, templateQuery)
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
