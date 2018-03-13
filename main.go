package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	s, err := newServer()
	if err != nil {
		fmt.Printf("aborting: %v\n", err)
		os.Exit(1)
	}
	http.ListenAndServe(":8080", s)
}
