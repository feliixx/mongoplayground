package main

import (
	"time"

	"github.com/feliixx/mgodatagen/datagen/generators"
	"github.com/globalsign/mgo/bson"
)

var (
	// use a constant time for objectId generation
	t = uint32(time.Date(2018, 02, 26, 0, 0, 0, 0, time.UTC).Unix())
)

func objectIDBytes(n int32) []byte {
	return []byte{
		byte(t >> 24),
		byte(t >> 16),
		byte(t >> 8),
		byte(t),
		byte(1), // 1,2,3 for hostname bytes
		byte(2),
		byte(3),
		byte(4), // 4,5 for pid bytes
		byte(5),
		byte(n >> 16), // Increment, 3 bytes, big endian
		byte(n >> 8),
		byte(n),
	}
}

// seededObjectIDGenerator generator creating always the same sequence of
// bson objectID
type seededObjectIDGenerator struct {
	key []byte
	idx int32
	buf *generators.DocBuffer
}

func (g *seededObjectIDGenerator) Key() []byte  { return g.key }
func (g *seededObjectIDGenerator) Exists() bool { return true }
func (g *seededObjectIDGenerator) Type() byte   { return bson.ElementObjectId }

// Value encode an objectId in the encoder
func (g *seededObjectIDGenerator) Value() {
	g.buf.Write(objectIDBytes(g.idx))
	g.idx++
}
