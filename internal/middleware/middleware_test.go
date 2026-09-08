package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/futomaru/todo-app-go/internal/apperror"
)

func newTestLogger(buf *bytes.Buffer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(buf, nil))
}

func TestRequestLogger_RecordsMethodPathStatus(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})
	handler := RequestLogger(logger)(next)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/todos", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var logLine map[string]any
	if err := json.Unmarshal(buf.Bytes(), &logLine); err != nil {
		t.Fatalf("unmarshal log line: %v (raw=%s)", err, buf.String())
	}
	if logLine["method"] != http.MethodPost {
		t.Errorf("method = %v, want POST", logLine["method"])
	}
	if logLine["path"] != "/api/v1/todos" {
		t.Errorf("path = %v, want /api/v1/todos", logLine["path"])
	}
	if status, ok := logLine["status"].(float64); !ok || int(status) != http.StatusCreated {
		t.Errorf("status = %v, want 201", logLine["status"])
	}
}

func TestRequestLogger_DefaultStatusIsOKWhenUnset(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handler writes a body without ever calling WriteHeader.
		w.Write([]byte("ok"))
	})
	handler := RequestLogger(logger)(next)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/todos", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var logLine map[string]any
	if err := json.Unmarshal(buf.Bytes(), &logLine); err != nil {
		t.Fatalf("unmarshal log line: %v", err)
	}
	if status, ok := logLine["status"].(float64); !ok || int(status) != http.StatusOK {
		t.Errorf("status = %v, want 200", logLine["status"])
	}
}

func TestRecoverer_TurnsPanicInto500(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})
	handler := Recoverer(logger)(next)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/todos", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req) // must not panic out of the test

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Errorf("Content-Type = %q, want application/problem+json", ct)
	}

	var pd apperror.ProblemDetail
	if err := json.NewDecoder(rec.Body).Decode(&pd); err != nil {
		t.Fatalf("decode ProblemDetail: %v", err)
	}
	if pd.Status != http.StatusInternalServerError {
		t.Errorf("pd.Status = %d, want 500", pd.Status)
	}
	if !strings.Contains(buf.String(), "panic recovered") {
		t.Errorf("log output missing panic record: %s", buf.String())
	}
}

func TestRecoverer_NoPanicPassesThrough(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	handler := Recoverer(logger)(next)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/todos/1", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rec.Code)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no log output for a non-panicking request, got: %s", buf.String())
	}
}

func TestChain_AppliesOutermostFirst(t *testing.T) {
	var order []string

	mw := func(name string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name+":enter")
				next.ServeHTTP(w, r)
				order = append(order, name+":exit")
			})
		}
	}

	base := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
	})

	handler := Chain(base, mw("A"), mw("B"))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	want := []string{"A:enter", "B:enter", "handler", "B:exit", "A:exit"}
	if len(order) != len(want) {
		t.Fatalf("order = %v, want %v", order, want)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Errorf("order[%d] = %q, want %q (full: %v)", i, order[i], want[i], order)
		}
	}
}
