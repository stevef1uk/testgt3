# Test Plan — Link Shelf

## Requirements → tests

### backend-setup
Requirement: The module and schema must match the SPEC data model exactly, with `Link` and idempotent `InitSchema` in `internal/store/schema.go`.
Level: unit
Test file:
Bead ID: te-q4t
Phase: backend-setup
Scenarios:
- Build the module and schema package with the prescribed Go toolchain.
- Verify the supplied schema implementation defines the required table and initialization behavior.
Assertions:
- `go build ./...` succeeds for the linkshelf module.
- The schema implementation remains the sole owner of the Link model, DDL, and InitSchema function.

### backend-store
Requirement: Package-level `store.DB`, `List`, `Create`, and `Delete` implement the exact store API, validation, ordering, and missing-ID behavior.
Level: unit
Test file: linkshelf/internal/store/store_test.go
Bead ID:
Phase: backend-store
Scenarios:
- Initialize an in-memory database with `InitSchema`, assign `store.DB`, and list an empty table.
- Create valid links and verify newest IDs are listed first.
- Try empty or over-200-rune titles and empty or non-HTTP(S) URLs.
- Delete an existing link and then delete a missing ID.
Assertions:
- Empty `List` returns a non-nil empty slice.
- Valid creation persists title, URL, ID, and an RFC3339 UTC timestamp.
- Invalid inputs return errors without creating rows.
- `List` uses descending ID order and `Delete` errors when the ID is absent.

### backend-api
Requirement: HTTP handlers and `main` implement the specified routes, JSON contracts, status codes, static traversal protection, and SQLite/server wiring.
Level: integration
Test file: linkshelf/internal/api/handlers_test.go
Bead ID:
Phase: backend-api
Scenarios:
- Exercise `/api/links` GET, valid and invalid POST requests, and DELETE success/missing-ID requests with `httptest`.
- Request `/`, valid and missing `/static/{file}`, and a static path containing `..`.
- Exercise unsupported methods, malformed JSON, invalid IDs, and JSON response bodies.
Assertions:
- GET returns 200 and JSON `[]` when empty; POST returns 201 with a link; DELETE returns 204 or 404 with `{"error":"..."}`.
- Invalid POST input returns 400 with a deterministic JSON error.
- Root and valid static files return 200, while missing/traversal paths return 404.
- Handler responses use the specified Link JSON fields and API routes do not fall through to static files.
- Server wiring opens `linkshelf.db`, calls `InitSchema` before assigning `store.DB`, registers `http.DefaultServeMux`, logs port 8080, and listens on `:8080`.

### e2e-testing
Requirement: The Playwright harness must verify the live UI load, add/delete flow, and static file serving against the running application.
Level: ui
Test file: linkshelf/tests/e2e.spec.js
Bead ID:
Phase: e2e-testing
Scenarios:
- Load the application and verify the title input, URL input, Add button, and links list are visible.
- Enter a valid title and URL, add the link, then delete it through the UI.
- Request a served static asset through `/static/`.
Assertions:
- The page loads successfully and controls are usable.
- The newly added link appears in the list and disappears after deletion.
- The static asset request succeeds with the expected content/status.
- Playwright configuration starts or targets the server at the configured base URL and `npx playwright test` passes.

### smoke-test
Requirement: The completed server remains buildable and serves the UI on port 8080 while preserving database and route wiring.
Level: integration
Test file: linkshelf/cmd/server/main.go
Bead ID:
Phase: smoke-test
Scenarios:
- Build the complete Go application after all phases.
- Start the server with a writable working directory and request `/`, `/api/links`, and a static asset.
Assertions:
- `go build ./...` succeeds.
- The process starts on `:8080` and logs that it is listening.
- The root UI, API, and static routes respond without breaking SQLite initialization.
---END WRITE
