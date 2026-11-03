# Test Plan — LinkShelf

## Requirements → tests

### backend-setup
Requirement: Initialize the LinkShelf backend and create the links database schema without failing on a fresh database or an already-initialized database.
Level: unit
Test file: linkshelf/internal/store/schema.go
Bead ID: te-q4t, te-t7g
Phase: backend-setup
Scenarios:
- Initialize a fresh in-memory database and verify the links table and required columns exist.
- Run schema initialization twice and verify it is idempotent.
Assertions:
- InitSchema returns no error and the table is queryable.
- The schema contains the required id, title, url, and created_at columns.

### backend-store
Requirement: Implement package-level DB, List, Create, and Delete with ordering, validation, timestamp, and missing-ID behavior.
Level: unit
Test file: linkshelf/internal/store/store_test.go
Bead ID:
Phase: backend-store
Scenarios:
- List an initialized empty database, create multiple links, and verify ORDER BY id DESC.
- Create valid links, including a Unicode title at 200 runes, then test empty/201-rune titles and non-http(s) URLs.
- Delete an existing link and attempt to delete an absent ID.
Assertions:
- List returns a non-nil [] and exact Link fields; Create returns a positive ID and RFC3339 UTC CreatedAt.
- Validation rejects empty or overlong titles and URLs without http:// or https:// without inserting rows.
- Delete removes an existing row and returns an error for a missing ID.

### backend-api
Requirement: Expose the exact root, static, list, create, and delete HTTP API contract with documented JSON responses, status codes, and static path safety.
Level: integration
Test file: linkshelf/internal/api/handlers_test.go
Bead ID:
Phase: backend-api
Scenarios:
- GET / returns index.html; GET /static/style.css returns 200, while a missing file and a path containing .. return 404.
- GET /api/links returns JSON [] for an empty store and returns newest-first links after creation.
- POST valid JSON returns 201 and the created link; malformed JSON, missing fields, invalid title length, and invalid URL return 400 with an error object.
- DELETE an existing numeric ID returns 204; deleting a missing or malformed ID returns 404 with an error object.
Assertions:
- Route methods, status codes, content types, JSON field names, and empty-array shape match SPEC.md.
- Static responses cannot escape linkshelf/web or fall through from API paths.
- Invalid POST requests do not create records and delete errors are deterministic JSON.



### e2e-testing
Requirement: Verify the complete running-application browser flow, including UI load, add/delete persistence, and static-file serving.
Level: ui
Test file: linkshelf/tests/e2e.spec.js
Bead ID:
Phase: e2e-testing
Scenarios:
- Start the configured server, load the page, and verify title and URL inputs, Add button, and links list are visible.
- Add a valid bookmark, reload, and verify its title and destination are rendered; then delete it and verify it disappears after refresh.
- Request a static asset through /static/ and verify a missing asset is not served.
Assertions:
- Playwright runs against the real server and SQLite-backed API, not mocks.
- Add and delete operations refresh the list and persist across page reload.
- Static assets are served beneath the web mapping with expected success or 404 status.

### smoke-test
Requirement: Verify the assembled LinkShelf application starts successfully and its primary user path is reachable.
Level: integration
Test file: linkshelf/cmd/server/main.go
Bead ID:
Phase: smoke-test
Scenarios:
- Start the built application using the documented command.
- Request the root UI and the primary links API route.
Assertions:
- The process starts without a schema or dependency error.
- The root responds successfully with the UI and the API responds with valid JSON.
