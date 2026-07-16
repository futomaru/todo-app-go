# Specification

Behavioral specification of the Todo REST API: endpoints, schemas, validation, errors, and
runtime behavior.

## Basics

| Item | Value |
|---|---|
| Base URL | `http://localhost:8080` |
| API prefix | `/api/v1/todos` |
| Resource | Todo (single entity, no relations) |
| Auth | none |
| Request body | `application/json` |
| Success body | `application/json` |
| Error body | `application/problem+json` (RFC 7807) |
| Persistence | SQLite file `./data/todo.db`; schema created idempotently at startup |

## Endpoints

| Method | Path | Purpose | Success | Errors |
|---|---|---|---|---|
| GET | `/api/v1/todos` | list (optional `?completed=` filter) | 200 | 400 |
| GET | `/api/v1/todos/{id}` | get one | 200 | 400 / 404 |
| POST | `/api/v1/todos` | create | 201 + `Location` | 400 |
| PATCH | `/api/v1/todos/{id}` | partial update | 200 | 400 / 404 |
| DELETE | `/api/v1/todos/{id}` | delete one | 204 | 400 / 404 |
| DELETE | `/api/v1/todos?completed=true` | delete all completed | 204 | 400 |

Endpoint semantics:

- **List** returns todos in `id` ascending order. With `?completed=true|false` it filters by
  completion state; an unparsable value is a 400. Zero matches yield an empty array `[]`,
  never `null`.
- **Create** ignores any client-supplied `id`, `completed`, or timestamps: `completed` always
  starts `false`, and `createdAt`/`updatedAt` are set to the same server-side instant. The
  response carries a `Location: /api/v1/todos/{id}` header.
- **Partial update** follows PATCH semantics: only fields present (non-null) in the body are
  changed; omitted fields keep their values. `updatedAt` is refreshed, `createdAt` never
  changes. Setting a field explicitly back to null is not supported — null means "not
  specified".
- **Delete one** returns 404 if the id does not exist.
- **Delete completed** requires `completed=true` explicitly; omitting it or passing any other
  value is a 400 (bulk-deleting *unfinished* todos is deliberately unsupported). Deleting
  when nothing matches is still 204 — an empty-set operation on a collection is a success,
  unlike the single-resource delete above.
- **Path `id`** must be an integer ≥ 1; anything else is rejected with 400 before any lookup.

## Schemas

### Todo response (all success bodies; lists are arrays of this)

| Field | Type | Notes |
|---|---|---|
| `id` | integer | server-assigned |
| `title` | string | |
| `description` | string \| null | |
| `completed` | boolean | |
| `createdAt` | string | RFC 3339, UTC |
| `updatedAt` | string | RFC 3339, UTC |

### Create request (POST)

| Field | Type | Required | Constraint |
|---|---|---|---|
| `title` | string | yes | not blank, at most 255 characters |
| `description` | string \| null | no | at most 1000 characters |

### Update request (PATCH) — every field optional

| Field | Type | Constraint when present |
|---|---|---|
| `title` | string | not blank, at most 255 characters |
| `description` | string | at most 1000 characters |
| `completed` | boolean | none (`false` is distinct from omitted) |

## Validation

- Length limits count **Unicode characters, not bytes** — a 255-character Japanese title is
  valid.
- "Blank" means empty or whitespace-only.
- All violations in a request are collected and reported together, not one at a time.

## Error Responses

Every error body is a problem-details object:

```json
{
  "type": "about:blank",
  "title": "Bad Request",
  "status": 400,
  "detail": "Validation failed",
  "instance": "/api/v1/todos",
  "errors": [
    { "field": "title", "message": "must not be blank" }
  ]
}
```

`errors` appears only on validation failures (400). Status mapping:

| Situation | Status |
|---|---|
| id not found (get / update / delete) | 404 |
| body validation failure, unparsable path/query value, malformed JSON | 400 |
| required `completed=true` missing on bulk delete | 400 |
| any unexpected internal failure | 500 — generic message only; details go to the log, never the response |

## Runtime Behavior

| Item | Behavior |
|---|---|
| Listen port | `:8080`, overridable via `PORT` environment variable |
| DB path | `./data/todo.db`, overridable via `DB_PATH` environment variable |
| Timestamps | assigned by the application (not DB defaults), UTC, RFC 3339 |
| Logging | structured JSON lines; one entry per request (method, path, status, duration) |
| Panics | recovered; logged with stack trace; client receives a plain 500 |
| Shutdown | on SIGINT/SIGTERM, stops accepting connections and waits up to 10 seconds for in-flight requests |
