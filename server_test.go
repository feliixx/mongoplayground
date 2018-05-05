package main

import (
	"bytes"
	"compress/gzip"
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
	"time"

	"github.com/dgraph-io/badger"
	"github.com/globalsign/mgo/bson"
)

const (
	templateResult = `[{"_id":ObjectId("5a934e000102030405000000"),"k":10},{"_id":ObjectId("5a934e000102030405000001"),"k":2},{"_id":ObjectId("5a934e000102030405000002"),"k":7},{"_id":ObjectId("5a934e000102030405000003"),"k":6},{"_id":ObjectId("5a934e000102030405000004"),"k":9},{"_id":ObjectId("5a934e000102030405000005"),"k":10},{"_id":ObjectId("5a934e000102030405000006"),"k":9},{"_id":ObjectId("5a934e000102030405000007"),"k":10},{"_id":ObjectId("5a934e000102030405000008"),"k":2},{"_id":ObjectId("5a934e000102030405000009"),"k":1}]`
	templateURL    = "p/snbIQ3uGHGq"
)

var (
	templateParams = url.Values{"mode": {"mgodatagen"}, "config": {templateConfig}, "query": {templateQuery}}
	testServer     *server
)

func TestMain(m *testing.M) {
	err := os.RemoveAll(badgerDir)
	if err != nil {
		fmt.Printf("aborting: %v\n", err)
		os.Exit(1)
	}
	log := log.New(ioutil.Discard, "", 0)
	s, err := newServer(log)
	if err != nil {
		fmt.Printf("aborting: %v\n", err)
		os.Exit(1)
	}
	testServer = s
	defer s.session.Close()
	defer s.storage.Close()

	retCode := m.Run()
	os.Exit(retCode)
}

func TestServeHTTP(t *testing.T) {

	t.Parallel()

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	resp := httptest.NewRecorder()
	testServer.ServeHTTP(resp, req)
	if http.StatusOK != resp.Code {
		t.Errorf("expected code %d but got %d", http.StatusOK, resp.Code)
	}
}

func TestRunCreateDB(t *testing.T) {

	testServer.clearDatabases(t)

	runCreateDBTests := []struct {
		name      string
		params    url.Values
		result    string
		createdDB int
		compact   bool
	}{
		{
			name:      "incorrect config",
			params:    url.Values{"mode": {"mgodatagen"}, "config": {"h"}, "query": {"h"}},
			result:    "fail to parse configuration: Error in configuration file: object / array / Date badly formatted: \n\n\t\tinvalid character 'h' looking for beginning of value",
			createdDB: 0,
			compact:   false,
		},
		{
			name:      "non existing collection",
			params:    url.Values{"mode": {"mgodatagen"}, "config": {templateConfig}, "query": {"db.c.find()"}},
			result:    noDocFound,
			createdDB: 1,
			compact:   false,
		},
		{
			name:      "deterministic list of objectId",
			params:    templateParams,
			result:    templateResult,
			createdDB: 0, // db already exists
			compact:   true,
		},
		{
			name: "deterministic results with generators",
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
			result:    `[{"_id":ObjectId("5a934e000102030405000000"),"k":"1jU"},{"_id":ObjectId("5a934e000102030405000001"),"k":"tBRWL"},{"_id":ObjectId("5a934e000102030405000002"),"k":"6Hch"},{"_id":ObjectId("5a934e000102030405000003"),"k":"ZWHW"},{"_id":ObjectId("5a934e000102030405000004"),"k":"RkMG"},{"_id":ObjectId("5a934e000102030405000005"),"k":"RIr"},{"_id":ObjectId("5a934e000102030405000006"),"k":"ru7"},{"_id":ObjectId("5a934e000102030405000007"),"k":"OB"},{"_id":ObjectId("5a934e000102030405000008"),"k":"ja"},{"_id":ObjectId("5a934e000102030405000009"),"k":"K307"}]`,
			createdDB: 1,
			compact:   true,
		},
		// same config, but aggregation
		{
			name: "basic aggregation",
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
			result:    `[{"k":"1jU"},{"k":"tBRWL"},{"k":"6Hch"},{"k":"ZWHW"},{"k":"RkMG"},{"k":"RIr"},{"k":"ru7"},{"k":"OB"},{"k":"ja"},{"k":"K307"}]`,
			createdDB: 0,
			compact:   true,
		},
		{
			name: "doc nb > 100",
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
			result:    `[{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{},{}]`,
			createdDB: 1,
			compact:   true,
		},
		{
			name: "invalid aggregation query (invalid json)",
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
			result:    "fail to parse content of query: invalid character ']' after object key:value pair",
			createdDB: 0,
			compact:   false,
		},
		{
			name: "invalid aggregation query (invalid syntax)",
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
			result:    "query failed: $project specification must be an object",
			createdDB: 0,
			compact:   false,
		},
		{
			name: "invalid find query (invalid syntax)",
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
			result:    "query failed: unknown top level operator: $set",
			createdDB: 1,
			compact:   false,
		},
		{
			name: "invalid find query (invalid json)",
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
			result:    "fail to parse content of query: invalid character ']' after object key:value pair",
			createdDB: 0,
			compact:   false,
		},
		{
			name: "two databases",
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
			result:    `[{"_id":ObjectId("5a934e000102030405000000"),"k":5},{"_id":ObjectId("5a934e000102030405000001"),"k":5},{"_id":ObjectId("5a934e000102030405000004"),"k":4},{"_id":ObjectId("5a934e000102030405000007"),"k":5},{"_id":ObjectId("5a934e000102030405000008"),"k":5},{"_id":ObjectId("5a934e000102030405000009"),"k":4}]`,
			createdDB: 2,
			compact:   true,
		},
		{
			name: "two databases invalid config",
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
			result:    "fail to create DB: fail to create collection coll2: fail to create DocumentGenerator:\n\tcause: invalid type  for field k",
			createdDB: 0,
			compact:   false,
		},
		{
			name: "basic json mode",
			params: url.Values{
				"mode":   {"json"},
				"config": {`[{"k": 1}]`},
				"query":  {`db.collection.find()`},
			},
			result:    `[{"_id":ObjectId("5a934e000102030405000000"),"k":1}]`,
			createdDB: 1,
			compact:   true,
		},
		{
			name: "empty json",
			params: url.Values{
				"mode":   {"json"},
				"config": {`[{}]`},
				"query":  {`db.collection.find()`},
			},
			result:    `[{"_id":ObjectId("5a934e000102030405000000")}]`,
			createdDB: 1,
			compact:   true,
		},
		{
			name: "invalid method",
			params: url.Values{
				"mode":   {"json"},
				"config": {`[{}]`},
				"query":  {`db.collection.findOne()`},
			},
			result:    "query failed: invalid method: findOne",
			createdDB: 0,
			compact:   false,
		},
		{
			name: "invalid query syntax",
			params: url.Values{
				"mode":   {"json"},
				"config": {`[{}]`},
				"query":  {`find()`},
			},
			result:    "invalid query: \nmust match db.coll.find(...) or db.coll.aggregate(...)",
			createdDB: 0,
			compact:   false,
		},
		{
			name: "require array of json documents",
			params: url.Values{
				"mode":   {"json"},
				"config": {`{"k": 1}, {"k": 2}`},
				"query":  {`db.collection.find()`},
			},
			result:    "fail to parse bson documents: json: cannot unmarshal object into Go value of type []bson.M",
			createdDB: 0,
			compact:   false,
		},
		{
			name: `json create only collection "collection"`,
			params: url.Values{
				"mode":   {"json"},
				"config": {`[{"k": 1}, {"k": 2}]`},
				"query":  {`db.otherCollection.find()`},
			},
			result:    noDocFound,
			createdDB: 1,
			compact:   false,
		},
		{
			name: `doc with "_id" should not be overwritten`,
			params: url.Values{
				"mode":   {"json"},
				"config": {`[{"_id": 1}, {"_id": 2}]`},
				"query":  {`db.collection.find()`},
			},
			result:    `[{"_id":1},{"_id":2}]`,
			createdDB: 1,
			compact:   true,
		},
		{
			name: `mixed doc with/without "_id"`,
			params: url.Values{
				"mode":   {"json"},
				"config": {`[{"_id": 1}, {}]`},
				"query":  {`db.collection.find()`},
			},
			result:    `[{"_id":1},{"_id":ObjectId("5a934e000102030405000001")}]`,
			createdDB: 1,
			compact:   true,
		},
		{
			name: `duplicate "_id" error`,
			params: url.Values{
				"mode":   {"json"},
				"config": {`[{"_id":1},{"_id":1}]`},
				"query":  {`db.collection.find()`},
			},
			result:    `E11000 duplicate key error collection: 57735364208e15b517d23e542088ed29.collection index: _id_ dup key: { : 1.0 }`,
			createdDB: 1,
			compact:   false,
		},
		{
			name: `bson "ObjectId" notation`,
			params: url.Values{
				"mode":   {"json"},
				"config": {`[{"_id": ObjectId("5a934e000102030405000001")},{"_id":1}]`},
				"query":  {`db.collection.find()`},
			},
			result:    `[{"_id":ObjectId("5a934e000102030405000001")},{"_id":1}]`,
			createdDB: 1,
			compact:   true,
		},
		{
			name: `bson unkeyed notation`,
			params: url.Values{
				"mode":   {"json"},
				"config": {`[{"_id": ObjectId("5a934e000102030405000001")},{"_id":1}]`},
				"query":  {`db.collection.find({_id: ObjectId("5a934e000102030405000001")})`},
			},
			result:    `[{"_id":ObjectId("5a934e000102030405000001")}]`,
			createdDB: 0,
			compact:   true,
		},
		{
			name: `unkeyed params in aggreagtion`,
			params: url.Values{
				"mode":   {"json"},
				"config": {`[{"_id": ObjectId("5a934e000102030405000001")},{"_id":1}]`},
				"query":  {`db.collection.aggregate([{$match: {_id: ObjectId("5a934e000102030405000001")}}])`},
			},
			result:    `[{"_id":ObjectId("5a934e000102030405000001")}]`,
			createdDB: 0,
			compact:   true,
		},
		{
			name: `doc with bson "ISODate"`,
			params: url.Values{
				"mode":   {"json"},
				"config": {`[{dt: ISODate("2000-01-01T00:00:00+00:00")}]`},
				"query":  {`db.collection.find()`},
			},
			result:    `[{"_id":ObjectId("5a934e000102030405000000"),"dt":ISODate("2000-01-01T00:00:00Z")}]`,
			createdDB: 1,
			compact:   true,
		},
		{
			name: `invalid "ObjectId" should not panic`,
			params: url.Values{
				"mode":   {"json"},
				"config": {`[{"_id": ObjectId("5a9")}]`},
				"query":  {`db.collection.find({_id: ObjectId("5a934e000102030405000001")})`},
			},
			result:    `fail to parse bson documents: invalid input to ObjectIdHex: "5a9"`,
			createdDB: 0,
			compact:   false,
		},
		{
			name: `regex parsing`, // TODO
			params: url.Values{
				"mode":   {"json"},
				"config": {`[{"k": "randompattern"}]`},
				"query":  {`db.collection.find({k: /pattern/})`},
			},
			result:    `fail to parse content of query: invalid character '/' looking for beginning of value`,
			createdDB: 1,
			compact:   false,
		},
		{
			name: `query with projection`,
			params: url.Values{
				"mode":   {"json"},
				"config": {`[{"k":1},{"k":2},{"k":3}]`},
				"query":  {`db.collection.find({}, {"_id": 0})`},
			},
			result:    `[{"k":1},{"k":2},{"k":3}]`,
			createdDB: 1,
			compact:   true,
		},
	}

	nbMongoDatabases := 0
	for _, tt := range runCreateDBTests {
		t.Run(tt.name, func(t *testing.T) {
			buf := httpBody(t, testServer.runHandler, http.MethodPost, "/run", tt.params)
			if tt.compact {
				comp, err := bson.CompactJSON(buf.Bytes())
				if err != nil {
					t.Errorf("fail to compact JSON: %v", err)
				}
				buf = bytes.NewBuffer(comp)
			}

			if want, got := tt.result, buf.String(); want != got {
				t.Errorf("expected '%s' but got '%s'", want, got)
			}
		})
		nbMongoDatabases += tt.createdDB
	}

	testStorageContent(t, nbMongoDatabases, 0)

}

func TestRunExistingDB(t *testing.T) {

	testServer.clearDatabases(t)

	// the first /run request should create the database
	buf := httpBody(t, testServer.runHandler, http.MethodPost, "/run", templateParams)
	comp, err := bson.CompactJSON(buf.Bytes())
	if err != nil {
		t.Error(err)
	}
	if want, got := templateResult, string(comp); want != got {
		t.Errorf("expected %s but got %s", want, got)
	}
	p := &page{
		Mode:   mgodatagenMode,
		Config: []byte(templateParams.Get("config")),
	}
	DBHash := p.dbHash()
	_, ok := testServer.activeDB.Load(DBHash)
	if !ok {
		t.Errorf("activeDb should contain DB %s", DBHash)
	}

	//  the second /run should produce the same result
	buf = httpBody(t, testServer.runHandler, http.MethodPost, "/run", templateParams)
	comp, err = bson.CompactJSON(buf.Bytes())
	if err != nil {
		t.Error(err)
	}
	if want, got := templateResult, string(comp); want != got {
		t.Errorf("expected %s but got %s", want, got)
	}

	testStorageContent(t, 1, 0)

}

func TestSave(t *testing.T) {

	testServer.clearDatabases(t)

	saveTests := []struct {
		name      string
		params    url.Values
		result    string
		newRecord bool
	}{
		{
			name:      "template config",
			params:    templateParams,
			result:    templateURL,
			newRecord: true,
		},
		{
			name:      "template config with new query",
			params:    url.Values{"mode": {"mgodatagen"}, "config": {templateConfig}, "query": {"db.collection.find({\"k\": 10})"}},
			result:    "p/DYlGRQeO0bX",
			newRecord: true,
		},
		{
			name:      "invalid config",
			params:    url.Values{"mode": {"mgodatagen"}, "config": {`[{}]`}, "query": {templateQuery}},
			result:    "p/EMmfQADkGcq",
			newRecord: true,
		},
		{
			name:      "save existing playground",
			params:    templateParams,
			result:    templateURL,
			newRecord: false,
		},
		{
			name:      "template query with new config",
			params:    url.Values{"mode": {"json"}, "config": {`[{}]`}, "query": {templateQuery}},
			result:    "p/4cOeA7NGLru",
			newRecord: true,
		},
	}

	nbBadgerRecords := 0
	for _, tt := range saveTests {
		t.Run(tt.name, func(t *testing.T) {
			buf := httpBody(t, testServer.saveHandler, http.MethodPost, "/save", tt.params)

			if want, got := tt.result, buf.String(); want != got {
				t.Errorf("expected %s, but got %s", want, got)
			}
		})
		if tt.newRecord {
			nbBadgerRecords++
		}
	}

	testStorageContent(t, 0, nbBadgerRecords)

}

func TestView(t *testing.T) {

	testServer.clearDatabases(t)

	viewTests := []struct {
		name         string
		params       url.Values
		url          string
		responseCode int
		newRecord    bool
	}{
		{
			name:         "template parameters",
			params:       templateParams,
			url:          templateURL,
			responseCode: http.StatusOK,
			newRecord:    true,
		},
		{
			name: "new config",
			params: url.Values{
				"mode":   {"json"},
				"config": {`[{"_id": 1}]`},
				"query":  {templateQuery},
			},
			url:          "p/DEz-pkpheLX",
			responseCode: http.StatusOK,
			newRecord:    true,
		},
		{
			name:         "non existing url",
			params:       templateParams,
			url:          "p/random",
			responseCode: http.StatusNotFound,
			newRecord:    false,
		},
	}

	nbBadgerRecords := 0
	for _, tt := range viewTests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.responseCode == http.StatusOK {
				buf := httpBody(t, testServer.saveHandler, http.MethodPost, "/save", tt.params)

				if want, got := tt.url, buf.String(); want != got {
					t.Errorf("expected %s but got %s", want, got)
				}
			}
			req, _ := http.NewRequest(http.MethodGet, "/"+tt.url, nil)
			resp := httptest.NewRecorder()
			testServer.viewHandler(resp, req)

			if tt.responseCode != resp.Code {
				t.Errorf("expected response code %d but got %d", tt.responseCode, resp.Code)
			}
		})
		if tt.newRecord {
			nbBadgerRecords++
		}
	}

	testStorageContent(t, 0, nbBadgerRecords)

}

func TestBasePage(t *testing.T) {

	t.Parallel()

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
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

	testServer.clearDatabases(t)

	params := url.Values{"mode": {"mgodatagen"}, "config": {templateConfig}, "query": {templateQuery}}
	buf := httpBody(t, testServer.runHandler, http.MethodPost, "/run", params)
	comp, err := bson.CompactJSON(buf.Bytes())
	if err != nil {
		t.Error(err)
	}

	if want, got := templateResult, string(comp); want != got {
		t.Errorf("expected %s but got %s", want, got)
	}

	p := &page{
		Mode:   mgodatagenMode,
		Config: []byte(params.Get("config")),
	}
	testServer.logger.Print(p.String())
	DBHash := p.dbHash()
	testServer.activeDB.Store(DBHash, time.Now().Add(-cleanupInterval).Unix())
	// this DB should not be removed
	configFormat := `[{"collection": "collection%v","count": 10,"content": {}}]`
	params.Set("config", fmt.Sprintf(configFormat, "other"))
	buf = httpBody(t, testServer.runHandler, http.MethodPost, "/run", params)

	if want, got := buf.String(), noDocFound; want != got {
		t.Errorf("expected %s but got %s", want, got)
	}

	testServer.removeExpiredDB()

	_, ok := testServer.activeDB.Load(DBHash)
	if ok {
		t.Errorf("DB %s should not be present in activeDB", DBHash)
	}

	dbNames, err := testServer.session.DatabaseNames()
	if err != nil {
		t.Error(err)
	}
	if dbNames[0] == DBHash {
		t.Errorf("%s should have been removed from mongodb", DBHash)
	}

	testStorageContent(t, 1, 0)

}

func TestStaticHandlers(t *testing.T) {

	staticFileTests := []struct {
		name         string
		url          string
		contentType  string
		responseCode int
	}{
		{
			name:         "css",
			url:          "/static/playground-min-3.css",
			contentType:  "text/css; charset=utf-8",
			responseCode: 200,
		},
		{
			name:         "documentation",
			url:          "/static/docs-3.html",
			contentType:  "text/html; charset=utf-8",
			responseCode: 200,
		},
		{
			name:         "documentation",
			url:          "/static/playground-min-3.js",
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

			t.Parallel()

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

func BenchmarkNewPage(b *testing.B) {
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		resp := httptest.NewRecorder()
		testServer.newPageHandler(resp, req)
	}
}

func BenchmarkView(b *testing.B) {
	req, _ := http.NewRequest(http.MethodPost, "/save", strings.NewReader(templateParams.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp := httptest.NewRecorder()
	testServer.saveHandler(resp, req)
	req, _ = http.NewRequest(http.MethodGet, "/"+templateURL, nil)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		resp := httptest.NewRecorder()
		testServer.viewHandler(resp, req)
	}
}

func BenchmarkSave(b *testing.B) {
	configFormat := `[{"collection": "coll%v","count": 10,"content": {}}]`
	params := url.Values{"mode": {"mgodatagen"}, "query": {templateQuery}}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		resp := httptest.NewRecorder()
		params.Set("config", fmt.Sprintf(configFormat, n))
		req, _ := http.NewRequest(http.MethodPost, "/save", strings.NewReader(params.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		testServer.saveHandler(resp, req)
	}
}

func BenchmarkRunExistingDB(b *testing.B) {
	req, _ := http.NewRequest(http.MethodPost, "/run", strings.NewReader(templateParams.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp := httptest.NewRecorder()
	testServer.runHandler(resp, req)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		resp = httptest.NewRecorder()
		testServer.runHandler(resp, req)
	}
}

func BenchmarkRunNonExistingDB(b *testing.B) {
	configFormat := `[{"collection": "collection","count": 10,"content": {"k": {"type": "int", "minInt": 0, "maxInt": %d}}}]`
	params := url.Values{"mode": {"mgodatagen"}, "query": {templateQuery}}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		resp := httptest.NewRecorder()
		params.Set("config", fmt.Sprintf(configFormat, n))
		req, _ := http.NewRequest(http.MethodPost, "/run", strings.NewReader(params.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		testServer.runHandler(resp, req)
	}
}

func BenchmarkServeStaticFile(b *testing.B) {
	for n := 0; n < b.N; n++ {
		resp := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/static/docs-3.html", nil)
		testServer.staticHandler(resp, req)
	}
}

func (s *server) clearDatabases(t *testing.T) error {
	dbNames, err := s.session.DatabaseNames()
	if err != nil {
		t.Error(err)
	}
	// dbNames are md5 hash, 32-char long
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
		t.Error(err)
	}

	deleteTxn := s.storage.NewTransaction(true)
	for i := 0; i < len(keys); i++ {
		err = deleteTxn.Delete(keys[i])
		if err != nil {
			t.Error(err)
		}
	}
	return deleteTxn.Commit(func(err error) {
		if err != nil {
			fmt.Printf("fail to delete: %v\n", err)
		}
	})
}

func testStorageContent(t *testing.T, nbMongoDatabases, nbBadgerRecords int) {
	dbNames, err := testServer.session.DatabaseNames()
	if err != nil {
		t.Error(err)
	}
	if want, got := nbMongoDatabases, len(filterDBNames(dbNames)); want != got {
		t.Errorf("expected %d DB, but got %d", want, got)
	}
	if want, got := nbBadgerRecords, testServer.savedPageNb(); want != got {
		t.Errorf("expected %d page saved, but got %d", want, got)
	}
}

// return only created db, and get rid of 'indexes', 'local' ect
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
