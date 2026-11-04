# Implementation plan

## Phase verification

Run from the repository root: `cd linkshelf && go test ./...`.
The implementation must satisfy the Link Shelf SPEC, not the placeholder `/api/items`
behavior: the public API is `/api/links` and the server listens on port 8080.

## Bead map

### te3-anl: linkshelf/internal/api/handlers.go
- Scope: Implement the API HTTP handler using the store layer described in
  `architecture.md`; keep the handler package focused on HTTP/JSON translation.
- Architecture: follow the `api` package design and the exported handler
  constructor/signatures in `architecture.md`; use the store's Link representation
  and operations rather than a second in-memory data model.
- Interfaces: provide the architecture-defined handler constructor and
  `ServeHTTP` behavior for `GET /api/links`, `POST /api/links`, and
  `DELETE /api/links/{id}`. Decode the SPEC request JSON, encode the SPEC link
  JSON, and set the documented status/content-type responses.
- Depends on: the store exports and schema supplied by
  `linkshelf/internal/store/schema.go` and `linkshelf/internal/store/store.go`
  in the completed backend foundation.
- Consumed by: `te3-0go` (`linkshelf/cmd/server/main.go`), which registers this
  handler under `/api/links`.
- Acceptance: GET returns persisted links as JSON; POST accepts the SPEC link
  fields, persists a link, and returns the created record; DELETE removes the
  requested ID and reports the documented result. Requests to unsupported
  methods/paths and malformed input receive the SPEC-required HTTP result.
- Acceptance: handlers use the shared store so a link created through POST is
  visible to a later GET and remains available across handler calls; no `/api/items`
  route or unrelated feature is introduced.
- Verify: `cd linkshelf && go test ./...`

### te3-0go: linkshelf/cmd/server/main.go
- Scope: Implement the server entrypoint after the store and API packages are
  available. Open the SQLite database, initialize its schema, construct the API
  handler, and serve the web assets.
- Architecture: follow the startup and dependency wiring in `architecture.md`;
  use the architecture-defined store schema initializer and API handler
  constructor, register the exact SPEC routes, and bind `:8080`.
- Interfaces: export no application API beyond the `main` package entrypoint;
  wire `GET /`, `GET /static/{file}`, `GET /api/links`, `POST /api/links`, and
  `DELETE /api/links/{id}` to the corresponding static/API handlers.
- Depends on: `te3-anl` (`linkshelf/internal/api/handlers.go`) plus the
  architecture-defined store/schema exports from the backend foundation.
- Consumed by: the Go runtime and the Playwright/browser workflow; it is the
  executable process used by the later frontend and e2e phases.
- Acceptance: `go run ./cmd/server` opens the configured SQLite database,
  initializes the links table before serving requests, listens on port 8080,
  serves the UI at `/`, serves files beneath `/static/`, and exposes the
  complete `/api/links` CRUD route table without route-name drift.
- Acceptance: an end-to-end sequence can load the page, create a link, observe
  it in the list, delete it, and observe its removal; startup does not require
  manually pre-created schema state.
- Verify: `cd linkshelf && go test ./...`

## Integration contract

Dependency order is schema/store first, API handler second, and server entrypoint
last. The entrypoint must obtain the SQLite-backed store, call the architecture's
schema initialization before registering handlers, pass that store to the API
constructor, and register the API at `/api/links` while the static file handler
owns `/` and `/static/`. The exact HTTP contract is:
`GET /`, `GET /static/{file}`, `GET /api/links`, `POST /api/links`, and
`DELETE /api/links/{id}`. The server must call `http.ListenAndServe(":8080", ...)`;
the API must not substitute `/api/items` or an in-memory-only implementation.

## Future considerations

None: authentication, pagination, alternate storage engines, and extra API
features are outside this SPEC phase.
