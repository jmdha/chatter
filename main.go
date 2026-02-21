package main

import (
	"fmt"
	"flag"
	"log"
	"net/http"
)

func main() {
	var port int

	// Parse arguments
	flag.IntVar(&port, "p", 8080, "port to operate on")
	flag.Parse()

	// Initialize stuff
	handler := NewHandler()

	// Run
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), &handler))
}
