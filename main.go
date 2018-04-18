package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	l := log.New(os.Stdout, "", log.LstdFlags)
	s, err := newServer(l)
	if err != nil {
		l.Fatalf("aborting: %v\n", err)
	}
	l.Fatal(http.ListenAndServe(":80", s))
}
