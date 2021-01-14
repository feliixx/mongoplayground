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

	// if the query is an update, the base collection will change, which can
	// mess up things if a find() is run after an update() with the same config
	// for example, when running the two following playgrounds successively
	//
	// {
	//  "config": {"n": 1},
	//  "query": "db.collection.update({},{"$set":{"updated":true}})"
	// }
	//
	// {
	//  "config": {"n": 1},
	//  "query": "db.collection.find()"
	// }
	//
	// output would be {"n": 1, "updated": true}, witch is incorrect for the
	// second playground.
	//
	// to avoid this, add an extra byte when computing the hashsum, so the two above
	// playgrounds get different database
	if bytes.Contains(p.Query, []byte(".update(")) {
		return fmt.Sprintf("%x", md5.Sum(append(p.Config, p.Mode, 0)))
	}

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
