# Architecture for testgt3

## Overview

This architecture implements the SPEC's Link Shelf application: a small same-origin
bookmark manager that persists links in SQLite, exposes the prescribed JSON HTTP API,
and serves a browser interface for listing, adding, and deleting links. The layout root
is `linkshelf/`; all application, browser, and end-to-end test files are rooted there.
The design deliberately keeps the dependency graph small: the store owns SQLite access,
the API owns HTTP validation and JSON responses, the server owns startup and static
serving, and the browser owns presentation and fetch calls.

The application must work with the SPEC's normal commands from the layout root:
`go mod tidy`, `go test ./...`, and `go run ./cmd/server`. The server listens on port
8080, initializes the database before registering handlers, and serves the web assets
without allowing a request to escape the web directory. SQLite errors, malformed input,
and missing records are converted to the status codes and response shapes specified by
the SPEC rather than being silently ignored.

## Planned file layout

Every file that the implementation creates is listed; there are no directory
placeholders or implied generated files.

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

The Go module uses the SQLite driver required by the SPEC. The web files are plain
HTML, JavaScript, and CSS rather than a second build system, and the Playwright files
are concrete runnable coverage for the required browser workflow.

## Delivery phases

| Phase / requirement ID | required_files | Delivery result |
|---|---|---|
| backend-core | `linkshelf/go.mod`; `linkshelf/internal/store/schema.go`; `linkshelf/internal/store/store.go`; `linkshelf/internal/api/handlers.go` | Supplies the Go module, exact schema, package-level store API, and exact JSON routes. |
| server-and-main | `linkshelf/cmd/server/main.go` | Opens the prescribed SQLite database, initializes it, connects the store to the default mux, serves static assets, and listens on port 8080. |
| frontend | `linkshelf/web/index.html`; `linkshelf/web/app.js`; `linkshelf/web/style.css`; `linkshelf/playwright.config.js`; `linkshelf/tests/e2e/link-shelf.spec.js` | Supplies the accessible form, link list, client refresh/delete behavior, styling, and browser test. |
| smoke-test | `linkshelf/go.mod`; `linkshelf/internal/store/schema.go`; `linkshelf/internal/store/store.go`; `linkshelf/internal/api/handlers.go`; `linkshelf/cmd/server/main.go`; `linkshelf/web/index.html`; `linkshelf/web/app.js`; `linkshelf/web/style.css`; `linkshelf/playwright.config.js`; `linkshelf/tests/e2e/link-shelf.spec.js` | Validates the complete assembled application without introducing an unlisted implementation file. |

## Requirements

### backend-core

`linkshelf/go.mod` declares module linkshelf, the Go version required by the SPEC,
and github.com/mattn/go-sqlite3 at v1.14.22. The schema package in
`linkshelf/internal/store/schema.go` is the sole owner of the Link model and schema
DDL. Link has ID int64, Title string, URL string, and CreatedAt string fields with the
JSON names required by the SPEC. It exports InitSchema(db *sql.DB) error. That function
executes CREATE TABLE IF NOT EXISTS links with an autoincrement integer primary key and
required title, url, and created_at text columns.

`linkshelf/internal/store/store.go` defines the package variable DB *sql.DB and does
not add a Store struct or a second initialization API. Its exported operations have
these exact signatures: List(ctx context.Context) ([]Link, error), Create(ctx
context.Context, title, url string) (Link, error), and Delete(ctx context.Context,
id int64) error. List selects the four Link columns and orders by id DESC; an empty
database returns a non-nil empty slice. Create rejects an empty title, a title over
200 runes, an empty URL, or a URL that does not begin with http:// or https://. It
stores the current time as RFC3339 UTC. Delete returns an error for an absent ID.
Database and validation errors remain distinguishable to the API, and this file does
not contain CREATE TABLE statements.

`linkshelf/internal/api/handlers.go` uses the package-level store API and registers
only the SPEC routes. GET /api/links returns 200 and a JSON array. POST /api/links
accepts JSON containing title and url, creates a link, and returns 201 with the
created Link JSON. Invalid JSON or validation failure returns 400 with a JSON error
object. DELETE /api/links/{id} deletes the numeric ID and returns 204 on success;
invalid IDs and missing links return 404 with a JSON error object. Unsupported
methods receive the appropriate 405 response. JSON responses set
application/json, and the handlers do not issue SQL directly.

### server-and-main

`linkshelf/cmd/server/main.go` is the only executable entrypoint. It parses the
database path and port configuration described by the SPEC (using the prescribed
defaults), opens SQLite, calls store.InitSchema(db) before serving requests, assigns
the opened connection to store.DB, and registers the API handlers on the default
HTTP mux. It then registers static serving for the web assets and starts the server
on port 8080 unless configuration overrides the port. Static requests serve
`linkshelf/web/index.html`, `linkshelf/web/app.js`, and `linkshelf/web/style.css`;
API routes are registered before the static fallback so they cannot collide. The
entrypoint closes the database on shutdown and fails fast on open, schema, or listen
errors.

### frontend

`linkshelf/web/index.html` contains the Link Shelf heading, an add-link form with
stable labels and selectors for title and URL, a submit control, and a list container
with a useful empty state. `linkshelf/web/app.js` loads GET /api/links on startup,
submits JSON to POST /api/links, renders title and URL safely as text, and places a
delete control beside every rendered link. Deletion calls DELETE /api/links/{id} and
refreshes the list. It reports failed requests visibly and prevents duplicate form
submission while a request is in progress. `linkshelf/web/style.css` provides readable
spacing, responsive layout, visible focus states, and distinct controls without
requiring external assets or a network CDN.

`linkshelf/playwright.config.js` starts the Go server for browser tests, uses the
SPEC's port 8080 base URL, waits for the server to be reachable, and runs the concrete
test in `linkshelf/tests/e2e/link-shelf.spec.js`. The test creates a uniquely named
HTTP(S) link through the UI, verifies it appears in the list, reloads to verify SQLite
persistence, and deletes it through its row-level control. Selectors should prefer
accessible labels, the form, and row text rather than CSS implementation details.

### smoke-test

The assembled implementation must pass `go test ./...` from `linkshelf/`, then the
Playwright command configured by `linkshelf/playwright.config.js`. The smoke path
checks schema creation on a fresh database, descending list order, create validation,
missing-ID deletion, all route status codes and JSON shapes, same-origin asset loading,
and the browser create/reload/delete flow. Tests use temporary databases or unique
records so they do not depend on a pre-existing database or test order.

## Go package / bead ownership

| File | Owns (exported) | Must not define |
|---|---|---|
| `linkshelf/internal/store/schema.go` | Link (struct); InitSchema(db *sql.DB) error | Store operations, HTTP handlers, or duplicate schema ownership |
| `linkshelf/internal/store/store.go` | DB (*sql.DB); List(ctx context.Context) ([]Link, error); Create(ctx context.Context, title, url string) (Link, error); Delete(ctx context.Context, id int64) error | Link redefinition, CREATE TABLE, route registration, or JSON response logic |
| `linkshelf/internal/api/handlers.go` | RegisterRoutes(mux *http.ServeMux) | Direct SQL, schema initialization, or static-file implementation |
| `linkshelf/cmd/server/main.go` | main() | Store SQL, handler business rules, or duplicate route definitions |

The API package reads the shared store DB only through the named store functions. The
entrypoint performs dependency wiring in this order: flags and database open, schema
initialization, DB assignment, API registration, static registration, then listen.

## Unit tests

Store-focused tests, whether kept alongside the specified store package during
implementation, cover schema initialization, the exact Link fields, empty-list
non-nil behavior, validation boundaries, RFC3339 timestamps, descending ordering,
context propagation, and deletion of existing versus missing IDs. API tests use an
httptest server and a temporary SQLite database to cover every route, malformed JSON,
unsupported methods, content types, status codes, and error objects. No test changes
the public signatures or creates a second schema.

## Integration and testing

From `linkshelf/`, run `go mod tidy` and `go test ./...`. Start the application with
`go run ./cmd/server`; confirm that GET /api/links is JSON while the root page and
its two referenced assets are static content. Run the Playwright suite through the
configuration in `linkshelf/playwright.config.js`. The e2e server uses an isolated
temporary database where possible, and the test data includes a valid https URL and a
distinct title so assertions cannot match stale rows.

## E2E / integration testing

The exact browser scenario in `linkshelf/tests/e2e/link-shelf.spec.js` opens the root
page, checks the Link Shelf heading and empty/list UI, fills the labeled title and URL
fields, submits, waits for the created title and URL, reloads and confirms persistence,
then clicks the delete control associated with that link and confirms it disappears.
The Playwright web server command is `go run ./cmd/server` from `linkshelf/`, and its
base URL is http://127.0.0.1:8080. The test must not call the API directly for the
create or delete assertions; those actions exercise the actual UI.

## Acceptance mapping

The SQLite schema and package-level store satisfy the SPEC persistence goal and keep
validation deterministic. The three API routes provide list, create, and delete
behavior with the prescribed contracts. Explicit startup ordering guarantees the
database is usable before the first request. Static routing and same-origin fetches
satisfy the no-build browser goal, while accessible controls and safe text rendering
satisfy usability and basic injection-safety expectations. The concrete Playwright
test demonstrates the complete acceptance path: create a link, observe it, reload
the page, and remove it.
