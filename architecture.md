# Architecture for testgt3 — Link Shelf MVP

## Overview
The Link Shelf application is a minimal bookmark manager (MVP) written in Go 1.22 with SQLite persistence via `mattn/go-sqlite3`. It exposes a REST API for listing, creating, and deleting links, and serves a single-page static frontend on the same port. The backend uses `database/sql` with a global `*sql.DB` variable, stores data in `linkshelf.db`, and listens on `:8080`. The frontend is plain HTML/CSS/JS that consumes the REST API.

The design follows a three-package structure under `linkshelf/internal/`: `store` holds the data model and DDL (`schema.go`) plus package-level CRUD functions (`store.go`), and `api` holds HTTP handlers (`handlers.go`). The entrypoint in `linkshelf/cmd/server/main.go` wires dependencies at startup.

## Planned file layout

All implement paths use the root prefix `linkshelf/` as specified in the SPEC.

| # | Path | Purpose |
|---|------|---------|
| 1 | `linkshelf/go.mod` | Go module definition; module `linkshelf`, go 1.22, require `github.com/mattn/go-sqlite3 v1.14.22` |
| 2 | `linkshelf/internal/store/schema.go` | `Link` struct type and `InitSchema` function; DDL (`CREATE TABLE IF NOT EXISTS links`) owned exclusively here |
| 3 | `linkshelf/internal/store/store.go` | Package-level var `DB *sql.DB`; functions `List`, `Create`, `Delete`; validation rules for `Create` |
| 4 | `linkshelf/internal/api/handlers.go` | HTTP handlers for `GET /api/links`, `POST /api/links`, `DELETE /api/links/{id}`, and static file serving for `/` and `/static/{file}` |
| 5 | `linkshelf/cmd/server/main.go` | Opens `linkshelf.db`, calls `InitSchema`, sets `store.DB`, registers routes on `http.DefaultServeMux`, starts `:8080` |
| 6 | `linkshelf/web/index.html` | Frontend HTML: title input, URL input, Add button, `<ul id="links"></ul>` |
| 7 | `linkshelf/web/app.js` | Frontend JS: `GET /api/links` on load, POST to add, DELETE to remove, refresh list after each change |
| 8 | `linkshelf/web/style.css` | Simple styling to make the UI readable; no framework |

No `*_test.go` files are required — tests are optional per SPEC.

## Go package / bead ownership

Files 2 (`schema.go`) and 3 (`store.go`) both live in Go package `store` at `linkshelf/internal/store/`. Symbol ownership must be strict to avoid duplicate definitions across beads:

| File | Owns (exported) | Must not define |
|------|----------------|-----------------|
| `linkshelf/internal/store/schema.go` | `type Link struct { ID int64; Title string; URL string; CreatedAt string }`, `func InitSchema(*sql.DB) error` | `var DB`, `func List`, `func Create`, `func Delete` |
| `linkshelf/internal/store/store.go` | `var DB *sql.DB`, `func List(context.Context) ([]Link, error)`, `func Create(context.Context, string, string) (Link, error)`, `func Delete(context.Context, int64) error` | `type Link`, `func InitSchema` |

The `Link` type is defined in `schema.go` and used by `store.go` within the same package (no import needed). The `api` package imports `store` and calls `store.List`, `store.Create`, `store.Delete`, and references `store.DB` — it does not import `schema` directly.

## HTTP + entrypoint integration

### HTTP API routes (matches SPEC verbatim)

| Method | Path | Success | Error |
|--------|------|---------|-------|
| GET | `/` | 200, serve `linkshelf/web/index.html` | — |
| GET | `/static/{file}` | 200, file under `linkshelf/web/` | 404 |
| GET | `/api/links` | 200, JSON array `[]` when empty | — |
| POST | `/api/links` | 201, JSON link | 400 `{"error":"..."}` |
| DELETE | `/api/links/{id}` | 204 | 404 `{"error":"..."}` |

POST body is `{"title":"...","url":"..."}`. Static paths containing `..` are rejected (404).

### Entrypoint wiring (`linkshelf/cmd/server/main.go`)

The startup sequence is:

1. `db, err := sql.Open("sqlite3", "linkshelf.db")` — opens SQLite file in current working directory.
2. `schema.InitSchema(db)` — ensures the `links` table exists (CREATE TABLE IF NOT EXISTS).
3. `store.DB = db` — assigns the connection to the package-level variable.
4. Register handlers: `http.HandleFunc("GET /", ...)`; similar for other routes.

## Unit tests
No unit test files are required for this MVP; however, the following structure could be used if tests were added:

- `linkshelf/internal/store/store_test.go`: unit tests for `List`, `Create`, `Delete`.
- `linkshelf/internal/api/handlers_test.go`: unit tests for handlers.

## Integration and testing
The continuous-integration stage runs two commands, both of which must succeed:

1. **Compilation & unit-test pass**

   cd linkshelf && go test ./...

2. **Runtime sanity check**

   cd linkshelf && go run ./cmd/server &
   SERVER_PID=$!
   # give the server a moment to start
   sleep 1
   # basic health-check: fetch the UI root
   curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/ | grep -q "^200$"
   kill $SERVER_PID

## Acceptance mapping
The architecture satisfies every acceptance criterion listed in the SPEC.

## Delivery phases
Based on the SPEC, the following phases are identified for delivery:

1. **Initial MVP**: Implement the MVP as specified, with all necessary backend and frontend components.
2. **Testing and Validation**: Perform thorough testing and validation of the MVP to ensure it meets the acceptance criteria.

