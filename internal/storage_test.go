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
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"testing"
	"time"

	"github.com/dgraph-io/badger/v2"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"go.mongodb.org/mongo-driver/bson"
)

func TestDeleteExistingDB(t *testing.T) {

	defer clearDatabases(t)

	p, _ := newPage("", "", "")
	testStorage.mongoSession.
		Database(p.dbHash()).
		Collection("c").
		InsertOne(context.Background(), bson.M{"_id": 1})

	testStorage.deleteExistingDB()

	testStorageContent(t, 0, 0, 0)
}

func TestRemoveExpiredDB(t *testing.T) {

	defer clearDatabases(t)

	params := url.Values{"mode": {"mgodatagen"}, "config": {templateConfigOld}, "query": {templateQuery}}
	want := templateResult
	got := httpBody(t, runEndpoint, http.MethodPost, params)
	if want != got {
		t.Errorf("expected %s but got %s", want, got)
	}

	p := &page{
		Mode:   mgodatagenMode,
		Config: []byte(params.Get("config")),
	}

	// this db should be removed: too old
	DBHash := p.dbHash()
	testStorage.activeDB.Lock()
	dbInfo := testStorage.activeDB.list[DBHash]
	dbInfo.lastUsed = time.Now().Add(-maxUnusedDuration).Unix()
	testStorage.activeDB.list[DBHash] = dbInfo
	testStorage.activeDB.Unlock()

	// this db should be removed: creation error
	params = url.Values{"mode": {"mgodatagen"}, "config": {"[{}]"}, "query": {templateQuery}}
	want = `error in configuration:
  error in configuration file: 
	'collection' and 'database' fields can't be empty`
	got = httpBody(t, runEndpoint, http.MethodPost, params)
	if want != got {
		t.Errorf("expected %s but got %s", want, got)
	}

	// this DB should not be removed
	params = url.Values{"mode": {"bson"}, "config": {"[{_id:1}]"}, "query": {templateQuery}}
	want = `[{"_id":1}]`
	got = httpBody(t, runEndpoint, http.MethodPost, params)
	if want != got {
		t.Errorf("expected %s but got %s", want, got)
	}

	testStorage.removeUnusedDB()

	_, ok := testStorage.activeDB.list[DBHash]
	if ok {
		t.Errorf("DB %s should not be present in activeDB", DBHash)
	}

	dbNames, err := testStorage.mongoSession.ListDatabaseNames(context.Background(), bson.D{})
	if err != nil {
		t.Error(err)
	}
	if dbNames[0] == DBHash {
		t.Errorf("%s should have been removed from mongodb", DBHash)
	}

	testStorageContent(t, 1, 1, 0)
}

func TestBackup(t *testing.T) {

	dir, _ := os.ReadDir(testStorage.backupDir)
	for _, d := range dir {
		os.RemoveAll(path.Join([]string{testStorage.backupDir, d.Name()}...))
	}

	testStorage.backup()

	dir, _ = os.ReadDir(testStorage.backupDir)
	if len(dir) != 1 {
		t.Error("a backup file should have been created, but there was none")
	}
}

func clearDatabases(t *testing.T) {
	dbNames, err := testStorage.mongoSession.ListDatabaseNames(context.Background(), bson.D{})
	if err != nil {
		t.Error(err)
	}

	for _, name := range filterDBNames(dbNames) {
		err = testStorage.mongoSession.Database(name).Drop(context.Background())
		if err != nil {
			fmt.Printf("fail to drop db: %v", err)
		}
		delete(testStorage.activeDB.list, name)
	}

	for _, db := range testStorage.activeDB.list {
		if db.err == nil {
			t.Errorf(`Database leaked: %+v`, db)
		}
	}
	testStorage.activeDB = &cache{
		list: map[string]dbMetaInfo{},
	}
	// reset prometheus metrics
	activeDatabasesCounter.Set(0)

	keys := make([][]byte, 0)
	err = testStorage.kvStore.View(func(txn *badger.Txn) error {
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

	deleteTxn := testStorage.kvStore.NewTransaction(true)
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

func testStorageContent(t *testing.T, cacheSize, nbMongoDatabases, nbBadgerRecords int) {
	dbNames, err := testStorage.mongoSession.ListDatabaseNames(context.Background(), bson.D{})
	if err != nil {
		t.Error(err)
	}
	if want, got := nbMongoDatabases, len(filterDBNames(dbNames)); want != got {
		t.Errorf("expected %d DB, but got %d", want, got)
	}
	if want, got := cacheSize, len(testStorage.activeDB.list); want != got {
		t.Errorf("expected %d db in map, but got %d", want, got)
	}
	if want, got := nbMongoDatabases, int(testutil.ToFloat64(activeDatabasesCounter)); want != got {
		t.Errorf("expected %d active db in prometheus counter, but got %d", want, got)
	}
	if want, got := nbBadgerRecords, countSavedPages(testStorage.kvStore); want != got {
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

func countSavedPages(kvStore *badger.DB) (count int) {
	kvStore.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			count++
		}
		return nil
	})
	return count
}
