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
	"bytes"
	"compress/gzip"
	"embed"
	"html/template"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/andybalholm/brotli"
)

const (
	templateConfig = `[{"key":1},{"key":2}]`
	templateQuery  = "db.collection.find()"

	staticDir        = "web/static"
	homeTemplateFile = "web/playground.html"

	brotliEncoding = "br"
	gzipEncoding   = "gzip"
)

var (
	//go:embed web/static web/playground.html
	assets embed.FS

	homeTemplate *template.Template
	// regex to strip file id
	fileIdxReg = regexp.MustCompile("-[0-9]+.")
)

// serve static resources (css/js/html)
func (s *staticContent) staticHandler(w http.ResponseWriter, r *http.Request) {

	// transform 'static/playground-min-10.css' to 'playground-min.css'
	// the numeric id is just used to force the browser to reload the new version
	name := strings.TrimPrefix(r.URL.Path, staticEndpoint)
	name = fileIdxReg.ReplaceAllString(name, ".")

	if name == "playground-min.js" {
		pageLoadCounter.WithLabelValues(firstVisit).Inc()
	}

	acceptedEncoding := gzipEncoding
	if strings.Contains(r.Header.Get("Accept-Encoding"), brotliEncoding) {
		acceptedEncoding = brotliEncoding
	}

	resource, ok := s.getResource(name, acceptedEncoding)
	if !ok {
		log.Printf("static resource %s doesn't exist", name)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", resource.contentType)
	w.Header().Set("Cache-Control", "public, max-age=31536000")
	w.Header().Set("Content-Length", strconv.Itoa(len(resource.content)))

	if resource.contentEncoding != "" {
		w.Header().Set("Content-Encoding", resource.contentEncoding)
		staticEncodingCounter.WithLabelValues(resource.contentEncoding).Inc()
	}

	w.Write(resource.content)
}

type staticResource struct {
	content         []byte
	contentType     string
	contentEncoding string
}

type staticContent struct {
	mongodbVersion  []byte
	compressedFiles map[string]staticResource
}

func (s *staticContent) addResource(content []byte, name, contentType, contentEncoding string) {

	b := make([]byte, len(content))
	copy(b, content)

	s.compressedFiles[name] = staticResource{
		content:         b,
		contentType:     contentType,
		contentEncoding: contentEncoding,
	}
}

func (s *staticContent) addResourceFromFile(fileName, contentType, contentEncoding string) {
	var content []byte
	if contentEncoding == brotliEncoding {
		content = compressFileWithBrotli(fileName)
	} else {
		content = compressFileWithGzip(fileName)
	}
	s.addResource(content, contentEncoding+"_"+fileName, contentType, contentEncoding)
}

func (s *staticContent) getResource(name, acceptedEncoding string) (staticResource, bool) {

	key := acceptedEncoding + "_" + name
	// favicon is not compressed
	if name == "favicon.png" {
		key = name
	}
	resource, ok := s.compressedFiles[key]

	return resource, ok
}

// load static resources (javascript, css, docs and default page)
// and compress them once at startup in order to serve them faster
func compressStaticResources(mongodbVersion []byte) (*staticContent, error) {

	staticContent := &staticContent{
		mongodbVersion:  mongodbVersion,
		compressedFiles: map[string]staticResource{},
	}

	content, err := assets.ReadFile(staticDir + "/favicon.png")
	if err != nil {
		return nil, err
	}
	staticContent.addResource(content, "favicon.png", "image/png", "")

	content, err = executeHomeTemplate(mongodbVersion)
	if err != nil {
		return nil, err
	}
	staticContent.addResource(compressContent(content, gzipEncoding), gzipEncoding+"_"+homeEndpoint, "text/html; charset=utf-8", gzipEncoding)
	staticContent.addResource(compressContent(content, brotliEncoding), brotliEncoding+"_"+homeEndpoint, "text/html; charset=utf-8", brotliEncoding)

	staticContent.addResourceFromFile("playground-min.css", "text/css; charset=utf-8", gzipEncoding)
	staticContent.addResourceFromFile("playground-min.css", "text/css; charset=utf-8", brotliEncoding)

	staticContent.addResourceFromFile("playground-min.js", "application/javascript; charset=utf-8", gzipEncoding)
	staticContent.addResourceFromFile("playground-min.js", "application/javascript; charset=utf-8", brotliEncoding)

	staticContent.addResourceFromFile("docs.html", "text/html; charset=utf-8", gzipEncoding)
	staticContent.addResourceFromFile("docs.html", "text/html; charset=utf-8", brotliEncoding)
	staticContent.addResourceFromFile("about.html", "text/html; charset=utf-8", gzipEncoding)
	staticContent.addResourceFromFile("about.html", "text/html; charset=utf-8", brotliEncoding)

	return staticContent, nil
}

func compressFileWithGzip(fileName string) []byte {
	b, _ := assets.ReadFile(staticDir + "/" + fileName)
	return compressContent(b, gzipEncoding)
}

func compressFileWithBrotli(fileName string) []byte {
	b, _ := assets.ReadFile(staticDir + "/" + fileName)
	return compressContent(b, brotliEncoding)
}

func compressContent(content []byte, encoding string) []byte {

	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	var wc io.WriteCloser

	if encoding == brotliEncoding {
		wc = brotli.NewWriterLevel(buf, brotli.BestCompression)
	} else {
		wc, _ = gzip.NewWriterLevel(buf, gzip.BestCompression)
	}

	wc.Write(content)
	wc.Close()
	return buf.Bytes()
}

func executeHomeTemplate(mongoVersion []byte) ([]byte, error) {
	w := bytes.NewBuffer(nil)
	homeTemplate = template.Must(template.ParseFS(assets, homeTemplateFile))

	p, _ := newPage(bsonLabel, templateConfig, templateQuery)
	p.MongoVersion = mongoVersion

	if err := homeTemplate.Execute(w, p); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}
