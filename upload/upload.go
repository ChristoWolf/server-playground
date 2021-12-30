// Package upload implements simple file upload to a web server.
// It consists of a REST API endpoint with an associated static web UI.
package upload

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
)

const (
	// Url is the URL of the upload service web UI.
	Url = "/upload/"
	// ApiUrl is the REST API endpoint of the upload service.
	ApiUrl = "/api" + Url
	// staticDir is the directory containing static content
	// associated to uploading, like index.html for the upload service web UI.
	staticDir = "./upload/static/"
	// uploadDir is the directory where uploaded files are stored.
	uploadDir = "./files/"
	// formName is the name of the form field containing the file.
	formName = "file"
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
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

// POST request wrapper function.
func handlePost(w http.ResponseWriter, r *http.Request) {
	if err := handleForm(w, r); err == nil ||
		(err != nil && !errors.Is(err, http.ErrNotMultipart)) {
		return // Either the form was handled correctly or the error was already written to the response.
	}
	if err := handleOther(w, r); err != nil && errors.Is(err, http.ErrNotSupported) {
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
	}
	// Otherwise, the error was already written to the response.
}

// Handles POST requests of form-data media to the API.
func handleForm(w http.ResponseWriter, r *http.Request) error {
	// Check if the media type is multipart/form-data.
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		return http.ErrNotMultipart
	}
	// Parse the form.
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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
	// Check if the media type is different from multipart/form-data
	// as that should have been handled beforehand.
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		return http.ErrNotSupported
	}
	// To provide an extension to the file name,
	// we will use the file extension
	// which matches the content-type header.
	// WE do NOT use http.DetectContentType here,
	// as that requires us to read from the body,
	// which would remove those read bytes from the io.Reader.
	contentType := r.Header.Get("Content-Type")
	typeCandidates, err := mime.ExtensionsByType(contentType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	var ext string
	if len(typeCandidates) > 0 {
		preferred := strings.SplitAfter(contentType, "/")[1]
		for _, candidate := range typeCandidates {
			// Match with the preferred one.
			if strings.HasSuffix(candidate, preferred) {
				ext = candidate
				break
			}
		}
		// If none matched, use the first one.
		if ext == "" {
			ext = typeCandidates[0]
		}
	}
	// generate a random UUID for the file name.
	uuid := uuid.NewString()
	path := uploadDir + uuid + ext
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
