// Package response provides types and functions for handling
// API response DTOs and their payloads, usually encoded as JSON.
package response

import (
	"encoding/json"
)

type MarshalUnmarshaler interface {
	json.Marshaler
	json.Unmarshaler
}

type JsonDto struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Nested  any    `json:"nested"`
	Error   error  `json:"error"`
}

// func (dto *JsonDto) MarshalJSON() ([]byte, error) {
// 	return json.Marshal(dto)
// }

// func (dto *JsonDto) UnmarshalJSON(data []byte) error {
// 	return json.Unmarshal(data, dto)
// }
