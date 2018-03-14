package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	s, err := newServer()
	if err != nil {
		fmt.Printf("aborting: %v\n", err)
		os.Exit(1)
	}
	log.Fatal(http.ListenAndServe(":80", s))
}
