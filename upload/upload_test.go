// Package upload_test provides a test suite for the upload package.
package upload_test

import (
	"bytes"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/christowolf/server-playground/upload"
)

const (
	ctMultipart = "multipart/form-data"
	uri         = upload.ApiUrl
)

// TestApiEndpointForm tests the upload API endpoint by posting a form containing a file.
func TestApiEndpointHandlerForm(t *testing.T) {
	// t.Parallel()
	fileName := "test.txt"
	expectedPath := upload.UploadDir + fileName
	// Register the file cleanup.
	fileCleanup(t, expectedPath)
	testContent := "test content form"
	// Create a new request.
	// For this, we also need an appropriate request body,
	// which can be built using mime/multipart.
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fileWriter, _ := writer.CreateFormFile("file", fileName)
	content := strings.NewReader(testContent)
	io.Copy(fileWriter, content)
	writer.Close()
	r := httptest.NewRequest(http.MethodPost, "http://testdomain.com"+uri, body)
	r.Header.Add("Content-Type", writer.FormDataContentType())
	// Create a new recorder.
	w := httptest.NewRecorder()
	// Call the API endpoint.
	sut := upload.ApiEndpoint()
	sut.ServeHTTP(w, r)
	// Check the response code.
	if w.Code != http.StatusCreated {
		t.Errorf("expected status code: %v, got: %v", http.StatusCreated, w.Code)
	}
	// Check if the file was uploaded correctly.
	if _, err := os.Stat(expectedPath); err != nil {
		t.Fatalf("expected file: %v, got error: %v", expectedPath, err)
	}
	// Check if the file content is correct.
	gotContent, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Errorf("expected file: %v, got error: %v", expectedPath, err)
	}
	if string(gotContent) != testContent {
		t.Errorf("expected file content: %v, got: %v", testContent, string(gotContent))
	}
}

// TestApiEndpointForm tests the upload API endpoint by posting binary content.
func TestApiEndpointHandlerOther(t *testing.T) {
	fileName := "test.txt"
	expectedPath := upload.UploadDir + fileName
	// Register the file cleanup.
	fileCleanup(t, expectedPath)
	testContent := "test content binary"
	// Create a new request.
	// For this, we also need an appropriate request body
	// containing binary data.
	content := strings.NewReader(testContent)
	r := httptest.NewRequest(http.MethodPost, "http://testdomain.com"+uri, content)
	ct := mime.TypeByExtension(filepath.Ext(fileName))
	r.Header.Add("Content-Type", ct)
	// Create a new recorder.
	w := httptest.NewRecorder()
	// Call the API endpoint.
	sut := upload.ApiEndpoint()
	sut.ServeHTTP(w, r)
	// Check the response code.
	if w.Code != http.StatusCreated {
		t.Errorf("expected status code: %v, got: %v", http.StatusCreated, w.Code)
	}
	// Check if the file was uploaded correctly.
	if _, err := os.Stat(expectedPath); err != nil {
		t.Fatalf("expected file: %v, got error: %v", expectedPath, err)
	}
	// Check if the file content is correct.
	gotContent, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Errorf("expected file: %v, got error: %v", expectedPath, err)
	}
	if string(gotContent) != testContent {
		t.Errorf("expected file content: %v, got: %v", testContent, string(gotContent))
	}
}

// fileCleanup executes file cleanup after test execution.
// If an error is encountered during os.Remove,
// it is communicated to the testing.T instance.
func fileCleanup(t *testing.T, path string) {
	t.Cleanup(func() {
		if err := os.Remove(path); err != nil {
			t.Errorf("cleanup failed for file: %v, got error: %v", path, err)
		}
	})
}
