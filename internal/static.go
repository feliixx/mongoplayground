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
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

const gzipEncoding = "gzip"

var (
	//go:embed web/static web/src/playground.html
	assets embed.FS

	// regex to match a md5 hash
	fileHashReg = regexp.MustCompile("-([0-9a-f]{32}|[0-9]+).")
)

// serve static resources (css/js/html)
//
// content is only compressed with gzip, as brotli compression from origin
// is not supported by cloudfare
//
// cloudfare will re-compress the resource using brotli for client that accept it
//
// see https://community.cloudflare.com/t/cloudfare-doesnt-serve-brolti-content-from-my-server/381662/5
// for details
func (s *staticContent) staticHandler(w http.ResponseWriter, r *http.Request) {

	// transform 'static/playground-min-c6c2fc9118ccb95919a78bc892d49228.css' to 'playground-min.css'
	// the numeric id is just used to force the browser to reload the new version
	name := strings.TrimPrefix(r.URL.Path, staticEndpoint)
	name = fileHashReg.ReplaceAllString(name, ".")

	resource, ok := s.resources[name]
	if !ok {
		log.Printf("static resource %s doesn't exist", name)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", resource.contentType)
	w.Header().Set("Cache-Control", "public, max-age=31536000")
	w.Header().Set("Content-Length", strconv.Itoa(len(resource.content)))

	if resource.compressed {
		w.Header().Set("Content-Encoding", gzipEncoding)
	}
	w.Write(resource.content)
}

type staticContent struct {
	resources map[string]staticResource
}

// load static resources (javascript, css, docs and default page)
// and compress them once at startup in order to serve them faster
func newStaticContent() *staticContent {

	return &staticContent{
		resources: map[string]staticResource{
			"favicon.png":        newResource("web/static/favicon.png", "image/png", false),
			"playground-min.css": newResource("web/static/playground-min.css", "text/css; charset=utf-8", true),
			"playground-min.js":  newResource("web/static/playground-min.js", "application/javascript; charset=utf-8", true),
			"docs.html":          newResource("web/static/docs.html", "text/html; charset=utf-8", true),
			"about.html":         newResource("web/static/about.html", "text/html; charset=utf-8", true),
		},
	}
}

type staticResource struct {
	content     []byte
	contentType string
	compressed  bool
}

func newResource(assetPath, contentType string, compressed bool) staticResource {

	content, err := assets.ReadFile(assetPath)
	if err != nil {
		panic(err)
	}
	if compressed {
		content = compressContent(content)
	}

	s := make([]byte, len(content))
	copy(s, content)

	return staticResource{
		content:     s,
		contentType: contentType,
		compressed:  compressed,
	}
}

func compressContent(content []byte) []byte {

	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	wc, _ := gzip.NewWriterLevel(buf, gzip.BestCompression)

	wc.Write(content)
	wc.Close()

	return buf.Bytes()
}
