package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/futomaru/todo-app-go/internal/apperror"
)

// statusWriter wraps http.ResponseWriter to remember the status code
// written, since the standard interface has no getter for it.
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// RequestLogger logs one line per request: method, path, status, duration.
func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			start := time.Now()
			next.ServeHTTP(sw, r)
			logger.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", sw.status,
				"dur_ms", time.Since(start).Milliseconds(),
			)
		})
	}
}

// Recoverer catches panics from downstream handlers, logs them, and
// responds with 500 instead of letting the panic crash the process.
//
// Limitation: if the handler already wrote a status/body before panicking,
// WriteHeader here is a no-op ("superfluous WriteHeader call") — recovery
// only produces a clean 500 for panics that happen before any output.
func Recoverer(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.Error("panic recovered",
						"err", rec,
						"path", r.URL.Path,
						"stack", string(debug.Stack()),
					)
					apperror.WriteError(w, r, fmt.Errorf("panic: %v", rec))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// Chain applies mws to h in order, outermost first: Chain(h, A, B) runs as
// A(B(h)) — a request passes through A, then B, then h.
func Chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}
