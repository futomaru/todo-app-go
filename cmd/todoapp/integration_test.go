package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/futomaru/todo-app-go/internal/dto"
	"github.com/futomaru/todo-app-go/internal/repository"
)

// newIntegrationServer builds the exact same stack as main() — real SQLite
// file, real repository/service/handler/middleware — against a temp DB, so
// this test exercises the actual wiring rather than fakes.
func newIntegrationServer(t *testing.T) *httptest.Server {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "todo.db")
	db, err := openDB(dbPath)
	if err != nil {
		t.Fatalf("openDB: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := repository.Migrate(context.Background(), db); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	srv := httptest.NewServer(newRouter(db, logger))
	t.Cleanup(srv.Close)
	return srv
}

func doJSON(t *testing.T, method, url string, body any) *http.Response {
	t.Helper()

	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		reader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	return resp
}

func decodeBody[T any](t *testing.T, resp *http.Response) T {
	t.Helper()
	defer resp.Body.Close()
	var v T
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	return v
}

// TestIntegration_TodoLifecycle drives create -> list -> update -> filter
// -> bulk delete -> verify through the real HTTP + DB stack, mirroring the
// curl sequence a human would run manually against `make run`.
func TestIntegration_TodoLifecycle(t *testing.T) {
	srv := newIntegrationServer(t)
	base := srv.URL + "/api/v1/todos"

	// 1. Create two todos.
	resp := doJSON(t, http.MethodPost, base, dto.TodoCreateRequest{Title: "buy milk"})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create #1: status = %d", resp.StatusCode)
	}
	first := decodeBody[dto.TodoResponse](t, resp)
	if first.Completed {
		t.Error("newly created todo should not be completed")
	}

	resp = doJSON(t, http.MethodPost, base, dto.TodoCreateRequest{Title: "walk dog"})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create #2: status = %d", resp.StatusCode)
	}
	second := decodeBody[dto.TodoResponse](t, resp)

	// 2. List returns both, ordered by id.
	resp = doJSON(t, http.MethodGet, base, nil)
	all := decodeBody[[]dto.TodoResponse](t, resp)
	if len(all) != 2 || all[0].ID != first.ID || all[1].ID != second.ID {
		t.Fatalf("list = %+v, want [%d, %d] in order", all, first.ID, second.ID)
	}

	// 3. Update: mark the first one completed via PATCH.
	completed := true
	resp = doJSON(t, http.MethodPatch, base+"/"+itoa(first.ID), dto.TodoUpdateRequest{Completed: &completed})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("update: status = %d", resp.StatusCode)
	}
	updated := decodeBody[dto.TodoResponse](t, resp)
	if !updated.Completed {
		t.Error("update did not persist completed=true")
	}
	if !updated.UpdatedAt.After(updated.CreatedAt) && !updated.UpdatedAt.Equal(updated.CreatedAt) {
		t.Errorf("UpdatedAt (%v) should be >= CreatedAt (%v)", updated.UpdatedAt, updated.CreatedAt)
	}

	// 4. Filter by completed=true returns only the updated one.
	resp = doJSON(t, http.MethodGet, base+"?completed=true", nil)
	doneOnly := decodeBody[[]dto.TodoResponse](t, resp)
	if len(doneOnly) != 1 || doneOnly[0].ID != first.ID {
		t.Fatalf("completed=true filter = %+v, want only id=%d", doneOnly, first.ID)
	}

	// 5. Bulk delete completed todos.
	resp = doJSON(t, http.MethodDelete, base+"?completed=true", nil)
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("deleteCompleted: status = %d", resp.StatusCode)
	}

	// 6. The completed todo is gone; the incomplete one remains.
	resp = doJSON(t, http.MethodGet, base+"/"+itoa(first.ID), nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("get deleted todo: status = %d, want 404", resp.StatusCode)
	}
	resp.Body.Close()

	resp = doJSON(t, http.MethodGet, base+"/"+itoa(second.ID), nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get remaining todo: status = %d, want 200", resp.StatusCode)
	}
	resp.Body.Close()

	// 7. Bulk delete again with nothing completed is still success (204).
	resp = doJSON(t, http.MethodDelete, base+"?completed=true", nil)
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("deleteCompleted (empty): status = %d, want 204", resp.StatusCode)
	}
}

func itoa(id int64) string {
	return strconv.FormatInt(id, 10)
}
