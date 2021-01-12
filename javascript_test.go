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

package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"testing"
)

func TestJavascriptIndentRoundTrip(t *testing.T) {

	t.Parallel()

	jsIndentTests := []struct {
		name    string
		input   string
		indent  string
		compact string
	}{
		{
			name:  `valid json`,
			input: `[{ "_id": 1, "key": {"field": "someValue"}}]`,
			indent: `[
  {
    "_id": 1,
    "key": {
      "field": "someValue"
    }
  }
]`,
			compact: `[{"_id":1,"key":{"field":"someValue"}}]`,
		}, {
			name:  `find() query`,
			input: `db.collection.find({ "_id": ObjectId("5a934e000102030405000000")}, { "_id":   0} )`,
			indent: `db.collection.find({
  "_id": ObjectId("5a934e000102030405000000")
},
{
  "_id": 0
})`,
			compact: `db.collection.find({"_id":ObjectId("5a934e000102030405000000")},{"_id":0})`,
		},
		{
			name: `valid json with tabs`,
			input: `[{	"_id":	1, "key": 	{"field": "someValue"}}]`,
			indent: `[
  {
    "_id": 1,
    "key": {
      "field": "someValue"
    }
  }
]`,
			compact: `[{"_id":1,"key":{"field":"someValue"}}]`,
		},
		{
			name: `new Date()`,
			input: `[ { "key": new Date(18384919)	}]`,
			indent: `[
  {
    "key": new Date(18384919)
  }
]`,
			compact: `[{"key":new Date(18384919)}]`,
		},
		{
			name: `empty json`,
			input: `[
{


}
]`,
			indent: `[
  {}
]`,
			compact: `[{}]`,
		},
		{
			name: `valid bson`,
			input: `[{_id: ObjectId("5a934e000102030405000000"), "date": ISODate("2000-01-01T00:00:00Z") }, 
			{ "_id": ObjectId("5a934e000102030405000001"), ts: Timestamp(1,1), newDate: new Date(1)}, 
			{"k": NumberInt(10), "k2": NumberLong(15), k3: NumberDecimal(177), f: 2.994499433}, 
			{"k": undefined, n: null,     
				
				bin: BinData(2,"ZmfjfjghhsjGSDHbdsj"), name: "some name"}]`,
			indent: `[
  {
    _id: ObjectId("5a934e000102030405000000"),
    "date": ISODate("2000-01-01T00:00:00Z")
  },
  {
    "_id": ObjectId("5a934e000102030405000001"),
    ts: Timestamp(1, 1),
    newDate: new Date(1)
  },
  {
    "k": NumberInt(10),
    "k2": NumberLong(15),
    k3: NumberDecimal(177),
    f: 2.994499433
  },
  {
    "k": undefined,
    n: null,
    bin: BinData(2, "ZmfjfjghhsjGSDHbdsj"),
    name: "some name"
  }
]`,
			compact: `[{_id:ObjectId("5a934e000102030405000000"),"date":ISODate("2000-01-01T00:00:00Z")},{"_id":ObjectId("5a934e000102030405000001"),ts:Timestamp(1,1),newDate:new Date(1)},{"k":NumberInt(10),"k2":NumberLong(15),k3:NumberDecimal(177),f:2.994499433},{"k":undefined,n:null,bin:BinData(2,"ZmfjfjghhsjGSDHbdsj"),name:"some name"}]`,
		},
		{
			name:  `replace single quote with double quote`,
			input: `[{ 'k': 'value 1', 'k2': "O'Neil" }]`,
			indent: `[
  {
    "k": "value 1",
    "k2": "O'Neil"
  }
]`,
			compact: `[{"k":"value 1","k2":"O'Neil"}]`,
		},
		{
			name: `javascript regex`,
			input: `db.col123.aggregate([ { "$match": {
				'k': /^db\..(\w+)\.(find|aggregate)\([\s\S]*\)$/
			}}])`,
			indent: `db.col123.aggregate([
  {
    "$match": {
      "k": /^db\..(\w+)\.(find|aggregate)\([\s\S]*\)$/
    }
  }
])`,
			compact: `db.col123.aggregate([{"$match":{"k":/^db\..(\w+)\.(find|aggregate)\([\s\S]*\)$/}}])`,
		},
		{
			name:    `invalid input missing '('`,
			input:   `db.coll.find{ })`,
			indent:  `db.coll.find{})`,
			compact: `db.coll.find{})`,
		},
		{
			name:  `unfinished regex`,
			input: `[{ k: /^db.*(\w)}]`,
			indent: `[
  {
    k: /^db.*(\w)}]`,
			compact: `[{k:/^db.*(\w)}]`,
		},
		{
			name:  `unfinished quoted string`,
			input: `[{k: "str}]`,
			indent: `[
  {
    k: "str}]`,
			compact: `[{k:"str}]`,
		},
		{
			name:  `unfinished new Date()`,
			input: `[{k: new Date(89928  }  ]`,
			indent: `[
  {
    k: new Date(89928  }  ]`,
			compact: `[{k:new Date(89928  }  ]`,
		},
		{
			name: `multiple collection bson`,
			input: `
 db  =   {
	 coll1: [
		 {k:NumberInt(1234)}
	]
}`,
			compact: `db={coll1:[{k:NumberInt(1234)}]}`,
			indent: `db={
  coll1: [
    {
      k: NumberInt(1234)
    }
  ]
}`,
		},
		{
			name: "single line comment",
			input: `
			db  =   {
				// first coll to create
				coll1: [
					{k:NumberInt(1234)}
			   ]
		   }`,
			compact: `db={/** first coll to create*/coll1:[{k:NumberInt(1234)}]}`,
			indent: `db={
  /** first coll to create*/
  coll1: [
    {
      k: NumberInt(1234)
    }
  ]
}`,
		},
		{
			name: "mutli line comment",
			input: `
			db  =   {
				/** the coll one
				*
				* that end here
				*/
				coll1: [
					{k:NumberInt(1234)}
			   ]
		   }`,
			compact: `db={/** the coll one** that end here*/coll1:[{k:NumberInt(1234)}]}`,
			indent: `db={
  /** the coll one
  *
  * that end here
  */
  coll1: [
    {
      k: NumberInt(1234)
    }
  ]
}`,
		},
		{
			name:    "unfinished comment",
			input:   `[{"k":/*}]`,
			compact: `[{"k":}]`,
			indent: `[
  {
    "k": 
  }
]`,
		},
		{
			name: "mixed single and multiple line",
			input: `
			db.collection.find({
				// the key
				k: 1
			})/**   comment
			*
			*/`,
			compact: `db.collection.find({/** the key*/k:1})/**   comment**/`,
			indent: `db.collection.find({
  /** the key*/
  k: 1
})/**   comment
*
*/
`,
		},
		{
			name: "multiline comment single star",
			input: `
			db.collection.find({
				/*
				the key
				*/
				k: 1
			})`,
			compact: `db.collection.find({/*** the key**/k:1})`,
			indent: `db.collection.find({
  /**
  * the key
  *
  */
  k: 1
})`,
		},
		{
			name: "multiline comment single star",
			input: `
			db.collection.find({
				/*some comment
	   on multiple line 
	                   with weird indentation 
				*/
				k: 1
			})`,
			compact: `db.collection.find({/**some comment* on multiple line* with weird indentation**/k:1})`,
			indent: `db.collection.find({
  /**some comment
  * on multiple line
  * with weird indentation
  *
  */
  k: 1
})`,
		},
		{
			name:    "mix single star multiple star",
			input:   `[{"k":1}]/**/`,
			compact: `[{"k":1}]`,
			indent: `[
  {
    "k": 1
  }
]`,
		},
		{
			name:    "query with trailing comma",
			input:   `db.collection.find();`,
			compact: `db.collection.find()`,
			indent:  `db.collection.find()`,
		},
		{
			name: "comment with no line return",
			input: `db.collection.find()
			// comment with no line return`,
			compact: `db.collection.find()/** comment with no line return*/`,
			indent: `db.collection.find()/** comment with no line return*/
`,
		},
		{
			name: "aggregate with explain",
			input: `db.collection.find().explain(
				
			)`,
			compact: `db.collection.find().explain()`,
			indent:  `db.collection.find().explain()`,
		},
	}

	buffer := loadPlaygroundJs(t)

	testFormat := `
	{
		"name": %s,
		"input": %s, 
		"expectedIndent": %s, 
		"expectedCompact": %s
	}
	`

	buffer.Write([]byte(`
		var tests = [`))
	for _, tt := range jsIndentTests {
		fmt.Fprintf(buffer, testFormat, strconv.Quote(tt.name), strconv.Quote(tt.input), strconv.Quote(tt.indent), strconv.Quote(tt.compact))
		buffer.WriteByte(',')
	}
	buffer.Write([]byte(`
	]
	
	`))

	// for each test case, indent/compact the input, and make sure results are correct.
	// Then, indent/compact the results, to make sure that re-indent/re-compact give the same
	// results

	buffer.Write([]byte(`
	for (let i in tests) {
		let tt = tests[i]

		let indentResult = indent(tt.input)
		if (indentResult !== tt.expectedIndent) {
			print("test " + tt.name + " ident failed, expected: \n" + tt.expectedIndent +  "\nbut got: \n" + indentResult)
		}
		let compactResult = compact(tt.input)
		if (compactResult !== tt.expectedCompact) {
			print("test " + tt.name + " compact failed, expected: \n" + tt.expectedCompact +  "\nbut got: \n" + compactResult)
		}

		indentResult = indent(indentResult)
		if (indentResult !== tt.expectedIndent) {
			print("test " + tt.name + " re-indent failed, expected: \n" + tt.expectedIndent +  "\nbut got: \n" + indentResult)
		}

		compactResult = compact(indentResult)
		if (compactResult !== tt.expectedCompact) {
			print("test " + tt.name + " compact-indent failed, expected: \n" + tt.expectedCompact +  "\nbut got: \n" + compactResult)
		}

		indentResult = indent(compactResult)
		if (indentResult !== tt.expectedIndent) {
			print("test " + tt.name + " indent-compact failed, expected: \n" + tt.expectedIndent +  "\nbut got: \n" + indentResult)
		}

		compactResult = compact(compactResult)
		if (compactResult !== tt.expectedCompact) {
			print("test " + tt.name + " re-compact failed, expected: \n" + tt.expectedCompact +  "\nbut got: \n" + compactResult)
		}
	}	
	`))

	runJsTest(t, buffer, "tests/testIndent.js")
}

func TestCompactAndRemoveComment(t *testing.T) {

	t.Parallel()

	removeCommentTests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "single line",
			input: `[{
				// some comment 
	
				// and other
				"key": 1
			}]`,
			expected: `[{"key":1}]`,
		},
		{
			name: "multi line",
			input: `[{
				"key": 1
			/**   comment 
			*
			*/}]`,
			expected: `[{"key":1}]`,
		},
		{
			name: "start of doc",
			input: `
			/**   comment 
			*
			*/db.collection.find({})`,
			expected: `db.collection.find({})`,
		},
		{
			name: "end of query",
			input: `
			db.collection.find({})/**   comment 
			*
			*/`,
			expected: `db.collection.find({})`,
		},
		{
			name: "mixed single and multiple line",
			input: `
			db.collection.find({
				// the key
				k: 1
			})/**   comment 
			*
			*/`,
			expected: `db.collection.find({k:1})`,
		},
		{
			name: "multiple line single star",
			input: `
			db.collection.find({
				/*
				 the key */

				k: 1
			})/*   comment 
			
			*/`,
			expected: `db.collection.find({k:1})`,
		},
		{
			name: "explain",
			input: `
			db.collection.find().explain( )`,
			expected: `db.collection.find().explain()`,
		},
		{
			name: "explain before aggregate",
			input: `
			db.collection.explain(
				"queryPlanner"
			).aggregate([])`,
			expected: `db.collection.explain("queryPlanner").aggregate([])`,
		},
	}

	buffer := loadPlaygroundJs(t)

	testFormat := `
	{
		"name": %s,
		"input": %s, 
		"expected": %s
	}
	`

	buffer.Write([]byte("var tests = ["))
	for _, tt := range removeCommentTests {
		fmt.Fprintf(buffer, testFormat, strconv.Quote(tt.name), strconv.Quote(tt.input), strconv.Quote(tt.expected))
		buffer.WriteByte(',')
	}
	buffer.Write([]byte(`
	]
	
	`))

	buffer.Write([]byte(`
	for (let i in tests) {
		let tt = tests[i]

		let want = tt.expected
		let got = compactAndRemoveComment(tt.input) 
		if (want !== got) {
			print("test " + tt.name + " compact and remove comment failed, expected: \n" + want +  "\nbut got: \n" + got)
		}
	}	
	`))

	runJsTest(t, buffer, "tests/testCompactAndRemoveComment.js")

}

func TestValidConfig(t *testing.T) {

	t.Parallel()

	formatTests := []struct {
		name             string
		input            string
		validModeBSON    bool
		validModeDatagen bool
	}{
		{
			name:             `valid config`,
			input:            `[{"k":1}]`,
			validModeBSON:    true,
			validModeDatagen: true,
		},
		{
			name:             `invalid config`,
			input:            `[{"k":1}`,
			validModeBSON:    false,
			validModeDatagen: false,
		},
		{
			name:             `multiple collections bson mode`,
			input:            `db={"collection1":[{"k":1}]}`,
			validModeBSON:    true,
			validModeDatagen: false,
		},
		{
			name:             `multiple collections bson mode starting with comment`,
			input:            `/** all db*/db={"collection1":[{"k":1}]}`,
			validModeBSON:    true,
			validModeDatagen: false,
		},
		{
			name:             `multiple collections bson mode ending with comment`,
			input:            `db={"collection1":[{"k":1}]}/** end*/`,
			validModeBSON:    true,
			validModeDatagen: false,
		},
	}

	buffer := loadPlaygroundJs(t)

	testFormat := `
	{
		"name": %s,
		"input": %s, 
		"validModeBSON": %v, 
		"validModeDatagen": %v
	}
	`

	buffer.Write([]byte("var tests = ["))
	for _, tt := range formatTests {
		fmt.Fprintf(buffer, testFormat, strconv.Quote(tt.name), strconv.Quote(tt.input), tt.validModeBSON, tt.validModeDatagen)
		buffer.WriteByte(',')
	}
	buffer.Write([]byte(`
	]
	
	`))

	buffer.Write([]byte(`
	for (let i in tests) {
		let tt = tests[i]

		let got = isConfigValid(tt.input, "bson") 
		if (got !== tt.validModeBSON) {
			print("test " + tt.name + " format mode bson failed, expected: " + tt.validModeBSON +  " but got: " + got)
		}

		got = isConfigValid(tt.input, "mgodatagen") 
		if (got !== tt.validModeDatagen) {
			print("test " + tt.name + " format mode mgodatagen failed, expected: " + tt.validModeDatagen +  " but got: " + got)
		}
	}	
	`))

	runJsTest(t, buffer, "tests/testFormatConfig.js")

}

func TestValidQuery(t *testing.T) {

	t.Parallel()

	formatTests := []struct {
		name  string
		input string
		valid bool
	}{
		{
			name:  `trailing comma`,
			input: `db.collection.find();`,
			valid: true,
		},
		{
			name: `correct aggregation query`,
			input: `db.collection.aggregate([
				{
					"$match": {
						_id: ObjectId("5a934e000102030405000000"), 
						k: {
							"$gt": 0.2323
						}
					}
				}
			])`,
			valid: true,
		},
		{
			name:  `wrong format`,
			input: `dbcollection.find()`,
			valid: false,
		},
		{
			name:  `invalid function`,
			input: `db.collection.findOne()`,
			valid: false,
		},
		{
			name:  `wrong format`,
			input: `dbcollection.find()`,
			valid: false,
		},
		{
			name:  `wrong format`,
			input: `db["collection"].find()`,
			valid: false,
		},
		{
			name:  `wrong format`,
			input: `db.getCollection("coll").find()`,
			valid: false,
		},
		{
			name:  `dot in query`,
			input: `db.collection.find({k: 1.123})`,
			valid: true,
		},
		{
			name:  `chained empty method`,
			input: `db.collection.find().toArray()`,
			valid: false,
		},
		{
			name:  `single letter collection name`,
			input: `db.k.find()`,
			valid: true,
		},
		{
			name:  `chained non-empty method`,
			input: `db.collection.aggregate([{"$match": { "_id": ObjectId("5a934e000102030405000000")}}]).pretty()`,
			valid: false,
		},
		{
			name: `query starting with single line comment`,
			input: `// the query
// that we want to debug
db.collection.aggregate([{"$match": { "_id": ObjectId("5a934e000102030405000000")}}])`,
			valid: true,
		},
		{
			name:  `query starting with multi line comment`,
			input: `/**  aggregation */db.collection.aggregate([{"$match": { "_id": 1}}])`,
			valid: true,
		},
		{
			name: `query ending with multi line comment`,
			input: `db.collection.aggregate([{"$match": { "_id": 1}}])/** tests
*
* ok
*/`,
			valid: true,
		},
		{
			name:  `empty comment`,
			input: `db.collection.find(/**/)`,
			valid: true,
		},
		{
			name:  `update`,
			input: `db.collection.update({"k":1},{"$set":{"a":true}})`,
			valid: true,
		},
		{
			name:  `explain`,
			input: `db.collection.find({"k":1}).explain()`,
			valid: true,
		},
		{
			name:  `explain with option`,
			input: `db.collection.find({"k":1}).explain("executionStats")`,
			valid: true,
		},
		{
			name:  `explain before find`,
			input: `db.collection.explain("queryPlanner").find({"k":1})`,
			valid: true,
		},
	}

	buffer := loadPlaygroundJs(t)

	testFormat := `
	{
		"name": %s,
		"input": %s, 
		"expected": %v, 
	}
	`

	buffer.Write([]byte("var tests = ["))
	for _, tt := range formatTests {
		fmt.Fprintf(buffer, testFormat, strconv.Quote(tt.name), strconv.Quote(tt.input), tt.valid)
		buffer.WriteByte(',')
	}
	buffer.Write([]byte(`
	]
	
	`))

	buffer.Write([]byte(`
	for (let i in tests) {
		let tt = tests[i]

		let want = tt.expected
		let got = isQueryValid(tt.input) 
		if (want != got) {
			print("test " + tt.name + " format mode bson failed, expected: " + want +  " but got: " + got)
		}
	}	
	`))

	runJsTest(t, buffer, "tests/testFormatQuery.js")
}

func TestExtractAvailableCollections(t *testing.T) {

	t.Parallel()

	formatTests := []struct {
		name                      string
		input                     string
		collectionsBsonMode       string
		collectionsMgodatagenMode string
	}{
		{
			name:                      `basic bson config default collection`,
			input:                     `[{"k":1}]`,
			collectionsBsonMode:       "collection",
			collectionsMgodatagenMode: "collection",
		},
		{
			name: `basic mgodatagen`,
			input: `[
  {
    "collection": "test",
	"count": 10
  }
]`,
			collectionsBsonMode:       "collection",
			collectionsMgodatagenMode: "test",
		},
		{
			name: `bson multiple collection`,
			input: `db={
  "orders": [
    {
      "_id": 3
    }
  ],
  "inventory": [
    {
      "_id": 1,
      "sku": "almonds",
      "description": "product 1",
      "instock": 120
    }
  ]
}`,
			collectionsBsonMode:       "orders,inventory",
			collectionsMgodatagenMode: "orders,inventory",
		},
		{
			name:                      "empty config",
			input:                     "",
			collectionsBsonMode:       "collection",
			collectionsMgodatagenMode: "collection",
		},
		{
			name: `mgodatagen config without ':'`,
			input: `[
  {
    "collection" "test",
	"count": 10
  }
]`,
			collectionsBsonMode:       "collection",
			collectionsMgodatagenMode: "collection",
		},
		{
			name: `bson multiple collection without ':'`,
			input: `db={
  "orders"
    {
      "_id": 3
    }
  ],
  "inventory": [
    {
      "_id": 1,
      "sku": "almonds",
      "description": "product 1",
      "instock": 120
    }
  ]
}`,
			collectionsBsonMode:       "inventory",
			collectionsMgodatagenMode: "inventory",
		},
		{
			name: `mgodatagen config ending with ':'`,
			input: `[
  {
    "collection":
  }
]`,
			collectionsBsonMode:       "collection",
			collectionsMgodatagenMode: "collection",
		},
	}

	buffer := loadPlaygroundJs(t)

	testFormat := `
	{
		"name": %s,
		"input": %s, 
		"expectedModeBson": %v, 
		"expectedModeMgodatagen": %v, 
	}
	`

	buffer.Write([]byte("var tests = ["))
	for _, tt := range formatTests {
		fmt.Fprintf(buffer, testFormat, strconv.Quote(tt.name), strconv.Quote(tt.input), strconv.Quote(tt.collectionsBsonMode), strconv.Quote(tt.collectionsMgodatagenMode))
		buffer.WriteByte(',')
	}
	buffer.Write([]byte(`
	]
	
	`))

	buffer.Write([]byte(`
	for (let i in tests) {
		let tt = tests[i]

		updateAvailableCollection(tt.input, "bson")

		let want = tt.expectedModeBson
		let got = availableCollections.map(c => c.value).join(",") 
		if (want != got) {
			print("test " + tt.name + " in bson mode failed, expected: " + want +  " but got: " + got)
		}

		updateAvailableCollection(tt.input, "mgodatagen")

		want = tt.expectedModeMgodatagen
		got = availableCollections.map(c => c.value).join(",") 
		if (want != got) {
			print("test " + tt.name + " in mgodatagen mode failed, expected: " + want +  " but got: " + got)
		}
	}	
	`))

	runJsTest(t, buffer, "tests/testExtractCollections.js")
}

func loadPlaygroundJs(t *testing.T) *bytes.Buffer {
	playgroundjs, err := ioutil.ReadFile("web/playground.js")
	if err != nil {
		t.Error(err)
	}
	buffer := bytes.NewBuffer(playgroundjs)
	buffer.WriteString("\n")
	return buffer
}

func runJsTest(t *testing.T, buffer *bytes.Buffer, filename string) {

	testFile, err := os.Create(filename)
	if err != nil {
		t.Error(err)
	}
	io.Copy(testFile, buffer)
	testFile.Close()
	// run the tests using mongodb javascript engine
	cmd := exec.Command("mongo", "--quiet", filename)
	var out bytes.Buffer
	cmd.Stdout = &out

	err = cmd.Run()
	if err != nil {
		t.Error(err)
	}
	result := out.String()
	if result != "" {
		t.Error(result)
	} else {
		os.Remove(filename)
	}
}
