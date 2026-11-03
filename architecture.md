# Architecture for testgt3

## Overview

Link Shelf is a deliberately small Go bookmark application. It stores links in
SQLite, exposes list/create/delete operations through JSON HTTP endpoints, and
serves a vanilla HTML, JavaScript, and CSS interface from the same process. The
implementation follows the SPEC literally: the module is linkshelf, the server
listens on port 8080, the database file is linkshelf.db in the current working
directory, and schema initialization happens before the package-level store
database is assigned.

The layout root is linkshelf. The store owns SQL access and validation, the API
owns HTTP decoding, status codes, and JSON errors, and the server entrypoint
owns process setup and static-file registration. The existing validated schema
implementation and schema test are authoritative and must be retained as
provided. The store must use package-level functions and the package variable
store.DB; it must not introduce a Store struct or a second schema definition.

## Planned file layout

Every required file is explicitly listed:
- `linkshelf/go.mod` — module declaration and SQLite dependency.
- `linkshelf/internal/store/schema.go` — Link, DDL, and InitSchema only.
- `linkshelf/internal/store/schema_test.go` — schema creation and idempotence tests.
- `linkshelf/internal/store/store.go` — package-level List, Create, and Delete.
- `linkshelf/internal/store/store_test.go` — store behavior tests.
- `linkshelf/internal/api/handlers.go` — HTTP handlers and JSON contract.
- `linkshelf/cmd/server/main.go` — SQLite setup, route registration, and server.
- `linkshelf/internal/api/handlers_test.go` — handler contract tests.
- `linkshelf/web/index.html` — title and URL form plus links list.
- `linkshelf/web/app.js` — load, render, add, and delete browser behavior.
- `linkshelf/web/style.css` — readable responsive presentation.
- `linkshelf/playwright.config.js` — Playwright configuration and server base URL.
- `linkshelf/package.json` — Playwright dependency and test metadata.
- `linkshelf/tests/e2e.spec.js` — required browser UI and static-file tests.

No wildcard, generated directory, or additional production file is required.

## Module / bead ownership

| File | Owns | Must not define |
|---|---|---|
| `linkshelf/internal/store/schema.go` | type Link; func InitSchema(db *sql.DB) error | store queries or HTTP behavior |
| `linkshelf/internal/store/store.go` | var DB *sql.DB; func List(ctx context.Context) ([]Link, error); func Create(ctx context.Context, title, url string) (Link, error); func Delete(ctx context.Context, id int64) error | CREATE TABLE, routes, or a Store struct |
| `linkshelf/internal/api/handlers.go` | HTTP handler registration and request/response helpers | database initialization or frontend markup |
| `linkshelf/cmd/server/main.go` | func main() and process wiring | duplicate store logic or route business rules |
| `linkshelf/internal/store/schema_test.go`, `linkshelf/internal/store/store_test.go` | store package verification | production implementation |
| `linkshelf/internal/api/handlers_test.go` | API verification with HTTP test requests | production implementation |
| `linkshelf/web/index.html`, `linkshelf/web/app.js`, `linkshelf/web/style.css` | browser UI | SQLite access |
| `linkshelf/playwright.config.js`, `linkshelf/package.json`, `linkshelf/tests/e2e.spec.js` | E2E harness | production API behavior |

The store validates a non-empty title of at most 200 runes and a non-empty URL
beginning with http:// or https://. List queries use ORDER BY id DESC and
return a non-nil empty slice when there are no rows. Create records an RFC3339
UTC timestamp and Delete returns an error when the ID is absent. All queries
use context-aware, parameterized database operations.

## HTTP + entrypoint integration

The HTTP contract is exactly:
- GET `/` returns HTTP 200 and linkshelf/web/index.html.
- GET `/static/{file}` returns HTTP 200 for a file under linkshelf/web/ and
  returns HTTP 404 for a missing file. Static paths containing `..` are
  rejected.
- GET `/api/links` returns HTTP 200 and a JSON array, including [] when empty.
- POST `/api/links` accepts {"title":"...","url":"..."} and returns HTTP 201
  with the created JSON link.
- DELETE `/api/links/{id}` returns HTTP 204 on success and HTTP 404 with
  {"error":"..."} when the link is absent.
- Invalid POST input returns HTTP 400 with {"error":"..."}.

The handlers use store.List, store.Create, and store.Delete only. They enforce
methods and parse numeric delete IDs, encode JSON with the Link field names
from schema.go, and provide deterministic JSON error objects. API paths are
registered on http.DefaultServeMux from the entrypoint. Static handling is
registered without allowing an API path to fall through to a file response.

The entrypoint opens SQLite at linkshelf.db in the current working directory,
calls store.InitSchema(db), assigns store.DB = db, registers API and static
routes, logs listening on :8080, and calls http.ListenAndServe(":8080", nil).
The root and static handlers resolve files below linkshelf/web. The database is
closed when the process exits.

## Unit tests

`linkshelf/internal/store/schema_test.go` preserves the supplied in-memory
SQLite tests for table creation and idempotent InitSchema. The store tests use
an in-memory database, initialize the schema, assign store.DB, and verify empty
list shape, validation, creation fields, newest-ID ordering, and successful and
missing-ID deletion. Tests isolate database state and close connections.

`linkshelf/internal/api/handlers_test.go` uses httptest and an initialized
in-memory store. It verifies GET, valid POST, invalid JSON and validation
errors, DELETE success and missing IDs, JSON content, and route/method behavior.
Tests check observable status and response contracts rather than implementation
details.

## Integration and testing

The backend verification command is cd linkshelf && go mod tidy && go test ./....
Each phase-specific Go verification uses cd linkshelf && go build ./....
The frontend phase has its explicitly specified no-automated-tests verification.
The E2E phase runs npm install --ignore-scripts followed by npx playwright test
from linkshelf. The final smoke verification builds the complete server and
checks that the entrypoint remains runnable with the configured static layout.

## E2E / integration testing

`linkshelf/playwright.config.js` defines the test directory, base URL, and
browser settings needed by npx playwright test. `linkshelf/package.json`
declares the Playwright package and a repeatable test script. The concrete
`linkshelf/tests/e2e.spec.js` scenarios verify that the page loads, the title
and URL controls are usable, adding a link renders it, deleting removes it, and
static files are served through the /static/ path. The E2E harness exercises a
running server and the real SQLite-backed API rather than mocked browser data.

## Delivery phases

### backend-setup

Required files are `linkshelf/go.mod`, `linkshelf/internal/store/schema.go`,
and `linkshelf/internal/store/schema_test.go`. Establish module linkshelf with
Go 1.22 and github.com/mattn/go-sqlite3 v1.14.22. Preserve the supplied Link
shape, links table DDL, and idempotent InitSchema behavior. Exit criterion:
cd linkshelf && go build ./... succeeds.

### backend-store

Required files are `linkshelf/internal/store/store.go` and
`linkshelf/internal/store/store_test.go`. Implement the exact package-level DB,
List, Create, and Delete API, including validation and ORDER BY id DESC, without
placing CREATE TABLE in the store. Exit criterion: cd linkshelf && go build ./...
succeeds and store tests cover the specified behavior.

### backend-api

Required files are `linkshelf/internal/api/handlers.go`,
`linkshelf/cmd/server/main.go`, and `linkshelf/internal/api/handlers_test.go`.
Implement the exact API routes and error/status contract, then wire SQLite,
InitSchema, store.DB, DefaultServeMux, and the server listener. Exit criterion:
cd linkshelf && go build ./... succeeds with handlers tested through HTTP.

### frontend-ui

Required files are `linkshelf/web/index.html`, `linkshelf/web/app.js`, and
`linkshelf/web/style.css`. Provide title and URL inputs, an Add button, and
ul id="links"; load through GET /api/links, POST additions, DELETE removals,
and refresh after changes. Exit criterion is the specified phase verification
and a complete readable UI using the live API.

### e2e-testing

Required files are `linkshelf/playwright.config.js`,
`linkshelf/package.json`, and `linkshelf/tests/e2e.spec.js`. Configure the
Playwright harness and test UI load, add/delete flows, and static serving.
Exit criterion: cd linkshelf && npm install --ignore-scripts && npx playwright
test succeeds.

### smoke-test

The required file is `linkshelf/cmd/server/main.go`, which must remain the
runnable entrypoint after all dependent phases. Verify the HTTP server can be
built and responds using the specified port, database setup, API routes, root,
and static paths. Exit criterion: cd linkshelf && go build ./... succeeds.

## Requirements

### backend-setup

The module and schema must match the SPEC data model exactly, with
linkshelf/internal/store/schema.go containing Link and InitSchema and no
additional schema owner. The concrete schema test must pass.

### backend-store

store.DB and the three package-level functions must use the exact signatures
from the SPEC. List returns [] for an empty table, Create enforces title and
URL rules, and Delete errors for a missing ID.

### backend-api

Handlers must match exactly GET /, GET /static/{file}, GET /api/links, POST
/api/links, and DELETE /api/links/{id}, including the stated success and error
responses. main must perform SQLite open, InitSchema, store.DB assignment,
route registration, log listening on :8080, and ListenAndServe.

### frontend-ui

The three concrete web files must implement the specified form, links list,
live API loading, add and delete operations, and simple readable styling.
Static asset URLs must remain under the linkshelf/web mapping.

### e2e-testing

The concrete Playwright files must run with npx playwright test and verify UI
load, add, delete, and static file serving against the running application.

### smoke-test

The completed server must build and serve the UI on :8080 while preserving the
database and route wiring required by the SPEC.

## Docker & Deployment

No Dockerfile or docker-compose file appears in the workflow required files or
the SPEC, so Docker artifacts are intentionally omitted. Deployment consists
of the Go executable, linkshelf/web static assets, and a writable linkshelf.db
in the current working directory. The prescribed port is 8080.

## Acceptance mapping

The database schema and store satisfy durable list/create/delete bookmark
behavior and exact validation. The API exposes the required JSON surface and
protects static path traversal. The entrypoint establishes dependencies in the
required order and serves the UI and API from one process. The frontend gives
users the complete add/list/delete flow, while the concrete unit, handler, and
Playwright files map to the workflow's backend, UI, E2E, and smoke acceptance
checks.
