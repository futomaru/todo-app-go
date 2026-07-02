package dto_test

import (
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
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
