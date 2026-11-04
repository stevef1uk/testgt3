# Test Plan — LinkShelf

## Requirements → tests

### backend-setup
Requirement: Deliver the Go module and Link schema with idempotent SQLite initialization.
Level: unit
Test file: linkshelf/internal/store/schema_test.go
Bead ID: te3-all
Phase: backend-setup
Scenarios:
- Open an in-memory SQLite database and call InitSchema twice.
- Inspect the links table columns and verify the Link JSON field contract.
Assertions:
- Both schema calls succeed without duplicate-table errors.
- The schema contains id, title, url, and created_at, and Link exposes the documented JSON names.

### backend-store
Requirement: Persist links through package-level List, Create, and Delete functions with validation and timestamps.
Level: unit
Test file: linkshelf/internal/store/store_test.go
Bead ID: 
Phase: backend-store
Scenarios:
- Verify an empty List is non-nil, valid links are listed in descending ID order, and timestamps are RFC3339 UTC.
- Reject empty or over-200-rune titles, empty URLs, and URLs without an http:// or https:// prefix.
- Delete one link, then attempt to delete a missing ID.
Assertions:
- Valid rows round-trip with stable IDs and required fields; empty List is [] rather than null.
- Invalid creates do not mutate the store; deletion removes only the selected row and missing deletion returns an error.

### backend-api
Requirement: Expose the documented root, static, list, create, and delete routes with exact JSON responses and HTTP statuses.
Level: integration
Test file: linkshelf/internal/api/handlers_test.go
Bead ID: 
Phase: backend-api
Scenarios:
- Exercise documented list, create, and delete routes with valid JSON.
- Send malformed and incomplete requests.
- Request the documented root and static asset routes, including a traversal-like path.
Assertions:
- Valid routes return documented statuses and JSON shapes.
- Created links are listable and deleted links are absent.
- Invalid input is rejected without mutation.
- Root and JavaScript asset are served while traversal does not escape static files.

### frontend-ui
Requirement: Render the required controls and perform browser add, list, and delete operations through the API.
Level: ui
Test file: linkshelf/tests/e2e.spec.js
Bead ID: 
Phase: frontend-ui
Scenarios:
- Load the page and inspect the links container and add form.
- Submit a title and URL.
- Delete a displayed link.
Assertions:
- The links list and controls render.
- The submitted title is rendered as a link with the submitted URL.
- Deleting removes the selected list item from the DOM.

### e2e-testing
Requirement: Provide Playwright verification of UI loading, static-file serving, adding, displaying, and deleting a link.
Level: ui
Test file: linkshelf/tests/e2e.spec.js
Bead ID: 
Phase: e2e-testing
Scenarios:
- Load the UI and request its JavaScript asset.
- Add a uniquely named link, reload, and find it.
- Add a uniquely named link and delete it.
Assertions:
- The page contains #links and the asset succeeds with client fetch behavior.
- An added link remains visible with the same href after reload.
- A deleted link is no longer visible.

### smoke-test
Requirement: Recheck the complete SPEC definition of done with Go tests, a running server, and Playwright.
Level: integration
Test file: linkshelf/tests/e2e.spec.js
Bead ID: 
Phase: smoke-test
Scenarios:
- Run the service and probe the documented health endpoint.
- Probe the root page and JavaScript asset over HTTP.
Assertions:
- Health returns the documented success response.
- Root and static requests return expected content.
- The service remains usable after the smoke probes.
