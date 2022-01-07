// Package response_test provides a test suite for the response package.
package response_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"testing/quick"

	"github.com/christowolf/server-playground/response"
)

// TestMarshalUnmarshal tests the marshaling/unmarshaling/from of a response DTO (JsonDto) to JSON.
//
// The test checks if
//
// - the marshalled JSON contains the status code, message, error string and file name,
//
// - and the unmarshalled DTO represents the original DTO.
//
// These units are tested in combination here to prevent redundant code and computation.
func TestMarshalUnmarshal(t *testing.T) {
	t.Parallel()
	var dtos = []*response.JsonDto{
		{
			Status:  http.StatusOK,
			Message: http.StatusText(http.StatusOK),
		},
		{
			Status:  http.StatusCreated,
			Message: http.StatusText(http.StatusCreated) + ": file created",
			File:    response.NewFileDto("test.txt"),
		},
		{
			Status:      http.StatusNotFound,
			ErrorString: errors.New("error content").Error(),
		},
	}
	for _, dto := range dtos {
		dto := dto
		t.Run(fmt.Sprintf("%v", dto), func(t *testing.T) {
			t.Parallel()
			// Act: Marshal the DTO as JSON.
			jsonData, err := json.Marshal(dto)
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			// Check if the marshaled JSON contains the correct values.
			failed := checkValues(t, string(jsonData), dto)
			if failed {
				t.Errorf("value(s) not found in JSON: %v", string(jsonData))
			}
			// Act: Unmarshal the JSON to a DTO.
			got := &response.JsonDto{}
			err = json.Unmarshal(jsonData, got)
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			// Check if the unmarshalled DTO represents the original DTO.
			if !reflect.DeepEqual(got, dto) {
				t.Errorf("expected: %v, got: %v", dto, got)
			}
		})
	}
}

// TestMarshalUnmarshalPropertySpec applies property based testing
// to probe for inputs which provoke marshalling errors or
// mismatches between the original and marshalled + unmarshalled DTO.
func TestMarshalUnmarshalProperty(t *testing.T) {
	t.Parallel()
	c := &quick.Config{MaxCount: 100000}
	f := marshalUnmarshalPropertySpec
	if err := quick.Check(f, c); err != nil {
		t.Error(err)
	}
}

// TestNewFileDto tests the creation of a new FileDto.
func TestNewFileDto(t *testing.T) {
	t.Parallel()
	paths := []string{
		"test.txt",
		"./Test.jpg",
		"/home/user/test.bsh",
		"C:\\Users\\user\\TEST.docx",
	}
	for _, path := range paths {
		path := path
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			got := response.NewFileDto(path)
			if got.Name != path {
				t.Errorf("expected: %v, got: %v", path, got.Name)
			}
			expType := mime.TypeByExtension(filepath.Ext(path))
			if got.MimeType != expType {
				t.Errorf("expected: %v, got: %v", expType, got.MimeType)
			}
		})
	}
}

// TestError tests the writing of error responses.
func TestError(t *testing.T) {
	t.Parallel()
	c := &quick.Config{MaxCount: 100000}
	f := errorPropertySpec
	if err := quick.Check(f, c); err != nil {
		t.Error(err)
	}
}

// checkValues is a test helper which checks if a given JSON contains given values.
func checkValues(t testing.TB, jsonString string, dto *response.JsonDto) (failed bool) {
	safeHelper(t)
	// Check if the JSON contains the status code.
	if !strings.Contains(jsonString, fmt.Sprint(dto.Status)) {
		failed = true
		safeErrorf(t, "expected status: %v", dto.Status)
	}
	// Check if the JSON contains the message.
	if !strings.Contains(jsonString, dto.Message) {
		failed = true
		safeErrorf(t, "expected message: %v", dto.Message)
	}
	// Check if the JSON contains the error string.
	if dto.ErrorString != "" && !strings.Contains(jsonString, dto.ErrorString) {
		failed = true
		safeErrorf(t, "expected error string: %v", dto.ErrorString)
	}
	if dto.File != nil {
		// Check if the JSON contains the file name.
		if !strings.Contains(jsonString, dto.File.Name) {
			failed = true
			safeErrorf(t, "expected file name: %v", dto.File.Name)
		}
		// Check if the JSON contains the file mime type.
		if !strings.Contains(jsonString, dto.File.MimeType) {
			failed = true
			safeErrorf(t, "expected file mime type: %v", dto.File.MimeType)
		}
	}
	return
}

// marshalUnmarshalPropertySpec is the property specification
// used for property based testing of marshal + unmarshalling.
func marshalUnmarshalPropertySpec(status uint16, message, errorString, name string) bool {
	dto := &response.JsonDto{
		Status:      status,
		Message:     message,
		ErrorString: errorString,
		File:        response.NewFileDto(name),
	}
	var got = &response.JsonDto{}
	jsonData, err := json.Marshal(dto)
	json.Unmarshal(jsonData, got)
	return err == nil && reflect.DeepEqual(got, dto)
}

// errorPropertySpec is the property specification
// used for property based testing of error responses.
func errorPropertySpec(status uint16, errorString string) bool {
	// Ensure that the generated status code
	// is in the range of HTTP status codes.
	status = (status % 412) + 100 // Transform to [100, 511].
	w := httptest.NewRecorder()
	response.Error(w, errorString, status)
	refDto := &response.JsonDto{
		Status:      status,
		ErrorString: errorString,
	}
	gotDto := &response.JsonDto{}
	err := json.Unmarshal(w.Body.Bytes(), gotDto)
	if err != nil {
		return false
	}
	hasStatus := w.Code == int(status)
	hasType := strings.Contains(w.Header().Get("Content-Type"), "application/json")
	matches := reflect.DeepEqual(gotDto, refDto)
	return hasStatus && hasType && matches
}

// safeErrorf is a helper function which wraps
// testing.T.Errorf to prevent panics.
func safeErrorf(t testing.TB, format string, args ...interface{}) {
	safeHelper(t)
	if t != nil {
		t.Errorf(format, args...)
	}
}

// safeHelper is a helper function which wraps
// testing.T.Helper to prevent panics.
func safeHelper(t testing.TB) {
	if t != nil {
		t.Helper()
	}
}
