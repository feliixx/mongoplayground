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
	"github.com/andybalholm/brotli"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStaticHandlers(t *testing.T) {

	staticFileTests := []struct {
		name         string
		url          string
		contentType  string
		responseCode int
	}{
		{
			name:         "css",
			url:          "/static/playground-min-10.css",
			contentType:  "text/css; charset=utf-8",
			responseCode: http.StatusOK,
		},
		{
			name:         "documentation",
			url:          "/static/docs-10.html",
			contentType:  "text/html; charset=utf-8",
			responseCode: http.StatusOK,
		},
		{
			name:         "documentation",
			url:          "/static/about.html",
			contentType:  "text/html; charset=utf-8",
			responseCode: http.StatusOK,
		},
		{
			name:         "documentation",
			url:          "/static/playground-min-11.js",
			contentType:  "application/javascript; charset=utf-8",
			responseCode: http.StatusOK,
		},
		{
			name:         "favicon",
			url:          "/static/favicon.png",
			contentType:  "image/png",
			responseCode: http.StatusOK,
		},
		{
			name:         "non existing file",
			url:          "/static/unknown.txt",
			contentType:  "",
			responseCode: http.StatusNotFound,
		},
		{
			name:         "file outside of static",
			url:          "/static/../README.md",
			contentType:  "",
			responseCode: http.StatusNotFound,
		},
	}
	for _, tt := range staticFileTests {
		t.Run(tt.name, func(t *testing.T) {

			checkHandlerResponse(t, testServer.staticHandler, tt.url, tt.responseCode, tt.contentType, brotliEncoding)
			checkHandlerResponse(t, testServer.staticHandler, tt.url, tt.responseCode, tt.contentType, gzipEncoding)
		})
	}
}

func checkHandlerResponse(t *testing.T, handler func(w http.ResponseWriter, r *http.Request), url string, expectedResponseCode int, expectedContentType, expectedEncoding string) {

	resp := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Accept-Encoding", expectedEncoding)

	handler(resp, req)

	if expectedResponseCode != resp.Code {
		t.Errorf("expected response code %d but got %d", expectedResponseCode, resp.Code)
	}

	if expectedResponseCode == http.StatusOK {

		if want, got := expectedEncoding, resp.Header().Get("Content-Encoding"); want != got {
			t.Errorf("expected Content-Encoding: %s, but got %s", want, got)
		}

		if want, got := expectedContentType, resp.Header().Get("Content-Type"); want != got {
			t.Errorf("expected Content-Type: %s, but got %s", want, got)
		}

		var reader io.Reader
		if expectedEncoding == brotliEncoding {
			reader = brotli.NewReader(resp.Body)
		} else {
			reader, _ = gzip.NewReader(resp.Body)
		}

		_, err := io.Copy(io.Discard, reader)
		if err != nil {
			t.Errorf("fail to read %s content: %v", expectedEncoding, err)
		}
	}
}
