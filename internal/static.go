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
	"fmt"
	"html/template"
	"io"
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
	reg          = regexp.MustCompile("-[0-9]+.")
)

// serve static ressources (css/js/html)
func (s *Server) staticHandler(w http.ResponseWriter, r *http.Request) {

	// transform 'static/playground-min-10.css' to 'playground-min.css'
	// the numeric id is juste used to force the browser to reload the new version
	name := strings.TrimPrefix(r.URL.Path, staticEndpoint)
	name = reg.ReplaceAllString(name, ".")

	content, ok := s.staticContent[name]
	if !ok {
		s.logger.Printf("static resource %s doesn't exist", name)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", contentTypeFromName(name))
	w.Header().Set("Cache-Control", "public, max-age=31536000")

	if !strings.Contains(r.Header.Get("Accept-Encoding"), brotliEncoding) {
		fallbackToGzip(w, fmt.Sprintf("%s/%s", staticDir, name))
		return
	}

	w.Header().Set("Content-Length", strconv.Itoa(len(content)))
	w.Header().Set("Content-Encoding", brotliEncoding)

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

func fallbackToGzip(w http.ResponseWriter, assetPath string) {
	w.Header().Set("Content-Encoding", gzipEncoding)
	zw := gzip.NewWriter(w)
	content, _ := assets.ReadFile(assetPath)
	zw.Write(content)
	zw.Close()
}

// load static resources (javascript, css, docs and default page)
// and compress them once at startup in order to serve them faster
func compressStaticResources(mongodbVersion []byte) (map[string][]byte, error) {

	staticContent := map[string][]byte{}

	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	br := brotli.NewWriterLevel(buf, brotli.BestCompression)

	homeTemplate = template.Must(template.ParseFS(assets, homeTemplateFile))
	err := executeHomeTemplate(br, mongodbVersion)
	if err != nil {
		return nil, err
	}
	addCompressedRessource(staticContent, homeEndpoint, buf)

	files, err := assets.ReadDir(staticDir)
	if err != nil {
		return nil, err
	}
	for _, f := range files {

		buf.Reset()
		br.Reset(buf)

		b, err := assets.ReadFile(staticDir + "/" + f.Name())
		if err != nil {
			return nil, err
		}
		if _, err = br.Write(b); err != nil {
			return nil, err
		}
		if err := br.Close(); err != nil {
			return nil, err
		}
		addCompressedRessource(staticContent, f.Name(), buf)
	}
	return staticContent, nil
}

func addCompressedRessource(staticContent map[string][]byte, fileName string, buf *bytes.Buffer) {
	c := make([]byte, buf.Len())
	copy(c, buf.Bytes())
	staticContent[fileName] = c
}

func executeHomeTemplate(writer io.WriteCloser, mongoVersion []byte) error {
	p, _ := newPage(bsonLabel, templateConfig, templateQuery)
	p.MongoVersion = mongoVersion

	if err := homeTemplate.Execute(writer, p); err != nil {
		return err
	}
	return writer.Close()
}
