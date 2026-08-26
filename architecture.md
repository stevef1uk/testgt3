# Architecture for testgt3

## Overview

This design implements the SPEC's tiny bookmark application: users can list, create, and
delete links, with SQLite persistence and a small browser UI. The success path is the
SPEC's `cd linkshelf && go mod tidy && go test ./...`, followed by
`cd linkshelf && go run ./cmd/server` serving the UI on port 8080. The layout root is
`linkshelf/`, and every planned file path below is explicitly rooted there.

The implementation remains literal. The schema package owns the Link model and DDL, the
store package owns package-level database operations, the API package translates HTTP
requests to store calls, and the server entrypoint performs startup wiring. The browser
uses same-origin API requests. Static assets are kept separate from API routes, and
static path traversal is rejected.

## Planned file layout

All implementation and required frontend-test files are enumerated; no directory
placeholder represents an additional file.

- `linkshelf/go.mod`
- `linkshelf/internal/store/schema.go`
- `linkshelf/internal/store/store.go`
- `linkshelf/internal/api/handlers.go`
- `linkshelf/cmd/server/main.go`
- `linkshelf/web/index.html`
- `linkshelf/web/app.js`
- `linkshelf/web/style.css`
- `linkshelf/playwright.config.js`
- `linkshelf/tests/e2e/link-shelf.spec.js`

The three Go package files and three web assets are the MVP application files specified
by the layout. The Playwright configuration and one concrete browser test file are
required by the SPEC's frontend e2e requirement and are kept separate from Go packages.

## Delivery phases

| Phase / requirement ID | required_files | Delivery result |
|---|---|---|
| backend-core | `linkshelf/go.mod`; `linkshelf/internal/store/schema.go`; `linkshelf/internal/store/store.go`; `linkshelf/internal/api/handlers.go` | Provides the module, exact SQLite schema and package-level store API, plus the exact JSON HTTP API. |
| server-and-main | `linkshelf/cmd/server/main.go` | Opens the prescribed database, initializes schema, assigns the package database, registers the default mux, and listens on port 8080. |
| frontend | `linkshelf/web/index.html`; `linkshelf/web/app.js`; `linkshelf/web/style.css`; `linkshelf/playwright.config.js`; `linkshelf/tests/e2e/link-shelf.spec.js` | Provides the form/list UI, client refresh behavior, readable styling, and required Playwright coverage. |
| smoke-test | `linkshelf/cmd/server/main.go`; `linkshelf/internal/store/schema.go`; `linkshelf/internal/store/store.go`; `linkshelf/internal/api/handlers.go`; `linkshelf/web/index.html`; `linkshelf/web/app.js`; `linkshelf/web/style.css`; `linkshelf/playwright.config.js`; `linkshelf/tests/e2e/link-shelf.spec.js` | Verifies the assembled server, persistence, API round trips, static assets, and browser flow without adding another implementation file. |

## Requirements

### backend-core

`linkshelf/go.mod` declares module linkshelf, Go 1.22, and
github.com/mattn/go-sqlite3 v1.14.22. The schema in
`linkshelf/internal/store/schema.go` alone defines Link with ID int64, Title string,
URL string, and CreatedAt string JSON fields, and exports
InitSchema(db *sql.DB) error. InitSchema runs CREATE TABLE IF NOT EXISTS links with an
autoincrement integer primary key and required title, url, and created_at text columns.

`linkshelf/internal/store/store.go` defines package variable DB *sql.DB and no Store
struct or NewStore function. Its exact exported functions are
List(ctx context.Context) ([]Link, error), Create(ctx context.Context, title, url string)
(Link, error), and Delete(ctx context.Context, id int64) error. List queries ORDER BY id
DESC and returns a non-nil empty []Link when there are no rows. Create rejects an empty
title, a title longer than 200 runes, an empty URL, or a URL not beginning with
http:// or https://; it stores an RFC3339 UTC creation time. Delete returns an error
when the ID is missing. SQL and database errors are returned rather than hidden, and
`linkshelf/internal/store/store.go` contains no CREATE TABLE statement.

`linkshelf/internal/api/handlers.go` registers only the SPEC API behavior and calls
store.List, store.Create, and store.Delete. GET `/api/links` returns HTTP 200 and a
JSON array, including `[]` for an empty database. POST `/api/links` accepts exactly a
JSON object with title and url, returns the created JSON Link with HTTP 201, and returns
HTTP 400 with `{"error":"..."}` for malformed or invalid input. DELETE
`/api/links/{id}` returns HTTP 204 for success and HTTP 404 with
`{"error":"..."}` for an invalid or missing link. Unsupported methods are rejected
without changing the database, and JSON responses use application/json.

### server-and-main

`linkshelf/cmd/server/main.go` opens SQLite file linkshelf.db in the current working
directory, calls store.InitSchema(db), then assigns store.DB = db. It registers API
handlers on http.DefaultServeMux and registers the root and static routes before
calling http.ListenAndServe(":8080", nil). Startup failures are logged and terminate;
successful startup logs listening on :8080. GET `/` returns
`linkshelf/web/index.html` with HTTP 200. GET `/static/{file}` serves only the requested
file under `linkshelf/web/`, returns HTTP 200 for valid existing assets, and returns
HTTP 404 for absent files or any static path containing `..`. The static handler must
not allow an API path or a traversal path to collide with the API.

The entrypoint wiring order is deliberate: open database, call schema initialization,
assign the package-level database, register handlers on the default mux, register
static serving, and start the server. No separate Store object, migration runner, or
frontend build process is introduced.

### frontend

`linkshelf/web/index.html` contains a title input, a URL input, an Add button, and
`<ul id="links"></ul>`, with script and stylesheet references matching the server's
static paths. `linkshelf/web/app.js` loads GET `/api/links` on page load, renders each
link, POSTs `{"title":"...","url":"..."}` when the form is submitted, DELETEs the
selected link, and refreshes the list after each change. Rendered user data is treated
as text and URLs are restricted to the supported HTTP schemes before being placed in
link attributes. Recoverable failures are shown in the page rather than silently
ignored. `linkshelf/web/style.css` supplies a simple readable layout with no framework.

`linkshelf/playwright.config.js` configures Playwright to run the concrete browser test
against the already-started server at http://127.0.0.1:8080. The test in
`linkshelf/tests/e2e/link-shelf.spec.js` verifies that the UI loads, the expected form
and list are visible, a valid link can be added and displayed, that link can be
deleted, and `/static/app.js` is served successfully. The test uses stable semantic
labels or IDs from `linkshelf/web/index.html` and cleans up its created data through
the UI.

### smoke-test

The assembled acceptance checks run `cd linkshelf && go mod tidy && go test ./...`,
then start with `cd linkshelf && go run ./cmd/server`. A smoke run confirms that an
empty database is initialized, GET `/api/links` produces a JSON array, POST creates a
link, GET lists it newest first, DELETE removes it and rejects a missing ID, and a
second server/database open preserves rows. It also confirms GET `/` serves the page,
GET `/static/app.js` and GET `/static/style.css` serve the assets, and traversal input
under `/static/` returns 404.

With the server running, `cd linkshelf && npx playwright test` executes the required
browser checks. The test environment uses a disposable working directory or removes
the generated linkshelf.db after the run; it does not rely on pre-existing records.
The final acceptance is the SPEC's UI flow: load, add one HTTP(S) link, observe it in
the list, and delete it.

## Go package / bead ownership

| File | Owns (exported symbols and signatures) | Must not define |
|---|---|---|
| `linkshelf/internal/store/schema.go` | Link; InitSchema(db *sql.DB) error | Store operations, HTTP routes, or UI |
| `linkshelf/internal/store/store.go` | DB *sql.DB; List(ctx context.Context) ([]Link, error); Create(ctx context.Context, title, url string) (Link, error); Delete(ctx context.Context, id int64) error | Link or DDL ownership, handlers, or static serving |
| `linkshelf/internal/api/handlers.go` | HTTP handler registration and route functions using store.List, store.Create, and store.Delete | SQL, database opening, or a second persistence API |
| `linkshelf/cmd/server/main.go` | main() and startup wiring | Duplicated SQL queries or alternate mux/server startup |
| `linkshelf/web/index.html`, `linkshelf/web/app.js`, `linkshelf/web/style.css` | Document, browser behavior, and styles | Go symbols or alternate route names |
| `linkshelf/playwright.config.js`, `linkshelf/tests/e2e/link-shelf.spec.js` | Playwright runner configuration and e2e assertions | Production server behavior or a second UI |

## HTTP + entrypoint integration

The API route is exact: GET `/api/links`, POST `/api/links`, and DELETE
`/api/links/{id}`. The root route is GET `/`, and the static route is GET
`/static/{file}`. The API handler receives the request context and uses only the
package-level store functions. The server uses http.DefaultServeMux as required, so
the entrypoint registers these handlers before calling http.ListenAndServe(":8080", nil).
Static serving is rooted at `linkshelf/web/`, while the browser requests `/static/app.js`
and `/static/style.css`; `/api/links` is never passed to the static handler.

## Unit tests

Backend tests are optional in the SPEC, so no extra Go test file is required in the
planned layout. If tests are added during implementation, they use an in-memory SQLite
database, call InitSchema, assign store.DB, and exercise List, Create, and Delete with
contexts. Handler checks use httptest and cover exact methods, payload validation,
status codes, JSON array shape, missing IDs, and traversal rejection without changing
the required production file layout.

## Integration and testing

The full Go verification command is `cd linkshelf && go mod tidy && go test ./...`.
The frontend verification command is `cd linkshelf && npx playwright test` while the
server is running on port 8080. Integration coverage crosses schema initialization,
package-level store state, handlers, default mux registration, static file serving,
and the browser. It checks that an empty result is encoded as `[]`, not null, and that
created links retain title, URL, ID, and RFC3339 UTC timestamp.

## E2E / integration testing

Start the application with `cd linkshelf && go run ./cmd/server`; the process must log
listening on :8080. Run `cd linkshelf && npx playwright test`. The Playwright test
navigates to http://127.0.0.1:8080/, fills the title and URL inputs, clicks Add, waits
for the new link in `ul#links`, clicks its delete control, and asserts it disappears.
It separately requests `/static/app.js` and expects HTTP 200. The test uses the IDs
provided by `linkshelf/web/index.html`, does not assume a frontend bundler, and leaves
the server responsible for all API and asset responses.

## Acceptance mapping

The backend-core phase maps directly to the SPEC's exact data model, validation rules,
package-level store API, and JSON routes. The server-and-main phase maps the required
SQLite filename, initialization order, default mux, static path safety, and port.
The frontend phase maps the required inputs, list element, add/delete refresh behavior,
readable CSS, and Playwright artifacts. The smoke-test phase demonstrates the complete
definition of done: Go compilation and tests, a live UI at port 8080, working CRUD, and
successful `npx playwright test` coverage.
