package main

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
)

const (
	mgodatagenMode byte = iota
	jsonMode
)

func modeByte(mode string) byte {
	if mode == "json" {
		return jsonMode
	}
	return mgodatagenMode
}

type page struct {
	Mode byte
	// configuration used to generate the sample database
	Config []byte
	// query to run against the collection / database
	Query []byte
	// mongodb version
	MongoVersion []byte
}

// generate an unique url for this page
func (p *page) ID() []byte {
	e := sha256.New()
	e.Write([]byte{p.Mode})
	e.Write(p.Query)
	e.Write(p.Config)
	sum := e.Sum(nil)
	b := make([]byte, base64.URLEncoding.EncodedLen(len(sum)))
	base64.URLEncoding.Encode(b, sum)
	return b[:11]
}

// generate an unique hash to identify the database used by the p page. Two pages with
// same config and mode should generate the same dbHash
func (p *page) dbHash() string {
	return fmt.Sprintf("%x", md5.Sum(append(p.Config, p.Mode)))
}

func (p *page) String() string {
	mode := "json"
	if p.Mode != jsonMode {
		mode = "mgodatagen"
	}
	return fmt.Sprintf("mode: %s\nconfig: %s\nquery: %s\n", mode, p.Config, p.Query)
}

// encode a page into a []byte
//
// v[0:4] -> an int32 to store the position of the last byte of the configuration
// v[4] -> the mode (mgodatagen / json) to use for building the database
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
func (p *page) decode(val []byte) {
	endConfig := binary.LittleEndian.Uint32(val[0:4])
	p.Mode = val[4]
	p.Config = val[5:endConfig]
	p.Query = val[endConfig:]
}
