# Link Shelf – MVP spec (pipeline-friendly)

## Goal

A tiny bookmark app: **list**, **create**, and **delete** links. Success for the automated pipeline means:

```bash
cd linkshelf && go mod tidy && go test ./...
```

passes, and `cd linkshelf && go run ./cmd/server` serves the UI on `:8080`.

Keep implementations small and literal. Prefer the exact names and shapes below over extra abstractions.

## Layout (implement beads only)

```
linkshelf/
├── go.mod
├── cmd/server/main.go
├── internal/store/schema.go    # Link + InitSchema + DDL only
├── internal/store/store.go     # List / Create / Delete (package-level)
├── internal/api/handlers.go    # HTTP handlers
└── web/
    ├── index.html
    ├── app.js
    └── style.css
```

**Do not** add `store_test.go` or `handlers_test.go` unless you want extras — they are **not** required for this MVP.

## Module

```
module linkshelf

go 1.22

require github.com/mattn/go-sqlite3 v1.14.22
```

(`go mod tidy` may adjust the sqlite driver version.)

## Data model

`Link` and table DDL live in **`schema.go`** only:

```go
type Link struct {
    ID        int64  `json:"id"`
    Title     string `json:"title"`
    URL       string `json:"url"`
    CreatedAt string `json:"created_at"` // RFC3339 UTC
}

func InitSchema(db *sql.DB) error // runs CREATE TABLE IF NOT EXISTS links (...)
```

`store.go` must **not** contain `CREATE TABLE` — call `InitSchema` from `main` and from any tests.

## Store API (`internal/store/store.go`)

Use **package-level** functions and a package variable — **not** a `Store` struct or `NewStore`:

```go
var DB *sql.DB

func List(ctx context.Context) ([]Link, error)       // ORDER BY id DESC; empty slice not nil
func Create(ctx context.Context, title, url string) (Link, error)
func Delete(ctx context.Context, id int64) error     // error if id missing
```

`Create` validation:

- Title: non-empty, max 200 runes
- URL: non-empty, must start with `http://` or `https://`

## HTTP API (`internal/api/handlers.go`)

Register on `http.DefaultServeMux` from `main`. Use `store.List` / `store.Create` / `store.Delete` only.

| Method | Path | Success | Error |
|--------|------|---------|-------|
| GET | `/` | 200, `linkshelf/web/index.html` | — |
| GET | `/static/{file}` | 200, file under `linkshelf/web/` | 404 |
| GET | `/api/links` | 200, JSON array `[]` when empty | — |
| POST | `/api/links` | 201, JSON link | 400 `{"error":"..."}` |
| DELETE | `/api/links/{id}` | 204 | 404 `{"error":"..."}` |

Reject static paths containing `..`.

POST body JSON: `{"title":"...","url":"..."}`.

## `linkshelf/cmd/server/main.go`

1. Open SQLite file `linkshelf.db` in the current working directory.
2. `InitSchema(db)` then `store.DB = db`.
3. Register handlers and static routes.
4. `http.ListenAndServe(":8080", nil)` and log `listening on :8080`.

## Frontend (`linkshelf/web/`)

- **index.html** — title input, URL input, Add button, `<ul id="links"></ul>`.
- **app.js** — on load, `GET /api/links` and render list; POST to add; DELETE to remove; refresh list after each change.
- **style.css** — simple readable layout (no framework).

## Optional tests

If you add tests: `:memory:` DB, `InitSchema`, `store.DB = db`, then call `List`/`Create`/`Delete`. Handler tests may use `httptest`.

## Definition of done

1. `cd linkshelf && go test ./...` — green (packages without `*_test.go` still compile).
2. `cd linkshelf && go run ./cmd/server` — UI loads; add one link, see it, delete it.
