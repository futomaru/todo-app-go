package handler

import "net/http"

// decodeJSON reads the request body into dst. A decode failure is surfaced as a
// *apperror.ValidationError (a malformed body is a client 400, not a 500).
//
// TODO: json.NewDecoder(r.Body).Decode(dst); on error return a ValidationError
// with a {field:"body", message:"invalid JSON"}.
func decodeJSON(r *http.Request, dst any) error {
	panic("not implemented")
}

// writeJSON sets the Content-Type, writes the status line, and encodes v.
//
// TODO: set "Content-Type: application/json"; w.WriteHeader(status);
// json.NewEncoder(w).Encode(v).
func writeJSON(w http.ResponseWriter, status int, v any) {
	panic("not implemented")
}
