package main

import (
	"strings"
	"testing"
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/stretchr/testify/assert"
)

func TestExtendedJSON(t *testing.T) {
	l := []struct {
		a interface{}
		b string
	}{
		{
			a: bson.ObjectIdHex("5a934e000102030405000000"),
			b: `ObjectId("5a934e000102030405000000")`,
		},
		{
			a: bson.MongoTimestamp(4294967298),
<<<<<<< HEAD
			b: `Timestamp(1, 2)`,
=======
			b: `Timestamp(1,2)`,
>>>>>>> acf1d42... allow extended JSON as input
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
<<<<<<< HEAD
			b: `BinData(2, "Zm9v")`,
=======
			b: `BinData(2,"Zm9v")`,
>>>>>>> acf1d42... allow extended JSON as input
		},
		{
			a: bson.Undefined,
			b: `undefined`,
		},
<<<<<<< HEAD
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
=======
>>>>>>> acf1d42... allow extended JSON as input
	}

	for _, c := range l {
		out, err := bson.MarshalIndentExtendedJSON(c.a)
		assert.Nil(t, err)
<<<<<<< HEAD
		assert.Equal(t, c.b, strings.TrimSuffix(string(out), "\n"))
=======
		compactOut, err := bson.CompactJSON(out)
		assert.Nil(t, err)
		assert.Equal(t, c.b, strings.TrimSuffix(string(compactOut), "\n"))
>>>>>>> acf1d42... allow extended JSON as input
	}
}
