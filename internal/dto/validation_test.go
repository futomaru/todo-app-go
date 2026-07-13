package dto_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/futomaru/todo-app-go/internal/apperror"
	"github.com/futomaru/todo-app-go/internal/dto"
)

func TestValidateCreateRequest(t *testing.T) {
	tests := []struct {
		name string
		req  dto.TodoCreateRequest
		want []apperror.FieldError
	}{
		{
			name: "valid request",
			req: dto.TodoCreateRequest{
				Title:       "Test Todo",
				Description: stringPtr("This is a test todo"),
			},
			want: nil,
		},
		{
			name: "invalid request - blank title",
			req: dto.TodoCreateRequest{
				Title:       "",
				Description: stringPtr("This is a test todo"),
			},
			want: []apperror.FieldError{
				{
					Field:   "title",
					Message: "must not be blank",
				},
			},
		},
		{
			name: "valid request - title at max length (255 runes)",
			req: dto.TodoCreateRequest{
				Title: strings.Repeat("あ", 255),
			},
			want: nil,
		},
		{
			name: "invalid request - title over max length (256 runes)",
			req: dto.TodoCreateRequest{
				Title: strings.Repeat("あ", 256),
			},
			want: []apperror.FieldError{
				{
					Field:   "title",
					Message: "size must be between 1 and 255",
				},
			},
		},
		{
			name: "valid request - description at max length (1000 runes)",
			req: dto.TodoCreateRequest{
				Title:       "Test Todo",
				Description: stringPtr(strings.Repeat("b", 1000)),
			},
			want: nil,
		},
		{
			name: "invalid request - description over max length (1001 runes)",
			req: dto.TodoCreateRequest{
				Title:       "Test Todo",
				Description: stringPtr(strings.Repeat("b", 1001)),
			},
			want: []apperror.FieldError{
				{
					Field:   "description",
					Message: "size must be between 0 and 1000",
				},
			},
		},
		{
			name: "invalid request - blank title over max length reports both errors",
			req: dto.TodoCreateRequest{
				Title: strings.Repeat(" ", 256),
			},
			want: []apperror.FieldError{
				{
					Field:   "title",
					Message: "must not be blank",
				},
				{
					Field:   "title",
					Message: "size must be between 1 and 255",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dto.ValidateCreateRequest(tt.req)
			if !equalFieldErrors(got, tt.want) {
				t.Errorf("ValidateCreateRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateUpdateRequest(t *testing.T) {
	tests := []struct {
		name string
		req  dto.TodoUpdateRequest
		want []apperror.FieldError
	}{
		{
			name: "valid request",
			req: dto.TodoUpdateRequest{
				Title:       stringPtr("Updated Todo"),
				Description: stringPtr("This is an updated todo"),
				Completed:   boolPtr(true),
			},
			want: nil,
		},
		{
			name: "valid request - all fields omitted (partial update)",
			req:  dto.TodoUpdateRequest{},
			want: nil,
		},
		{
			name: "invalid request - blank title",
			req: dto.TodoUpdateRequest{
				Title: stringPtr(""),
			},
			want: []apperror.FieldError{
				{
					Field:   "title",
					Message: "must not be blank",
				},
			},
		},
		{
			name: "invalid request - title over max length (256 runes)",
			req: dto.TodoUpdateRequest{
				Title: stringPtr(strings.Repeat("あ", 256)),
			},
			want: []apperror.FieldError{
				{
					Field:   "title",
					Message: "size must be between 1 and 255",
				},
			},
		},
		{
			name: "invalid request - description over max length (1001 runes)",
			req: dto.TodoUpdateRequest{
				Description: stringPtr(strings.Repeat("b", 1001)),
			},
			want: []apperror.FieldError{
				{
					Field:   "description",
					Message: "size must be between 0 and 1000",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dto.ValidateUpdateRequest(tt.req)
			if !equalFieldErrors(got, tt.want) {
				t.Errorf("ValidateUpdateRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func equalFieldErrors(a, b []apperror.FieldError) bool {
	return reflect.DeepEqual(a, b)
}
