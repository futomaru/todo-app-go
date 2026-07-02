package apperror_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/futomaru/todo-app-go/internal/apperror"
)

func TestWriteError(t *testing.T) {
	// Define test cases for different error scenarios
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantError  int // Number of Errors
		wantDetail string
	}{
		// Test case for ErrNotFound
		{
			name:       "ErrNotFound",
			err:        apperror.ErrNotFound,
			wantStatus: http.StatusNotFound,
			wantError:  0,
			wantDetail: "not found",
		},

		// Test case for ValidationError
		{
			name: "ValidationError",
			err: &apperror.ValidationError{
				Fields: []apperror.FieldError{{Field: "title", Message: "must not be blank"}},
				Detail: "Invalid input",
			},
			wantStatus: http.StatusBadRequest,
			wantError:  1,
			wantDetail: "Invalid input",
		},
		{
			name: "ValidationErrorWithoutDetail",
			err: &apperror.ValidationError{
				Fields: []apperror.FieldError{{Field: "title", Message: "must not be blank"}},
			},
			wantStatus: http.StatusBadRequest,
			wantError:  1,
			wantDetail: "Validation failed",
		},

		// Test case for an unhandled error
		{
			name:       "UnhandledError",
			err:        errors.New("unexpected error"),
			wantStatus: http.StatusInternalServerError,
			wantError:  0,
			wantDetail: "An unexpected error occurred",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/todos/1", nil)
			apperror.WriteError(rec, req, tt.err)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d; want %d", rec.Code, tt.wantStatus)
			}

			if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
				t.Errorf("Content-Type = %q; want %q", ct, "application/problem+json")
			}

			// Decode the response body into a ProblemDetail struct
			var pd apperror.ProblemDetail
			if err := json.NewDecoder(rec.Body).Decode(&pd); err != nil {
				t.Fatalf("failed to decode response body: %v", err)
			}

			if pd.Detail != tt.wantDetail {
				t.Errorf("Detail = %q; want %q", pd.Detail, tt.wantDetail)
			}

			if len(pd.Errors) != tt.wantError {
				t.Errorf("len(Errors) = %d; want %d", len(pd.Errors), tt.wantError)
			}
		})
	}
}
