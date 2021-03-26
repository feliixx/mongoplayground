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
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/dgraph-io/badger/v2"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"go.mongodb.org/mongo-driver/bson"
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
	testServer     *Server
)

func TestMain(m *testing.M) {

	storage, _ := ioutil.TempDir(os.TempDir(), "storage")
	backups, _ := ioutil.TempDir(os.TempDir(), "backups")

	logger := log.New(io.Discard, "", 0)
	s, err := NewServer(logger, storage, backups)
	if err != nil {
		fmt.Printf("aborting: %v\n", err)
		os.Exit(1)
	}
	testServer = s

	defer s.session.Disconnect(context.Background())
	defer s.storage.Close()

	retCode := m.Run()
	os.Exit(retCode)
}

func TestBasePage(t *testing.T) {

	t.Parallel()

	checkServerResponse(t, homeEndpoint, http.StatusOK, "text/html; charset=utf-8", brotliEncoding)
	checkServerResponse(t, homeEndpoint, http.StatusOK, "text/html; charset=utf-8", gzipEncoding)
}

func TestRemoveOldDB(t *testing.T) {

	defer testServer.clearDatabases(t)

	params := url.Values{"mode": {"mgodatagen"}, "config": {templateConfigOld}, "query": {templateQuery}}
	buf := httpBody(t, runEndpoint, http.MethodPost, params)
	if want, got := templateResult, buf.String(); want != got {
		t.Errorf("expected %s but got %s", want, got)
	}

	p := &page{
		Mode:   mgodatagenMode,
		Config: []byte(params.Get("config")),
	}

	DBHash := p.dbHash()
	dbInfo := testServer.activeDB[DBHash]
	dbInfo.lastUsed = time.Now().Add(-cleanupInterval).Unix()
	testServer.activeDB[DBHash] = dbInfo

	// this DB should not be removed
	configFormat := `[{"collection": "collection%v","count": 10,"content": {}}]`
	params.Set("config", fmt.Sprintf(configFormat, "other"))
	buf = httpBody(t, runEndpoint, http.MethodPost, params)

	if want, got := `collection "collection" doesn't exist`, buf.String(); want != got {
		t.Errorf("expected %s but got %s", want, got)
	}

	testServer.removeExpiredDB()

	_, ok := testServer.activeDB[DBHash]
	if ok {
		t.Errorf("DB %s should not be present in activeDB", DBHash)
	}

	dbNames, err := testServer.session.ListDatabaseNames(context.Background(), bson.D{})
	if err != nil {
		t.Error(err)
	}
	if dbNames[0] == DBHash {
		t.Errorf("%s should have been removed from mongodb", DBHash)
	}

	testStorageContent(t, 1, 0)
}

func TestBackup(t *testing.T) {

	dir, _ := os.ReadDir(testServer.backupDir)
	for _, d := range dir {
		os.RemoveAll(path.Join([]string{testServer.backupDir, d.Name()}...))
	}

	testServer.backup()

	dir, _ = os.ReadDir(testServer.backupDir)
	if len(dir) != 1 {
		t.Error("a backup file should have been created, but there was none")
	}
}

func (s *Server) clearDatabases(t *testing.T) {
	dbNames, err := s.session.ListDatabaseNames(context.Background(), bson.D{})
	if err != nil {
		t.Error(err)
	}

	for _, name := range filterDBNames(dbNames) {
		err = s.session.Database(name).Drop(context.Background())
		if err != nil {
			fmt.Printf("fail to drop db: %v", err)
		}
		delete(s.activeDB, name)
	}

	if len(s.activeDB) > 0 {
		t.Errorf("activeDB map content and databases doesn't match. Remaining keys: %v", s.activeDB)
		s.activeDB = map[string]dbMetaInfo{}
	}

	// reset prometheus metrics
	activeDatabases.Set(0)

	keys := make([][]byte, 0)
	err = s.storage.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := make([]byte, len(item.Key()))
			copy(key, item.Key())
			keys = append(keys, key)
		}
		return err
	})
	if err != nil {
		t.Error(err)
	}

	deleteTxn := s.storage.NewTransaction(true)
	for i := 0; i < len(keys); i++ {
		err = deleteTxn.Delete(keys[i])
		if err != nil {
			t.Error(err)
		}
	}
	err = deleteTxn.Commit()
	if err != nil {
		t.Errorf("fail to commit delete transcation: %v", err)
	}
}

func testStorageContent(t *testing.T, nbMongoDatabases, nbBadgerRecords int) {
	dbNames, err := testServer.session.ListDatabaseNames(context.Background(), bson.D{})
	if err != nil {
		t.Error(err)
	}
	if want, got := nbMongoDatabases, len(filterDBNames(dbNames)); want != got {
		t.Errorf("expected %d DB, but got %d", want, got)
	}
	if want, got := nbMongoDatabases, len(testServer.activeDB); want != got {
		t.Errorf("expected %d db in map, but got %d", want, got)
	}
	if want, got := nbMongoDatabases, int(testutil.ToFloat64(activeDatabases)); want != got {
		t.Errorf("expected %d active db in prometheus counter, but got %d", want, got)
	}
	if want, got := nbBadgerRecords, testServer.countSavedPages(); want != got {
		t.Errorf("expected %d page saved, but got %d", want, got)
	}
}

// return only created db, and get rid of 'indexes', 'local'
func filterDBNames(dbNames []string) []string {
	r := make([]string, 0)
	for _, n := range dbNames {
		if len(n) == 32 {
			r = append(r, n)
		}
	}
	return r
}

func (s *Server) countSavedPages() (count int) {
	s.storage.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			count++
		}
		return nil
	})
	return count
}

func httpBody(t *testing.T, url string, method string, params url.Values) *bytes.Buffer {
	req, err := http.NewRequest(method, url, strings.NewReader(params.Encode()))
	if err != nil {
		t.Error(err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp := httptest.NewRecorder()
	testServer.ServeHTTP(resp, req)
	return resp.Body
}
