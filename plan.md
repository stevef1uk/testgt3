# Backend API implementation plan

Active phase: `backend-api`. Implement only the two required files in this phase. Use the exact contracts in SPEC.md and architecture.md; do not add frontend or later-phase files.

## Integration contract

Dependency order is the existing store package, then `linkshelf/internal/api/handlers.go`, then `linkshelf/cmd/server/main.go`. The API handler obtains persistence through package-level `store.DB` and calls only `store.List(ctx) ([]store.Link, error)`, `store.Create(ctx, title, url string) (store.Link, error)`, and `store.Delete(ctx, id int64) error`. The entrypoint opens `linkshelf.db` in the current working directory, calls `store.InitSchema(db)`, assigns `store.DB = db`, constructs `api.NewHandler() *api.Handler`, registers the API on `http.DefaultServeMux`, and calls `http.ListenAndServe(":8080", nil)` while logging `listening on :8080`.

The exact SPEC routes and contracts are: `GET /` serves `linkshelf/web/index.html`; `GET /static/{file}` serves a file under `linkshelf/web/` and rejects any path containing `..` with 404; `GET /api/links` returns 200 and a JSON array, including `[]` when empty; `POST /api/links` accepts `{"title":"...","url":"..."}`, returns 201 with the created link, and returns 400 with `{"error":"..."}` for malformed or invalid input; `DELETE /api/links/{id}` returns 204 on success and 404 with `{"error":"..."}` when the ID is missing. The API must use store CRUD rather than placeholder or in-memory behavior.

## Bead map

### te3-xe4: linkshelf/internal/api/handlers.go
- Scope: Implement the HTTP handler, exact API routing, JSON encoding, status codes, and JSON error responses in `linkshelf/internal/api/handlers.go`.
- Architecture reference: The file owns `Handler`, `NewHandler() *Handler`, `ServeHTTP`, `list`, `create`, and `writeJSON`; it must not initialize the database.
- Interfaces / Depends on: `linkshelf/internal/store` exports `List`, `Create`, `Delete`, `Link`, and `DB`; this file consumes those package-level APIs and is consumed by `linkshelf/cmd/server/main.go` through `api.NewHandler() *api.Handler`.
- Acceptance: `GET /api/links` calls store listing and returns status 200 with a JSON array, never a nil JSON value for an empty database. `POST /api/links` decodes the SPEC `title` and `url` body, calls store creation, and returns status 201 with the created link.
- Acceptance: invalid or malformed POST JSON and store validation failures return status 400 with a JSON object shaped `{"error":"..."}`. `DELETE /api/links/{id}` parses the ID, calls store deletion, returns 204 on success, and returns status 404 with `{"error":"..."}` for a missing or invalid ID.
- Acceptance: unsupported methods and unmatched API paths do not become successful CRUD responses; all API success and error bodies use JSON content type. The implementation must use the exact `/api/links` and `/api/links/{id}` routes from SPEC.md.
- Verify: `cd linkshelf && go test ./...`

### te3-20x: linkshelf/cmd/server/main.go
- Scope: Implement startup and HTTP wiring in `linkshelf/cmd/server/main.go`.
- Architecture reference: This file owns SQLite opening, schema initialization, assignment to `store.DB`, API and static route registration, the `main()` entrypoint, and the `:8080` listener; it must not implement store CRUD.
- Interfaces / Depends on: depends on `store.InitSchema(db) error`, `store.DB`, and `api.NewHandler() *api.Handler`; consumed by the executable and integration/runtime checks.
- Acceptance: running from `linkshelf/` opens `linkshelf.db`, initializes the schema before serving, assigns the opened database to `store.DB`, registers the API handler on `http.DefaultServeMux`, logs `listening on :8080`, and calls `http.ListenAndServe(":8080", nil)`.
- Acceptance: the root route serves `linkshelf/web/index.html` for `GET /`; `GET /static/{file}` serves files beneath `linkshelf/web/` only, rejects any requested static path containing `..`, and returns 404 for unknown or non-GET paths.
- Acceptance: server wiring preserves the exact API routes and statuses delegated to the handler: `GET /api/links` 200 JSON array, `POST /api/links` 201 JSON link or 400 JSON error, and `DELETE /api/links/{id}` 204 or 404 JSON error.
- Verify: `cd linkshelf && go test ./...`

## Phase verification

From the rig working directory, the polecat must run `cd linkshelf && go test ./...`. Completion requires both required files to exist at their exact paths and the concrete behavior above to pass without implementing files from later phases.
