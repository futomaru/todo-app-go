package dto

import (
	"strings"
	"unicode/utf8"

	"github.com/futomaru/todo-app-go/internal/apperror"
)

const (
	maxTitleLength       = 255
	maxDescriptionLength = 1000
)

// ValidateCreateRequest validates the TodoCreateRequest and returns a slice of FieldError if there are validation errors.
func ValidateCreateRequest(req TodoCreateRequest) []apperror.FieldError {
	var errs []apperror.FieldError

	if strings.TrimSpace(req.Title) == "" {
		errs = append(errs, apperror.FieldError{
			Field:   "title",
			Message: "must not be blank",
		})
	}
	if utf8.RuneCountInString(req.Title) > maxTitleLength {
		errs = append(errs, apperror.FieldError{
			Field:   "title",
			Message: "size must be between 1 and 255",
		})
	}

	if req.Description != nil && utf8.RuneCountInString(*req.Description) > maxDescriptionLength {
		errs = append(errs, apperror.FieldError{
			Field:   "description",
			Message: "size must be between 0 and 1000",
		})
	}

	return errs
}

// ValidateUpdateRequest validates the TodoUpdateRequest and returns a slice of FieldError if there are validation errors.
func ValidateUpdateRequest(req TodoUpdateRequest) []apperror.FieldError {
	var errs []apperror.FieldError

	if req.Title != nil {
		if strings.TrimSpace(*req.Title) == "" {
			errs = append(errs, apperror.FieldError{
				Field:   "title",
				Message: "must not be blank",
			})
		}
		if utf8.RuneCountInString(*req.Title) > maxTitleLength {
			errs = append(errs, apperror.FieldError{
				Field:   "title",
				Message: "size must be between 1 and 255",
			})
		}
	}

	if req.Description != nil {
		if utf8.RuneCountInString(*req.Description) > maxDescriptionLength {
			errs = append(errs, apperror.FieldError{
				Field:   "description",
				Message: "size must be between 0 and 1000",
			})
		}
	}

	return errs
}
