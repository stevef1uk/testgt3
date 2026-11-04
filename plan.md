# Backend API implementation plan

## Bead map

### te3-anl: linkshelf/internal/api/handlers.go
- Scope: Implement the API, root, and static-file HTTP behavior owned by this file.
- Architecture: Follow architecture.md §backend-api and §HTTP + entrypoint integration. Use package `store` and only its package-level `List`, `Create`, and `Delete` operations.
- Interfaces: Register/serve the exact SPEC routes: `GET /api/links`, `POST /api/links`, `DELETE /api/links/{id}`, `GET /`, and `GET /static/{file}`. The API returns a JSON array for list, 201 plus a JSON Link for valid create, 400 with `{"error":"..."}` for invalid or malformed create input, 204 for successful delete, and 404 with `{"error":"..."}` for a missing delete. Root serves `linkshelf/web/index.html`; static files come from `linkshelf/web/` and paths containing `..` are rejected.
- Depends on: `linkshelf/internal/store/schema.go` and `linkshelf/internal/store/store.go` from completed phases; use `store.Link`, `store.List`, `store.Create`, and `store.Delete` exactly as defined in architecture.md.
- Consumed by: `linkshelf/cmd/server/main.go` (te3-0go), which wires the handlers on `http.DefaultServeMux`.
- Acceptance: Empty `GET /api/links` encodes `[]`; create decodes `{"title":"...","url":"..."}`, delegates validation to `store.Create`, and returns the exact success/error statuses and JSON contracts in SPEC.md. Delete parses the path ID, calls `store.Delete`, and distinguishes 204 success from the specified 404 error. Root, valid static files, and traversal rejection match SPEC.md.
- Verify: `cd linkshelf && go test ./...`

### te3-0go: linkshelf/cmd/server/main.go
- Scope: Implement the executable startup and dependency wiring for the backend-api phase.
- Architecture: Follow architecture.md §backend-api and §HTTP + entrypoint integration; main owns SQLite opening, schema initialization, store assignment, route registration, logging, and listening, not store implementation.
- Interfaces: Open `linkshelf.db` in the current working directory, call `store.InitSchema(db)`, assign `store.DB = db`, register the handler/root/static routes on `http.DefaultServeMux`, log `listening on :8080`, and call `http.ListenAndServe(":8080", nil)`.
- Depends on: `linkshelf/internal/store/schema.go`, `linkshelf/internal/store/store.go`, and `linkshelf/internal/api/handlers.go` (te3-anl); wire these in that order and use only symbols documented by architecture.md.
- Consumed by: The `go run ./cmd/server` process and later frontend/e2e phases.
- Acceptance: `cd linkshelf && go run ./cmd/server` opens the database relative to `linkshelf/`, initializes the schema before serving, exposes all five exact SPEC routes through the real store and handlers, serves the root/static UI, logs the required message, and listens on port 8080. No extra CRUD routes or invented store constructors are introduced.
- Verify: `cd linkshelf && go test ./...`

## Integration contract

Implement te3-anl before te3-0go. The entrypoint opens `linkshelf.db`, calls `store.InitSchema`, assigns `store.DB`, then invokes the handler registration owned by `linkshelf/internal/api/handlers.go` on `http.DefaultServeMux`. The exact routes are `GET /`, `GET /static/{file}`, `GET /api/links`, `POST /api/links`, and `DELETE /api/links/{id}`. The handler calls only `store.List`, `store.Create`, and `store.Delete`; main does not duplicate persistence logic. The listener is `:8080`.

## Verification and YAGNI

The active phase gate is `cd linkshelf && go test ./...`. Implement only the SPEC and architecture contracts above; backend tests, frontend files, and Playwright files belong to later or optional work and are not additional beads in this phase.
