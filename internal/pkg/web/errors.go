package web

import (
	"errors"
	"net/http"
)

// Error type for web responses.
type Error struct {
	Status  int
	Message string
	Fields  map[string]string
}

func (e *Error) Error() string {
	return e.Message
}

// NewRequestError creates a new Error with a status code.
func NewRequestError(status int, message string) error {
	return &Error{Status: status, Message: message}
}

// NewValidationError creates a new Error for validation failures.
func NewValidationError(fields map[string]string) error {
	return &Error{
		Status:  http.StatusBadRequest,
		Message: "validation failed",
		Fields:  fields,
	}
}

// IsRequestError checks if the error is a RequestError.
func IsRequestError(err error) (*Error, bool) {
	var re *Error
	if errors.As(err, &re) {
		return re, true
	}
	return nil, false
}

// FieldError is used to indicate an error with a specific request field.
type FieldError struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

// ErrorResponse is the form used for API responses when an error occurs.
type ErrorResponse struct {
	Error  string            `json:"error"`
	Fields map[string]string `json:"fields,omitempty"`
}

// ShutdownError is used to signal that the app is shutting down.
type ShutdownError struct {
	Message string
}

func (se *ShutdownError) Error() string {
	return se.Message
}

// NewShutdownError creates a new ShutdownError.
func NewShutdownError(message string) error {
	return &ShutdownError{Message: message}
}

// IsShutdown determines if the error is a shutdown error.
func IsShutdown(err error) bool {
	var se *ShutdownError
	if errors.As(err, &se) {
		return true
	}
	return false
}
