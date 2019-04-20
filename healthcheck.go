package main

import (
	"bytes"
	"fmt"
	"net/http"
)

var statusOK = []byte(`{"status":"ok"}"`)

func (s *server) healthcheckHandler(w http.ResponseWriter, r *http.Request) {

	p := &page{
		Mode:   bsonMode,
		Config: []byte(`[{"_id":1}]`),
		Query:  []byte(templateQuery),
	}

	w.Header().Set("Content-Type", "encoding/json")

	result, err := s.run(p)
	if err != nil || bytes.Compare(bytes.TrimSuffix(result, []byte("\n")), p.Config) != 0 {
		fmt.Fprintf(w, `{"status":"unexpected result: (err: %v, result: %s"}`, err, result)
		return
	}
	w.Write(statusOK)
}
