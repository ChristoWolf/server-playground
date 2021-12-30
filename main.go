package main

import (
	"log"
	"net/http"

	"github.com/christowolf/server-playground/upload"
)

// Main entry point of server playground.
func main() {
	// Instantiate a mux for registering handlers.
	mux := http.NewServeMux()
	// Register the upload API route.
	mux.Handle(upload.ApiUrl, upload.ApiEndpoint())
	// Start the server.
	log.Fatal(http.ListenAndServe(":8080", mux))
}
