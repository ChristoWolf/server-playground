package main

import (
	"io"
	"log"
	"net/http"
	"os"
)

// Main entry point of server playground.
func main() {
	// Register the file server route.
	http.Handle("/", http.FileServer(http.Dir("./static")))
	// Register the upload route.
	http.HandleFunc("/upload/", handleUpload)
	// Start the server.
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Handles requests for uploading files.
func handleUpload(w http.ResponseWriter, r *http.Request) {
	// Parse the form.
	// Of course, this only works if the content-type is multipart/form-data.
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	}
	// Get the file from the form.
	// Closing of the file is deferred.
	file, handler, err := r.FormFile("myFile")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()
	fname := handler.Filename
	// Create a new file.
	// Closing of the file is deferred.
	f, err := os.OpenFile("./static/"+fname, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()
	// Copy the parsed content to the new file.
	io.Copy(f, file)
	// Redirect back to the beginning.
	http.Redirect(w, r, "/", http.StatusFound)
}
