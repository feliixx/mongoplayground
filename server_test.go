package main

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/stretchr/testify/assert"
)

const (
	templateResult = `[{"_id":{"$oid":"5a934e000102030405000000"},"k":10},{"_id":{"$oid":"5a934e000102030405000001"},"k":2},{"_id":{"$oid":"5a934e000102030405000002"},"k":7},{"_id":{"$oid":"5a934e000102030405000003"},"k":6},{"_id":{"$oid":"5a934e000102030405000004"},"k":9},{"_id":{"$oid":"5a934e000102030405000005"},"k":10},{"_id":{"$oid":"5a934e000102030405000006"},"k":9},{"_id":{"$oid":"5a934e000102030405000007"},"k":10},{"_id":{"$oid":"5a934e000102030405000008"},"k":2},{"_id":{"$oid":"5a934e000102030405000009"},"k":1}]
`
)

var (
	templateParams = url.Values{"mode": {"mgodatagen"}, "config": {templateConfig}, "query": {templateQuery}}
	templateURL    = "p/eYtYmPq-C4J"
	srv            *server
)

func TestMain(m *testing.M) {
	err := os.RemoveAll(badgerDir)
	if err != nil {
		fmt.Printf("aborting: %v\n", err)
		os.Exit(1)
	}
	s, err := newServer()
	if err != nil {
		fmt.Printf("aborting: %v\n", err)
		os.Exit(1)
	}
	srv = s
	defer s.session.Close()
	defer s.storage.Close()

	err = s.clearDatabases()
	if err != nil {
		fmt.Printf("fail to remove old db: %v\n", err)
		os.Exit(1)
	}

	retCode := m.Run()
	os.Exit(retCode)
}

func (s *server) clearDatabases() error {
	dbNames, err := s.session.DatabaseNames()
	if err != nil {
		return err
	}
	for _, name := range dbNames {
		if len(name) == 32 {
			s.session.DB(name).DropDatabase()
		}
	}
	s.activeDB.Range(func(k, v interface{}) bool {
		s.activeDB.Delete(k)
		return true
	})

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
		return err
	}

	deleteTxn := s.storage.NewTransaction(true)
	for i := 0; i < len(keys); i++ {
		err = deleteTxn.Delete(keys[i])
		if err != nil {
			return err
		}
	}
	return deleteTxn.Commit(func(err error) {
		if err != nil {
			fmt.Printf("fail to delete: %v\n", err)
		}
	})
}

// return only 32 long name
func filterDBNames(dbNames []string) []string {
	r := make([]string, 0)
	for _, n := range dbNames {
		if len(n) == 32 {
			r = append(r, n)
		}
	}
	return r
}

func (s *server) savedPageNb() int {
	count := 0
	s.storage.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			count++
		}
		return nil
	})
	return count
}

func getDBHash(mode byte, config []byte) string {
	return fmt.Sprintf("%x", md5.Sum(append(config, mode)))
}

func TestServeHTTP(t *testing.T) {
	req, _ := http.NewRequest("GET", "/static/docs.html", nil)
	resp := httptest.NewRecorder()
	srv.ServeHTTP(resp, req)
	assert.Equal(t, resp.Code, http.StatusOK)
}

func TestHomePage(t *testing.T) {
	err := precompile([]byte("3.6.3"))
	assert.Nil(t, err)
}

func TestRunCreateDB(t *testing.T) {
	l := []struct {
		params    url.Values
		result    string
		createdDB int
	}{
		// incorrect config should not create db
		{
			params:    url.Values{"mode": {"mgodatagen"}, "config": {"h"}, "query": {"h"}},
			result:    "fail to parse configuration: Error in configuration file: object / array / Date badly formatted: \n\n\t\tinvalid character 'h' looking for beginning of value",
			createdDB: 0,
		},
		// correct config, but collection 'c' doesn't exists
		{
			params:    url.Values{"mode": {"mgodatagen"}, "config": {templateConfig}, "query": {"db.c.find()"}},
			result:    NoDocFound,
			createdDB: 1,
		},
		// make sure that we always get the same list of "_id"
		{
			params:    templateParams,
			result:    templateResult,
			createdDB: 0, // db already exists
		},
		// make sure other generators produce the same output
		{
			params: url.Values{
				"mode": {"mgodatagen"},
				"config": {`[
				{
					"collection": "collection",
					"count": 10,
					"content": {
						"k": {
							"type": "string", 
							"minLength": 2,
							"maxLength": 5
						}
					}
				}
			]`}, "query": {templateQuery}},
			result: `[{"_id":{"$oid":"5a934e000102030405000000"},"k":"1jU"},{"_id":{"$oid":"5a934e000102030405000001"},"k":"tBRWL"},{"_id":{"$oid":"5a934e000102030405000002"},"k":"6Hch"},{"_id":{"$oid":"5a934e000102030405000003"},"k":"ZWHW"},{"_id":{"$oid":"5a934e000102030405000004"},"k":"RkMG"},{"_id":{"$oid":"5a934e000102030405000005"},"k":"RIr"},{"_id":{"$oid":"5a934e000102030405000006"},"k":"ru7"},{"_id":{"$oid":"5a934e000102030405000007"},"k":"OB"},{"_id":{"$oid":"5a934e000102030405000008"},"k":"ja"},{"_id":{"$oid":"5a934e000102030405000009"},"k":"K307"}]
`,
			createdDB: 1,
		},
		// same config, but aggregation
		{
			params: url.Values{
				"mode": {"mgodatagen"},
				"config": {`[
				{
					"collection": "collection",
					"count": 10,
					"content": {
						"k": {
							"type": "string", 
							"minLength": 2,
							"maxLength": 5
						}
					}
				}
			]`}, "query": {`db.collection.aggregate([{"$project": {"_id": 0}}])`}},
			result: `[{"k":"1jU"},{"k":"tBRWL"},{"k":"6Hch"},{"k":"ZWHW"},{"k":"RkMG"},{"k":"RIr"},{"k":"ru7"},{"k":"OB"},{"k":"ja"},{"k":"K307"}]
`,
			createdDB: 0,
		},
		// same query/config, number of doc too big
		{
			params: url.Values{
				"mode": {"mgodatagen"},
				"config": {`[
				{
					"collection": "collection",
					"count": 1000,
					"content": {
						"k": {
							"type": "string", 
							"minLength": 2,
							"maxLength": 5
						}
					}
				}
			]`}, "query": {`db.collection.aggregate([{"$project": {"_id": 0, "k": 0}}])`}},
			result: `[{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{}]
`,
			createdDB: 1,
		},
		// invalid aggregation query
		{
			params: url.Values{
				"mode": {"mgodatagen"},
				"config": {`[
				{
					"collection": "collection",
					"count": 10,
					"content": {
						"k": {
							"type": "string", 
							"minLength": 2,
							"maxLength": 5
						}
					}
				}
			]`}, "query": {`db.collection.aggregate([{"$project": {"_id": 0}])`}},
			result:    "Fail to parse aggregate() query: invalid character ']' after object key:value pair",
			createdDB: 0,
		},
		// aggregation query should be parsed correctly, but fail to run
		{
			params: url.Values{
				"mode": {"mgodatagen"},
				"config": {`[
				{
					"collection": "collection",
					"count": 10,
					"content": {
						"k": {
							"type": "string", 
							"minLength": 2,
							"maxLength": 5
						}
					}
				}
			]`}, "query": {`db.collection.aggregate([{"$project": "_id"}])`}},
			result:    "Aggregate query failed: $project specification must be an object",
			createdDB: 0,
		},
		// valid config, invalid query, valid json inside 'db.collection.find(...)' should create
		// a db, but fail to run the query
		{
			params: url.Values{
				"mode": {"mgodatagen"},
				"config": {`[
				{
					"collection": "collection",
					"count": 10,
					"content": {
						"k": {
							"type": "string", 
							"minLength": 1,
							"maxLength": 5
						}
					}
				}
			]`}, "query": {`db.collection.find({"$set": 12})`}},
			result:    "Find query failed: unknown top level operator: $set",
			createdDB: 1,
		},
		// valid config but invalid json in query
		{
			params: url.Values{
				"mode": {"mgodatagen"},
				"config": {`[
				{
					"collection": "collection",
					"count": 10,
					"content": {
						"k": {
							"type": "string", 
							"minLength": 1,
							"maxLength": 5
						}
					}
				}
			]`}, "query": {`db.collection.find({"k": "tJ")`}},
			result:    "Fail to parse find() query: unexpected EOF",
			createdDB: 0,
		},
		// two database, valid config
		{
			params: url.Values{
				"mode": {"mgodatagen"},
				"config": {`[
				{
					"collection": "coll1",
					"count": 10,
					"content": {
						"k": {
							"type": "string", 
							"minLength": 1,
							"maxLength": 5
						}
					}
				}, {
					"collection": "coll2",
					"count": 10,
					"content": {
						"k": {
							"type": "int", 
							"minInt": 1,
							"maxInt": 5
						}
					}
				}
			]`}, "query": {`db.coll2.find({"k": {"$gt": 3}})`}},
			result: `[{"_id":{"$oid":"5a934e000102030405000000"},"k":5},{"_id":{"$oid":"5a934e000102030405000001"},"k":5},{"_id":{"$oid":"5a934e000102030405000004"},"k":4},{"_id":{"$oid":"5a934e000102030405000007"},"k":5},{"_id":{"$oid":"5a934e000102030405000008"},"k":5},{"_id":{"$oid":"5a934e000102030405000009"},"k":4}]
`,
			createdDB: 2,
		},
		// two databases, one with invalid config (missing 'type')
		{
			params: url.Values{
				"mode": {"mgodatagen"},
				"config": {`[
				{
					"collection": "coll1",
					"count": 10,
					"content": {
						"k": {
							"type": "string", 
							"minLength": 1,
							"maxLength": 5
						}
					}
				}, {
					"collection": "coll2",
					"count": 10,
					"content": {
						"k": {
							"minInt": 1,
							"maxInt": 5
						}
					}
				}
			]`}, "query": {`db.coll2.find({"k": {"$gt": 3}})`}},
			result:    "fail to create DB: fail to create collection coll2: error while creating generators from configuration file:\n\tcause: invalid type  for field k",
			createdDB: 0,
		},
		// valid json
		{
			params: url.Values{
				"mode":   {"json"},
				"config": {`[{"k": 1}]`},
				"query":  {`db.collection.find()`},
			},
			result: `[{"_id":{"$oid":"5a934e000102030405000000"},"k":1}]
`,
			createdDB: 1,
		},
		// empty json
		{
			params: url.Values{
				"mode":   {"json"},
				"config": {`[{}]`},
				"query":  {`db.collection.find()`},
			},
			result: `[{"_id":{"$oid":"5a934e000102030405000000"}}]
`,
			createdDB: 1,
		},
		// invalid method
		{
			params: url.Values{
				"mode":   {"json"},
				"config": {`[{}]`},
				"query":  {`db.collection.findOne()`},
			},
			result:    "invalid method: findOne",
			createdDB: 0,
		},
		{
			params: url.Values{
				"mode":   {"json"},
				"config": {`[{}]`},
				"query":  {`find()`},
			},
			result:    "invalid query: \nmust match db.coll.find(...) or db.coll.aggregate(...)",
			createdDB: 0,
		},
		// json shoud be an array of documents
		{
			params: url.Values{
				"mode":   {"json"},
				"config": {`{"k": 1}, {"k": 2}`},
				"query":  {`db.collection.find()`},
			},
			result:    "fail to parse bson documents: json: cannot unmarshal object into Go value of type []bson.M",
			createdDB: 0,
		},
		// only work on 'collection' collection
		{
			params: url.Values{
				"mode":   {"json"},
				"config": {`[{"k": 1}, {"k": 2}]`},
				"query":  {`db.otherCollection.find()`},
			},
			result:    NoDocFound,
			createdDB: 1,
		},
		// doc with '_id' should not be overwritten
		{
			params: url.Values{
				"mode":   {"json"},
				"config": {`[{"_id": 1}, {"_id": 2}]`},
				"query":  {`db.collection.find()`},
			},
			result: `[{"_id":1},{"_id":2}]
`,
			createdDB: 1,
		},
		// mixed doc with / without '_id'
		{
			params: url.Values{
				"mode":   {"json"},
				"config": {`[{"_id": 1}, {}]`},
				"query":  {`db.collection.find()`},
			},
			result: `[{"_id":1},{"_id":{"$oid":"5a934e000102030405000001"}}]
`,
			createdDB: 1,
		},
		// duplicate ID err
		{
			params: url.Values{
				"mode":   {"json"},
				"config": {`[{"_id":1},{"_id":1}]`},
				"query":  {`db.collection.find()`},
			},
			result:    `E11000 duplicate key error collection: 57735364208e15b517d23e542088ed29.collection index: _id_ dup key: { : 1.0 }`,
			createdDB: 1,
		},
		// bson notation test input ObjectId
		{
			params: url.Values{
				"mode":   {"json"},
				"config": {`[{"_id": ObjectId("5a934e000102030405000001")},{"_id":1}]`},
				"query":  {`db.collection.find()`},
			},
			result: `[{"_id":{"$oid":"5a934e000102030405000001"}},{"_id":1}]
`,
			createdDB: 1,
		},
		// bson notation unkeyed query + objectId
		{
			params: url.Values{
				"mode":   {"json"},
				"config": {`[{"_id": ObjectId("5a934e000102030405000001")},{"_id":1}]`},
				"query":  {`db.collection.find({_id: ObjectId("5a934e000102030405000001")})`},
			},
			result: `[{"_id":{"$oid":"5a934e000102030405000001"}}]
`,
			createdDB: 0,
		},
		// aggregation + unkeyed
		{
			params: url.Values{
				"mode":   {"json"},
				"config": {`[{"_id": ObjectId("5a934e000102030405000001")},{"_id":1}]`},
				"query":  {`db.collection.aggregate([{$match: {_id: ObjectId("5a934e000102030405000001")}}])`},
			},
			result: `[{"_id":{"$oid":"5a934e000102030405000001"}}]
`,
			createdDB: 0,
		},
		// ISODate test
		{
			params: url.Values{
				"mode":   {"json"},
				"config": {`[{dt: ISODate("2000-01-01T00:00:00+00:00")}]`},
				"query":  {`db.collection.find()`},
			},
			result: `[{"_id":{"$oid":"5a934e000102030405000000"},"dt":{"$date":"2000-01-01T00:00:00Z"}}]
`,
			createdDB: 1,
		},
		// invalid objectId should not panic
		{
			params: url.Values{
				"mode":   {"json"},
				"config": {`[{"_id": ObjectId("5a9")}]`},
				"query":  {`db.collection.find({_id: ObjectId("5a934e000102030405000001")})`},
			},
			result:    `fail to parse bson documents: invalid ObjectId: 5a9`,
			createdDB: 0,
		},
		// TODO implement regex parsing
		{
			params: url.Values{
				"mode":   {"json"},
				"config": {`[{"k": "randompattern"}]`},
				"query":  {`db.collection.find({k: /pattern/})`},
			},
			result:    `Fail to parse find() query: invalid character '/' looking for beginning of value`,
			createdDB: 1,
		},
	}

	expectedDbNumber := 0
	for _, c := range l {
		r := assert.HTTPBody(srv.runHandler, http.MethodPost, "/run/", c.params)
		assert.Equal(t, c.result, r)
		expectedDbNumber += c.createdDB
	}

	dbNames, err := srv.session.DatabaseNames()
	assert.Nil(t, err)
	assert.Equal(t, expectedDbNumber, len(filterDBNames(dbNames)))

	assert.Equal(t, 0, srv.savedPageNb())

	err = srv.clearDatabases()
	assert.Nil(t, err)
}

func TestRunExistingDB(t *testing.T) {
	// the first /run/ request should create the database
	r := assert.HTTPBody(srv.runHandler, http.MethodPost, "/run/", templateParams)
	assert.Equal(t, templateResult, r)
	// the DBHash should be in the map
	DBHash := getDBHash(mgodatagenMode, []byte(templateParams.Get("config")))
	_, ok := srv.activeDB.Load(DBHash)
	assert.True(t, ok)
	//  the second /run/ should produce the same result
	r = assert.HTTPBody(srv.runHandler, http.MethodPost, "/run/", templateParams)
	assert.Equal(t, templateResult, r)

	dbNames, err := srv.session.DatabaseNames()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(filterDBNames(dbNames)))

	assert.Equal(t, 0, srv.savedPageNb())

	err = srv.clearDatabases()
	assert.Nil(t, err)
}

func TestSave(t *testing.T) {
	l := []struct {
		params    url.Values
		result    string
		newRecord bool
	}{
		// template params
		{
			params:    templateParams,
			result:    templateURL,
			newRecord: true,
		},
		// same config but different query should produce distinct url
		{
			params:    url.Values{"mode": {"mgodatagen"}, "config": {templateConfig}, "query": {"db.collection.find({\"k\": 10})"}},
			result:    "p/JExTWG6zk6K",
			newRecord: true,
		},
		// invalid config should be saved to
		{
			params:    url.Values{"mode": {"mgodatagen"}, "config": {`[{}]`}, "query": {templateQuery}},
			result:    "p/adv9VNjZGf-",
			newRecord: true,
		},
		// re-saving an existing playground should return same url
		{
			params:    templateParams,
			result:    templateURL,
			newRecord: false,
		},
		// different mode should create different url
		{
			params:    url.Values{"mode": {"json"}, "config": {`[{}]`}, "query": {templateQuery}},
			result:    "p/vkTDdT0z08q",
			newRecord: true,
		},
	}

	expectedRecordsNb := 0
	for _, c := range l {
		r := assert.HTTPBody(srv.saveHandler, http.MethodPost, "/save/", c.params)
		assert.Equal(t, c.result, r)
		if c.newRecord {
			expectedRecordsNb++
		}
	}
	assert.Equal(t, expectedRecordsNb, srv.savedPageNb())

	// /save/ should not try to generate samples
	dbNames, err := srv.session.DatabaseNames()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(filterDBNames(dbNames)))

	err = srv.clearDatabases()
	assert.Nil(t, err)
}

func TestView(t *testing.T) {
	r := assert.HTTPBody(srv.saveHandler, http.MethodPost, "/save/", templateParams)
	assert.Equal(t, templateURL, r)

	req, _ := http.NewRequest(http.MethodGet, "/"+templateURL, nil)
	resp := httptest.NewRecorder()
	srv.viewHandler(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)

	// save in json mode
	jsonParams := url.Values{
		"mode":   {"json"},
		"config": {`[{"_id": 1}]`},
		"query":  {templateQuery},
	}
	r = assert.HTTPBody(srv.saveHandler, http.MethodPost, "/save/", jsonParams)
	assert.Equal(t, "p/VxLxAzh9Uv9", r)

	req, _ = http.NewRequest(http.MethodGet, "/"+r, nil)
	resp = httptest.NewRecorder()
	srv.viewHandler(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)

	// if the page does not exists in storage, return 404 err
	req, _ = http.NewRequest(http.MethodGet, "/p/random", nil)
	resp = httptest.NewRecorder()
	srv.viewHandler(resp, req)
	assert.Equal(t, http.StatusNotFound, resp.Code)

	assert.Equal(t, 2, srv.savedPageNb())

	err := srv.clearDatabases()
	assert.Nil(t, err)
}

func TestBasePage(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	resp := httptest.NewRecorder()

	srv.newPageHandler(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestRemoveOldDB(t *testing.T) {
	params := url.Values{"mode": {"mgodatagen"}, "config": {templateConfig}, "query": {templateQuery}}
	r := assert.HTTPBody(srv.runHandler, http.MethodPost, "/run/", params)
	assert.Equal(t, templateResult, r)

	DBHash := getDBHash(mgodatagenMode, []byte(params.Get("config")))
	srv.activeDB.Store(DBHash, time.Now().Add(-cleanupInterval).Unix())
	// this DB should not be removed
	configFormat := `[{"collection": "collection%v","count": 10,"content": {}}]`
	params.Set("config", fmt.Sprintf(configFormat, "other"))
	r = assert.HTTPBody(srv.runHandler, http.MethodPost, "/run/", params)
	assert.Equal(t, NoDocFound, r)

	srv.removeExpiredDB()

	_, ok := srv.activeDB.Load(DBHash)
	assert.False(t, ok)
	dbNames, err := srv.session.DatabaseNames()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(filterDBNames(dbNames)))
	assert.NotEqual(t, dbNames[0], DBHash)

	err = srv.clearDatabases()
	assert.Nil(t, err)

}

func TestStaticHandlers(t *testing.T) {
	l := []struct {
		url        string
		statusCode int
	}{
		{
			url:        "/static/playground-min.js",
			statusCode: 200,
		},
		{
			url:        "/static/playground-min.css",
			statusCode: 200,
		},
		{
			url:        "/static/docs.html",
			statusCode: 200,
		},
		{
			url:        "/static/unknown.txt",
			statusCode: 404,
		},
		{
			url:        "/static/../README.md",
			statusCode: 404,
		},
	}
	for _, c := range l {
		resp := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", c.url, nil)
		srv.staticHandler(resp, req)
		assert.Equal(t, c.statusCode, resp.Code)
	}
}

func BenchmarkComputeID(b *testing.B) {

	config := []byte(`[
  {
    "collection": "collectionName",
    "count": 100,
    "content": {
      "k": {
        "type": "string",
        "minLength": 5,
        "maxLength": 5
      },
      "k2": {
        "type": "string",
        "maxLength": 4
      }
    }
  }
]`)
	query := []byte("db.collectionName.find()")
	mode := []byte("json")

	for n := 0; n < b.N; n++ {
		_ = computeID(mode, config, query)
	}

}

func BenchmarkNewPage(b *testing.B) {
	b.StopTimer()
	req, _ := http.NewRequest("GET", "/", nil)
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		resp := httptest.NewRecorder()
		srv.newPageHandler(resp, req)
	}
}

func BenchmarkView(b *testing.B) {
	b.StopTimer()
	assert.HTTPBody(srv.saveHandler, "POST", "/save/", templateParams)
	req, _ := http.NewRequest("GET", "/"+templateURL, nil)
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		resp := httptest.NewRecorder()
		srv.viewHandler(resp, req)
	}
}

func BenchmarkSave(b *testing.B) {
	b.StopTimer()
	configFormat := `[{"collection": "coll%v","count": 10,"content": {}}]`
	params := url.Values{"mode": {"mgodatagen"}, "query": {templateQuery}}
	b.StartTimer()

	for n := 0; n < b.N; n++ {
		resp := httptest.NewRecorder()
		params.Set("config", fmt.Sprintf(configFormat, n))
		req, _ := http.NewRequest("POST", "/save/", strings.NewReader(params.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		srv.saveHandler(resp, req)
	}
}

func BenchmarkRunExistingDB(b *testing.B) {
	b.StopTimer()
	req, _ := http.NewRequest("POST", "/run/", strings.NewReader(templateParams.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp := httptest.NewRecorder()
	srv.runHandler(resp, req)
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		resp = httptest.NewRecorder()
		srv.runHandler(resp, req)
	}
}

func BenchmarkRunNonExistingDB(b *testing.B) {
	b.StopTimer()

	configFormat := `[{"collection": "collection","count": 10,"content": {"k": {"type": "int", "minInt": 0, "maxInt": %d}}}]`
	params := url.Values{"mode": {"mgodatagen"}, "query": {templateQuery}}
	b.StartTimer()

	for n := 0; n < b.N; n++ {
		resp := httptest.NewRecorder()
		params.Set("config", fmt.Sprintf(configFormat, n))
		req, _ := http.NewRequest("POST", "/run/", strings.NewReader(params.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		srv.runHandler(resp, req)
	}
}

func BenchmarkServeStaticFile(b *testing.B) {
	for n := 0; n < b.N; n++ {
		resp := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/static/docs.html", nil)
		srv.staticHandler(resp, req)
	}
}
