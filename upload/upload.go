// Package upload implements simple file upload to a web server.
// It consists of a REST API endpoint with an associated static web UI.
package upload

import (
	"crypto/sha256"
	"io"
	"mime"
	"net/http"
	"os"
)

const (
	// Url is the URL of the upload service web UI.
	Url = "/upload/"
	// ApiUrl is the REST API endpoint of the upload service.
	ApiUrl = "/api" + Url
	// staticDir is the directory containing static content
	// like index.html and the upload directory.
	staticDir = "./static/"
	// uploadDir is the directory where uploaded files are stored.
	uploadDir = staticDir + "upload/"
	// formName is the name of the form field containing the file.
	formName = "inputFile"
)

// Api returns an http.Handler that serves the upload API.
func ApiEndpoint() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		// case "GET":
		// 	http.ServeFile(w, r, staticDir + "index.html")
		case http.MethodPost:
			handlePost(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

// POST request wrapper function.
func handlePost(w http.ResponseWriter, r *http.Request) {
	if err := handleForm(w, r); err != nil && err != http.ErrNotMultipart {
		return // Error should already be handled.
	}
	if err := handleOther(w, r); err != nil {
		if err == http.ErrNotSupported {
			http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		}
	}
}

// Handles POST requests of form-data media to the API.
func handleForm(w http.ResponseWriter, r *http.Request) error {
	// Check if the media type is multipart/form-data.
	if r.Header.Get("Content-Type") != "multipart/form-data" {
		return http.ErrNotMultipart
	}
	// Parse the form.
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return err
	}
	// Get the file from the form and open it.
	file, handler, err := r.FormFile(formName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	defer file.Close()
	path := uploadDir + handler.Filename
	// Write the form file to a new file.
	if err := handleFile(w, path, file); err != nil {
		return err
	}
	return nil
}

// Handles POST requests of non-form-data media to the API.
func handleOther(w http.ResponseWriter, r *http.Request) error {
	// Check if the media type is different from multipart/form-data.
	if r.Header.Get("Content-Type") == "multipart/form-data" {
		return http.ErrNotSupported
	}
	// Detect the file type.
	// For that, only the first 512 bytes are needed.
	buffer := make([]byte, 512)
	_, err := r.Body.Read(buffer)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	contentType := http.DetectContentType(buffer)
	// We will use an arbitrary file extension
	// which matches the detected content type.
	typeCandidates, err := mime.ExtensionsByType(contentType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	var ext string
	if typeCandidates != nil || len(typeCandidates) <= 0 {
		ext = typeCandidates[0]
	}
	// Compute the file's hash for the file name.
	hash := sha256.New()
	if _, err := io.Copy(hash, r.Body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	path := uploadDir + string(hash.Sum(nil)) + ext
	// Write the request body to a new file.
	if err := handleFile(w, path, r.Body); err != nil {
		return err
	}
	return nil
}

// Re-usable file handling function which takes care of
// writing a container's content to a new file in the upload directory.
// If the directory does not exist, it is created.
func handleFile(w http.ResponseWriter, filePath string, container io.ReadCloser) error {
	// Create the upload directory if it does not exist yet.
	if err := createUploadDir(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	// Create a new file in the upload directory.
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	defer f.Close()
	// Copy the container's content to the new file.
	if _, err := io.Copy(f, container); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	// Set an appropriate status code and return.
	w.WriteHeader(http.StatusCreated)
	return nil
}

// Creates the upload directory if it does not exist yet.
func createUploadDir() error {
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		err := os.Mkdir(uploadDir, 0777)
		if err != nil {
			return err
		}
	}
	return nil
}
