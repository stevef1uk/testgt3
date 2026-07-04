# Architecture for testgt3 — Link Shelf MVP

## Overview

The Link Shelf MVP is a minimal bookmark service implemented in Go. It provides a REST API to list, create, and delete bookmarks (links), and serves a static web UI. The system uses a SQLite database for persistence, package-level store functions for data access, and the standard `net/http` mux for routing. The entire Go module lives under `linkshelf/` as defined in the project layout.

## Planned file layout

All implement paths are relative to the `linkshelf/` root directory as required by the pipeline.

- `linkshelf/go.mod` — Go module definition (`module linkshelf`, Go 1.22, dependency `github.com/mattn/go-sqlite3`)
- `linkshelf/cmd/server/main.go` — application entrypoint: opens database, inits schema, assigns DB to store, registers handlers, starts HTTP server on :8080
- `linkshelf/internal/store/schema.go` — data model (`Link` struct) and `InitSchema` DDL function
- `linkshelf/internal/store/store.go` — package-level store API (`List`, `Create`, `Delete`) and package variable `DB`
- `linkshelf/internal/api/handlers.go` — HTTP handler functions registered on `http.DefaultServeMux`
- `linkshelf/web/index.html` — HTML page with form to add links and a list element
- `linkshelf/web/app.js` — client-side JavaScript to interact with the REST API
- `linkshelf/web/style.css` — minimal styling for the UI

## Go package / bead ownership

The `linkshelf/internal/store/` directory contains two `.go` files sharing the `store` package. To avoid duplication of exported symbols, ownership is explicitly assigned:

| File | Owns (exported) | Must not define |
|------|----------------|-----------------|
| `linkshelf/internal/store/schema.go` | type `Link` struct, func `InitSchema(db *sql.DB) error` — DDL and type definition only | store functions (`List`, `Create`, `Delete`), package variable `DB` |
| `linkshelf/internal/store/store.go` | var `DB *sql.DB`, func `List(context.Context) ([]Link, error)`, func `Create(context.Context, string, string) (Link, error)`, func `Delete(context.Context, int64) error` — all store behavior | type `Link`, func `InitSchema` — must not redefine them |

The `linkshelf/internal/api/` directory contains a single file `handlers.go` (package `api`). It depends on the `store` package symbols (List, Create, Delete, Link) and defines HTTP handlers. No exported types beyond those needed for testing (if any) are required.

## HTTP + entrypoint integration

### Route table (copied verbatim from SPEC)

| Method | Path | Success | Error |
|--------|------|---------|-------|
| GET | `/` | 200, `linkshelf/web/index.html` | — |
| GET | `/static/{file}` | 200, file under `linkshelf/web/` | 404 |
| GET | `/api/links` | 200, JSON array `[]` when empty | — |
| POST | `/api/links` | 201, JSON link | 400 `{"error":"..."}` |
| DELETE | `/api/links/{id}` | 204 | 404 `{"error":"..."}` |

Static file serving must reject paths containing `..`. POST body is JSON with `title` and `url` fields. DELETE path parameter `{id}` is parsed as an integer.

### Entrypoint wiring (`linkshelf/cmd/server/main.go`)

   - `GET /` → serve `linkshelf/web/index.html` as the default page (either via `http.FileServer` or a custom handler that redirects to `/static/index.html` or serves it directly).
   - `GET /static/{file}` → `http.StripPrefix` + `http.FileServer` pointing to the `linkshelf/web/` directory.
   - `GET /api/links` → handler calling `store.List(ctx)` and writing JSON.
   - `POST /api/links` → handler reading body JSON, calling `store.Create(ctx, title, url)`, and writing 201 with link JSON, or 400 on validation failure.
   - `DELETE /api/links/{id}` → handler parsing id, calling `store.Delete(ctx, id)`, and writing 204 or 404.

All handlers are registered on `http.DefaultServeMux` (no third-party router needed). The entrypoint imports `linkshelf/internal/store`, `linkshelf/internal/api`, and `linkshelf/internal/store` (for `InitSchema`).

## Unit tests

The SPEC states that tests are optional for this MVP, but the pipeline requires `go test ./...` to pass. The following test files may be added (not required but recommended for completeness):

- `linkshelf/internal/store/store_test.go` — test `List`, `Create`, `Delete` against an in-memory SQLite database. Steps: open `:memory:`, call `InitSchema`, assign `store.DB = db`. Test each function with valid and invalid inputs (empty title, too-long title, missing URL, bad scheme, missing ID for delete).
- `linkshelf/internal/api/handlers_test.go` — test HTTP handlers using `httptest.NewRecorder` and `httptest.NewRequest`. Tests for each endpoint: GET `/api/links` returns `[]`, POST creates a link and returns 201, DELETE returns 204, 404 for unknown ID, 400 for invalid input.

If no test files exist, `go test ./...` must still pass because Go does not fail on missing test files.

## Integration and testing


Data flow: HTTP request → handler → store function → SQLite → response. SQLite database file `linkshelf.db` is created in the current working directory on startup.

## Acceptance mapping


## Delivery phases

The rig profile specifies a single delivery phase (bead), but architecture accounts for a potential multi-file implementation order:


Each phase is independently testable with `go test ./linkshelf/internal/store/...` etc. The full integration test uses `go test ./...` from the `linkshelf/` directory.
