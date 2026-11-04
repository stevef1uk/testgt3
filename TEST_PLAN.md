# Test Plan — Linkshelf

## Requirements → tests

### backend-setup
Requirement: Deliver the Go module and Link schema with idempotent SQLite initialization.
Level: unit
Test file: linkshelf/internal/store/schema_test.go
Bead ID: te3-all
Phase: backend-setup
Scenarios:
- Open an in-memory SQLite database and call InitSchema twice.
- Inspect the Link fields and JSON names, and confirm the links table can be used.
Assertions:
- Both schema calls succeed without duplicate-table errors.
- Link has id, title, url, and created_at fields with the documented JSON names.

### backend-store
Requirement: Persist links through package-level List, Create, and Delete functions with validation and timestamps.
Level: unit
Test file: linkshelf/internal/store/store_test.go
Bead ID: 
Phase: backend-store
Scenarios:
- Verify an empty List is non-nil, create valid links, list them in descending ID order, and confirm fields and RFC3339 UTC timestamps.
- Reject empty or over-200-rune titles, empty URLs, and URLs without an http:// or https:// prefix.
- Delete one link, then attempt to delete a missing ID.
Assertions:
- Valid rows round-trip with stable IDs and the required fields; empty List is [] rather than null.
- Invalid creates return errors without mutation; deletion removes only the selected row and missing deletion returns an error.

### backend-api
Requirement: Expose the documented root, static, list, create, and delete routes with exact JSON responses and HTTP statuses.
Level: integration
Test file: linkshelf/internal/api/handlers_test.go
Bead ID: te3-xe4
Phase: backend-api
Scenarios:
- GET /api/links on an empty store, POST a valid JSON link, then GET /api/links and DELETE the created ID using httptest.
- Send malformed JSON, missing or invalid fields, unsupported methods, unknown API paths, invalid IDs, and deletion of a missing ID.
- Exercise GET / and safe versus traversal or unknown static paths through the server handler.
Assertions:
- List returns 200 with JSON [] initially and includes the created link; create returns 201 JSON and delete returns 204.
- Invalid creates return 400 JSON errors without store mutation; missing deletes return 404 JSON errors; unsupported or unmatched routes do not succeed.
- Root and safe static GETs return 200, while traversal, unknown, and non-GET static requests return 404.

### frontend-ui
Requirement: Render the required link shelf controls and perform browser add, list, and delete operations through the API.
Level: ui
Test file: linkshelf/tests/e2e.spec.js
Bead ID: 
Phase: frontend-ui
Scenarios:
- Load the page and verify title input, URL input, Add button, and ul#links are present.
- Submit a valid link, verify its title and URL render, then activate its delete control.
Assertions:
- The page loads and renders the API list; adding refreshes the list and displays the link.
- Deleting refreshes the list and removes the selected link without a full-page navigation.

### e2e-testing
Requirement: Provide Playwright verification of UI loading, static-file serving, adding, displaying, and deleting a link.
Level: integration
Test file: linkshelf/tests/e2e.spec.js
Bead ID: 
Phase: e2e-testing
Scenarios:
- Run Playwright against the configured server, load the root page, and verify a static asset response.
- Add a link through the UI, verify it appears in #links, and delete it.
Assertions:
- Playwright reports successful root and static delivery.
- The created link is visible after adding and absent after deletion.

### smoke-test
Requirement: Recheck the complete SPEC definition of done with Go tests, a running server, and Playwright.
Level: integration
Test file: linkshelf/tests/e2e.spec.js
Bead ID: 
Phase: smoke-test
Scenarios:
- Run go mod tidy and go test ./..., start the server from linkshelf/, and wait for port 8080.
- Request / and /api/links, then run the primary add/list/delete browser flow and Playwright suite.
Assertions:
- Go verification passes and the server serves the UI and API on :8080.
- Playwright completes the documented UI load, static asset, add, and delete checks.
