package apperror

import (
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("not found")

// FieldError represents a validation error for a specific field.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationError represents a validation error for multiple fields.
type ValidationError struct {
	Fields []FieldError
	Detail string
}

// Error implements the error interface for ValidationError.
func (e *ValidationError) Error() string {
	if e.Detail != "" {
		return e.Detail
	}
	return fmt.Sprintf("validation failed: %d fields", len(e.Fields))
}

// NotFoundf returns an error that wraps ErrNotFound with a formatted message.
func NotFoundf(format string, args ...any) error {
	return fmt.Errorf(format+": %w", append(args, ErrNotFound)...)
}
