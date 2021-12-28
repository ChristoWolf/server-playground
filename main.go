package main

import (
	"log"
	"net/http"

	"github.com/christowolf/server-playground/upload"
)

// Main entry point of server playground.
func main() {
	// // Register the file server route.
	// http.Handle("/", http.FileServer(http.Dir("./static")))
	// Register the upload API route.
	http.Handle(upload.ApiUrl, upload.ApiEndpoint())
	// Start the server.
	log.Fatal(http.ListenAndServe(":8080", nil))
}
