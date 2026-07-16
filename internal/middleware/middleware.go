package middleware

import (
	"log/slog"
	"net/http"
)

// Middleware is the standard net/http decorator shape: it wraps a handler and
// returns a new one, so several can be composed into a chain.
type Middleware = func(http.Handler) http.Handler

// statusWriter wraps http.ResponseWriter to remember the status code written by
// the inner handler (needed for logging, since ResponseWriter exposes no getter).
// Embedding the interface means every other method is forwarded unchanged.
type statusWriter struct {
	http.ResponseWriter
	status int
}

// WriteHeader records the code before delegating to the wrapped writer.
//
// TODO: w.status = code; w.ResponseWriter.WriteHeader(code).
func (w *statusWriter) WriteHeader(code int) {
	panic("not implemented")
}

// RequestLogger logs one structured line per request (method, path, status,
// duration) using the injected logger rather than the global one.
//
// TODO: return a Middleware that wraps next in a statusWriter (default 200),
// times next.ServeHTTP, then logger.Info("request", ...fields...).
func RequestLogger(logger *slog.Logger) Middleware {
	panic("not implemented")
}

// Recoverer catches panics from downstream handlers, logs them with a stack, and
// converts them to a 500 via apperror.WriteError so one bad request cannot crash
// the process. Place it OUTERMOST in the chain.
//
// TODO: return a Middleware whose handler defers a recover(); on non-nil,
// logger.Error with debug.Stack() and apperror.WriteError(w, r, fmt.Errorf(...)).
func Recoverer(logger *slog.Logger) Middleware {
	panic("not implemented")
}

// Chain applies middlewares so the first argument ends up outermost (i.e. runs
// first on the way in), wrapping h from the inside out.
//
// TODO: iterate mws in reverse, h = mws[i](h); return h.
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	panic("not implemented")
}
