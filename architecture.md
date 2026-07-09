# Architecture for testgt3 — Link Shelf MVP

## Overview

Link Shelf is a minimal Go web application that stores and serves bookmark links. It uses a SQLite‑backed data store, a clean HTTP API, and a small static frontend. The design follows the SPEC exactly: package‑level store functions, a single `Link` struct defined in `schema.go`, and HTTP handlers registered on the default mux. The project root is `linkshelf/`.

## Planned file layout (implement paths with `linkshelf/` prefix)

| Path (within `linkshelf/`) | Purpose |
|----------------------------|---------|
| `linkshelf/go.mod` | Go module declaration (`module linkshelf`) with Go 1.22 and the `github.com/mattn/go-sqlite3` driver |
| `linkshelf/cmd/server/main.go` | Entrypoint: opens `linkshelf.db`, calls `schema.InitSchema`, assigns the DB to `store.DB`, registers routes, and serves on `:8080` |
| `linkshelf/internal/store/schema.go` | Defines the `Link` struct and `InitSchema(db *sql.DB) error` – the single source of DDL (`CREATE TABLE IF NOT EXISTS links`) |
| `linkshelf/internal/store/store.go` | Package‑level functions `List`, `Create`, `Delete` that operate on the global `var DB *sql.DB`; includes validation (title ≤ 200 runes, URL must start with `http://` or `https://`) |
| `linkshelf/internal/api/handlers.go` | HTTP handlers registered on `http.DefaultServeMux`; each handler calls the appropriate `store` function |
| `linkshelf/web/index.html` | Static HTML with title & URL inputs, an **Add** button, and `<ul id="links"></ul>` for rendering the list |
| `linkshelf/web/app.js` | Front‑end JavaScript: on load `GET /api/links`, renders list, `POST /api/links` to add, `DELETE /api/links/{id}` to remove, and refreshes after each mutation |
| `linkshelf/web/style.css` | Simple CSS for a readable layout (no external frameworks) |

## Go package / bead ownership

`linkshelf/internal/store/` contains two source files that share a package. Symbol ownership is split deliberately:

| File | Owns (exported) | Must not define |
|------|-----------------|-----------------|
| `linkshelf/internal/store/schema.go` | `type Link struct { ID int64; Title string; URL string; CreatedAt string }` ; `func InitSchema(db *sql.DB) error` | `var DB *sql.DB` ; `func List` ; `func Create` ; `func Delete` ; any DDL other than inside `InitSchema` |
| `linkshelf/internal/store/store.go` | `var DB *sql.DB` ; `func List(ctx context.Context) ([]Link, error)` ; `func Create(ctx context.Context, title, url string) (Link, error)` ; `func Delete(ctx context.Context, id int64) error` | `type Link` ; `func InitSchema` ; `CREATE TABLE` SQL statements |

The `linkshelf/internal/api/` package contains only `handlers.go` and therefore owns the HTTP handler functions. It imports `linkshelf/internal/store` and calls the exported `store.List`, `store.Create`, and `store.Delete`.

## HTTP + entrypoint integration

### Route table (exactly as SPEC specifies)

| Method | Path | Success | Error |
|--------|------|---------|-------|
| GET | `/` | 200, serves `linkshelf/web/index.html` | — |
| GET | `/static/{file}` | 200, returns file under `linkshelf/web/` | 404 if not found |
| GET | `/api/links` | 200, JSON array `[]` when no links exist | — |
| POST | `/api/links` | 201, JSON representation of the newly created link | 400 `{"error":"..."}` on validation failure |
| DELETE | `/api/links/{id}` | 204, no body | 404 `{"error":"..."}` if the id does not exist |

Static route handling must reject any path that contains `..` to prevent directory traversal.

### Entrypoint wiring (`linkshelf/cmd/server/main.go`)

   * `http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "linkshelf/web/index.html") })`
   * `http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("linkshelf/web"))))`
   * `http.HandleFunc("/api/links", apiLinksHandler)` – handles both GET (list) and POST (create)
   * `http.HandleFunc("/api/links/", apiLinkDeleteHandler)` – extracts the `{id}` segment and calls `store.Delete`

No separate `Store` struct, no `NewStore`, and no `RegisterHandlers` helper; all wiring is performed directly inside `main.go` using the default mux.

## Data model

The `Link` struct lives in `linkshelf/internal/store/schema.go`:

type Link struct {
    ID        int64  `json:"id"`
    Title     string `json:"title"`
    URL       string `json:"url"`
    CreatedAt string `json:"created_at"` // RFC3339 UTC timestamp
}

func InitSchema(db *sql.DB) error // runs: CREATE TABLE IF NOT EXISTS links (id INTEGER PRIMARY KEY AUTOINCREMENT, title TEXT NOT NULL, url TEXT NOT NULL, created_at TEXT NOT NULL DEFAULT (datetime('now')))

Only `InitSchema` contains DDL. `store.go` never runs any `CREATE TABLE` statements; it simply assumes the schema exists.

## Store API — package‑level functions

Implemented in `linkshelf/internal/store/store.go`:

var DB *sql.DB

func List(ctx context.Context) ([]Link, error)       // ORDER BY id DESC; returns an empty slice (not nil) when there are no rows
func Create(ctx context.Context, title, url string) (Link, error)
func Delete(ctx context.Context, id int64) error     // Returns an error when the provided id does not exist

**Create** validation rules (exactly as SPEC requires):
* `title` must be non‑empty and contain at most 200 Unicode runes.
* `url` must be non‑empty and begin with the literal prefix `http://` or `https://`.

If validation fails, `Create` returns a Go error; the HTTP handler converts this to a `400` response with a JSON payload `{"error":"…"}`.

## Unit tests (optional but supported)

Although the SPEC says test files are not required, the following test layout would satisfy the pipeline if developers choose to add them:

* **Store tests** – `linkshelf/internal/store/store_test.go`  

* **Handler tests** – `linkshelf/internal/api/handlers_test.go`  

The CI pipeline runs `cd linkshelf && go test ./...`; packages without test files still compile, satisfying the “no missing imports” requirement.

## Integration and testing workflow

* **Compilation & unit testing**  
  cd linkshelf && go mod tidy && go test ./...
  All packages must compile, and any optional tests must pass.

* **Running the server**  
  cd linkshelf && go run ./cmd/server
  The server must start, log `listening on :8080`, and serve the static UI. Visiting `http://localhost:8080` should load `linkshelf/web/index.html`. Adding a link via the UI must result in a POST to `/api/links`, receive a `201` response with the created link JSON, and display the new entry. Deleting the entry must send a DELETE request, receive a `204`, and remove the item from the UI list.

* **Persistence**  
  The SQLite file `linkshelf.db` lives in the process’s working directory. All store operations (`List`, `Create`, `Delete`) use the global `store.DB`, ensuring the same DB is used for the lifetime of the server.

## Acceptance mapping

| SPEC requirement | Architecture fulfillment |
|------------------|--------------------------|
| **List links** | `GET /api/links` → `store.List` returns all rows ordered by `id DESC`; when empty, returns `[]` (empty slice) |
| **Create link** | `POST /api/links` → `store.Create` validates inputs, inserts a row, returns the new `Link` with `201` status |
| **Delete link** | `DELETE /api/links/{id}` → `store.Delete` deletes by primary key, returns `204` on success, `404` if not found |
| **Static UI** | `/` serves `linkshelf/web/index.html`; `/static/{file}` serves files from `linkshelf/web/` (rejects `..` traversals) |
| **SQLite persistence** | `linkshelf.db` opened in `main.go`; `schema.InitSchema` creates the `links` table; all store functions use the global `DB` |
| **Server listening** | `main.go` calls `http.ListenAndServe(":8080", nil)` after wiring dependencies |
| **Build & test** | `go test ./...` compiles every package; optional tests fully exercise store and handlers |
| **End‑to‑end UI flow** | Front‑end JS (`app.js`) performs the exact API calls defined above, refreshes the list after each mutation, and thus satisfies the user‑experience goal |

## Delivery phases

The MVP is delivered in a single phase. All beads listed in the layout are implemented together, and the full test suite must succeed. No further incremental phases are required.

## Summary

The architecture mirrors the SPEC precisely: a flat Go module under `linkshelf/` with package‑level store functions, a single `Link` struct in `schema.go`, DDL confined to `InitSchema`, HTTP handlers on the default mux, and a static frontend in `linkshelf/web/`. All file paths are prefixed with `linkshelf/` as mandated by the workflow. The design is minimal, literal, and pipeline‑friendly, ensuring that `go test ./...` and `go run ./cmd/server` are sufficient for validation.
