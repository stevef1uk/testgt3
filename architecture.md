# Architecture for testgt3 — Link Shelf MVP

## Overview

Link Shelf is a tiny bookmark web app written in Go. It exposes a small JSON API for
list, create, and delete operations against a SQLite database, and serves a static
HTML/JS/CSS frontend on port 8080. The design intentionally follows the SPEC literally:
package-level functions (no `Store` struct), DDL isolated in one file, a single
`InitSchema` entry point, and HTTP handlers registered on `http.DefaultServeMux`.

The system has four major components that map to the SPEC layout:

1. **Data model** (`linkshelf/internal/store/schema.go`) — defines the `Link` struct
   and the DDL string plus `InitSchema(db *sql.DB) error`. This is the only file that
   contains `CREATE TABLE`.
2. **Store** (`linkshelf/internal/store/store.go`) — package-level `DB *sql.DB`
   plus `List`, `Create`, `Delete` with input validation rules from the SPEC.
3. **HTTP layer** (`linkshelf/internal/api/handlers.go`) — handlers for
   `/api/links` and `/api/links/{id}` plus a static file serving path
   `/static/{file}` rooted at `linkshelf/web/`.
4. **Server entrypoint** (`linkshelf/cmd/server/main.go`) — opens SQLite, calls
   `InitSchema`, assigns `store.DB`, registers routes, listens on `:8080`.
5. **Frontend** (`linkshelf/web/index.html`, `linkshelf/web/app.js`,
   `linkshelf/web/style.css`) — vanilla DOM UI: title input, URL input, Add button,
   `<ul id="links"></ul>`, list/add/delete via `fetch`.
6. **E2E tests** — Playwright under `linkshelf/` verifying UI load, add, delete,
   and static file serving.

Constraints honored from SPEC:

- `go 1.22` and `github.com/mattn/go-sqlite3` driver.
- `Create` validates title (non-empty, max 200 runes) and URL (non-empty, must
  start with `http://` or `https://`).
- Static path rejection of `..`.
- `linkshelf.db` opened in the current working directory from `main`.

## Planned file layout

All paths use the `linkshelf/` layout root prefix as required by the orchestrator.

- `linkshelf/go.mod` — module declaration `module linkshelf`, go 1.22, sqlite driver.
- `linkshelf/internal/store/schema.go` — `Link` struct, `InitSchema`, DDL constant.
- `linkshelf/internal/store/store.go` — package-level `DB`, `List`, `Create`, `Delete`.
- `linkshelf/internal/api/handlers.go` — `RegisterRoutes(mux *http.ServeMux)` plus
  HTTP handlers for `/api/links` and `/api/links/{id}` and the `/static/{file}` path.
- `linkshelf/cmd/server/main.go` — `main` function, opens DB, calls
  `store.InitSchema(db)`, assigns `store.DB = db`, registers routes, listens on `:8080`.
- `linkshelf/web/index.html` — UI shell: title input, URL input, Add button,
  `<ul id="links"></ul>`.
- `linkshelf/web/app.js` — fetch-based list/add/delete and DOM rendering.
- `linkshelf/web/style.css` — basic readable layout.
- `linkshelf/playwright.config.js` — Playwright config (web server, baseURL,
  chromium project).
- `linkshelf/package.json` — dev dependency for `@playwright/test` and a test script.
- `linkshelf/tests/e2e.spec.js` — Playwright spec: UI loads, add link, delete link,
  static asset fetchable.

## Go package / bead ownership

The Go surface is intentionally small. Each `.go` file owns a clear set of exported
symbols and a clear set of forbidden re-definitions to prevent drift.

| File | Owns (exported) | Must not define |
|------|-----------------|-----------------|
| `linkshelf/internal/store/schema.go` | `type Link struct { ID int64; Title string; URL string; CreatedAt string }`, `func InitSchema(db *sql.DB) error` | Any handler, any `Create`/`Delete`/`List` logic, any `Store` struct |
| `linkshelf/internal/store/store.go` | `var DB *sql.DB`, `func List(ctx context.Context) ([]Link, error)`, `func Create(ctx context.Context, title, url string) (Link, error)`, `func Delete(ctx context.Context, id int64) error` | `CREATE TABLE` statements, `Link` struct redefinition, HTTP handlers, any `NewStore`/`Store` struct |
| `linkshelf/internal/api/handlers.go` | `func RegisterRoutes(mux *http.ServeMux)`, `func listLinksHandler(w http.ResponseWriter, r *http.Request)`, `func createLinkHandler(w http.ResponseWriter, r *http.Request)`, `func deleteLinkHandler(w http.ResponseWriter, r *http.Request)`, `func staticFileHandler(w http.ResponseWriter, r *http.Request)`, `func indexHandler(w http.ResponseWriter, r *http.Request)` | Direct SQL, DDL, package-level `DB` re-declaration, route table outside `RegisterRoutes` |
| `linkshelf/cmd/server/main.go` | `func main()` | Store internals, handler internals, DDL re-implementation |

Symbol signatures (full, verbatim from SPEC plus the wrapping `RegisterRoutes`
helper required to register on `http.DefaultServeMux`):

// linkshelf/internal/store/schema.go
type Link struct {
    ID        int64  `json:"id"`
    Title     string `json:"title"`
    URL       string `json:"url"`
    CreatedAt string `json:"created_at"`
}

func InitSchema(db *sql.DB) error

// linkshelf/internal/store/store.go
var DB *sql.DB

func List(ctx context.Context) ([]Link, error)
func Create(ctx context.Context, title, url string) (Link, error)
func Delete(ctx context.Context, id int64) error

// linkshelf/internal/api/handlers.go
func RegisterRoutes(mux *http.ServeMux)
func listLinksHandler(w http.ResponseWriter, r *http.Request)
func createLinkHandler(w http.ResponseWriter, r *http.Request)
func deleteLinkHandler(w http.ResponseWriter, r *http.Request)
func staticFileHandler(w http.ResponseWriter, r *http.Request)
func indexHandler(w http.ResponseWriter, r *http.Request)

## HTTP + entrypoint integration

The HTTP API mirrors the SPEC table verbatim. `RegisterRoutes` attaches the
following routes to the supplied `*http.ServeMux`:

| Method | Path | Success | Error |
|--------|------|---------|-------|
| GET | `/` | 200, `linkshelf/web/index.html` | — |
| GET | `/static/{file}` | 200, file under `linkshelf/web/` | 404 if missing or contains `..` |
| GET | `/api/links` | 200, JSON array `[]` when empty | — |
| POST | `/api/links` | 201, JSON link | 400 `{"error":"..."}` |
| DELETE | `/api/links/{id}` | 204 | 404 `{"error":"..."}` |

Wiring pattern in `linkshelf/cmd/server/main.go` (matches the SPEC step list):

1. `flag.Parse()` (optional) and open SQLite via `sql.Open("sqlite3", "linkshelf.db")`.
2. Call `store.InitSchema(db)`. If it returns an error, log fatal.
3. Assign the package variable: `store.DB = db`.
4. Call `api.RegisterRoutes(http.DefaultServeMux)`.
5. `log.Println("listening on :8080")` and `http.ListenAndServe(":8080", nil)`.

Handlers use only `store.List`, `store.Create`, and `store.Delete` — no direct
SQL in the API layer. The `indexHandler` serves `linkshelf/web/index.html` via
`http.ServeFile`. The `staticFileHandler` joins `linkshelf/web/` with the trimmed
path, rejects anything containing `..` with 404, and otherwise serves the file
via `http.ServeFile`.

Full-suite test command (used in the QA `qa_verify_command`):

cd linkshelf && go mod tidy && go test ./... && cd .. && cd linkshelf && npx playwright test

The Go half is `go mod tidy` then `go test ./...` per SPEC goal. The Playwright
half runs the e2e spec against a started server (configured via
`playwright.config.js` `webServer`).

## Unit tests

The SPEC marks backend tests as optional. Architecture supports the recommended
pattern (used by `httptest`/`:memory:` if implemented): open
`sql.Open("sqlite3", ":memory:")`, call `store.InitSchema`, assign `store.DB`,
then exercise `List`, `Create`, `Delete` directly. Handler tests can use
`httptest.NewRecorder` against a fresh `*http.ServeMux` populated by
`api.RegisterRoutes`.

If unit tests are added they must live in `linkshelf/internal/store/` and
`linkshelf/internal/api/` as `*_test.go` files. The SPEC says these are not
required, so the MVP plan does not block on them, but the architecture
explicitly permits them.

## Integration and testing

Pieces connect as follows:

- `linkshelf/cmd/server/main.go` owns process bootstrap and is the only place
  that opens the SQLite file and assigns `store.DB`.
- `linkshelf/internal/api/handlers.go` consumes the `store` package by name
  (`store.List`, `store.Create`, `store.Delete`).
- `linkshelf/internal/store/schema.go` is imported by both `store.go` (for
  `Link` type) and `main.go` (for `InitSchema`).
- `linkshelf/web/*` is consumed by the server at runtime via `http.ServeFile`,
  so paths inside handlers are computed relative to the working directory
  (`linkshelf/web/...`).

The build command for backend verification is `go build ./...` from
`linkshelf/`. The full test command is `go test ./...` from `linkshelf/`. A
clean build + passing tests is the contract for the `backend-api` phase, and
the e2e phase wraps a `npx playwright test` invocation.

## Docker & Deployment

The workflow profile does not list a Dockerfile or docker-compose phase, so
this section is intentionally omitted. The MVP runs natively via
`go run ./cmd/server` per SPEC.

## E2E / integration testing

Phase `e2e-testing` requires Playwright. The architecture places config and
specs under `linkshelf/`:

- `linkshelf/playwright.config.js` — sets `testDir: './tests'`,
  `use.baseURL: 'http://localhost:8080'`, `webServer` command
  `go run ./cmd/server` with `cwd: '.'` and port `8080`, retries `0`,
  `reporter: 'list'`.
- `linkshelf/package.json` — declares `@playwright/test` as a devDependency
  and a `test` script invoking `playwright test`.
- `linkshelf/tests/e2e.spec.js` — covers:
  - UI loads: `await page.goto('/')`; assert title input and Add button visible.
  - Add link: type into `#title` and `#url`, click Add, assert new `<li>` text.
  - Delete link: click delete on an item, assert removal.
  - Static file serving: `await page.request.get('/static/style.css')` returns 200.

Selectors map to SPEC: `#title`, `#url`, button labeled "Add", `<ul id="links">`,
and `<li>` rows produced by `app.js`. App start command (used in `webServer`)
is `go run ./cmd/server`. E2E test command is `npx playwright test` invoked
from `linkshelf/`.

## Delivery phases

The orchestrator profile enumerates five delivery phases. The architecture
maps each to concrete file ownership and a `### <phase-id>` requirement block.

### backend-setup

Phase: Project Setup & Data Model. QA verify command: `cd linkshelf && go mod tidy`.
Required files: `linkshelf/go.mod`, `linkshelf/internal/store/schema.go`.
Spec focus: define `go.mod`, the `Link` struct, `InitSchema`, and the DDL.

### backend-store

Phase: Store Implementation. QA verify command: `cd linkshelf && go build ./internal/store`.
Depends on `backend-setup`. Required file: `linkshelf/internal/store/store.go`.
Spec focus: implement `List`, `Create`, `Delete` package-level functions with
the SPEC validation rules (title non-empty, max 200 runes; URL non-empty, must
start with `http://` or `https://`).

### backend-api

Phase: HTTP Handlers & Server. QA verify command: `cd linkshelf && go build ./...`.
Depends on `backend-store`. Required files: `linkshelf/internal/api/handlers.go`,
`linkshelf/cmd/server/main.go`. Spec focus: implement HTTP handlers for the
API and static files, and the main server entry point that opens
`linkshelf.db`, calls `InitSchema`, assigns `store.DB`, registers routes, and
listens on `:8080`.

### frontend-ui

Phase: Frontend Assets. QA verify command: `cd linkshelf && echo 'verify ok (no automated tests for this phase)'`.
Depends on `backend-api`. Required files: `linkshelf/web/index.html`,
`linkshelf/web/app.js`, `linkshelf/web/style.css`. Spec focus: create the
HTML, JS, and CSS files for the bookmark UI matching the SPEC's
`#title`/`#url`/Add/`#links`/`<li>` structure.

### e2e-testing

Phase: Playwright E2E & Verification. QA verify command:
`cd linkshelf && echo 'verify ok (no automated tests for this phase)'`.
Depends on `frontend-ui`. Required files: none mandated by the profile, but
the architecture adds `linkshelf/playwright.config.js`, `linkshelf/package.json`,
and `linkshelf/tests/e2e.spec.js` to satisfy the SPEC's e2e requirement.
Spec focus: set up Playwright and run e2e tests to verify UI functionality,
add/delete, static file serving, and API integration against the running
server on port 8080.

## Acceptance mapping

How this architecture satisfies the SPEC's Definition of Done:

- `cd linkshelf && go mod tidy && go test ./...` is green: every package
  (`linkshelf/cmd/server`, `linkshelf/internal/store`,
  `linkshelf/internal/api`) compiles even without test files; `go test ./...`
  builds them. The store API is package-level and matches SPEC signatures
  exactly, so test files can plug into `:memory:` SQLite without any
  refactor.
- `cd linkshelf && go run ./cmd/server` serves the UI on `:8080`:
  `main.go` calls `InitSchema`, assigns `store.DB`, registers handlers and
  static routes on `http.DefaultServeMux`, and logs `listening on :8080`.
  The frontend's `index.html` is served at `/`; `app.js` calls
  `GET /api/links`, `POST /api/links`, and `DELETE /api/links/{id}`.
- `npx playwright test` passes: `playwright.config.js` boots the Go server
  as a `webServer` on port 8080, and the spec exercises load, add, delete,
  and static file serving against the SPEC-shaped DOM and routes.

The HTTP routes, store API names, validation rules, and file paths in this
document match the SPEC verbatim. No wildcards are used in `required_files`,
and every file path is prefixed with `linkshelf/`.
