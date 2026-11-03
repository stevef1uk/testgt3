# Test Plan — Linkshelf

## Requirements → tests

### backend-setup
Requirement: The backend project and server entry point are initialized so the application can be built and started.
Level: integration
Test file: linkshelf/internal/store/schema.go
Bead ID: 
Phase: backend-setup
Scenarios:
- Build the server from the repository root.
- Start the server using the documented command and verify it binds successfully.
Assertions:
- The build exits successfully.
- The process starts without a missing-package, missing-entrypoint, or configuration error.

### backend-store
Requirement: The store persists links and supports the specified link operations and validation rules.
Level: unit
Test file: linkshelf/internal/store/store_test.go
Bead ID: 
Phase: backend-store
Scenarios:
- Add a valid link, retrieve it, list it, update it, and delete it.
- Reject invalid link data and return the specified not-found result for an unknown link.
- Verify that identifiers and persisted values remain stable across operations.
Assertions:
- Successful operations return the documented values and mutate only the intended record.
- Invalid input is rejected and unknown records do not produce a false success.
- A fresh store has an empty list and deletion removes the record.

### backend-api
Requirement: The HTTP API exposes the specified routes, methods, status codes, JSON shapes, and error behavior over the store.
Level: integration
Test file: linkshelf/internal/api/handlers_test.go
Bead ID: 
Phase: backend-api
Scenarios:
- Exercise every route from the SPEC route table with valid requests.
- Send malformed JSON, invalid values, unsupported methods, and unknown identifiers.
- Verify response headers and bodies for success and failure cases.
Assertions:
- Each route uses the documented method and returns the documented status code.
- JSON responses have the required fields and content type.
- API errors are deterministic and do not modify store state.



### frontend-ui
Requirement: The frontend provides the specified form, link list, live API loading, add/delete actions, and readable styling.
Level: ui
Test file: linkshelf/web/index.html
Bead ID: 
Phase: frontend-ui
Scenarios:
- Load the page and verify title input, URL input, Add button, and links list are present.
- Add a valid link and delete it through the rendered controls.
- Reload the page and verify the UI reflects the server-backed list.
Assertions:
- The required controls and `ul#links` are visible.
- Add sends the specified POST request and renders the created link.
- Delete sends the specified DELETE request and removes the link after refresh.
- Styling is loaded from the served static asset without a framework dependency.

### e2e-testing
Requirement: The end-to-end suite verifies the complete running-server link workflow in a real browser.
Level: ui
Test file: linkshelf/tests/e2e.spec.js
Bead ID: 
Phase: e2e-testing
Scenarios:
- Run the Playwright suite against the configured server and Chromium project.
- Complete the happy path from opening the application through creating and managing a link.
- Repeat the workflow from a clean state to detect leaked test data or ordering dependence.
Assertions:
- Playwright starts the configured server and all required browser assertions pass.
- The browser-visible state agrees with the API-backed state after each operation.
- Tests are isolated and pass with retries disabled.

### smoke-test
Requirement: The delivered application passes the documented build, server, API, and browser smoke checks.
Level: integration
Test file: linkshelf/cmd/server/main.go
Bead ID: 
Phase: smoke-test
Scenarios:
- Build and start the server using the documented commands.
- Check the documented baseline endpoint/page and run the focused browser smoke flow.
- Stop cleanly and verify the smoke command reports success.
Assertions:
- The server is reachable at the documented address and returns the expected baseline response.
- The initial page renders and the primary link workflow completes.
- The smoke command exits zero without unrelated test failures.
