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

	name := strings.TrimPrefix(r.URL.Path, "/static/")
	sub := strings.Split(name, ".")

	contentType := "text/html; charset=utf-8"
	if len(sub) > 0 {
		switch sub[len(sub)-1] {
		case "css":
			contentType = "text/css; charset=utf-8"
		case "js":
			contentType = "application/javascript; charset=utf-8"
		}
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Cache-Control", "public, max-age=31536000")
	pos, ok := s.staticContentMap[name]
	if !ok {
		s.logger.Printf("static resource %s doesn't exist", name)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Write(s.staticContent[pos])
}

// load static ressources (javascript, css, docs and default page)
// and compress them in order to serve them faster
func (s *server) precompile() error {

	var buf bytes.Buffer
	zw, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	zw.Name = "playground.html"
	zw.ModTime = time.Now()
	p := &page{
		Mode:         bsonMode,
		Config:       []byte(templateConfig),
		Query:        []byte(templateQuery),
		MongoVersion: s.mongodbVersion,
	}
	if err := templates.Execute(zw, p); err != nil {
		return err
	}
	if err := s.add(zw, &buf, 0); err != nil {
		return err
	}

	files, err := ioutil.ReadDir(staticDir)
	if err != nil {
		return err
	}
	for i, f := range files {
		buf.Reset()
		zw.Reset(&buf)
		zw.Name = f.Name()
		zw.ModTime = time.Now()
		b, err := ioutil.ReadFile(staticDir + "/" + f.Name())
		if err != nil {
			return err
		}
		if _, err = zw.Write(b); err != nil {
			return err
		}
		if err := s.add(zw, &buf, i+1); err != nil {
			return err
		}
	}
	return nil
}

func (s *server) add(zw *gzip.Writer, buf *bytes.Buffer, index int) error {
	if s.staticContent == nil {
		s.staticContent = make([][]byte, 0)
		s.staticContentMap = map[string]int{}
	}
	if err := zw.Close(); err != nil {
		return err
	}
	c := make([]byte, buf.Len())
	copy(c, buf.Bytes())
	s.staticContentMap[zw.Name] = index
	s.staticContent = append(s.staticContent, c)
	return nil
}
