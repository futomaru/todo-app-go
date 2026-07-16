// Command todoapp is the composition root: it constructs every layer once and
// starts the HTTP server. This is the only place that knows the concrete types.
package main

// _ "modernc.org/sqlite" registers the pure-Go "sqlite" driver with
// database/sql as a side effect (no CGO). sql.Open("sqlite", dsn) needs it.
import (
	_ "modernc.org/sqlite"
)

func main() {
	// TODO: assemble the application, step by step:
	//
	//  1. Config: read PORT (default ":8080") and DB_PATH (default
	//     "./data/todo.db") from the environment via a small envOr helper.
	//  2. Logger: slog.New(slog.NewJSONHandler(os.Stderr, ...)).
	//  3. DB: os.MkdirAll(filepath.Dir(dbPath)); sql.Open("sqlite", dbPath +
	//     "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)");
	//     db.SetMaxOpenConns(1) to serialise writes; repository.Migrate(ctx, db).
	//  4. Wiring (bottom-up, constructor injection):
	//        repo := repository.New(db)
	//        svc  := service.New(repo, clock.System())
	//        h    := handler.New(svc)
	//        root := middleware.Chain(h.Routes(),
	//                    middleware.Recoverer(logger),
	//                    middleware.RequestLogger(logger))
	//  5. Server: &http.Server{Addr, Handler: root, ReadHeaderTimeout: 5s};
	//     run it with graceful shutdown (signal.NotifyContext + srv.Shutdown).
	panic("not implemented")
}
