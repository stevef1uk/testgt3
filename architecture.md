# Architecture for testgt3 — Link Shelf MVP

## Overview

Link Shelf is a minimal bookmark manager application designed for an automated pipeline. It provides a Go/SQLite backend serving a REST API and a static frontend for CRUD operations on links. The entire project lives under `linkshelf/`. The backend consists of three internal packages: `store` (schema + persistence), `api` (HTTP handlers), and the entrypoint at `linkshelf/cmd/server`. The frontend is static HTML/JS/CSS served from `linkshelf/web/`. The pipeline verifies `cd linkshelf && go test ./...` passes and `cd linkshelf && go run ./linkshelf/cmd/server` serves the UI on `:8080`.

## Planned file layout

All implement paths use the `linkshelf/` prefix as defined in the SPEC:

- `linkshelf/go.mod` — Go module definition (module `linkshelf`, go 1.22, dependency `github.com/mattn/go-sqlite3`).
- `linkshelf/cmd/server/main.go` — Entrypoint: open SQLite file `linkshelf.db`, call `store.InitSchema`, set `store.DB = db`, register handlers via `http.DefaultServeMux`, listen on `:8080`.
- `linkshelf/internal/store/schema.go` — Data model (`Link` struct) and `InitSchema` function (DDL only, no store logic).
- `linkshelf/internal/store/store.go` — Package-level `DB` variable, functions `List`, `Create`, `Delete` (no DDL, no struct redefinition).
- `linkshelf/internal/api/handlers.go` — HTTP handlers for `/api/links` GET/POST/DELETE and static route helpers.
- `linkshelf/web/index.html` — UI: title input, URL input, Add button, `<ul id="links"></ul>`.
- `linkshelf/web/app.js` — JavaScript: on load GET /api/links, render list; POST to add; DELETE to remove; refresh after each change.
- `linkshelf/web/style.css` — Simple readable layout (no framework).

## Go package / bead ownership

The `store` package is split across two implement files (`linkshelf/internal/store/schema.go` and `linkshelf/internal/store/store.go`). To avoid duplicate exported symbols, ownership is strictly assigned per file:

| File | Owns (exported) | Must not define |
|------|-----------------|-----------------|
| `linkshelf/internal/store/schema.go` | type `Link struct`, `func InitSchema(db *sql.DB) error` | `var DB`, `func List`, `func Create`, `func Delete` |
| `linkshelf/internal/store/store.go` | `var DB *sql.DB`, `func List(ctx context.Context) ([]Link, error)`, `func Create(ctx context.Context, title, url string) (Link, error)`, `func Delete(ctx context.Context, id int64) error` | `Link struct`, `InitSchema` |

The `api` package lives in `linkshelf/internal/api/handlers.go` as a single bead. It references `store.List`, `store.Create`, `store.Delete`, and `store.DB`. The entrypoint `linkshelf/cmd/server/main.go` belongs to package `main` and must not define types from `store` or `api`.

## HTTP API table (from SPEC verbatim)

| Method | Path | Success | Error |
|--------|------|---------|-------|
| GET | `/` | 200, `linkshelf/web/index.html` | — |
| GET | `/static/{file}` | 200, file under `linkshelf/web/` | 404 |
| GET | `/api/links` | 200, JSON array `[]` when empty | — |
| POST | `/api/links` | 201, JSON link | 400 `{"error":"..."}` |
| DELETE | `/api/links/{id}` | 204 | 404 `{"error":"..."}` |

POST body JSON: `{"title":"...","url":"..."}`. Validation in `Create`: title non-empty, max 200 runes; URL non-empty, must start with `http://` or `https://`. Reject static paths containing `..`.

## Entrypoint wiring (`linkshelf/cmd/server/main.go`)

5. `http.ListenAndServe(":8080", nil)` and log `listening on :8080`.

## Data model and storage

### SQLite table (DDL in `linkshelf/internal/store/schema.go`)

sql
CREATE TABLE IF NOT EXISTS links (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    url TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

### Go struct in `linkshelf/internal/store/schema.go`

type Link struct {
    ID        int64  `json:"id"`
    Title     string `json:"title"`
    URL       string `json:"url"`
    CreatedAt string `json:"created_at"` // RFC3339 UTC
}

func InitSchema(db *sql.DB) error

`InitSchema` runs the CREATE TABLE statement only. It is called once at startup and in any tests.

### Store functions in `linkshelf/internal/store/store.go`

var DB *sql.DB

func List(ctx context.Context) ([]Link, error)
func Create(ctx context.Context, title, url string) (Link, error)
func Delete(ctx context.Context, id int64) error


## Unit tests

The SPEC states tests are optional and not required for this MVP. If implement beads add tests, they belong in `linkshelf/internal/store/store_test.go` and `linkshelf/internal/api/handlers_test.go`. Test patterns:


The full-suite command is `cd linkshelf && go test ./...`. Polecat runs this during implementation.

## Integration and testing

Pipeline verification sequence:

1. `cd linkshelf && go mod tidy` — resolve dependencies.
2. `cd linkshelf && go test ./...` — all packages compile and tests pass (packages without `*_test.go` still compile).
3. `cd linkshelf && go run ./cmd/server` — starts HTTP server on `:8080`. UI loads in browser; user can add a link via form, see it in the list, and delete it.

## Acceptance mapping

| SPEC requirement | Architecture coverage |
|------------------|-----------------------|
| `cd linkshelf && go test ./...` passes | All internal packages compile; tests (if added) use `:memory:` DB and pass. |
| `cd linkshelf && go run ./cmd/server` serves UI on `:8080` | Entrypoint opens `linkshelf.db`, inits schema, registers handlers with `http.DefaultServeMux`, listens on `:8080`. |
| GET /api/links returns JSON array | Handler calls `store.List(ctx)`, writes JSON via `json.NewEncoder`. Empty → `[]`. |
| POST /api/links creates with validation | Handler reads `{"title":"...","url":"..."}`, validates per Create rules, calls `store.Create`, returns 201 with link JSON, 400 on error. |
| DELETE /api/links/{id} deletes | Handler extracts ID via path parsing, calls `store.Delete`, returns 204 on success, 404 on missing. |
| Static files served at `/static/{file}` from `linkshelf/web/` | File server at `/static/` stripped prefix, serves from `linkshelf/web/`, rejects `..` paths. |
| Root `/` serves `linkshelf/web/index.html` | Handler serves index.html from web directory. |
| SQLite storage, DDL only in schema.go | `linkshelf/internal/store/schema.go` owns `Link` struct and `InitSchema`; `linkshelf/internal/store/store.go` has no CREATE TABLE. |
| Package-level store API (no Store struct) | `var DB *sql.DB` and public functions `List`, `Create`, `Delete` in `linkshelf/internal/store/store.go`. |
| Frontend CRUD via JS | `linkshelf/web/app.js` calls GET /api/links on load, POST to add, DELETE to remove; refreshes list. |

## Dependency on SPEC symbols

All exported identifiers match SPEC verbatim:
