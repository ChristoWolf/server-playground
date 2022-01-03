// Package response_test provides a test suite for the response package.
package response_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"

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
			Status:  200,
			Message: http.StatusText(200),
		},
		{
			Status:  201,
			Message: http.StatusText(201) + ": file created",
			File:    response.NewFileDto("test.txt"),
		},
		{
			Status:      404,
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

func checkValues(t *testing.T, jsonString string, dto *response.JsonDto) (failed bool) {
	t.Helper()
	// Check if the JSON contains the status code.
	if !strings.Contains(jsonString, fmt.Sprint(dto.Status)) {
		failed = true
		t.Errorf("expected status: %v", dto.Status)
	}
	// Check if the JSON contains the message.
	if !strings.Contains(jsonString, dto.Message) {
		failed = true
		t.Errorf("expected message: %v", dto.Message)
	}
	// Check if the JSON contains the error string.
	if dto.ErrorString != "" && !strings.Contains(jsonString, dto.ErrorString) {
		failed = true
		t.Errorf("expected error string: %v", dto.ErrorString)
	}
	if dto.File != nil {
		// Check if the JSON contains the file name.
		if !strings.Contains(jsonString, dto.File.Name) {
			failed = true
			t.Errorf("expected file name: %v", dto.File.Name)
		}
		// Check if the JSON contains the file mime type.
		if !strings.Contains(jsonString, dto.File.MimeType) {
			failed = true
			t.Errorf("expected file mime type: %v", dto.File.MimeType)
		}
	}
	return
}
