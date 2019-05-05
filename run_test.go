package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/globalsign/mgo/bson"
)

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
			result:    "error in configuration:\n  Error in configuration file: object / array / Date badly formatted: \n\n\t\tinvalid character 'h' looking for beginning of value",
			createdDB: 0,
			compact:   false,
		},
		{
			name:      "non existing collection",
			params:    url.Values{"mode": {"mgodatagen"}, "config": {templateConfig}, "query": {"db.c.find()"}},
			result:    `collection "c" doesn't exist`,
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
			result:    `[{"_id":ObjectId("5a934e00010203040500000a"),"k":5},{"_id":ObjectId("5a934e00010203040500000b"),"k":5},{"_id":ObjectId("5a934e00010203040500000e"),"k":4},{"_id":ObjectId("5a934e000102030405000011"),"k":5},{"_id":ObjectId("5a934e000102030405000012"),"k":5},{"_id":ObjectId("5a934e000102030405000013"),"k":4}]`,
			createdDB: 1,
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
			result:    "error in configuration:\n  fail to create collection coll2: fail to create DocumentGenerator:\n\tcause: invalid type  for field k",
			createdDB: 0,
			compact:   false,
		},
		{
			name: "basic json mode",
			params: url.Values{
				"mode":   {"bson"},
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
				"mode":   {"bson"},
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
				"mode":   {"bson"},
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
				"mode":   {"bson"},
				"config": {`[{}]`},
				"query":  {`find()`},
			},
			result:    "invalid query: \nmust match db.coll.find(...) or db.coll.aggregate(...)",
			createdDB: 0,
			compact:   false,
		},
		{
			name: "require array of bson documents or a single document",
			params: url.Values{
				"mode":   {"bson"},
				"config": {`{"k": 1}, {"k": 2}`},
				"query":  {`db.collection.find()`},
			},
			result:    fmt.Sprintf("error in configuration:\n  %v", invalidConfig),
			createdDB: 0,
			compact:   false,
		},
		{
			name: "multiple collection in bson mode",
			params: url.Values{
				"mode":   {"bson"},
				"config": {`db={"collection1":[{"_id":1,"k":8}],"collection2":[{"_id":1,"k2":10}]}`},
				"query":  {`db.collection1.find()`},
			},
			result:    `[{"_id":1,"k":8}]`,
			createdDB: 1,
			compact:   true,
		},
		{
			name: "multiple collection in json mode without _id",
			params: url.Values{
				"mode":   {"bson"},
				"config": {`db={"collection1":[{"k":8}],"collection2":[{"k2":8},{"k2":8}]}`},
				"query":  {`db.collection1.aggregate({"$lookup":{"from":"collection2","localField":"k",foreignField:"k2","as":"lookupDoc"}})`},
			},
			result:    `[{"_id":ObjectId("5a934e000102030405000000"),"k":8,"lookupDoc":[{"_id":ObjectId("5a934e000102030405000001"),"k2":8},{"_id":ObjectId("5a934e000102030405000002"),"k2":8}]}]`,
			createdDB: 1,
			compact:   true,
		},
		{
			name: "multiple collection in bson mode with lookup",
			params: url.Values{
				"mode":   {"bson"},
				"config": {`db={"collection1":[{"_id":1,"k":8}],"collection2":[{"_id":1,"k2":1}]}`},
				"query":  {`db.collection1.aggregate({"$lookup":{"from":"collection2","localField":"_id",foreignField:"_id","as":"lookupDoc"}})`},
			},
			result:    `[{"_id":1,"k":8,"lookupDoc":[{"_id":1,"k2":1}]}]`,
			createdDB: 1,
			compact:   true,
		},
		{
			name: `bson old syntax create only collection "collection"`,
			params: url.Values{
				"mode":   {"bson"},
				"config": {`[{"k": 1}, {"k": 2}]`},
				"query":  {`db.otherCollection.find()`},
			},
			result:    `collection "otherCollection" doesn't exist`,
			createdDB: 1,
			compact:   false,
		},
		{
			name: `doc with "_id" should not be overwritten`,
			params: url.Values{
				"mode":   {"bson"},
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
				"mode":   {"bson"},
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
				"mode":   {"bson"},
				"config": {`[{"_id":1},{"_id":1}]`},
				"query":  {`db.collection.find()`},
			},
			result:    `E11000 duplicate key error collection: 57735364208e15b517d23e542088ed29.collection index: _id_ dup key: { : 1.0 }`,
			createdDB: 0, // the config is incorrect, no db should be created
			compact:   false,
		},
		{
			name: `bson "ObjectId" notation`,
			params: url.Values{
				"mode":   {"bson"},
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
				"mode":   {"bson"},
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
				"mode":   {"bson"},
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
				"mode":   {"bson"},
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
				"mode":   {"bson"},
				"config": {`[{"_id": ObjectId("5a9")}]`},
				"query":  {`db.collection.find({_id: ObjectId("5a934e000102030405000001")})`},
			},
			result:    "error in configuration:\n  invalid input to ObjectIdHex: \"5a9\"",
			createdDB: 0,
			compact:   false,
		},
		{
			name: `regex parsing`, // TODO
			params: url.Values{
				"mode":   {"bson"},
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
				"mode":   {"bson"},
				"config": {`[{"k":1},{"k":2},{"k":3}]`},
				"query":  {`db.collection.find({}, {"_id": 0})`},
			},
			result:    `[{"k":1},{"k":2},{"k":3}]`,
			createdDB: 1,
			compact:   true,
		},
		{
			name: `empty config`,
			params: url.Values{
				"mode":   {"bson"},
				"config": {""},
				"query":  {"db.c.find()"},
			},
			result:    fmt.Sprintf("error in configuration:\n  %v", invalidConfig),
			createdDB: 0,
			compact:   false,
		},
		{
			name: `too many collections`,
			params: url.Values{
				"mode":   {"bson"},
				"config": {`db={"a":[],"b":[],"c":[],"d":[],"e":[],"f":[],"g":[],"h":[],"i":[],"j":[],"k":[]}`},
				"query":  {"db.c.find()"},
			},
			result:    "max number of collection in a database is 10, but was 11",
			createdDB: 0,
			compact:   false,
		},
		{
			name: `no documents found`,
			params: url.Values{
				"mode":   {"bson"},
				"config": {`db={"a":[]}`},
				"query":  {"db.a.find()"},
			},
			result:    noDocFound,
			createdDB: 1,
			compact:   false,
		},
	}

	nbMongoDatabases := 0

	t.Run("parallel run", func(t *testing.T) {
		for _, tt := range runCreateDBTests {

			test := tt // capture range variable
			t.Run(test.name, func(t *testing.T) {

				t.Parallel()

				buf := httpBody(t, testServer.runHandler, http.MethodPost, "/run", test.params)
				if test.compact {
					comp, err := bson.CompactJSON(buf.Bytes())
					if err != nil {
						t.Errorf("could not compact result: %s (%v)", buf.Bytes(), err)
					}
					buf = bytes.NewBuffer(comp)
				}

				if want, got := test.result, buf.String(); want != got {
					t.Errorf("expected\n '%s'\n but got\n '%s'", want, got)
				}
			})
			nbMongoDatabases += test.createdDB
		}
	})

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
	_, ok := testServer.activeDB[DBHash]
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

func TestConsistentError(t *testing.T) {

	testServer.clearDatabases(t)

	params := url.Values{"mode": {"mgodatagen"}, "config": {`[{"k":1}]`}, "query": {templateQuery}}
	buf := httpBody(t, testServer.runHandler, http.MethodPost, "/run", params)

	errorMsg := "error in configuration:\n  Error in configuration file: \n\t'collection' and 'database' fields can't be empty"

	if want, got := errorMsg, buf.String(); want != got {
		t.Errorf("expected %s but got %s", want, got)
	}

	buf = httpBody(t, testServer.runHandler, http.MethodPost, "/run", params)

	if want, got := errorMsg, buf.String(); want != got {
		t.Errorf("expected %s but got %s", want, got)
	}
}
