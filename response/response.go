// Package response provides types and functions for handling
// API response DTOs and their payloads, usually encoded as JSON.
package response

import (
	"mime"
	"path/filepath"
)

type JsonDto struct {
	Status      int      `json:"status"`
	Message     string   `json:"message"`
	ErrorString string   `json:"error,omitempty"`
	File        *FileDto `json:"file,omitempty"`
}

type FileDto struct {
	Name     string `json:"name"`
	MimeType string `json:"mime_type"`
}

func NewFileDto(fileName string) *FileDto {
	name := filepath.Clean(filepath.Base(fileName))
	mimeType := mime.TypeByExtension(filepath.Ext(name))
	return &FileDto{Name: name, MimeType: mimeType}
}
