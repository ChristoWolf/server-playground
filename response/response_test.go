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

// TestMarshal tests the marshaling of a response DTO (JsonDto) as JSON.
func TestMarshal(t *testing.T) {
	t.Parallel()
	var dtos = []*response.JsonDto{
		&response.JsonDto{
			Status:  200,
			Message: http.StatusText(200),
		},
		&response.JsonDto{
			Status: 404,
			Nested: &response.JsonDto{Status: 500, Message: "nested content"},
			Error:  errors.New("error content"),
		},
	}
	for _, dto := range dtos {
		dto := dto
		t.Run(fmt.Sprintf("%v", dto), func(t *testing.T) {
			t.Parallel()
			// Marshal the DTO as JSON.
			json, err := json.Marshal(dto)
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			// Check if the marshaled JSON contains the correct values.
			failed := checkValues(t, string(json), dto)
			if failed {
				t.Fatalf("value(s) not found in JSON: %v", string(json))
			}
		})
	}
}

// Inspired by https://stackoverflow.com/a/18927729.
func checkValues(t *testing.T, jsonString string, referenceDto *response.JsonDto) (failed bool) {
	v := reflect.ValueOf(*referenceDto)
	for i := 0; i < v.NumField(); i++ {
		var stringValue string
		value := v.Field(i).Interface()
		if value == nil {
			stringValue = "null"
		} else {
			stringValue = fmt.Sprintf("%v", value)
		}
		if !strings.Contains(jsonString, stringValue) {
			failed = true
			t.Errorf("expected value: %v, got: %v", stringValue, jsonString)
		}
	}
	return
}
