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
	"net/http"
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
			name:         "css old version",
			url:          "/static/playground-min-10.css",
			contentType:  "text/css; charset=utf-8",
			responseCode: http.StatusOK,
		},
		{
			name:         "css with md5 hash",
			url:          "/static/playground-min-62345cc3aaee366e7ea51bd732975c6b.css",
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
			name:         "about",
			url:          "/static/about.html",
			contentType:  "text/html; charset=utf-8",
			responseCode: http.StatusOK,
		},
		{
			name:         "js",
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
			name:         "file outside of static redirect to home page",
			url:          "/static/../README.md",
			contentType:  "",
			responseCode: http.StatusMovedPermanently,
		},
	}
	for _, tt := range staticFileTests {

		test := tt // capture range variable
		t.Run(test.name, func(t *testing.T) {

			t.Parallel()

			checkServerResponse(t, test.url, test.responseCode, test.contentType, gzipEncoding)
		})
	}
}
