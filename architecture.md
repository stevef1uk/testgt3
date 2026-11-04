# Architecture for testgt3

## Overview
Link Shelf is the SPEC-defined minimal bookmark CRUD service: a Go 1.22 backend, SQLite persistence, exact HTTP API routes, and a vanilla browser UI served on port 8080. The design keeps schema ownership, store behavior, HTTP handlers, entrypoint wiring, static assets, and Playwright verification in separate delivery phases. The runtime database is linkshelf/linkshelf.db when the server is launched from linkshelf/, while tests use isolated in-memory databases. All implementation paths below are repository-relative and explicitly rooted at linkshelf/.

## Planned file layout
- `linkshelf/go.mod` — Go module manifest and SQLite dependency.
- `linkshelf/internal/store/schema.go` — Link data model and idempotent schema initialization.
- `linkshelf/internal/store/store.go` — package-level DB and List, Create, Delete persistence API.
- `linkshelf/internal/api/handlers.go` — exact API, root, and static-file handlers.
- `linkshelf/cmd/server/main.go` — executable startup and dependency wiring.
- `linkshelf/web/index.html` — UI structure and required controls.
- `linkshelf/web/app.js` — browser CRUD client.
- `linkshelf/web/style.css` — framework-free presentation styles.
- `linkshelf/package.json` — Playwright test dependency and script.
- `linkshelf/playwright.config.js` — browser test configuration for port 8080.
- `linkshelf/tests/e2e.spec.js` — end-to-end add and delete coverage.

## Module / bead ownership
`linkshelf/internal/store/schema.go` owns the Link struct and InitSchema function and owns all links-table DDL. `linkshelf/internal/store/store.go` owns the package-level DB plus List, Create, and Delete; it must not duplicate schema DDL. `linkshelf/internal/api/handlers.go` owns HTTP registration and request/response behavior and calls only the store API. `linkshelf/cmd/server/main.go` owns process startup, SQLite opening, schema initialization, store assignment, static registration, and listening. Frontend files own only browser presentation and fetch behavior; Playwright files own only browser orchestration.

| File | Owns | Must not define |
|---|---|---|
| `linkshelf/internal/store/schema.go` | Link and InitSchema | HTTP handlers or CRUD queries |
| `linkshelf/internal/store/store.go` | DB, List, Create, Delete | schema DDL or route registration |
| `linkshelf/internal/api/handlers.go` | route handlers and JSON/status contracts | database initialization |
| `linkshelf/cmd/server/main.go` | startup wiring and :8080 listener | store implementation |

## HTTP + entrypoint integration
The service exposes exactly the SPEC routes: GET /api/links lists links, POST /api/links creates a link, and DELETE /api/links/{id} removes a link. GET / returns `linkshelf/web/index.html`, and GET /static/{file} serves a file under `linkshelf/web/` while rejecting paths containing `..`. The entrypoint parses its configured startup state, opens `linkshelf/linkshelf.db` relative to its working directory, calls InitSchema, assigns the resulting connection to store.DB, registers the handlers on http.DefaultServeMux, registers static serving, and listens on :8080.


The service exposes exactly the SPEC routes: GET /api/links lists links, POST /api/links creates a link, and DELETE /api/links/{id} removes a link. GET / returns linkshelf/web/index.html, and GET /static/{file} serves a file under linkshelf/web/ while rejecting paths containing ..

## Integration and testing

The backend unit-test strategy is package-local and deterministic. Optional concrete tests in linkshelf/internal/store/store_test.go and linkshelf/internal/store/schema_test.go open an isolated :memory: SQLite database, call InitSchema, assign store.DB, and verify List, Create, and Delete behavior including validation, descending ID order, timestamps, empty non-nil results, and missing-delete errors. Optional concrete tests in linkshelf/internal/api/handlers_test.go use httptest with the same isolated store to verify the exact API statuses, JSON bodies, malformed POST input, and traversal rejection. The required backend smoke command is cd linkshelf && go mod tidy && go test ./..., which also proves every delivered Go package compiles.

The runtime smoke strategy starts the application with go run linkshelf/cmd/server, waits for port 8080, requests the root and API endpoints, and terminates the process cleanly. The server must use linkshelf.db in its current working directory, initialize the schema before serving, and log listening on :8080.

## E2E / integration testing

The Playwright integration harness uses linkshelf/playwright.config.js to target the running server on port 8080. linkshelf/tests/e2e.spec.js verifies that GET / loads linkshelf/web/index.html, a static asset is served successfully, a user can submit the title and URL inputs, the created link appears in ul id links, and the delete action removes it. The suite uses linkshelf/package.json for its Playwright dependency and test script, starts from a clean database or controlled test state, waits for server readiness, and reports failures rather than silently skipping them. The e2e command is cd linkshelf && npx playwright test.


## Requirements

### backend-setup

Deliver linkshelf/go.mod and linkshelf/internal/store/schema.go. The module declares module linkshelf, Go 1.22, and the SQLite dependency. The schema file owns the exact Link struct and InitSchema function, including idempotent links-table DDL. No schema DDL belongs in linkshelf/internal/store/store.go.

### backend-store

Deliver linkshelf/internal/store/store.go. Optional tests may use an in-memory database, call InitSchema, assign store.DB, and verify the store contract. The store uses package-level DB, List, Create, and Delete exactly as specified, returns a non-nil empty list, orders by id descending, validates title and URL, records RFC3339 UTC timestamps, and reports missing deletes. Tests use an in-memory database and call InitSchema.

### backend-api

Deliver linkshelf/internal/api/handlers.go and linkshelf/cmd/server/main.go. Optional handler tests may use httptest to verify the HTTP contract. Handlers use only store.List, store.Create, and store.Delete and implement the exact root, static, list, create, and delete routes and status/error bodies. The entrypoint opens linkshelf.db, initializes the schema, assigns store.DB, registers http.DefaultServeMux, and listens on :8080.

### frontend-ui

Deliver linkshelf/web/index.html, linkshelf/web/app.js, and linkshelf/web/style.css. The page includes title and URL inputs, an Add button, and ul id links. The client loads, creates, renders, deletes, and refreshes links through the API, while the stylesheet remains a small readable framework-free design.

### e2e-testing

Deliver linkshelf/playwright.config.js, linkshelf/package.json, and linkshelf/tests/e2e.spec.js. The Playwright harness runs against the server and verifies UI loading, static file serving, adding links, and deleting links.

### smoke-test

Recheck the complete application using the SPEC definition of done: go mod tidy and go test ./... from linkshelf, followed by npx playwright test from linkshelf against the runnable server on :8080.


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
The backend phases implement the SPEC Link model, idempotent SQLite schema, package-level store API, validation, timestamps, descending listing, and missing-delete behavior. The API phase preserves the exact root, static, list, create, and delete routes, including statuses, JSON error bodies, traversal rejection, current-working-directory database path, schema initialization, and port 8080 startup. The frontend phase supplies the required controls and refresh-based browser CRUD. The E2E phase verifies UI loading, static file serving, adding, and deleting; the smoke phase verifies the complete definition of done.
