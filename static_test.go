package main

import (
	"compress/gzip"
	"io"
	"io/ioutil"
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
			url:          "/static/playground-min-5.css",
			contentType:  "text/css; charset=utf-8",
			responseCode: 200,
		},
		{
			name:         "documentation",
			url:          "/static/docs-5.html",
			contentType:  "text/html; charset=utf-8",
			responseCode: 200,
		},
		{
			name:         "documentation",
			url:          "/static/playground-min-5.js",
			contentType:  "application/javascript; charset=utf-8",
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
				_, err = io.Copy(ioutil.Discard, zr)
				if err != nil {
					t.Errorf("fail to read gzip content: %v", err)
				}
				zr.Close()
			}
		})
	}
}
