package handler

import (
	"encoding/json"
	"net/http"

	"github.com/futomaru/todo-app-go/internal/apperror"
)

func decodeJSON(r *http.Request, dst any) error {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return &apperror.ValidationError{Fields: []apperror.FieldError{{
			Field: "body", Message: "invalid JSON",
		}}}
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
