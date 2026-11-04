# Implementation plan

## Scope and verification

Active phase: backend-api. Implement only the two required paths below, preserving the package/module layout in `architecture.md`. The polecat must run `cd linkshelf && go test ./...` and ensure the backend behavior is compatible with the SPEC HTTP contract.

## Bead map

### te3-anl: linkshelf/internal/api/handlers.go
- Scope: Implement the API and static/UI HTTP handlers in the `api` package, using the store package functions rather than introducing a Store struct or NewStore.
- Architecture: `architecture.md` backend API ownership for `linkshelf/internal/api/handlers.go`; routes and response semantics are defined by SPEC.
- Interfaces: Handler registration/route handling for `GET /`, `GET /static/{file}`, `GET /api/links`, `POST /api/links`, and `DELETE /api/links/{id}`; call `store.List`, `store.Create`, and `store.Delete` as the store interface.
- Depends on: the store/schema implementation described by `architecture.md` (including `store.Link`, `store.List`, `store.Create`, and `store.Delete`); `<none>` among active beads.
- Consumed by: `linkshelf/cmd/server/main.go` (`te3-0go`), which registers the API handlers.
- Acceptance: Exact routes are served: `GET /` serves the UI, `GET /static/{file}` serves an allowed static file, `GET /api/links` lists links, `POST /api/links` creates a link, and `DELETE /api/links/{id}` deletes one.
- Acceptance: An empty GET list is valid JSON `[]`; a valid POST body `{"title","url"}` returns HTTP 201 and the created link as JSON, while invalid or malformed input returns HTTP 400 with an `{"error":"..."}` JSON response.
- Acceptance: DELETE returns HTTP 204 for an existing link and HTTP 404 with an `{"error":"..."}` JSON response when the id is missing; handlers invoke only `store.List`, `store.Create`, and `store.Delete` for persistence and do not add a Store struct or NewStore.
- Acceptance: Static file requests reject `..` traversal instead of reading files outside the static/UI tree.
- Verify: `cd linkshelf && go test ./...`

### te3-0go: linkshelf/cmd/server/main.go
- Scope: Wire the SQLite database, schema initialization, API handlers, static routes, and HTTP server startup exactly as specified.
- Architecture: `architecture.md` server entrypoint ownership for `linkshelf/cmd/server/main.go` and its dependency order.
- Interfaces: Startup wiring exports no application API; it opens `linkshelf.db`, calls `store.InitSchema(db)`, assigns `store.DB = db`, registers routes on `http.DefaultServeMux`, logs `listening on :8080`, and calls `http.ListenAndServe(":8080", nil)`.
- Depends on: `te3-anl` (`linkshelf/internal/api/handlers.go`) and the store symbols owned by the architecture's store package.
- Consumed by: the executable/runtime only; `<none>` among later active-phase beads.
- Acceptance: The process opens `linkshelf.db` in the current working directory, invokes `store.InitSchema(db)` before serving, and sets `store.DB = db` before requests are accepted.
- Acceptance: Routes are registered on `http.DefaultServeMux`; startup logs the text `listening on :8080`; serving uses exactly `http.ListenAndServe(":8080", nil)`.
- Acceptance: The entrypoint wires the handlers without changing the SPEC paths or adding alternate server configuration, and `cd linkshelf && go test ./...` passes.
- Verify: `cd linkshelf && go test ./...`

## Integration contract

Dependency order is store schema/database symbols first, then `linkshelf/internal/api/handlers.go`, then `linkshelf/cmd/server/main.go`. The entrypoint obtains the database by opening `linkshelf.db` in the current working directory, calls `store.InitSchema(db)`, assigns `store.DB = db`, then registers the API/static handlers on `http.DefaultServeMux` and serves the default mux with `http.ListenAndServe(":8080", nil)`.

The exact HTTP contract is: `GET /`, `GET /static/{file}`, `GET /api/links`, `POST /api/links`, and `DELETE /api/links/{id}`. GET list returns `[]` when empty; valid POST returns 201 and a JSON link, invalid/malformed POST returns 400 with JSON `error`; successful DELETE returns 204 and missing-id DELETE returns 404 with JSON `error`; static `..` traversal is rejected.

Per-file exports and ownership from the architecture: `linkshelf/internal/store/schema.go` owns `Link` and `InitSchema`; `linkshelf/internal/store/store.go` owns `DB`, `List`, `Create`, and `Delete`; `linkshelf/internal/api/handlers.go` owns the route handlers; `linkshelf/cmd/server/main.go` owns startup wiring. The entrypoint consumes the store and API packages in that order and does not invent a Store struct or NewStore.

## Future considerations

None for this phase; frontend and other phase files remain out of scope.
