package main

import (
	"bytes"
	"context"
	"fmt"
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
	"go.mongodb.org/mongo-driver/bson"
)

const (
	templateResult = `[{"_id":ObjectId("5a934e000102030405000000"),"k":10},{"_id":ObjectId("5a934e000102030405000001"),"k":2},{"_id":ObjectId("5a934e000102030405000002"),"k":7},{"_id":ObjectId("5a934e000102030405000003"),"k":6},{"_id":ObjectId("5a934e000102030405000004"),"k":9},{"_id":ObjectId("5a934e000102030405000005"),"k":10},{"_id":ObjectId("5a934e000102030405000006"),"k":9},{"_id":ObjectId("5a934e000102030405000007"),"k":10},{"_id":ObjectId("5a934e000102030405000008"),"k":2},{"_id":ObjectId("5a934e000102030405000009"),"k":1}]`
	templateURL    = "p/snbIQ3uGHGq"
	testDir        = "tests"
)

var (
	templateParams = url.Values{"mode": {"mgodatagen"}, "config": {templateConfigOld}, "query": {templateQuery}}
	testServer     *server
)

func TestMain(m *testing.M) {
	err := os.RemoveAll(badgerDir)
	if err != nil {
		fmt.Printf("aborting: %v\n", err)
		os.Exit(1)
	}
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		os.Mkdir(testDir, os.ModePerm)
	}
	log := log.New(ioutil.Discard, "", 0)
	s, err := newServer(log)
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

func TestServeHTTP(t *testing.T) {

	t.Parallel()

	req, _ := http.NewRequest(http.MethodGet, homeEndpoint, nil)
	resp := httptest.NewRecorder()
	testServer.ServeHTTP(resp, req)
	if http.StatusOK != resp.Code {
		t.Errorf("expected code %d but got %d", http.StatusOK, resp.Code)
	}
}

func TestBasePage(t *testing.T) {

	t.Parallel()

	req, _ := http.NewRequest(http.MethodGet, homeEndpoint, nil)
	resp := httptest.NewRecorder()

	testServer.newPageHandler(resp, req)

	if http.StatusOK != resp.Code {
		t.Errorf("expected response code %d but got %d", http.StatusOK, resp.Code)
	}

	if want, got := "text/html; charset=utf-8", resp.Header().Get("Content-Type"); want != got {
		t.Errorf("expected Content-Type: %s but got %s", want, got)
	}

	if want, got := "gzip", resp.Header().Get("Content-Encoding"); want != got {
		t.Errorf("expected Content-Encoding: %s but got %s", want, got)
	}
}

func TestRemoveOldDB(t *testing.T) {

	defer testServer.clearDatabases(t)

	params := url.Values{"mode": {"mgodatagen"}, "config": {templateConfigOld}, "query": {templateQuery}}
	buf := httpBody(t, testServer.runHandler, http.MethodPost, runEndpoint, params)
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
	buf = httpBody(t, testServer.runHandler, http.MethodPost, runEndpoint, params)

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

	dir, _ := ioutil.ReadDir(backupDir)
	for _, d := range dir {
		os.RemoveAll(path.Join([]string{backupDir, d.Name()}...))
	}

	testServer.backup()

	dir, _ = ioutil.ReadDir(backupDir)
	if len(dir) != 1 {
		t.Error("a backup file should have been created, but there was none")
	}
}

func (s *server) clearDatabases(t *testing.T) {
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
		t.Errorf("fail to commit delete trnascation: %v", err)
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

func (s *server) countSavedPages() (count int) {
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

func httpBody(t *testing.T, handler func(http.ResponseWriter, *http.Request), method string, url string, params url.Values) *bytes.Buffer {
	req, err := http.NewRequest(method, url, strings.NewReader(params.Encode()))
	if err != nil {
		t.Error(err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp := httptest.NewRecorder()
	handler(resp, req)
	return resp.Body
}
