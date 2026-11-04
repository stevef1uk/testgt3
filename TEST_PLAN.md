# Test Plan — testgt3

## Requirements → tests

### backend-setup
Requirement: Deliver `linkshelf/go.mod` and `linkshelf/internal/store/schema.go`; declare module linkshelf with Go 1.22 and SQLite, define the exact Link fields/JSON names, and make InitSchema create the links table idempotently.
Level: unit
Test file: linkshelf/internal/store/schema_test.go
Bead ID: te3-6tp
Phase: backend-setup
Scenarios:
- Open an isolated in-memory SQLite database and call InitSchema twice.
- Inspect the Link JSON behavior and verify the links table exists after initialization.
- Run the backend setup verification so the module and schema package compile.
Assertions:
- InitSchema succeeds on both calls without duplicating or corrupting the table.
- Link fields and JSON names are id, title, url, and created_at.
- The schema test and `go test ./...` pass with the declared module dependency.

### backend-store
Requirement: Implement the persistent store and its domain operations defined by the specification.
Level: unit
Test file: linkshelf/internal/store/store_test.go
Bead ID:
Phase: backend-store
Scenarios:
- Create, retrieve, update, and delete a valid record.
- Query an empty store and request a missing record.
- Reject invalid input without corrupting existing data.
Assertions:
- Successful operations round-trip all required fields.
- Missing data returns the specified not-found result.
- Invalid operations return the specified error and preserve state.

### backend-api
Requirement: Expose the specified HTTP API routes and connect them to the store.
Level: integration
Test file: linkshelf/internal/api/handlers_test.go
Bead ID:
Phase: backend-api
Scenarios:
- Exercise every documented route with valid input.
- Exercise malformed payloads, unknown identifiers, unsupported methods, and unknown routes.
- Verify JSON response headers, status codes, and response bodies.
Assertions:
- Each route returns the documented status and schema.
- Successful mutations are observable through subsequent reads.
- Client errors do not panic or mutate state.



### frontend-ui
Requirement: Deliver the required bookmark UI markup, browser CRUD behavior, and readable styling.
Level: ui
Test file: linkshelf/tests/e2e.spec.js
Bead ID:
Phase: frontend-ui
Scenarios:
- Load the page and verify title input, URL input, Add button, and links list are present.
- Load existing links, add a valid link, and delete it through the browser.
- Verify the client refreshes the list after loading and after each mutation.
Assertions:
- The required controls and `#links` list are visible.
- Added links render with their title and URL, and deleted links disappear.
- Browser requests use the documented list, create, and delete API operations.

### e2e-testing
Requirement: Verify the complete user workflow across the running frontend and backend.
Level: ui
Test file: linkshelf/tests/e2e.spec.js
Bead ID:
Phase: e2e-testing
Scenarios:
- Start the application, load the UI, and complete the principal end-to-end workflow.
- Refresh after a mutation and confirm persisted state remains visible.
- Cover a failed API response in the browser.
Assertions:
- The workflow succeeds across real HTTP boundaries.
- Persisted data survives navigation or reload.
- The UI shows a recoverable error and remains usable.

### smoke-test
Requirement: Recheck the complete SPEC definition of done, including Go tests, server startup, UI CRUD, and Playwright execution.
Level: integration
Test file: linkshelf/tests/e2e.spec.js
Bead ID:
Phase: smoke-test
Scenarios:
- Run the documented smoke command against a clean application.
- Check the primary documented endpoint and a minimal UI/backend connection.
Assertions:
- The command exits successfully.
- The application responds on the configured interface with the documented result.
