# Backend API implementation plan

## Bead map

### te3-anl: linkshelf/internal/api/handlers.go
- Scope: Implement HTTP handlers and route registration in `linkshelf/internal/api/handlers.go`.
- Architecture: `architecture.md` backend-api requirements and HTTP + entrypoint integration; `SPEC.md` HTTP API table.
- Interfaces: Use only `store.List`, `store.Create`, and `store.Delete`; implement `GET /`, `GET /static/{file}`, `GET /api/links`, `POST /api/links`, and `DELETE /api/links/{id}`.
- Depends on: `linkshelf/internal/store/schema.go` and `linkshelf/internal/store/store.go` exports.
- Consumed by: `linkshelf/cmd/server/main.go` (`te3-0go`).
- Acceptance: `GET /` serves `linkshelf/web/index.html`; `GET /static/{file}` serves files beneath `linkshelf/web/` and returns 404 for every path containing `..`.
- Acceptance: `GET /api/links` returns 200 and a JSON array, including `[]` when empty; valid `POST /api/links` returns 201 and created-link JSON.
- Acceptance: `POST /api/links` returns 400 with an error JSON body for empty title, title over 200 runes, malformed JSON, or a URL not starting with `http://` or `https://`.
- Acceptance: successful `DELETE /api/links/{id}` returns 204; missing ID returns 404 with an error JSON body; persistence calls are limited to the specified store functions.
- Verify: `cd linkshelf && go test ./...`

### te3-0go: linkshelf/cmd/server/main.go
- Scope: Implement startup and dependency wiring in `linkshelf/cmd/server/main.go`.
- Architecture: `architecture.md` backend-api requirements and HTTP + entrypoint integration; `SPEC.md` startup steps.
- Interfaces: Consume `InitSchema(db *sql.DB) error`, `store.DB`, and route registration owned by `linkshelf/internal/api/handlers.go`; call `http.ListenAndServe(":8080", nil)`.
- Depends on: `linkshelf/internal/api/handlers.go` (`te3-anl`) and `linkshelf/internal/store/schema.go` / `linkshelf/internal/store/store.go`.
- Consumed by: The running service and later frontend/e2e phases.
- Acceptance: Open `linkshelf.db` in the current working directory, call `InitSchema(db)` before serving, then assign `store.DB = db`.
- Acceptance: Register root, static, and API routes on `http.DefaultServeMux`, log exactly `listening on :8080`, and call `http.ListenAndServe(":8080", nil)`.
- Acceptance: The server exposes exactly `GET /`, `GET /static/{file}`, `GET /api/links`, `POST /api/links`, and `DELETE /api/links/{id}`.
- Verify: `cd linkshelf && go test ./...`

## Integration contract

The module is `linkshelf`. Dependency order is
`linkshelf/internal/store/schema.go`, then
`linkshelf/internal/store/store.go`, then
`linkshelf/internal/api/handlers.go`, then
`linkshelf/cmd/server/main.go`.

Per-file exported ownership from `architecture.md`:
- `linkshelf/internal/store/schema.go`: `Link`; `InitSchema(db *sql.DB) error`.
- `linkshelf/internal/store/store.go`: `DB *sql.DB`; `List(ctx context.Context) ([]Link, error)`; `Create(ctx context.Context, title, url string) (Link, error)`; `Delete(ctx context.Context, id int64) error`.
- `linkshelf/internal/api/handlers.go`: architecture-defined HTTP route registration and handlers.
- `linkshelf/cmd/server/main.go`: executable startup wiring and no store abstraction.

The entrypoint opens `linkshelf.db`, calls `InitSchema(db *sql.DB) error`,
assigns `store.DB`, registers routes on `http.DefaultServeMux`, logs
`listening on :8080`, and calls `http.ListenAndServe(":8080", nil)`.

The exact SPEC HTTP route contract is:
- `GET /` serves `linkshelf/web/index.html`.
- `GET /static/{file}` serves a file under `linkshelf/web/` and rejects any path containing `..`.
- `GET /api/links` returns HTTP 200 with a JSON array, including `[]` when empty.
- `POST /api/links` returns HTTP 201 with the created link JSON, or HTTP 400 with `{"error":"..."}` for invalid input.
- `DELETE /api/links/{id}` returns HTTP 204 on success, or HTTP 404 with `{"error":"..."}` when the ID is missing.

The handlers call only `store.List`, `store.Create`, and `store.Delete`; the
entrypoint performs database initialization and wiring but does not implement
store behavior.

## Verification

The active phase verification is `cd linkshelf && go test ./...`. Frontend,
Playwright, and other later-phase files are intentionally excluded.
