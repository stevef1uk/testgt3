# Smoke-test implementation plan

This active smoke-test phase rechecks the runnable Link Shelf stack using only the
four required paths. Earlier phases own the store, schema, API handlers, web assets,
and Playwright configuration; these beads must preserve their interfaces and add no
new features.

## Bead map

### te3-0l4: linkshelf/go.mod
- Scope: Declare the `linkshelf` Go module, Go 1.22, and the
  `github.com/mattn/go-sqlite3 v1.14.22` dependency required by the existing backend.
- Architecture: Follow architecture.md §backend-setup and the module contract; keep
  the module path exactly `linkshelf` so imports from `linkshelf/cmd/server/main.go`
  resolve.
- Interfaces / Depends on / Consumed by: Provides module identity and dependency
  resolution. Depends on the previously delivered store, schema, and API packages.
  Consumed by `linkshelf/cmd/server/main.go` and Go verification.
- Acceptance: `go mod tidy` succeeds from linkshelf; `go test ./...` resolves the
  SQLite driver and compiles every delivered Go package without changing package
  ownership.
- Verify: `cd linkshelf && go mod tidy && go test ./...`

### te3-6uq: linkshelf/cmd/server/main.go
- Scope: Open `linkshelf.db` in the current working directory, call
  `store.InitSchema(db)`, assign `store.DB = db`, register the existing API and
  static handlers, log `listening on :8080`, and call
  `http.ListenAndServe(":8080", nil)`.
- Architecture: Follow architecture.md §backend-api and §HTTP + entrypoint
  integration. The entrypoint owns startup wiring only, not CRUD or schema DDL.
- Interfaces / Depends on / Consumed by: Depends on `linkshelf/internal/store`
  exported `InitSchema(db *sql.DB) error` and `DB *sql.DB`, plus the handler
  registration contract from the delivered API package. Consumed by
  `go run ./cmd/server` and `linkshelf/tests/e2e.spec.js`.
- Acceptance: Running from linkshelf serves GET `/` with
  `linkshelf/web/index.html`; GET `/static/{file}` serves a file under
  `linkshelf/web/` and rejects paths containing `..`; GET `/api/links` returns a
  JSON array; POST `/api/links` returns 201 for valid JSON
  `{"title":"...","url":"..."}`; DELETE `/api/links/{id}` returns 204 for an
  existing link and 404 with `{"error":"..."}` for a missing link. The server
  listens on :8080 and uses the SQLite file in the current working directory.
- Verify: `cd linkshelf && go mod tidy && go test ./... && npx playwright test`

### te3-0i8: linkshelf/package.json
- Scope: Declare the Node package metadata, Playwright dependency, and test script
  required by the existing browser harness.
- Architecture: Follow architecture.md §e2e-testing; use the existing
  `linkshelf/playwright.config.js` and `linkshelf/tests/e2e.spec.js` rather than
  creating alternate routes or server commands.
- Interfaces / Depends on / Consumed by: Depends on
  `linkshelf/playwright.config.js` and `linkshelf/tests/e2e.spec.js`. Provides the
  npm/Playwright invocation consumed by smoke-test verification.
- Acceptance: The declared Playwright dependency is installable and
  `npx playwright test` runs the real browser suite against port 8080, including
  UI load, static `/static/app.js` delivery, add, persistence after reload, and
  delete assertions.
- Verify: `cd linkshelf && npx playwright test`

### te3-wmm: linkshelf/tests/e2e.spec.js
- Scope: Exercise the real Link Shelf UI through Playwright; do not mock API
  calls or replace the server with a fixture.
- Architecture: Follow architecture.md §E2E / integration testing and
  §e2e-testing. Consume the markup and browser behavior from
  `linkshelf/web/index.html` and `linkshelf/web/app.js`.
- Interfaces / Depends on / Consumed by: Depends on
  `linkshelf/package.json`, `linkshelf/playwright.config.js`,
  `linkshelf/cmd/server/main.go`, and the delivered web assets. Consumed by
  `npx playwright test`.
- Acceptance: Assert the page loads with the title and URL controls, the links
  list is present, and GET `/static/app.js` succeeds. Submit a valid title and
  URL, assert the created link's visible name and href, reload the page and
  assert the link remains (proving POST/GET persistence), then delete it and
  assert the list contains zero matching links. Use real browser requests against
  `GET /api/links`, `POST /api/links`, and `DELETE /api/links/{id}`.
- Verify: `cd linkshelf && npx playwright test`

## Integration contract

Implement and verify in dependency order: `linkshelf/go.mod` first, then the
already-delivered store/schema/API packages, then
`linkshelf/cmd/server/main.go`; `linkshelf/package.json` and
`linkshelf/tests/e2e.spec.js` consume the running entrypoint. The entrypoint must
obtain the SQLite database through `store.InitSchema(db *sql.DB) error`, assign
`store.DB`, and register the exact SPEC routes: GET `/`, GET `/static/{file}`,
GET `/api/links`, POST `/api/links`, and DELETE `/api/links/{id}`. It must serve
the UI from `linkshelf/web/`, reject static paths containing `..`, listen on
`:8080`, and be started by Playwright as `go run ./cmd/server` from
linkshelf. The e2e test must exercise those real routes and persistence rather
than inventing helpers or alternate paths.

## Verification

From the linkshelf project directory run exactly:
`go mod tidy && go test ./... && npx playwright test`.
The phase is complete only when Go compilation, server wiring, static delivery,
browser CRUD, persistence after reload, and deletion all pass.
