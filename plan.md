# Backend API implementation plan

## Bead map

### te3-xe4: linkshelf/internal/api/handlers.go
- Scope: Implement the API package's HTTP registration and handlers, using only the package-level store API (`store.List`, `store.Create`, and `store.Delete`).
- Architecture: `architecture.md` §backend-api and §HTTP + entrypoint integration; this file owns API route behavior and JSON/status contracts, not database initialization.
- Interfaces: Export `type Handler` and `func NewHandler() *Handler` for the entrypoint; provide `GET /api/links`, `POST /api/links`, and `DELETE /api/links/{id}` behavior.
- Depends on: `linkshelf/internal/store/schema.go` and `linkshelf/internal/store/store.go`, including `Link`, `List`, `Create`, and `Delete`.
- Consumed by: `te3-20x` (`linkshelf/cmd/server/main.go`), which mounts the handler below `/api/`.
- Acceptance:
  - GET `/api/links` returns 200 and a JSON array, specifically `[]` when empty.
  - POST `/api/links` accepts `{"title":"...","url":"..."}`, returns 201 and the created JSON link, and returns 400 `{"error":"..."}` for malformed JSON or invalid title/URL input.
  - DELETE `/api/links/{id}` returns 204 for an existing link and 404 `{"error":"..."}` for invalid or missing IDs.
  - API methods and paths match the SPEC exactly; extra ID path segments are not accepted.
- Verify: `cd linkshelf && go test ./...`

### te3-20x: linkshelf/cmd/server/main.go
- Scope: Implement executable startup and wiring: open `linkshelf.db` in the current working directory, initialize schema, assign `store.DB`, register the API and static/root routes, log startup, and listen.
- Architecture: `architecture.md` §backend-api, §HTTP + entrypoint integration, and §backend-api requirement; this file owns process startup and `:8080` listener, not store implementation.
- Interfaces: Consume `api.NewHandler()`, `store.InitSchema(db)`, and `store.DB`; register the API beneath `/api/`; serve `GET /` from `linkshelf/web/index.html` and `GET /static/{file}` from `linkshelf/web/`.
- Depends on: `te3-xe4` (`api.Handler` and `api.NewHandler`) and the existing store schema/store contracts.
- Consumed by: the Go executable, runtime smoke checks, and later Playwright testing.
- Acceptance:
  - Open SQLite at `linkshelf.db`, call `InitSchema(db)` before assigning `store.DB` and serving requests, and call `http.ListenAndServe(":8080", nil)`.
  - Log `listening on :8080` and register the API on `http.DefaultServeMux`.
  - Serve the root UI and static files under `linkshelf/web`; return 404 for unsupported methods or static paths containing `..`.
  - `cd linkshelf && go test ./...` passes with the entrypoint and dependency wiring.
- Verify: `cd linkshelf && go test ./...`

## Integration contract

Implement in dependency order: the existing store schema/store contracts, then `te3-xe4`, then `te3-20x`. The entrypoint obtains an `api.Handler` through `api.NewHandler()`, mounts it at `/api/`, and initializes `store.InitSchema(db)` before assigning `store.DB`. Exact API routes are `GET /api/links`, `POST /api/links`, and `DELETE /api/links/{id}`; exact UI routes are `GET /` and `GET /static/{file}`. The API bead owns `Handler` and `NewHandler`; the entrypoint consumes those symbols and owns startup wiring.

## Verification

The active phase verification is `cd linkshelf && go test ./...`. Acceptance must cover the SPEC route paths, statuses, JSON error shape, empty-list encoding, input validation, deletion behavior, SQLite initialization, `:8080` listener configuration, root/static serving, and traversal rejection without adding out-of-scope features.
