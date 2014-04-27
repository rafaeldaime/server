package main

import (
	"fmt"
)

const (
	// Error codes here, not implemented yet
	ErrorCodeDefault = 1
)

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// The serializable Error structure.
type APIError struct {
	Error Error `json:"error"`
}

func (e *APIError) String() string {
	return fmt.Sprintf("Error: [%d] %s", e.Error.Code, e.Error.Message)
}

// NewError creates an error instance with the specified code and message.
func NewError(code int, msg string) APIError {
	return APIError{
		Error: Error{
			Code:    code,
			Message: msg,
		},
	}
}
