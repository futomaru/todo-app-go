package handler

import "net/http"

// Routes builds the ServeMux mapping method+path patterns to handlers. Go 1.22+
// lets a single mux dispatch by method, so the collection path /api/v1/todos is
// shared by list (GET), create (POST) and deleteCompleted (DELETE); the item
// path carries the {id} wildcard read via r.PathValue("id").
//
// TODO: register on http.NewServeMux():
//
//	GET    /api/v1/todos        -> h.list
//	POST   /api/v1/todos        -> h.create
//	DELETE /api/v1/todos        -> h.deleteCompleted
//	GET    /api/v1/todos/{id}   -> h.get
//	PATCH  /api/v1/todos/{id}   -> h.update
//	DELETE /api/v1/todos/{id}   -> h.delete
func (h *TodoHandler) Routes() *http.ServeMux {
	panic("not implemented")
}
