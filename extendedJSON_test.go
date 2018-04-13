package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/globalsign/mgo/bson"
)

func TestExtendedJSON(t *testing.T) {
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

func TestJavascriptIndent(t *testing.T) {

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

	var buffer bytes.Buffer
	playgroundjs, err := os.Open("web/playground.js")
	if err != nil {
		t.Error(err)
	}
	io.Copy(&buffer, playgroundjs)
	playgroundjs.Close()

	testFuncFormat := `
	{
		"name": %s,
		"input": %s, 
		"expectedIndent": %s, 
		"expectedCompact": %s
	}
	`

	buffer.Write([]byte("var tests = ["))
	for i, tt := range jsIndentTests {
		fmt.Fprintf(&buffer, testFuncFormat, strconv.Quote(tt.name), strconv.Quote(tt.input), strconv.Quote(tt.indent), strconv.Quote(tt.compact))
		if i != len(jsIndentTests) {
			buffer.WriteByte(',')
		}
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

		let indentResult = indent(tt.input, indentMode)
		if (indentResult !== tt.expectedIndent) {
			print("test " + tt.name + " failed, expected: \n" + tt.expectedIndent +  "\nbut got: \n" + indentResult)
		}
		let compactResult = indent(tt.input, compactMode)
		if (compactResult !== tt.expectedCompact) {
			print("test " + tt.name + " failed, expected: \n" + tt.expectedCompact +  "\nbut got: \n" + compactResult)
		}

		indentResult = indent(indentResult, indentMode)
		if (indentResult !== tt.expectedIndent) {
			print("test " + tt.name + " re-indent failed, expected: \n" + tt.expectedIndent +  "\nbut got: \n" + indentResult)
		}

		compactResult = indent(indentResult, compactMode)
		if (compactResult !== tt.expectedCompact) {
			print("test " + tt.name + " compact-indent failed, expected: \n" + tt.expectedCompact +  "\nbut got: \n" + compactResult)
		}

		indentResult = indent(compactResult, indentMode)
		if (indentResult !== tt.expectedIndent) {
			print("test " + tt.name + " indent-compact failed, expected: \n" + tt.expectedIndent +  "\nbut got: \n" + indentResult)
		}

		compactResult = indent(compactResult, compactMode)
		if (compactResult !== tt.expectedCompact) {
			print("test " + tt.name + " re-compact failed, expected: \n" + tt.expectedCompact +  "\nbut got: \n" + compactResult)
		}
	}	
	`))

	testFile, err := os.Create("tests/test.js")
	if err != nil {
		t.Error(err)
	}
	io.Copy(testFile, &buffer)
	testFile.Close()

	// run the tests using mongodb javascript engine
	cmd := exec.Command("mongo", "--quiet", "tests/test.js")
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
		os.Remove("tests/test.js")
	}
}
