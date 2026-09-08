package handler

import "net/http"

// Routes wires the Todo resource's HTTP surface using the standard library
// ServeMux (Go 1.22+ method + path patterns). Method mismatches on a known
// path are rejected with 405 by the mux itself.
func (h *TodoHandler) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/todos", h.list)
	mux.HandleFunc("POST /api/v1/todos", h.create)
	mux.HandleFunc("DELETE /api/v1/todos", h.deleteCompleted)
	mux.HandleFunc("GET /api/v1/todos/{id}", h.get)
	mux.HandleFunc("PATCH /api/v1/todos/{id}", h.update)
	mux.HandleFunc("DELETE /api/v1/todos/{id}", h.delete)
	return mux
}
