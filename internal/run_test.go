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
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

type runTest struct {
	name          string
	params        url.Values
	result        string
	nbDBexcpected int
}

var runTests = []runTest{
	{
		name:          "incorrect config",
		params:        url.Values{"mode": {"mgodatagen"}, "config": {"h"}, "query": {"db.c.find()"}},
		result:        "error in configuration:\n  error in configuration file: object / array / Date badly formatted: \n\n\t\tinvalid character 'h' looking for beginning of value",
		nbDBexcpected: 0,
	},
	{
		name:          "non existing collection",
		params:        url.Values{"mode": {"mgodatagen"}, "config": {templateConfigOld}, "query": {"db.c.find()"}},
		result:        `collection "c" doesn't exist`,
		nbDBexcpected: 1,
	},
	{
		name:          "deterministic list of objectId",
		params:        templateParams,
		result:        templateResult,
		nbDBexcpected: 0, // db already exists
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
			}]`},
			"query": {templateQuery}},
		result:        `[{"_id":ObjectId("5a934e000102030405000000"),"k":"1jU"},{"_id":ObjectId("5a934e000102030405000001"),"k":"tBRWL"},{"_id":ObjectId("5a934e000102030405000002"),"k":"6Hch"},{"_id":ObjectId("5a934e000102030405000003"),"k":"ZWHW"},{"_id":ObjectId("5a934e000102030405000004"),"k":"RkMG"},{"_id":ObjectId("5a934e000102030405000005"),"k":"RIr"},{"_id":ObjectId("5a934e000102030405000006"),"k":"ru7"},{"_id":ObjectId("5a934e000102030405000007"),"k":"OB"},{"_id":ObjectId("5a934e000102030405000008"),"k":"ja"},{"_id":ObjectId("5a934e000102030405000009"),"k":"K307"}]`,
		nbDBexcpected: 1,
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
			}]`},
			"query": {`db.collection.aggregate([{"$project": {"_id": 0}}])`}},
		result:        `[{"k":"1jU"},{"k":"tBRWL"},{"k":"6Hch"},{"k":"ZWHW"},{"k":"RkMG"},{"k":"RIr"},{"k":"ru7"},{"k":"OB"},{"k":"ja"},{"k":"K307"}]`,
		nbDBexcpected: 0,
	},
	{
		name: "doc nb > 100 mgodatagen mode",
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
			}]`},
			"query": {`db.collection.aggregate([{"$group": {"_id": 0, "nbDoc": {"$sum":1}}}])`}},
		result:        `[{"_id":0,"nbDoc":100}]`,
		nbDBexcpected: 1,
	},
	{
		name: "doc nb > 100 bson mode",
		params: url.Values{
			"mode":   {"bson"},
			"config": {`[{"_id":1},{"_id":2},{"_id":3},{"_id":4},{"_id":5},{"_id":6},{"_id":7},{"_id":8},{"_id":9},{"_id":10},{"_id":11},{"_id":12},{"_id":13},{"_id":14},{"_id":15},{"_id":16},{"_id":17},{"_id":18},{"_id":19},{"_id":20},{"_id":21},{"_id":22},{"_id":23},{"_id":24},{"_id":25},{"_id":26},{"_id":27},{"_id":28},{"_id":29},{"_id":30},{"_id":31},{"_id":32},{"_id":33},{"_id":34},{"_id":35},{"_id":36},{"_id":37},{"_id":38},{"_id":39},{"_id":40},{"_id":41},{"_id":42},{"_id":43},{"_id":44},{"_id":45},{"_id":46},{"_id":47},{"_id":48},{"_id":49},{"_id":50},{"_id":51},{"_id":52},{"_id":53},{"_id":54},{"_id":55},{"_id":56},{"_id":57},{"_id":58},{"_id":59},{"_id":60},{"_id":61},{"_id":62},{"_id":63},{"_id":64},{"_id":65},{"_id":66},{"_id":67},{"_id":68},{"_id":69},{"_id":70},{"_id":71},{"_id":72},{"_id":73},{"_id":74},{"_id":75},{"_id":76},{"_id":77},{"_id":78},{"_id":79},{"_id":80},{"_id":81},{"_id":82},{"_id":83},{"_id":84},{"_id":85},{"_id":86},{"_id":87},{"_id":88},{"_id":89},{"_id":90},{"_id":91},{"_id":92},{"_id":93},{"_id":94},{"_id":95},{"_id":96},{"_id":97},{"_id":98},{"_id":99},{"_id":100},{"_id":101},{"_id":102},{"_id":103},{"_id":104},{"_id":105},{"_id":106}]`},
			"query":  {`db.collection.aggregate([{"$group": {"_id": 0, "nbDoc": {"$sum":1}}}])`}},
		result:        `[{"_id":0,"nbDoc":100}]`,
		nbDBexcpected: 1,
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
			}]`},
			"query": {`db.collection.aggregate([{"$project": {"_id": 0}])`}},
		result:        "error in query:\n  fail to parse content of query: invalid character ']' after object key:value pair",
		nbDBexcpected: 0,
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
			}]`},
			"query": {`db.collection.aggregate([{"$project": "_id"}])`}},
		result:        "query failed: (Location15969) $project specification must be an object",
		nbDBexcpected: 0,
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
			}]`},
			"query": {`db.collection.find({"$set": 12})`}},
		result:        "query failed: (BadValue) unknown top level operator: $set. If you have a field name that starts with a '$' symbol, consider using $getField or $setField.",
		nbDBexcpected: 1,
	},
	{
		name: "invalid find query (invalid json)",
		params: url.Values{
			"mode": {"mgodatagen"},
			"config": {`[
			{
				"collection": "collection",
				"count": 1,
				"content": {
					"k": {
						"type": "string", 
						"minLength": 11,
						"maxLength": 12
					}
				}
			}]`},
			"query": {`db.collection.find({"k": "tJ")`}},
		result:        "error in query:\n  fail to parse content of query: invalid character ']' after object key:value pair",
		nbDBexcpected: 0,
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
			}]`},
			"query": {`db.coll2.find({"k": {"$gt": 3}})`}},
		result:        `[{"_id":ObjectId("5a934e00010203040500000a"),"k":5},{"_id":ObjectId("5a934e00010203040500000b"),"k":5},{"_id":ObjectId("5a934e00010203040500000e"),"k":4},{"_id":ObjectId("5a934e000102030405000011"),"k":5},{"_id":ObjectId("5a934e000102030405000012"),"k":5},{"_id":ObjectId("5a934e000102030405000013"),"k":4}]`,
		nbDBexcpected: 1,
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
			}]`},
			"query": {`db.coll2.find({"k": {"$gt": 3}})`}},
		result:        "error in configuration:\n  fail to create collection coll2: invalid generator for field 'k'\n  cause: invalid type ''",
		nbDBexcpected: 0,
	},
	{
		name: "basic json mode",
		params: url.Values{
			"mode":   {"bson"},
			"config": {`[{"k": 1}]`},
			"query":  {`db.collection.find()`},
		},
		result:        `[{"_id":ObjectId("5a934e000102030405000000"),"k":1}]`,
		nbDBexcpected: 1,
	},
	{
		name: "empty json",
		params: url.Values{
			"mode":   {"bson"},
			"config": {`[{}]`},
			"query":  {`db.collection.find()`},
		},
		result:        `[{"_id":ObjectId("5a934e000102030405000000")}]`,
		nbDBexcpected: 1,
	},
	{
		name: "invalid method",
		params: url.Values{
			"mode":   {"bson"},
			"config": {`[{"_id":356636}]`},
			"query":  {`db.collection.findOne()`},
		},
		result:        "invalid method: 'findOne'",
		nbDBexcpected: 1,
	},
	{
		name: "invalid query syntax",
		params: url.Values{
			"mode":   {"bson"},
			"config": {`[{}]`},
			"query":  {`find()`},
		},
		result:        fmt.Sprintf("error in query:\n  %v", errInvalidQuery),
		nbDBexcpected: 0,
	},
	{
		name: "require array of bson documents or a single document",
		params: url.Values{
			"mode":   {"bson"},
			"config": {`{"k": 1}, {"k": 2}`},
			"query":  {`db.collection.find()`},
		},
		result:        fmt.Sprintf("error in configuration:\n  %v", errInvalidConfig),
		nbDBexcpected: 0,
	},
	{
		name: "multiple collection in bson mode",
		params: url.Values{
			"mode":   {"bson"},
			"config": {`db={"collection1":[{"_id":1,"k":8}],"collection2":[{"_id":1,"k2":10}]}`},
			"query":  {`db.collection1.find()`},
		},
		result:        `[{"_id":1,"k":8}]`,
		nbDBexcpected: 1,
	},
	{
		name: "multiple collection in json mode without _id",
		params: url.Values{
			"mode":   {"bson"},
			"config": {`db={"collection1":[{"k":8}],"collection2":[{"k2":8},{"k2":8}]}`},
			"query":  {`db.collection1.aggregate({"$lookup":{"from":"collection2","localField":"k",foreignField:"k2","as":"lookupDoc"}})`},
		},
		result:        `[{"_id":ObjectId("5a934e000102030405000000"),"k":8,"lookupDoc":[{"_id":ObjectId("5a934e000102030405000001"),"k2":8},{"_id":ObjectId("5a934e000102030405000002"),"k2":8}]}]`,
		nbDBexcpected: 1,
	},
	{
		name: "multiple collection in bson mode with lookup",
		params: url.Values{
			"mode":   {"bson"},
			"config": {`db={"collection1":[{"_id":1,"k":8}],"collection2":[{"_id":1,"k2":1}]}`},
			"query":  {`db.collection1.aggregate({"$lookup":{"from":"collection2","localField":"_id",foreignField:"_id","as":"lookupDoc"}})`},
		},
		result:        `[{"_id":1,"k":8,"lookupDoc":[{"_id":1,"k2":1}]}]`,
		nbDBexcpected: 1,
	},
	{
		name: `bson old syntax create only collection "collection"`,
		params: url.Values{
			"mode":   {"bson"},
			"config": {`[{"k": 1.1111}, {"k": 2.2323}]`},
			"query":  {`db.otherCollection.find()`},
		},
		result:        `collection "otherCollection" doesn't exist`,
		nbDBexcpected: 1,
	},
	{
		name: `doc with "_id" should not be overwritten`,
		params: url.Values{
			"mode":   {"bson"},
			"config": {`[{"_id": 1}, {"_id": 2}]`},
			"query":  {`db.collection.find()`},
		},
		result:        `[{"_id":1},{"_id":2}]`,
		nbDBexcpected: 1,
	},
	{
		name: `mixed doc with/without "_id"`,
		params: url.Values{
			"mode":   {"bson"},
			"config": {`[{"_id": 1}, {}]`},
			"query":  {`db.collection.find()`},
		},
		result:        `[{"_id":1},{"_id":ObjectId("5a934e000102030405000001")}]`,
		nbDBexcpected: 1,
	},
	{
		name: `duplicate "_id" error`,
		params: url.Values{
			"mode":   {"bson"},
			"config": {`[{"_id":1},{"_id":1}]`},
			"query":  {`db.collection.find()`},
		},
		result:        "error in configuration:\n  bulk write exception: write errors: [E11000 duplicate key error collection: 57735364208e15b517d23e542088ed29.collection index: _id_ dup key: { _id: 1.0 }]",
		nbDBexcpected: 0, // the config is incorrect, no db should be created
	},
	{
		name: `bson "ObjectId" notation`,
		params: url.Values{
			"mode":   {"bson"},
			"config": {`[{"_id": ObjectId("5a934e000102030405000001")},{"_id":1}]`},
			"query":  {`db.collection.find()`},
		},
		result:        `[{"_id":ObjectId("5a934e000102030405000001")},{"_id":1}]`,
		nbDBexcpected: 1,
	},
	{
		name: `bson unkeyed notation`,
		params: url.Values{
			"mode":   {"bson"},
			"config": {`[{"_id": ObjectId("5a934e000102030405000001")},{"_id":1}]`},
			"query":  {`db.collection.find({_id: ObjectId("5a934e000102030405000001")})`},
		},
		result:        `[{"_id":ObjectId("5a934e000102030405000001")}]`,
		nbDBexcpected: 0,
	},
	{
		name: `unkeyed params in aggreagtion`,
		params: url.Values{
			"mode":   {"bson"},
			"config": {`[{"_id": ObjectId("5a934e000102030405000001")},{"_id":1}]`},
			"query":  {`db.collection.aggregate([{$match: {_id: ObjectId("5a934e000102030405000001")}}])`},
		},
		result:        `[{"_id":ObjectId("5a934e000102030405000001")}]`,
		nbDBexcpected: 0,
	},
	{
		name: `doc with bson "ISODate"`,
		params: url.Values{
			"mode":   {"bson"},
			"config": {`[{dt: ISODate("2000-01-01T00:00:00+00:00")}]`},
			"query":  {`db.collection.find()`},
		},
		result:        `[{"_id":ObjectId("5a934e000102030405000000"),"dt":ISODate("2000-01-01T00:00:00Z")}]`,
		nbDBexcpected: 1,
	},
	{
		name: `invalid "ObjectId" should not panic`,
		params: url.Values{
			"mode":   {"bson"},
			"config": {`[{"_id": ObjectId("5a9")}]`},
			"query":  {`db.collection.find({_id: ObjectId("5a934e000102030405000001")})`},
		},
		result:        "error in configuration:\n  the provided hex string is not a valid ObjectID",
		nbDBexcpected: 0,
	},
	{
		name: `regex parsing`, // TODO
		params: url.Values{
			"mode":   {"bson"},
			"config": {`[{"k": "randompattern"}]`},
			"query":  {`db.collection.find({k: /pattern/})`},
		},
		result:        "error in query:\n  fail to parse content of query: invalid character '/' looking for beginning of value",
		nbDBexcpected: 0,
	},
	{
		name: `query with projection`,
		params: url.Values{
			"mode":   {"bson"},
			"config": {`[{"k":1},{"k":2},{"k":3}]`},
			"query":  {`db.collection.find({}, {"_id": 0})`},
		},
		result:        `[{"k":1},{"k":2},{"k":3}]`,
		nbDBexcpected: 1,
	},
	{
		name: `empty config`,
		params: url.Values{
			"mode":   {"bson"},
			"config": {""},
			"query":  {"db.c.find()"},
		},
		result:        fmt.Sprintf("error in configuration:\n  %v", errInvalidConfig),
		nbDBexcpected: 0,
	},
	{
		name: `empty query`,
		params: url.Values{
			"mode":   {"bson"},
			"config": {templateConfigOld},
			"query":  {""},
		},
		result:        fmt.Sprintf("error in query:\n  %v", errInvalidQuery),
		nbDBexcpected: 0,
	},
	{
		name: `too many collections`,
		params: url.Values{
			"mode":   {"bson"},
			"config": {`db={"a":[],"b":[],"c":[],"d":[],"e":[],"f":[],"g":[],"h":[],"i":[],"j":[],"k":[]}`},
			"query":  {"db.c.find()"},
		},
		result:        "error in configuration:\n  max number of collection in a database is 10, but was 11",
		nbDBexcpected: 0,
	},
	{
		name: `no documents found`,
		params: url.Values{
			"mode":   {"bson"},
			"config": {`db={"a":[]}`},
			"query":  {"db.a.find()"},
		},
		result:        noDocFound,
		nbDBexcpected: 0,
	},
	{
		name: `invalid query with 3 '.'`,
		params: url.Values{
			"mode":   {"bson"},
			"config": {templateConfigOld},
			"query":  {`[{"key.path.test":{"$match":10}}])`},
		},
		result:        fmt.Sprintf("error in query:\n  %v", errInvalidQuery),
		nbDBexcpected: 0,
	},
	{
		name: `playground too big`,
		params: url.Values{
			"mode":   {"bson"},
			"config": {string(make([]byte, maxByteSize))},
			"query":  {"db.collection.find()"},
		},
		result:        errPlaygroundToBig,
		nbDBexcpected: 0,
	},
	{
		name: `basic update one`,
		params: url.Values{
			"mode":   {"bson"},
			"config": {`[{"_id":1,"k":3},{"_id":2,"k":3}]`},
			"query":  {`db.collection.update({"k":3}, {"$set": {"k":0}}, {"multi": false})`},
		},
		result:        `[{"_id":1,"k":0},{"_id":2,"k":3}]`,
		nbDBexcpected: 1,
	},
	{
		name: `basic update many`,
		params: url.Values{
			"mode":   {"bson"},
			"config": {`[{"_id":1,"n":5},{"_id":2,"n":2}]`},
			"query":  {`db.collection.update({}, {"$inc": {"n":10}}, {"multi": true})`},
		},
		result:        `[{"_id":1,"n":15},{"_id":2,"n":12}]`,
		nbDBexcpected: 1,
	},
	{
		name: `update without option`,
		params: url.Values{
			"mode":   {"bson"},
			"config": {`[{"_id":1,"name":"ke"},{"_id":2,"name":"lme"}]`},
			"query":  {`db.collection.update({}, {"$rename": {"name":"new"}})`},
		},
		result:        `[{"_id":1,"new":"ke"},{"_id":2,"name":"lme"}]`,
		nbDBexcpected: 1,
	},
	{
		name: `update with upsert`,
		params: url.Values{
			"mode":   {"bson"},
			"config": {`[{"field":2.334}]`},
			"query":  {`db.collection.update({"field":2}, {"$set": {"_id":2}}, {"upsert": true})`},
		},
		result:        `[{"_id":ObjectId("5a934e000102030405000000"),"field":2.334},{"_id":2,"field":2}]`,
		nbDBexcpected: 1,
	},
	{
		name: `update with arrayFilter`,
		params: url.Values{
			"mode":   {"bson"},
			"config": {`[{"_id":1,"grades":[95,92,90]},{"_id":2,"grades":[98,100,102]},{"_id":3,"grades":[95,110,100]}]`},
			"query":  {`db.collection.update({grades:{$gte:100}},{$set:{"grades.$[element]":100}}, {"multi": true, arrayFilters: [{"element": { $gte: 100 }}]})`},
		},
		result:        `[{"_id":1,"grades":[95,92,90]},{"_id":2,"grades":[98,100,100]},{"_id":3,"grades":[95,100,100]}]`,
		nbDBexcpected: 1,
	},
	{
		name: `empty update`,
		params: url.Values{
			"mode":   {"bson"},
			"config": {`[{"_id":1,"g":95},{"_id":2,"g":98}]`},
			"query":  {`db.collection.update()`},
		},
		result:        `fail to run update: update document must have at least one element`,
		nbDBexcpected: 1,
	},
	{
		name: `upsert with empty db`,
		params: url.Values{
			"mode":   {"bson"},
			"config": {`[]`},
			"query":  {`db.collection.update({},{"$set":{"_id":"new"}},{"upsert":true})`},
		},
		result:        `[{"_id":"new"}]`,
		nbDBexcpected: 1, // this should create a db even if config is empty, because of the upsert
	},
	{
		name: `update with pipeline`,
		params: url.Values{
			"mode":   {"bson"},
			"config": {`[{"_id":1,"username":"moshe","health":0,"maxHealth":200}]`},
			"query":  {`db.collection.update({},[{"$set": { "health": "$maxHealth" }}])`},
		},
		result:        `[{"_id":1,"health":200,"maxHealth":200,"username":"moshe"}]`,
		nbDBexcpected: 1,
	},
	{
		name: `explain default`,
		params: url.Values{
			"mode":   {"bson"},
			"config": {`[{"_id":1,"username":"greta"}]`},
			"query":  {`db.collection.find().explain()`},
		},
		result:        `{"command":{"$db":"433c2ef8cb26c90dd962d047dea315de","filter":{},"find":"collection","maxTimeMS":NumberLong(20000),"projection":{}},"explainVersion":"1","queryPlanner":{"indexFilterSet":false,"maxIndexedAndSolutionsReached":false,"maxIndexedOrSolutionsReached":false,"maxScansToExplodeReached":false,"namespace":"433c2ef8cb26c90dd962d047dea315de.collection","parsedQuery":{},"planCacheKey":"D542626C","queryHash":"8B3D4AB8","rejectedPlans":[],"winningPlan":{"direction":"forward","stage":"COLLSCAN"}},"serverParameters":{"internalDocumentSourceGroupMaxMemoryBytes":104857600,"internalDocumentSourceSetWindowFieldsMaxMemoryBytes":104857600,"internalLookupStageIntermediateDocumentMaxSizeBytes":104857600,"internalQueryFacetBufferSizeBytes":104857600,"internalQueryFacetMaxOutputDocSizeBytes":104857600,"internalQueryMaxAddToSetBytes":104857600,"internalQueryMaxBlockingSortMemoryUsageBytes":104857600,"internalQueryProhibitBlockingMergeOnMongoS":0}}`,
		nbDBexcpected: 1,
	},
	{
		name: `explain executionStats`,
		params: url.Values{
			"mode":   {"bson"},
			"config": {`[{"_id":1,"username":"tim"}]`},
			"query":  {`db.collection.find().explain("executionStats")`},
		},
		result:        `{"command":{"$db":"d0eaaeabc460c11f6f70b605a70c50d8","filter":{},"find":"collection","maxTimeMS":NumberLong(20000),"projection":{}},"executionStats":{"executionStages":{"advanced":1,"direction":"forward","docsExamined":1,"executionTimeMillisEstimate":0,"isEOF":1,"nReturned":1,"needTime":1,"needYield":0,"restoreState":0,"saveState":0,"stage":"COLLSCAN","works":3},"executionSuccess":true,"executionTimeMillis":0,"nReturned":1,"totalDocsExamined":1,"totalKeysExamined":0},"explainVersion":"1","queryPlanner":{"indexFilterSet":false,"maxIndexedAndSolutionsReached":false,"maxIndexedOrSolutionsReached":false,"maxScansToExplodeReached":false,"namespace":"d0eaaeabc460c11f6f70b605a70c50d8.collection","parsedQuery":{},"rejectedPlans":[],"winningPlan":{"direction":"forward","stage":"COLLSCAN"}},"serverParameters":{"internalDocumentSourceGroupMaxMemoryBytes":104857600,"internalDocumentSourceSetWindowFieldsMaxMemoryBytes":104857600,"internalLookupStageIntermediateDocumentMaxSizeBytes":104857600,"internalQueryFacetBufferSizeBytes":104857600,"internalQueryFacetMaxOutputDocSizeBytes":104857600,"internalQueryMaxAddToSetBytes":104857600,"internalQueryMaxBlockingSortMemoryUsageBytes":104857600,"internalQueryProhibitBlockingMergeOnMongoS":0}}`,
		nbDBexcpected: 1,
	},
	{
		name: `explain executionStats before find`,
		params: url.Values{
			"mode":   {"bson"},
			"config": {`[{"_id":1,"username":"tim"}]`},
			"query":  {`db.collection.explain("executionStats").find()`},
		},
		result:        `{"command":{"$db":"d0eaaeabc460c11f6f70b605a70c50d8","filter":{},"find":"collection","maxTimeMS":NumberLong(20000),"projection":{}},"executionStats":{"executionStages":{"advanced":1,"direction":"forward","docsExamined":1,"executionTimeMillisEstimate":0,"isEOF":1,"nReturned":1,"needTime":1,"needYield":0,"restoreState":0,"saveState":0,"stage":"COLLSCAN","works":3},"executionSuccess":true,"executionTimeMillis":0,"nReturned":1,"totalDocsExamined":1,"totalKeysExamined":0},"explainVersion":"1","queryPlanner":{"indexFilterSet":false,"maxIndexedAndSolutionsReached":false,"maxIndexedOrSolutionsReached":false,"maxScansToExplodeReached":false,"namespace":"d0eaaeabc460c11f6f70b605a70c50d8.collection","parsedQuery":{},"rejectedPlans":[],"winningPlan":{"direction":"forward","stage":"COLLSCAN"}},"serverParameters":{"internalDocumentSourceGroupMaxMemoryBytes":104857600,"internalDocumentSourceSetWindowFieldsMaxMemoryBytes":104857600,"internalLookupStageIntermediateDocumentMaxSizeBytes":104857600,"internalQueryFacetBufferSizeBytes":104857600,"internalQueryFacetMaxOutputDocSizeBytes":104857600,"internalQueryMaxAddToSetBytes":104857600,"internalQueryMaxBlockingSortMemoryUsageBytes":104857600,"internalQueryProhibitBlockingMergeOnMongoS":0}}`,
		nbDBexcpected: 0, // same config as above
	},
	{
		name: `explain allPlansExecution before find`,
		params: url.Values{
			"mode":   {"bson"},
			"config": {`[{"_id":1,"username":"TP"}]`},
			"query":  {`db.collection.explain("allPlansExecution").find()`},
		},
		result:        `{"command":{"$db":"40dd3ef1cd82a6d68d98fdcd3ddf4242","filter":{},"find":"collection","maxTimeMS":NumberLong(20000),"projection":{}},"executionStats":{"allPlansExecution":[],"executionStages":{"advanced":1,"direction":"forward","docsExamined":1,"executionTimeMillisEstimate":0,"isEOF":1,"nReturned":1,"needTime":1,"needYield":0,"restoreState":0,"saveState":0,"stage":"COLLSCAN","works":3},"executionSuccess":true,"executionTimeMillis":0,"nReturned":1,"totalDocsExamined":1,"totalKeysExamined":0},"explainVersion":"1","queryPlanner":{"indexFilterSet":false,"maxIndexedAndSolutionsReached":false,"maxIndexedOrSolutionsReached":false,"maxScansToExplodeReached":false,"namespace":"40dd3ef1cd82a6d68d98fdcd3ddf4242.collection","parsedQuery":{},"rejectedPlans":[],"winningPlan":{"direction":"forward","stage":"COLLSCAN"}},"serverParameters":{"internalDocumentSourceGroupMaxMemoryBytes":104857600,"internalDocumentSourceSetWindowFieldsMaxMemoryBytes":104857600,"internalLookupStageIntermediateDocumentMaxSizeBytes":104857600,"internalQueryFacetBufferSizeBytes":104857600,"internalQueryFacetMaxOutputDocSizeBytes":104857600,"internalQueryMaxAddToSetBytes":104857600,"internalQueryMaxBlockingSortMemoryUsageBytes":104857600,"internalQueryProhibitBlockingMergeOnMongoS":0}}`,
		nbDBexcpected: 1,
	},
	{
		name: `malformatted explain`,
		params: url.Values{
			"mode":   {"bson"},
			"config": {`[{"_id":1,"username":"unfinished"}]`},
			"query":  {`db.collection.find().explain(`},
		},
		result:        `[{"_id":1,"username":"unfinished"}]`,
		nbDBexcpected: 1,
	},
	{
		name: `mgodatagen $text query without index`,
		params: url.Values{
			"mode": {"mgodatagen"},
			"config": {`[
				{
				  "collection": "collection",
				  "count": 10,
				  "content": {
					"question": {
					  "type": "faker",
					  "method": "MimeType"
					}
				  }
				}
			  ]`},
			"query": {`db.collection.find({
				$text: {
				  $search: "application -rtf -pkcs"
				}
			  })`},
		},
		result:        `query failed: (IndexNotFound) text index required for $text query`,
		nbDBexcpected: 1,
	},
	{
		name: `mgodatagen $text query with index`,
		params: url.Values{
			"mode": {"mgodatagen"},
			"config": {`[
				{
				  "collection": "collection",
				  "count": 10,
				  "content": {
					"word": {
					  "type": "string",
					  "minLength": 3,
					  "maxLength": 10
					}
				  },
				  "indexes": [
					{
					  "name": "word_text",
					  "key": {
						"word": "text"
					  }
					}
				  ]
				}
			  ]`},
			"query": {`db.collection.find({
				$text: {
				  $search: "RIre"
				}
			  })`},
		},
		result:        `[{"_id":ObjectId("5a934e000102030405000005"),"word":"RIre"}]`,
		nbDBexcpected: 1,
	},
	{
		name: `aggregation batch size greater than 100 ( defaut )`,
		params: url.Values{
			"mode": {"mgodatagen"},
			"config": {`[
				{
				  "collection": "collection",
				  "count": 1,
				  "content": {
					"array": {
					  "type": "array",
					  "size": 110,
					  "arrayContent": {
						"type": "autoincrement",
						"autoType": "int",
						"startInt": 1
					  }
					}
				  }
				}
			  ]`},
			"query": {`db.collection.aggregate([{"$unwind": "$array"},{"$project":{"_id":"$array"}}])`},
		},
		result:        `[{"_id":1},{"_id":2},{"_id":3},{"_id":4},{"_id":5},{"_id":6},{"_id":7},{"_id":8},{"_id":9},{"_id":10},{"_id":11},{"_id":12},{"_id":13},{"_id":14},{"_id":15},{"_id":16},{"_id":17},{"_id":18},{"_id":19},{"_id":20},{"_id":21},{"_id":22},{"_id":23},{"_id":24},{"_id":25},{"_id":26},{"_id":27},{"_id":28},{"_id":29},{"_id":30},{"_id":31},{"_id":32},{"_id":33},{"_id":34},{"_id":35},{"_id":36},{"_id":37},{"_id":38},{"_id":39},{"_id":40},{"_id":41},{"_id":42},{"_id":43},{"_id":44},{"_id":45},{"_id":46},{"_id":47},{"_id":48},{"_id":49},{"_id":50},{"_id":51},{"_id":52},{"_id":53},{"_id":54},{"_id":55},{"_id":56},{"_id":57},{"_id":58},{"_id":59},{"_id":60},{"_id":61},{"_id":62},{"_id":63},{"_id":64},{"_id":65},{"_id":66},{"_id":67},{"_id":68},{"_id":69},{"_id":70},{"_id":71},{"_id":72},{"_id":73},{"_id":74},{"_id":75},{"_id":76},{"_id":77},{"_id":78},{"_id":79},{"_id":80},{"_id":81},{"_id":82},{"_id":83},{"_id":84},{"_id":85},{"_id":86},{"_id":87},{"_id":88},{"_id":89},{"_id":90},{"_id":91},{"_id":92},{"_id":93},{"_id":94},{"_id":95},{"_id":96},{"_id":97},{"_id":98},{"_id":99},{"_id":100},{"_id":101},{"_id":102},{"_id":103},{"_id":104},{"_id":105},{"_id":106},{"_id":107},{"_id":108},{"_id":109},{"_id":110}]`,
		nbDBexcpected: 1,
	},
	{
		name: `explain too short`,
		params: url.Values{
			"mode":   {"bson"},
			"config": {`[{"_id":1,"username":"singleQuote"}]`},
			"query":  {`db.collection.find().explain(")`},
		},
		result:        `{"command":{"$db":"7fb2bc41534140cadc0bb68d1377cc2a","filter":{},"find":"collection","maxTimeMS":NumberLong(20000),"projection":{}},"explainVersion":"1","queryPlanner":{"indexFilterSet":false,"maxIndexedAndSolutionsReached":false,"maxIndexedOrSolutionsReached":false,"maxScansToExplodeReached":false,"namespace":"7fb2bc41534140cadc0bb68d1377cc2a.collection","parsedQuery":{},"planCacheKey":"D542626C","queryHash":"8B3D4AB8","rejectedPlans":[],"winningPlan":{"direction":"forward","stage":"COLLSCAN"}},"serverParameters":{"internalDocumentSourceGroupMaxMemoryBytes":104857600,"internalDocumentSourceSetWindowFieldsMaxMemoryBytes":104857600,"internalLookupStageIntermediateDocumentMaxSizeBytes":104857600,"internalQueryFacetBufferSizeBytes":104857600,"internalQueryFacetMaxOutputDocSizeBytes":104857600,"internalQueryMaxAddToSetBytes":104857600,"internalQueryMaxBlockingSortMemoryUsageBytes":104857600,"internalQueryProhibitBlockingMergeOnMongoS":0}}`,
		nbDBexcpected: 1,
	},
	{
		name: `aggregation with $out`,
		params: url.Values{
			"mode":   {"bson"},
			"config": {`[{"_id":"yellow"}]`},
			"query":  {`db.collection.aggregate([{$out: "ouptut"}])`},
		},
		result:        `[{"_id":"yellow"}]`,
		nbDBexcpected: 1,
	},
	{
		name: `aggregation with "$out" quoted`,
		params: url.Values{
			"mode":   {"bson"},
			"config": {`[{"_id":1},{"_id":2},{"_id":3}]`},
			"query":  {`db.collection.aggregate([{"$match":{"_id":1}},{"$out": {db: "ouptut", collection: "y"}}])`},
		},
		result:        `[{"_id":1}]`,
		nbDBexcpected: 1,
	},
	{
		name: `aggregation with $merge`,
		params: url.Values{
			"mode":   {"bson"},
			"config": {`[{"_id":"abcde"}]`},
			"query":  {`db.collection.aggregate([{$merge: "ouptut-merge"}])`},
		},
		result:        `[{"_id":"abcde"}]`,
		nbDBexcpected: 1,
	},
	{
		name: `aggregation invalid pipeline`,
		params: url.Values{
			"mode":   {"bson"},
			"config": {`[{"_id":0.000122}]`},
			"query":  {`db.collection.aggregate([1,2])`},
		},
		result:        `query failed: (TypeMismatch) Each element of the 'pipeline' array must be an object`,
		nbDBexcpected: 1,
	},
	{
		name: `aggregation with $merge in first pos`,
		params: url.Values{
			"mode":   {"bson"},
			"config": {`[{"_id":1}]`},
			"query":  {`db.collection.aggregate([{$merge: "ouptut-merge"},{$match:{_id:1}},{$project:{_id:0}}])`},
		},
		result:        `[{}]`,
		nbDBexcpected: 1,
	},
	{
		name: `aggregation with $out and $merge`,
		params: url.Values{
			"mode":   {"bson"},
			"config": {`[{"_id":1},{"_id":2},{"_id":3},{"_id":38294834}]`},
			"query":  {`db.collection.aggregate([{$merge: "ouptut-merge"},{"$match":{"_id":1}},{"$out": {db: "ouptut", collection: "y"}}])`},
		},
		result:        `[{"_id":1}]`,
		nbDBexcpected: 1,
	},
	{
		name: `fuzz entry 1`,
		params: url.Values{
			"mode":   {"bson"},
			"config": {`[]`},
			"query":  {`..)(`},
		},
		result:        "error in query:\n  query must match db.coll.find(...) or db.coll.aggregate(...) or db.coll.update()",
		nbDBexcpected: 0,
	},
}

func TestRunCreateDB(t *testing.T) {

	defer clearDatabases(t)

	t.Run("parallel run", func(t *testing.T) {
		for _, tt := range runTests {

			test := tt // capture range variable
			t.Run(test.name, func(t *testing.T) {

				t.Parallel()

				got := httpBody(t, runEndpoint, http.MethodPost, test.params)

				if want := test.result; want != got {
					t.Errorf("expected\n '%s'\n but got\n '%s'", want, got)
				}
			})
		}
	})

	// run only should not save anything in badger
	nbBadgerRecords := 0
	nbMongoDatabases := 0
	for _, tt := range runTests {
		nbMongoDatabases += tt.nbDBexcpected
	}
	testStorageContent(t, nbMongoDatabases, nbBadgerRecords)
}

func TestRunExistingDB(t *testing.T) {

	defer clearDatabases(t)

	// the first /run request should create the database
	want := templateResult
	got := httpBody(t, runEndpoint, http.MethodPost, templateParams)
	if want != got {
		t.Errorf("expected %s but got %s", want, got)
	}
	p := &page{
		Mode:   mgodatagenMode,
		Config: []byte(templateParams.Get("config")),
	}
	DBHash := p.dbHash()
	_, ok := testStorage.activeDB[DBHash]
	if !ok {
		t.Errorf("activeDb should contain DB %s", DBHash)
	}

	//  the second /run should produce the same result
	got = httpBody(t, runEndpoint, http.MethodPost, templateParams)
	if want != got {
		t.Errorf("expected %s but got %s", want, got)
	}

	testStorageContent(t, 1, 0)
}

func TestRunUpdateTwice(t *testing.T) {

	defer clearDatabases(t)

	params := url.Values{"mode": {"bson"}, "config": {`[]`}, "query": {`db.collection.update({},{"$set":{"_id":0}},{"upsert":true})`}}
	want := `[{"_id":0}]`
	got := httpBody(t, runEndpoint, http.MethodPost, params)
	if want != got {
		t.Errorf("expected %s but got %s", want, got)
	}
	// re-run the same run query, activeDatabase counter should not be
	// incremented because it's the same db, even if we re-create it
	// every time
	got = httpBody(t, runEndpoint, http.MethodPost, params)
	if want != got {
		t.Errorf("expected %s but got %s", want, got)
	}

	testStorageContent(t, 1, 0)
}

func TestRunFindAfterUpdate(t *testing.T) {
	defer clearDatabases(t)

	params := url.Values{"mode": {"bson"}, "config": {`[{_id:1}]`}, "query": {`db.collection.update({},{"$set":{"updated":true}})`}}
	want := `[{"_id":1,"updated":true}]`
	got := httpBody(t, runEndpoint, http.MethodPost, params)
	if want != got {
		t.Errorf("expected %s but got %s", want, got)
	}
	// change query to be a find(), but keep mode and config the same as for
	// the previous update(). This should create a distinct db
	params.Set("query", "db.collection.find()")
	want = `[{"_id":1}]`
	got = httpBody(t, runEndpoint, http.MethodPost, params)
	if want != got {
		t.Errorf("expected %s but got %s", want, got)
	}

	testStorageContent(t, 2, 0)
}

func TestConsistentError(t *testing.T) {

	defer clearDatabases(t)

	params := url.Values{"mode": {"mgodatagen"}, "config": {`[{}]`}, "query": {templateQuery}}
	want := "error in configuration:\n  error in configuration file: \n\t'collection' and 'database' fields can't be empty"
	got := httpBody(t, runEndpoint, http.MethodPost, params)

	if want != got {
		t.Errorf("expected\n'%s'\n but got\n'%s'", want, got)
	}

	got = httpBody(t, runEndpoint, http.MethodPost, params)
	if want != got {
		t.Errorf("expected\n'%s'\n but got\n'%s'", want, got)
	}
}

// for https://github.com/feliixx/mongoplayground/issues/120
func TestUniqueBinaryUUID(t *testing.T) {

	defer clearDatabases(t)

	params := url.Values{"mode": {"mgodatagen"}, "config": {`[
		{
		  "collection": "collection",
		  "count": 3,
		  "content": {
			"uuid": {
			  "type": "uuid",
			  "format": "binary"
			}
		  }
		}
	  ]`}, "query": {`db.collection.aggregate([{"$project": {"_id":0}}])`}}

	got := httpBody(t, runEndpoint, http.MethodPost, params)
	got = strings.ReplaceAll(got, "[", "")
	got = strings.ReplaceAll(got, "]", "")
	got = strings.ReplaceAll(got, "4,", "")

	uuids := strings.Split(got, ",")
	if uuids[0] == uuids[1] || uuids[1] == uuids[2] {
		t.Errorf("expected unique UUIDs in db but got same multiple times: %v", uuids)
	}
}

func FuzzRun(f *testing.F) {

	for _, tt := range runTests {
		f.Add(
			tt.params.Get("mode"),
			tt.params.Get("config"),
			tt.params.Get("query"),
		)
	}
	f.Fuzz(func(t *testing.T, mode, config, query string) {

		params := url.Values{"mode": {mode}, "config": {config}, "query": {query}}
		req, err := http.NewRequest(http.MethodPost, runEndpoint, strings.NewReader(params.Encode()))
		if err != nil {
			t.Errorf("fail to create request with params %+v", params)
		}
		resp := httptest.NewRecorder()
		testStorage.runHandler(resp, req)
	})
}
