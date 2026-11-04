# Test Plan — Linkshelf

## Requirements → tests

### backend-setup
Requirement: Provide the Go server entry point, SQLite dependency, configurable listen address, and static-file serving setup.
Level: integration
Test file: linkshelf/internal/store/schema_test.go
Bead ID: 
Phase: backend-setup
Scenarios:
- Start the server with the default address and with ADDR set to a test address.
- Request `/` and an existing `/static/` asset, and request an unknown path and a path containing `..`.
Assertions:
- The server is constructible and serves the index and safe static assets with GET.
- Unknown, traversal, and unsupported-method requests return 404; the configured address is honored.

### backend-store
Requirement: Persist links in SQLite and provide schema initialization and link CRUD behavior.
Level: unit
Test file: linkshelf/internal/store/store_test.go
Bead ID: 
Phase: backend-store
Scenarios:
- Initialize an in-memory database, create links, list them, retrieve them, and delete one.
- Exercise empty storage, missing IDs, invalid required values, and schema initialization more than once.
Assertions:
- Schema creation is idempotent and persisted rows round-trip with stable IDs and fields.
- Delete removes only the selected row; missing records and invalid inputs return the specified errors without corrupting data.

### backend-api
Requirement: Expose the documented link API routes with JSON responses, validation, and correct HTTP status codes.
Level: integration
Test file: linkshelf/internal/api/handlers_test.go
Bead ID: 
Phase: backend-api
Scenarios:
- GET the collection, POST a valid link, GET the created link, and DELETE it through an httptest server.
- Send malformed JSON, missing or invalid fields, unknown IDs, unsupported methods, and invalid JSON requests.
Assertions:
- Successful responses have the documented status, JSON content type, and response shape.
- Bad input is rejected with 4xx without a store mutation, missing links return 404, and delete is scoped to the requested ID.

### frontend-ui
Requirement: Render the link shelf in the browser and let a user add and remove links through the API.
Level: ui
Test file: linkshelf/tests/e2e.spec.js
Bead ID: 
Phase: frontend-ui
Scenarios:
- Load the page with an empty collection, submit a valid link, and reload/list the created link.
- Use the delete control and exercise API failure and empty-state rendering.
Assertions:
- The page displays links and their title/URL, updates without a full-page navigation, and removes the selected item.
- Empty, loading, and error states are visible and user input is escaped rather than interpreted as markup.

### e2e-testing
Requirement: Verify the complete running application across browser, HTTP API, persistence, and static assets.
Level: integration
Test file: linkshelf/tests/e2e.spec.js
Bead ID: 
Phase: e2e-testing
Scenarios:
- Start the compiled server against a temporary database, open the UI, add a link, verify it through the API, restart, and delete it.
- Check root/static asset delivery and the documented invalid-request paths.
Assertions:
- A link created in the UI is persisted and returned by the API after restart, and deletion is reflected in both.
- The browser receives usable assets and all documented success and error status codes match the API contract.

### smoke-test
Requirement: Provide a repeatable smoke check that proves the built server starts and the primary shelf flow works.
Level: integration
Test file: linkshelf/tests/e2e.spec.js
Bead ID: 
Phase: smoke-test
Scenarios:
- Launch the built binary with a temporary database and wait for the configured port.
- Probe `/`, the primary links endpoint, and one create/list/delete flow, then terminate the process.
Assertions:
- The process becomes ready, serves the expected content, and exits cleanly after the smoke run.
- The primary happy path succeeds end to end and failures produce actionable test output.
