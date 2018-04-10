package main

import (
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
