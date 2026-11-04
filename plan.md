# Backend API implementation plan

## Bead map

### te3-xg3: linkshelf/internal/api/handlers.go
- Scope: Implement the HTTP API handler at `linkshelf/internal/api/handlers.go` for the exact Link Shelf CRUD endpoints, using only the package-level store functions.
- Architecture: Follow architecture.md § backend-api and § HTTP + entrypoint integration; expose `NewHandler() http.Handler` for the entrypoint and use `store.List`, `store.Create`, and `store.Delete`.
- Interfaces / Depends on: Depends on `linkshelf/internal/store/store.go` exports `List(ctx context.Context) ([]Link, error)`, `Create(ctx context.Context, title, url string) (Link, error)`, and `Delete(ctx context.Context, id int64) error`; consumed by `linkshelf/cmd/server/main.go` through `api.NewHandler()`.
- Acceptance: `GET /api/links` returns HTTP 200 and a JSON array, including `[]` when empty. `POST /api/links` accepts `{"title":"...","url":"..."}`, returns HTTP 201 and the created JSON link, and returns HTTP 400 with `{"error":"..."}` for malformed or invalid input. `DELETE /api/links/{id}` returns HTTP 204 on success and HTTP 404 with an error object for a missing or invalid target. The handler owns the exact `/api/links` and `/api/links/{id}` paths and does not initialize the database.
- Verify: `cd linkshelf && go test ./...`

### te3-h2v: linkshelf/cmd/server/main.go
- Scope: Implement the executable entrypoint at `linkshelf/cmd/server/main.go`, wiring SQLite, schema initialization, API registration, static files, and the HTTP listener.
- Architecture: Follow architecture.md § backend-api and § HTTP + entrypoint integration; open `linkshelf.db`, call `store.InitSchema`, assign `store.DB`, register `api.NewHandler()`, and listen on `:8080`.
- Interfaces / Depends on: Depends on `store.InitSchema(db *sql.DB) error`, `store.DB`, and `api.NewHandler() http.Handler`; consumed by the running application and browser clients. Implement after `linkshelf/internal/api/handlers.go`.
- Acceptance: From `linkshelf/`, startup opens `linkshelf.db` in the current working directory, calls `InitSchema` before serving, assigns `store.DB`, logs `listening on :8080`, and calls `http.ListenAndServe(":8080", nil)`. Registration serves `GET /` with `linkshelf/web/index.html`, serves `GET /static/{file}` from `linkshelf/web/`, rejects static paths containing `..`, and returns 404 for unsupported paths or methods. The API handler is registered for the exact `/api/links` routes without duplicating store implementation.
- Verify: `cd linkshelf && go test ./...`

## Integration contract

The dependency order is `linkshelf/internal/store/schema.go` and `linkshelf/internal/store/store.go` first, then `linkshelf/internal/api/handlers.go`, then `linkshelf/cmd/server/main.go`. The entrypoint obtains `api.NewHandler() http.Handler`, registers it for `/api/`, and relies on the store exports `InitSchema(db *sql.DB) error`, `DB`, `List`, `Create`, and `Delete`.

The server must register these exact SPEC routes: `GET /`, `GET /static/{file}`, `GET /api/links`, `POST /api/links`, and `DELETE /api/links/{id}`. `GET /` serves `linkshelf/web/index.html`; `GET /static/{file}` serves a file under `linkshelf/web/` and rejects any path containing `..`. API creation returns 201 or 400 with `{"error":"..."}`, listing returns 200 JSON, and deletion returns 204 or 404 with an error object. Active-phase verification is `cd linkshelf && go test ./...`.
