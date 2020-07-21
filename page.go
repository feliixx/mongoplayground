package main

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	// main modes for playground
	mgodatagenMode byte = iota
	bsonMode
	// detail modes for bson playground
	bsonSingleCollection
	bsonMultipleCollection
	unknown

	mgodatagenLabel             = "mgodatagen"
	bsonLabel                   = "bson"
	bsonSingleCollectionLabel   = "bson_single_collection"
	bsonMultipleCollectionLabel = "bson_multiple_collection"
	unknownLabel                = "unknown"

	// max size of a playground. This value is the minimum we can
	// set to avoid breaking already saved playground
	maxByteSize = 350 * 1000
	// length of the id of a page. Do not change this value
	pageIDLength = 11
)

type page struct {
	Mode byte
	// configuration used to generate the sample database
	Config []byte
	// query to run against the collection / database
	Query []byte
	// mongodb version
	MongoVersion []byte
}

func newPage(modeName, config, query string) (*page, error) {

	if (len(config) + len(query)) > maxByteSize {
		return nil, errors.New(errPlaygroundToBig)
	}
	mode := bsonMode
	if modeName == mgodatagenLabel {
		mode = mgodatagenMode
	}
	return &page{
		Mode:   mode,
		Config: []byte(config),
		Query:  []byte(query),
	}, nil
}

// generate an unique id for this page
func (p *page) ID() []byte {
	e := sha256.New()
	e.Write([]byte{p.Mode})
	e.Write(p.Query)
	e.Write(p.Config)
	sum := e.Sum(nil)
	b := make([]byte, base64.URLEncoding.EncodedLen(len(sum)))
	base64.URLEncoding.Encode(b, sum)
	return b[:pageIDLength]
}

// generate an unique hash to identify the database used by the p page. Two pages with
// same config and mode should generate the same dbHash
func (p *page) dbHash() string {
	return fmt.Sprintf("%x", md5.Sum(append(p.Config, p.Mode)))
}

// encode a page into a byte slice
//
// v[0:4] -> an int32 to store the position of the last byte of the configuration
// v[4] -> the mode (mgodatagen / bson) to use for building the database
// v[5:endConfig] -> the configuration
// v[endConfig:] -> the query
func (p *page) encode() []byte {
	v := make([]byte, 5+len(p.Config)+len(p.Query))

	endConfig := len(p.Config) + 5
	binary.LittleEndian.PutUint32(v[0:4], uint32(endConfig))

	v[4] = p.Mode
	copy(v[5:endConfig], p.Config)
	copy(v[endConfig:], p.Query)
	return v
}

// decode a slice of byte into the p page
func (p *page) decode(v []byte) {
	endConfig := binary.LittleEndian.Uint32(v[0:4])
	p.Mode = v[4]
	p.Config = v[5:endConfig]
	p.Query = v[endConfig:]
}

// returns a label for the page for prometheus metrics
func (p *page) label() string {

	if p.Mode == mgodatagenMode {
		return mgodatagenLabel
	}
	if p.Mode == bsonMode {

		switch detailBsonMode(p.Config) {
		case bsonSingleCollection:
			return bsonSingleCollectionLabel
		case bsonMultipleCollection:
			return bsonMultipleCollectionLabel
		}
	}
	return unknownLabel
}

func detailBsonMode(config []byte) byte {
	if bytes.HasPrefix(config, []byte{'['}) {
		return bsonSingleCollection
	}
	if bytes.HasPrefix(config, []byte("db={")) {
		return bsonMultipleCollection
	}
	return unknown
}
