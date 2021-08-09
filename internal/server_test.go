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
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/andybalholm/brotli"
)

const (
	templateResult    = `[{"_id":ObjectId("5a934e000102030405000000"),"k":10},{"_id":ObjectId("5a934e000102030405000001"),"k":2},{"_id":ObjectId("5a934e000102030405000002"),"k":7},{"_id":ObjectId("5a934e000102030405000003"),"k":6},{"_id":ObjectId("5a934e000102030405000004"),"k":9},{"_id":ObjectId("5a934e000102030405000005"),"k":10},{"_id":ObjectId("5a934e000102030405000006"),"k":9},{"_id":ObjectId("5a934e000102030405000007"),"k":10},{"_id":ObjectId("5a934e000102030405000008"),"k":2},{"_id":ObjectId("5a934e000102030405000009"),"k":1}]`
	templateURL       = "p/snbIQ3uGHGq"
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
)

var (
	templateParams = url.Values{"mode": {"mgodatagen"}, "config": {templateConfigOld}, "query": {templateQuery}}
	testServer     *http.Server
	testStorage    *storage
)

func TestMain(m *testing.M) {

	log.SetOutput(os.Stdout)

	storageDir, _ := ioutil.TempDir(os.TempDir(), "storage")
	backupsDir, _ := ioutil.TempDir(os.TempDir(), "backups")

	ts, err := newStorage("mongodb://localhost:27017", storageDir, backupsDir)
	if err != nil {
		fmt.Printf("aborting: %v\n", err)
		os.Exit(1)
	}
	testStorage = ts

	s, err := newHttpServerWithStorage(testStorage)
	if err != nil {
		fmt.Printf("aborting: %v\n", err)
		os.Exit(1)
	}
	testServer = s

	defer testStorage.mongoSession.Disconnect(context.Background())
	defer testStorage.kvStore.Close()

	retCode := m.Run()
	os.Exit(retCode)
}

func TestBasePage(t *testing.T) {

	t.Parallel()

	checkServerResponse(t, homeEndpoint, http.StatusOK, "text/html; charset=utf-8", brotliEncoding)
	checkServerResponse(t, homeEndpoint, http.StatusOK, "text/html; charset=utf-8", gzipEncoding)

	checkServerResponse(t, "/robots.txt", http.StatusNotFound, "", gzipEncoding)
}

func checkServerResponse(t *testing.T, url string, expectedResponseCode int, expectedContentType, expectedEncoding string) {

	resp := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Accept-Encoding", expectedEncoding)

	testServer.Handler.ServeHTTP(resp, req)

	if expectedResponseCode != resp.Code {
		t.Errorf("expected response code %d but got %d", expectedResponseCode, resp.Code)
	}

	if expectedResponseCode == http.StatusOK {

		if want, got := expectedContentType, resp.Header().Get("Content-Type"); want != got {
			t.Errorf("expected Content-Type: %s, but got %s", want, got)
		}

		// only for favicon
		encoding := resp.Header().Get("Content-Encoding")
		if encoding == "" {
			if resp.Body.Len() == 0 {
				t.Errorf("invalid empty body")
			}
			return
		}

		if want, got := expectedEncoding, encoding; want != got {
			t.Errorf("expected Content-Encoding: %s, but got %s", want, got)
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

func httpBody(t *testing.T, url string, method string, params url.Values) string {
	req, err := http.NewRequest(method, url, strings.NewReader(params.Encode()))
	if err != nil {
		t.Error(err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp := httptest.NewRecorder()
	testServer.Handler.ServeHTTP(resp, req)
	return resp.Body.String()
}
