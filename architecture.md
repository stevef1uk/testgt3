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
| 8 | `linkshelf/web/style.css` | Simple readable layout, no framework |

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
4. Register handlers: `http.HandleFunc("GET /api/links", handleList)`, `http.HandleFunc("POST /api/links", handleCreate)`, `http.HandleFunc("DELETE /api/links/{id}", handleDelete)`, `http.HandleFunc("GET /static/{file}", handleStatic)`, `http.HandleFunc("GET /", handleIndex)`.
5. `log.Println("listening on :8080")` then `http.ListenAndServe(":8080", nil)`.

All store calls in handlers use the package-level `store.DB` variable. The `handlers.go` file defines handler functions and exports them (or registers them) so `main` can wire them.

### Static file serving detail

The `/` handler serves `linkshelf/web/index.html` by reading and writing the file bytes (or using `http.ServeFile`). The `/static/{file}` path extracts the file name, rejects paths containing `..`, and serves from `linkshelf/web/` directory (e.g., `http.FileServer` with `http.Dir("web")`).

## Unit tests

Unit tests are optional per SPEC. If added:

- `linkshelf/internal/store/store_test.go`: Uses `:memory:` SQLite; calls `InitSchema(db)`, sets `store.DB = db`; tests `List` (empty returns `[]`, after insert returns ordered), `Create` (valid, title empty, title >200 runes, URL empty, URL not http/https), `Delete` (existing succeeds, missing returns error).
- `linkshelf/internal/api/handlers_test.go`: Uses `httptest.NewRecorder`; tests each route with in-memory DB.

No test files are in the required list — the bead implementation produces only the 8 files above plus `go.sum` (generated by `go mod tidy`).

## Integration and testing

The full test suite command (from the `linkshelf/` directory):

cd linkshelf && go mod tidy && go test ./...

If no test files exist, `go test ./...` still compiles and passes. The pipeline runs:

1. `go mod tidy` — resolves dependencies, generates `go.sum`.
2. `go test ./...` — all packages must compile; tests (if any) must pass.
3. `go run ./cmd/server` — must serve on `:8080`.

No Docker, docker-compose, or deployment scripts are involved in this MVP.

## Acceptance mapping

| SPEC requirement | Architecture satisfaction |
|------------------|--------------------------|
| `cd linkshelf && go mod tidy && go test ./...` passes | Package structure compiles with `mattn/go-sqlite3`; test files optional |
| `go run ./cmd/server` serves on `:8080` | `main` opens DB, calls `InitSchema`, sets global DB, registers handlers, starts listener |
| List links → `GET /api/links` returns `[]` when empty | `store.List` returns empty slice not nil; JSON serializes as `[]` |
| Create links with validation | `store.Create` validates title (non-empty, ≤200 runes) and URL (non-empty, must start with `http://` or `https://`) |
| Delete links → `DELETE /api/links/{id}` returns 204 or 404 | `store.Delete` removes by id or returns error if missing |
| Data model: `Link` with ID, Title, URL, CreatedAt (RFC3339 UTC) | `schema.go` defines struct and `CREATE TABLE IF NOT EXISTS links (id INTEGER PRIMARY KEY AUTOINCREMENT, title TEXT NOT NULL, url TEXT NOT NULL, created_at TEXT NOT NULL)` |
| Frontend loads and interacts | `index.html`, `app.js`, `style.css` served by `/` and `/static/{file}` |
| Order by id DESC | `store.List` uses `ORDER BY id DESC` |
| Static path safety | `/static/` handler rejects paths with `..` |
