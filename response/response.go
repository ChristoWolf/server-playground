// Package response provides types and functions for handling
// API response DTOs and their payloads, usually encoded as JSON.
package response

import (
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"path/filepath"
)

// JsonDto is the response DTO for JSON responses.
// It contains a request's status code, some message,
// an error string and a pointer to a FileDto, if applicable.
// This struct defines the JSON schema of the response
// and can be marshalled/unmarshalled using
// the json standard library package.
type JsonDto struct {
	Status      uint16   `json:"status"`
	Message     string   `json:"message"`
	ErrorString string   `json:"error,omitempty"`
	File        *FileDto `json:"file,omitempty"`
}

// FileDto is a DTO for defining the JSON schema of
// certain file information (e.g. of an uploaded file)
// which can be optionally nested in a JsonDto
// to provide relevant file information in a response.
type FileDto struct {
	Name     string `json:"name"`
	MimeType string `json:"mime_type"`
}

// NewFileDto creates a new FileDto instance.
// The media/mime-type of the file is automatically determined
// via the file extension of the given file name.
func NewFileDto(fileName string) *FileDto {
	name := filepath.Clean(filepath.Base(fileName))
	mimeType := mime.TypeByExtension(filepath.Ext(name))
	return &FileDto{Name: name, MimeType: mimeType}
}

// Error writes various error information to the response in JSON format.
// Based on http.Error.
func Error(w http.ResponseWriter, error string, code uint16) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(int(code))
	resp := &JsonDto{
		Status:      code,
		ErrorString: error,
	}
	jsonData, err := json.Marshal(resp)
	if err != nil {
		panic(err)
	}
	fmt.Fprint(w, string(jsonData))
}
