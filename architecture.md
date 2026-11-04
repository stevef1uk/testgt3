# Architecture for testgt3

## Overview
Link Shelf is the SPEC-defined minimal bookmark CRUD service: a Go 1.22 backend, SQLite persistence, exact HTTP API routes, and a vanilla browser UI served on port 8080. The design keeps schema ownership, store behavior, HTTP handlers, entrypoint wiring, static assets, and Playwright verification in separate delivery phases. When run with `cd linkshelf && go run ./cmd/server`, the server opens `linkshelf.db` in the current working directory. Optional backend tests use isolated in-memory databases initialized through InitSchema. All implementation paths below are repository-relative and explicitly rooted at linkshelf/.

## Planned file layout
- `linkshelf/go.mod` — Go module manifest and SQLite dependency.
- `linkshelf/internal/store/schema.go` — Link data model and idempotent schema initialization.
- `linkshelf/internal/store/store.go` — package-level DB and List, Create, Delete persistence API.
- `linkshelf/internal/api/handlers.go` — exact API, root, and static-file handlers.
- `linkshelf/cmd/server/main.go` — executable startup and dependency wiring.
- `linkshelf/web/index.html` — UI structure, title and URL inputs, Add button, and links list.
- `linkshelf/web/app.js` — browser CRUD client and refresh behavior.
- `linkshelf/web/style.css` — framework-free readable presentation styles.
- `linkshelf/package.json` — Playwright dependency and test script.
- `linkshelf/playwright.config.js` — browser test configuration for port 8080.
- `linkshelf/tests/e2e.spec.js` — end-to-end UI, static-serving, add, and delete coverage.

## Module / bead ownership
`linkshelf/internal/store/schema.go` owns the Link struct and InitSchema function and owns all links-table DDL. `linkshelf/internal/store/store.go` owns the package-level DB plus List, Create, and Delete; it must not duplicate schema DDL. `linkshelf/internal/api/handlers.go` owns HTTP registration and request/response behavior and calls only the store API. `linkshelf/cmd/server/main.go` owns process startup, SQLite opening, schema initialization, store assignment, static registration, and listening. Frontend files own only browser presentation and fetch behavior; Playwright files own only browser orchestration.

| File | Owns | Must not define |
|---|---|---|
| `linkshelf/internal/store/schema.go` | Link and InitSchema | HTTP handlers or CRUD queries |
| `linkshelf/internal/store/store.go` | DB, List, Create, Delete | schema DDL or route registration |
| `linkshelf/internal/api/handlers.go` | route handlers and JSON/status contracts | database initialization |
| `linkshelf/cmd/server/main.go` | startup wiring and :8080 listener | store implementation |
| `linkshelf/web/index.html`, `linkshelf/web/app.js`, `linkshelf/web/style.css` | UI markup, browser behavior, and presentation | server-side persistence |
| `linkshelf/package.json`, `linkshelf/playwright.config.js`, `linkshelf/tests/e2e.spec.js` | E2E dependencies, configuration, and browser assertions | application routes or store behavior |

The exported Go contracts are exact: `Link` and `InitSchema(db *sql.DB) error` in `linkshelf/internal/store/schema.go`; package variable `DB *sql.DB` and `List(ctx context.Context) ([]Link, error)`, `Create(ctx context.Context, title, url string) (Link, error)`, and `Delete(ctx context.Context, id int64) error` in `linkshelf/internal/store/store.go`. No Store struct or NewStore constructor is introduced.

## HTTP + entrypoint integration
The service exposes exactly the SPEC routes: GET /api/links lists links, POST /api/links creates a link, and DELETE /api/links/{id} removes a link. GET / returns `linkshelf/web/index.html`, and GET /static/{file} serves a file under `linkshelf/web/` while rejecting paths containing `..`. API creation accepts JSON with title and url, returns 201 with the created link, and returns 400 with an error object for invalid or malformed input. Missing deletes return 404 with an error object; successful deletes return 204.

When invoked with `cd linkshelf && go run ./cmd/server`, `linkshelf/cmd/server/main.go` opens `linkshelf.db` in the current working directory. It calls InitSchema before serving, assigns the connection to store.DB, registers handlers on http.DefaultServeMux including the root and static routes, logs `listening on :8080`, and calls http.ListenAndServe(":8080", nil). The static handler serves only files beneath `linkshelf/web/` and rejects any requested static path containing `..`.

## Integration and testing
The backend contract is verified by the required command `cd linkshelf && go mod tidy && go test ./...`, which compiles every delivered Go package. Optional tests, if added, use a separate in-memory SQLite database, call InitSchema, assign store.DB, and verify the package-level store API without changing production ownership. Store coverage should check a non-nil empty List result, descending ID order, title and URL validation, RFC3339 UTC timestamps, successful deletion, and an error for a missing ID. Handler coverage may use httptest to check exact statuses, JSON bodies, malformed POST input, and static traversal rejection.

The runtime smoke procedure starts the server from `linkshelf/` with `go run ./cmd/server`, waits for port 8080, requests the root and API endpoints, and terminates the process cleanly. Because the working directory is `linkshelf/`, the only runtime database path is `linkshelf.db`; schema initialization occurs before the listener accepts requests.

## E2E / integration testing
The Playwright harness uses `linkshelf/playwright.config.js` to target the server on port 8080 and uses `linkshelf/package.json` for its dependency and test script. `linkshelf/tests/e2e.spec.js` verifies that the UI loads, a static asset is served successfully, a user can submit the title and URL inputs, the created link appears in the list element with id links, and the delete action removes it. The configuration starts or targets the server from `linkshelf/`, waits for server readiness, and reports failures rather than silently skipping them. The E2E command is `cd linkshelf && npx playwright test`.

## Requirements

### backend-setup
Deliver `linkshelf/go.mod` and `linkshelf/internal/store/schema.go`. The module declares module linkshelf, Go 1.22, and the SQLite dependency. The schema file defines the exact Link fields and JSON names and the InitSchema function with idempotent links-table DDL. No CREATE TABLE statement belongs in `linkshelf/internal/store/store.go`. Verify with `cd linkshelf && go test ./...`.

### backend-store
Deliver `linkshelf/internal/store/store.go`. The store uses package-level DB and the exact List, Create, and Delete signatures. List orders by id descending and returns a non-nil empty slice. Create rejects an empty title, titles over 200 runes, an empty URL, and URLs not beginning with http:// or https://; it records an RFC3339 UTC timestamp. Delete reports an error when the requested ID does not exist. Verify with `cd linkshelf && go test ./...`.

### backend-api
Deliver `linkshelf/internal/api/handlers.go` and `linkshelf/cmd/server/main.go`. Handlers use only store.List, store.Create, and store.Delete and implement the exact root, static, list, create, and delete routes, statuses, JSON bodies, and traversal rejection. The entrypoint opens linkshelf.db in the current working directory when run from linkshelf/, initializes the schema, assigns store.DB, registers http.DefaultServeMux, logs listening on :8080, and listens on :8080. Verify with `cd linkshelf && go test ./...`.

### frontend-ui
This phase declares `linkshelf/tests/e2e.spec.js` as the conventional browser-UI test companion for the delivered markup, client behavior, and styles; it is included in the phase contract so UI work has an explicit test artifact.
Deliver `linkshelf/web/index.html`, `linkshelf/web/app.js`, and `linkshelf/web/style.css`. The page includes title and URL inputs, an Add button, and `<ul id="links"></ul>`. The client loads links on page load, renders them, posts new links, deletes them, and refreshes after each change. The stylesheet is small, readable, and framework-free. Verify all three files are non-empty and run `node --check linkshelf/web/app.js && node --check linkshelf/tests/e2e.spec.js`.

### e2e-testing
Deliver `linkshelf/package.json`, `linkshelf/playwright.config.js`, and `linkshelf/tests/e2e.spec.js`. The Playwright harness targets the running application and verifies UI loading, static file serving, adding a link, displaying it, and deleting it. Verify with `cd linkshelf && npx playwright test`.

### smoke-test
Recheck the complete SPEC definition of done: run `cd linkshelf && go mod tidy && go test ./...`, start the server with `cd linkshelf && go run ./cmd/server`, confirm the UI and CRUD behavior on :8080, and run `cd linkshelf && npx playwright test` against that server.

## Delivery phases
```yaml
delivery_phases:
  - id: backend-setup
    title: backend-setup
    spec_focus: backend-setup
    required_files:
      - linkshelf/go.mod
      - linkshelf/internal/store/schema.go
    qa_verify_command: 'cd linkshelf && go test ./...'
    depends_on: []
  - id: backend-store
    title: backend-store
    spec_focus: backend-store
    required_files:
      - linkshelf/internal/store/store.go
    qa_verify_command: 'cd linkshelf && go test ./...'
    depends_on:
      - backend-setup
  - id: backend-api
    title: backend-api
    spec_focus: backend-api
    required_files:
      - linkshelf/internal/api/handlers.go
      - linkshelf/cmd/server/main.go
    qa_verify_command: 'cd linkshelf && go test ./...'
    depends_on:
      - backend-store
  - id: frontend-ui
    title: frontend-ui
    spec_focus: frontend-ui
    required_files:
      - linkshelf/web/index.html
      - linkshelf/web/app.js
      - linkshelf/web/style.css
    qa_verify_command: 'test -s linkshelf/web/index.html && test -s linkshelf/web/app.js && test -s linkshelf/web/style.css && node --check linkshelf/web/app.js'
    depends_on:
      - backend-api
  - id: e2e-testing
    title: e2e-testing
    spec_focus: e2e-testing
    required_files:
      - linkshelf/package.json
      - linkshelf/playwright.config.js
      - linkshelf/tests/e2e.spec.js
    qa_verify_command: 'cd linkshelf && npx playwright test'
    depends_on:
      - frontend-ui
  - id: smoke-test
    title: smoke-test
    spec_focus: smoke-test
    required_files:
      - linkshelf/go.mod
      - linkshelf/cmd/server/main.go
      - linkshelf/package.json
      - linkshelf/tests/e2e.spec.js
    qa_verify_command: 'cd linkshelf && go mod tidy && go test ./... && npx playwright test'
    depends_on:
      - e2e-testing
```

## Acceptance mapping
The backend phases satisfy the SPEC Link model, idempotent SQLite schema, package-level store API, validation, timestamps, descending listing, and missing-delete behavior. The API phase preserves the exact root, static, list, create, and delete routes, including statuses, JSON error bodies, traversal rejection, current-working-directory database path, schema initialization, and port 8080 startup. The frontend phase supplies the required controls and refresh-based browser CRUD. The E2E phase verifies UI loading, static file serving, adding, and deleting; the smoke phase verifies the complete definition of done.
