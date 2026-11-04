# Test Plan — LinkShelf

## Requirements → tests

### backend-setup
Requirement: Define the Link model and idempotent SQLite schema initialization, with the Go module supplying the declared SQLite dependency.
Level: unit
Test file: linkshelf/internal/store/schema_test.go
Bead ID: te3-all
Phase: backend-setup
Scenarios:
- Open an in-memory SQLite database and call InitSchema twice.
- Inspect the links table columns and verify the Link JSON field contract.
- Compile and run the store package through the linkshelf Go module.
Assertions:
- Both schema calls succeed without duplicate-table errors.
- The schema contains id, title, url, and created_at, and Link exposes the documented JSON names.
- The linkshelf module resolves its SQLite dependency and the store package compiles.

### backend-store
Requirement: Persist links through the package-level List, Create, and Delete API with validation, ordering, and timestamps.
Level: unit
Test file: linkshelf/internal/store/store_test.go
Bead ID: 
Phase: backend-store
Scenarios:
- Initialize an in-memory SQLite database and verify the links table can be used.
- Call initialization twice and pass a nil database.
Assertions:
- The table has the id, title, url, description, and created_at columns with the required non-null/default behavior.
- Repeated initialization is safe and nil input returns an error rather than panicking.

### backend-api
Requirement: Expose HTTP operations for listing, creating, and deleting links and reject malformed or unsupported requests.
Level: integration
Test file: linkshelf/internal/api/handlers_test.go
Bead ID: te3-xe4
Phase: backend-api
Scenarios:
- GET /api/links on an empty store returns an empty JSON array.
- POST a valid title and URL, then GET the collection and DELETE the created ID.
- POST malformed or incomplete JSON, use an unsupported method, and delete a missing ID.
Assertions:
- Successful list/create/delete responses use the documented 200, 201, and 204 statuses and JSON contains the created link.
- Invalid JSON or missing required data returns 400, unsupported methods return 405, and an unknown link returns 404.

### frontend-ui
Requirement: Serve the LinkShelf page and client asset, and render links with controls for adding and deleting them.
Level: ui
Test file: linkshelf/tests/e2e.spec.js
Bead ID: 
Phase: frontend-ui
Scenarios:
- Load / and verify the links container and JavaScript asset are available.
- Add a title and URL through the form and observe the link in the list.
- Delete a displayed link using its delete control.
Assertions:
- The page exposes #links, /static/app.js is served, and the asset contains the client fetch behavior.
- Added links show the submitted title and href; deleting removes the corresponding list item.

### e2e-testing
Requirement: Verify the complete browser flow persists a created link across a fresh page load.
Level: ui
Test file: linkshelf/tests/e2e.spec.js
Bead ID: 
Phase: e2e-testing
Scenarios:
- Start the server, create a uniquely titled link in the browser, and reload the page.
- Locate the same link after reload and verify its delete flow.
Assertions:
- The created link is visible with the submitted URL before and after reload.
- The browser flow completes through the running HTTP server without relying on mocked API responses.

### smoke-test
Requirement: Recheck the complete SPEC definition of done with Go tests, a running server, and Playwright.
Level: integration
Test file: linkshelf/tests/e2e.spec.js
Bead ID: 
Phase: smoke-test
Scenarios:
- Start the server using its normal entry point and request the root page.
- Request GET /api/links from a clean server instance.
Assertions:
- The server starts without error, root responds successfully with the LinkShelf UI, and its documented static asset is reachable.
- GET /api/links responds successfully with a JSON array; no undocumented health endpoint is required.

