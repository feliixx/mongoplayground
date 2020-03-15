package extendedjson_test

import (
	"testing"
	"time"

	"github.com/feliixx/mongoplayground/extendedjson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestExtendedJSON(t *testing.T) {

	t.Parallel()
	objectID, _ := primitive.ObjectIDFromHex("5a934e000102030405000000")

	extendedJSONTests := []struct {
		a interface{}
		b string
	}{
		{
			a: objectID,
			b: `ObjectId("5a934e000102030405000000")`,
		},
		{
			a: primitive.Timestamp{T: 1, I: 2},
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
			a: primitive.Binary{Subtype: 2, Data: []byte("foo")},
			b: `BinData(2,"Zm9v")`,
		},
		{
			a: primitive.Undefined{},
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
		b, err := extendedjson.Marshal(tt.a)
		if err != nil {
			t.Errorf("fail to unmarshal %v: %v", tt.a, err)
		}

		if want, got := tt.b, string(b); want != got {
			t.Errorf("expected %s, but got %s", want, got)
		}
	}
}
