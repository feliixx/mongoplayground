package main

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
)

func TestRunCreateDB(t *testing.T) {

	defer testServer.clearDatabases(t)

	runCreateDBTests := []struct {
		name      string
		params    url.Values
		result    string
		createdDB int
	}{
		{
			name:      "incorrect config",
			params:    url.Values{"mode": {"mgodatagen"}, "config": {"h"}, "query": {"h"}},
			result:    "error in configuration:\n  Error in configuration file: object / array / Date badly formatted: \n\n\t\tinvalid character 'h' looking for beginning of value",
			createdDB: 0,
		},
		{
			name:      "non existing collection",
			params:    url.Values{"mode": {"mgodatagen"}, "config": {templateConfig}, "query": {"db.c.find()"}},
			result:    `collection "c" doesn't exist`,
			createdDB: 1,
		},
		{
			name:      "deterministic list of objectId",
			params:    templateParams,
			result:    templateResult,
			createdDB: 0, // db already exists
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
			result:    "error in query:\n  fail to parse content of query: invalid character ']' after object key:value pair",
			createdDB: 0,
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
			result:    "query failed: (Location15969) $project specification must be an object",
			createdDB: 0,
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
			result:    "query failed: (BadValue) unknown top level operator: $set",
			createdDB: 1,
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
			result:    "error in query:\n  fail to parse content of query: invalid character ']' after object key:value pair",
			createdDB: 0,
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
			result:    "error in configuration:\n  fail to create collection coll2: fail to create DocumentGenerator:\n\tcause: for field k, invalid type ",
			createdDB: 0,
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
		},
		{
			name: "invalid query syntax",
			params: url.Values{
				"mode":   {"bson"},
				"config": {`[{}]`},
				"query":  {`find()`},
			},
			result:    fmt.Sprintf("error in query:\n  %v", errInvalidQuery),
			createdDB: 0,
		},
		{
			name: "require array of bson documents or a single document",
			params: url.Values{
				"mode":   {"bson"},
				"config": {`{"k": 1}, {"k": 2}`},
				"query":  {`db.collection.find()`},
			},
			result:    fmt.Sprintf("error in configuration:\n  %v", errInvalidConfig),
			createdDB: 0,
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
		},
		{
			name: `duplicate "_id" error`,
			params: url.Values{
				"mode":   {"bson"},
				"config": {`[{"_id":1},{"_id":1}]`},
				"query":  {`db.collection.find()`},
			},
			result:    "error in configuration:\n  bulk write error: [{[{E11000 duplicate key error collection: 57735364208e15b517d23e542088ed29.collection index: _id_ dup key: { : 1.0 }}]}, {<nil>}]",
			createdDB: 0, // the config is incorrect, no db should be created
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
		},
		{
			name: `invalid "ObjectId" should not panic`,
			params: url.Values{
				"mode":   {"bson"},
				"config": {`[{"_id": ObjectId("5a9")}]`},
				"query":  {`db.collection.find({_id: ObjectId("5a934e000102030405000001")})`},
			},
			result:    "error in configuration:\n  encoding/hex: odd length hex string",
			createdDB: 0,
		},
		{
			name: `regex parsing`, // TODO
			params: url.Values{
				"mode":   {"bson"},
				"config": {`[{"k": "randompattern"}]`},
				"query":  {`db.collection.find({k: /pattern/})`},
			},
			result:    "error in query:\n  fail to parse content of query: invalid character '/' looking for beginning of value",
			createdDB: 1,
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
		},
		{
			name: `empty config`,
			params: url.Values{
				"mode":   {"bson"},
				"config": {""},
				"query":  {"db.c.find()"},
			},
			result:    fmt.Sprintf("error in configuration:\n  %v", errInvalidConfig),
			createdDB: 0,
		},
		{
			name: `empty query`,
			params: url.Values{
				"mode":   {"bson"},
				"config": {templateConfig},
				"query":  {""},
			},
			result:    fmt.Sprintf("error in query:\n  %v", errInvalidQuery),
			createdDB: 1,
		},
		{
			name: `too many collections`,
			params: url.Values{
				"mode":   {"bson"},
				"config": {`db={"a":[],"b":[],"c":[],"d":[],"e":[],"f":[],"g":[],"h":[],"i":[],"j":[],"k":[]}`},
				"query":  {"db.c.find()"},
			},
			result:    "error in configuration:\n  max number of collection in a database is 10, but was 11",
			createdDB: 0,
		},
		{
			name: `no documents found`,
			params: url.Values{
				"mode":   {"bson"},
				"config": {`db={"a":[]}`},
				"query":  {"db.a.find()"},
			},
			result:    noDocFound,
			createdDB: 0,
		},
		{
			name: `invalid query with 3 '.'`,
			params: url.Values{
				"mode":   {"bson"},
				"config": {templateConfig},
				"query":  {`[{"key.path.test":{"$match":10}}])`},
			},
			result:    fmt.Sprintf("error in query:\n  %v", errInvalidQuery),
			createdDB: 0,
		},
		{
			name: `playground too big`,
			params: url.Values{
				"mode":   {"bson"},
				"config": {string(make([]byte, maxByteSize))},
				"query":  {"db.collection.find()"},
			},
			result:    errPlaygroundToBig,
			createdDB: 0,
		},
	}

	t.Run("parallel run", func(t *testing.T) {
		for _, tt := range runCreateDBTests {

			test := tt // capture range variable
			t.Run(test.name, func(t *testing.T) {

				t.Parallel()

				buf := httpBody(t, testServer.runHandler, http.MethodPost, runEndpoint, test.params)

				if want, got := test.result, buf.String(); want != got {
					t.Errorf("expected\n '%s'\n but got\n '%s'", want, got)
				}
			})
		}
	})

	// run only should not save anything in badger
	nbBadgerRecords := 0
	nbMongoDatabases := 0
	for _, tt := range runCreateDBTests {
		nbMongoDatabases += tt.createdDB
	}
	testStorageContent(t, nbMongoDatabases, nbBadgerRecords)
}

func TestRunExistingDB(t *testing.T) {

	defer testServer.clearDatabases(t)

	// the first /run request should create the database
	buf := httpBody(t, testServer.runHandler, http.MethodPost, runEndpoint, templateParams)
	if want, got := templateResult, buf.String(); want != got {
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
	buf = httpBody(t, testServer.runHandler, http.MethodPost, runEndpoint, templateParams)
	if want, got := templateResult, buf.String(); want != got {
		t.Errorf("expected %s but got %s", want, got)
	}

	testStorageContent(t, 1, 0)

}

func TestConsistentError(t *testing.T) {

	defer testServer.clearDatabases(t)

	params := url.Values{"mode": {"mgodatagen"}, "config": {`[{"k":1}]`}, "query": {templateQuery}}
	buf := httpBody(t, testServer.runHandler, http.MethodPost, runEndpoint, params)

	errorMsg := "error in configuration:\n  Error in configuration file: \n\t'collection' and 'database' fields can't be empty"

	if want, got := errorMsg, buf.String(); want != got {
		t.Errorf("expected %s but got %s", want, got)
	}

	buf = httpBody(t, testServer.runHandler, http.MethodPost, runEndpoint, params)

	if want, got := errorMsg, buf.String(); want != got {
		t.Errorf("expected %s but got %s", want, got)
	}
}
