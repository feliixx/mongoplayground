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
			responseCode: 200,
		},
		{
			name:         "documentation",
			url:          "/static/docs-10.html",
			contentType:  "text/html; charset=utf-8",
			responseCode: 200,
		},
		{
			name:         "documentation",
			url:          "/static/about.html",
			contentType:  "text/html; charset=utf-8",
			responseCode: 200,
		},
		{
			name:         "documentation",
			url:          "/static/playground-min-11.js",
			contentType:  "application/javascript; charset=utf-8",
			responseCode: 200,
		},
		{
			name:         "favicon",
			url:          "/static/favicon.png",
			contentType:  "image/png",
			responseCode: 200,
		},
		{
			name:         "non existing file",
			url:          "/static/unknown.txt",
			contentType:  "",
			responseCode: 404,
		},
		{
			name:         "file outside of static",
			url:          "/static/../README.md",
			contentType:  "",
			responseCode: 404,
		},
	}
	for _, tt := range staticFileTests {
		t.Run(tt.name, func(t *testing.T) {

			resp := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, tt.url, nil)
			testServer.staticHandler(resp, req)

			if tt.responseCode != resp.Code {
				t.Errorf("expected response code %d but got %d", tt.responseCode, resp.Code)
			}

			if tt.responseCode == http.StatusOK {

				if want, got := "gzip", resp.Header().Get("Content-Encoding"); want != got {
					t.Errorf("expected Content-Encoding: %s, but got %s", want, got)
				}

				if want, got := tt.contentType, resp.Header().Get("Content-Type"); want != got {
					t.Errorf("expected Content-Type: %s, but got %s", want, got)
				}

				zr, err := gzip.NewReader(resp.Body)
				if err != nil {
					t.Errorf("coulnd't read response body: %v", err)
				}
				_, err = io.Copy(io.Discard, zr)
				if err != nil {
					t.Errorf("fail to read gzip content: %v", err)
				}
				zr.Close()
			}
		})
	}
}
