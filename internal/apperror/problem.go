package apperror

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

// ProblemDetail represents a problem detail object.
type ProblemDetail struct {
	Type     string       `json:"type"`
	Title    string       `json:"title"`
	Status   int          `json:"status"`
	Detail   string       `json:"detail,omitempty"`
	Instance string       `json:"instance"`
	Errors   []FieldError `json:"errors,omitempty"`
}

// WriteError writes an error response in the Problem Details format.
func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	var pd ProblemDetail
	var ve *ValidationError

	switch {
	case errors.Is(err, ErrNotFound):
		pd = ProblemDetail{
			Title:  "Not Found",
			Status: http.StatusNotFound,
			Detail: err.Error(),
		}
	case errors.As(err, &ve):
		detail := ve.Detail
		if detail == "" {
			detail = "Validation failed"
		}
		pd = ProblemDetail{
			Title:  "Bad Request",
			Status: http.StatusBadRequest,
			Detail: detail,
			Errors: ve.Fields,
		}
	default:
		slog.Error("unhandled error", "err", err, "path", r.URL.Path)
		pd = ProblemDetail{
			Title:  "Internal Server Error",
			Status: http.StatusInternalServerError,
			Detail: "An unexpected error occurred",
		}
	}
	pd.Type = "about:blank"
	pd.Instance = r.URL.Path

	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(pd.Status)
	_ = json.NewEncoder(w).Encode(pd)
}
