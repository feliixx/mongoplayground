package main

import (
	"github.com/feliixx/mgodatagen/generators"
	"time"
)

var (
	t = uint32(time.Date(2018, 02, 26, 0, 0, 0, 0, time.UTC).Unix())
)

func objectIDBytes(n int32) []byte {
	return []byte{
		byte(t >> 24),
		byte(t >> 16),
		byte(t >> 8),
		byte(t),
		byte(1), // Machine, first 3 bytes of md5(hostname)
		byte(2),
		byte(3),
		byte(4), // Pid, 2 bytes, specs don't specify endianness, but we use big endian.
		byte(5),
		byte(n >> 16), // Increment, 3 bytes, big endian
		byte(n >> 8),
		byte(n),
	}
}

// SeededObjectIDGenerator generator creating always the same sequence of
// bson objectID
type SeededObjectIDGenerator struct {
	generators.EmptyGenerator
	idx int32
}

// Value encode an objectId in the encoder
func (g *SeededObjectIDGenerator) Value() {
	g.Out.Write(objectIDBytes(g.idx))
	g.idx++
}
