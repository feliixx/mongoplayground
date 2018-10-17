package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/globalsign/mgo/bson"
)

func TestExtendedJSON(t *testing.T) {

	t.Parallel()

	extendedJSONTests := []struct {
		a interface{}
		b string
	}{
		{
			a: bson.ObjectIdHex("5a934e000102030405000000"),
			b: `ObjectId("5a934e000102030405000000")`,
		},
		{
			a: bson.MongoTimestamp(4294967298),
			b: `Timestamp(1,2)`,
		},
		{
			a: time.Date(2016, 5, 15, 1, 2, 3, 4000000, time.UTC),
			b: `ISODate("2016-05-15T01:02:03.004Z")`,
		}, {
			a: time.Date(2016, 5, 15, 1, 2, 3, 4000000, time.FixedZone("CET", 60*60)),
			b: `ISODate("2016-05-15T01:02:03.004+01:00")`,
		},
		{
			a: bson.Binary{Kind: 2, Data: []byte("foo")},
			b: `BinData(2,"Zm9v")`,
		},
		{
			a: bson.Undefined,
			b: `undefined`,
		},
		{
			a: int64(10),
			b: `10`,
		}, {
			a: int(1),
			b: `1`,
		}, {
			a: int32(26),
			b: `NumberInt(26)`,
		},
	}

	for _, tt := range extendedJSONTests {
		b, err := bson.MarshalExtendedJSON(tt.a)
		if err != nil {
			t.Errorf("fail to unmarshal %v: %v", tt.a, err)
		}
		if want, got := tt.b, strings.TrimSuffix(string(b), "\n"); want != got {
			t.Errorf("expected %s, but got %s", want, got)
		}
	}
}

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
			name: `extended JSON`,
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

func TestFormatConfig(t *testing.T) {

	t.Parallel()

	formatTests := []struct {
		name                 string
		input                string
		formattedModeJSON    string
		formattedModeDatagen string
	}{
		{
			name:                 `valid config`,
			input:                `[{"k":1}]`,
			formattedModeJSON:    `[{"k":1}]`,
			formattedModeDatagen: `[{"k":1}]`,
		},
		{
			name:                 `invalid config`,
			input:                `[{"k":1}`,
			formattedModeJSON:    `invalid`,
			formattedModeDatagen: `invalid`,
		},
		{
			name:                 `multiple collections json mode`,
			input:                `{"collection1":[{"k":1}]}`,
			formattedModeJSON:    `{"collection1":[{"k":1}]}`,
			formattedModeDatagen: `invalid`,
		},
	}

	buffer := loadPlaygroundJs(t)

	testFormat := `
	{
		"name": %s,
		"input": %s, 
		"formattedModeJSON": %s, 
		"formattedModeDatagen": %s 
	}
	`

	buffer.Write([]byte("var tests = ["))
	for _, tt := range formatTests {
		fmt.Fprintf(buffer, testFormat, strconv.Quote(tt.name), strconv.Quote(tt.input), strconv.Quote(tt.formattedModeJSON), strconv.Quote(tt.formattedModeDatagen))
		buffer.WriteByte(',')
	}
	buffer.Write([]byte(`
	]
	
	`))

	buffer.Write([]byte(`
	for (let i in tests) {
		let tt = tests[i]

		let got = formatConfig(tt.input, "json") 
		if (got !== tt.formattedModeJSON) {
			print("test " + tt.name + " format mode json failed, expected: \n" + tt.formattedModeJSON +  "\nbut got: \n" + got)
		}

		got = formatConfig(tt.input, "mgodatagen") 
		if (got !== tt.formattedModeDatagen) {
			print("test " + tt.name + " format mode mgodatagen failed, expected: \n" + tt.formattedModeDatagen +  "\nbut got: \n" + got)
		}
	}	
	`))

	runJsTest(t, buffer, "tests/testFormatConfig.js")

}

func TestFormatQuery(t *testing.T) {

	t.Parallel()

	formatTests := []struct {
		name                 string
		input                string
		formattedModeJSON    string
		formattedModeDatagen string
	}{
		{
			name:                 `trailing comma`,
			input:                `db.collection.find();`,
			formattedModeJSON:    `db.collection.find()`,
			formattedModeDatagen: `db.collection.find()`,
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
			formattedModeJSON: `db.collection.aggregate([
				{
					"$match": {
						_id: ObjectId("5a934e000102030405000000"), 
						k: {
							"$gt": 0.2323
						}
					}
				}
			])`,
			formattedModeDatagen: `db.collection.aggregate([
				{
					"$match": {
						_id: ObjectId("5a934e000102030405000000"), 
						k: {
							"$gt": 0.2323
						}
					}
				}
			])`,
		},
		{
			name:                 `wrong format`,
			input:                `dbcollection.find()`,
			formattedModeJSON:    `invalid`,
			formattedModeDatagen: `invalid`,
		},
		{
			name:                 `invalid function`,
			input:                `db.collection.findOne()`,
			formattedModeJSON:    `invalid`,
			formattedModeDatagen: `invalid`,
		},
		{
			name:                 `wrong format`,
			input:                `dbcollection.find()`,
			formattedModeJSON:    `invalid`,
			formattedModeDatagen: `invalid`,
		},
		{
			name:                 `wrong format`,
			input:                `db["collection"].find()`,
			formattedModeJSON:    `invalid`,
			formattedModeDatagen: `invalid`,
		},
		{
			name:                 `wrong format`,
			input:                `db.getCollection("coll").find()`,
			formattedModeJSON:    `invalid`,
			formattedModeDatagen: `invalid`,
		},
		{
			name:                 `dot in query`,
			input:                `db.collection.find({k: 1.123})`,
			formattedModeJSON:    `db.collection.find({k: 1.123})`,
			formattedModeDatagen: `db.collection.find({k: 1.123})`,
		},
		{
			name:                 `chained empty method`,
			input:                `db.collection.find().toArray()`,
			formattedModeJSON:    `invalid`,
			formattedModeDatagen: `invalid`,
		},
		{
			name:                 `single letter collection name`,
			input:                `db.k.find()`,
			formattedModeJSON:    `db.k.find()`,
			formattedModeDatagen: `db.k.find()`,
		},
		{
			name:                 `chained non-empty method`,
			input:                `db.collection.aggregate([{"$match": { "_id": ObjectId("5a934e000102030405000000")}}]).explain("executionTimeMillis")`,
			formattedModeJSON:    `invalid`,
			formattedModeDatagen: `invalid`,
		},
	}

	buffer := loadPlaygroundJs(t)

	testFormat := `
	{
		"name": %s,
		"input": %s, 
		"formattedModeJSON": %s, 
		"formattedModeDatagen": %s 
	}
	`

	buffer.Write([]byte("var tests = ["))
	for _, tt := range formatTests {
		fmt.Fprintf(buffer, testFormat, strconv.Quote(tt.name), strconv.Quote(tt.input), strconv.Quote(tt.formattedModeJSON), strconv.Quote(tt.formattedModeDatagen))
		buffer.WriteByte(',')
	}
	buffer.Write([]byte(`
	]
	
	`))

	buffer.Write([]byte(`
	for (let i in tests) {
		let tt = tests[i]

		let got = formatQuery(tt.input, "json") 
		if (got !== tt.formattedModeJSON) {
			print("test " + tt.name + " format mode json failed, expected: \n" + tt.formattedModeJSON +  "\nbut got: \n" + got)
		}

		got = formatQuery(tt.input, "mgodatagen") 
		if (got !== tt.formattedModeDatagen) {
			print("test " + tt.name + " format mode mgodatagen failed, expected: \n" + tt.formattedModeDatagen +  "\nbut got: \n" + got)
		}
	}	
	`))

	runJsTest(t, buffer, "tests/testFormatQuery.js")

}

func loadPlaygroundJs(t *testing.T) *bytes.Buffer {
	playgroundjs, err := ioutil.ReadFile("web/playground.js")
	if err != nil {
		t.Error(err)
	}
	return bytes.NewBuffer(playgroundjs)
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
