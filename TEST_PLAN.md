# Test Plan — testgt3

## Requirements → tests

### backend-setup
Requirement: Establish the backend project and runnable test configuration required by the backend phases.
Level: unit
Test file: linkshelf/internal/store/schema_test.go
Bead ID: te3-6tp
Phase: backend-setup
Scenarios:
- Build the backend from a clean checkout.
- Run the configured test command with no application data present.
Assertions:
- Setup succeeds without missing dependencies or packages.
- The test command exits successfully.

### backend-store
Requirement: Provide the specified persistent store behavior and store API.
Level: unit
Test file: linkshelf/internal/store/store_test.go
Bead ID:
Phase: backend-store
Scenarios:
- Create, read, update, and delete each specified domain record.
- Exercise empty-store, missing-record, duplicate, and invalid-input cases.
Assertions:
- Valid operations persist and return the specified values.
- Missing and invalid operations return the specified errors without corrupting data.
- Repeated reads and writes preserve the documented invariants.

### backend-api
Requirement: Expose the documented HTTP routes and connect them to the store.
Level: integration
Test file: linkshelf/internal/api/handlers_test.go
Bead ID:
Phase: backend-api
Scenarios:
- Call every documented route with valid requests.
- Call routes with malformed payloads, unknown identifiers, unsupported methods, and missing resources.
Assertions:
- Status codes, response bodies, headers, and content types match the route table.
- Successful requests invoke the store and return persisted results.
- Invalid requests return documented errors and do not mutate state.

### frontend-ui
Requirement: Implement the specified user-facing UI and its backend interactions.
Level: ui
Test file: linkshelf/web/app_test.js
Bead ID:
Phase: frontend-ui
Scenarios:
- Load the UI and complete each documented primary user flow.
- Verify loading, empty, success, validation-error, and server-error states.
Assertions:
- Required controls, labels, data, and navigation are visible and usable.
- User actions issue the documented API requests and render their results.
- Errors are surfaced without losing valid user input or showing stale success state.

### e2e-testing
Requirement: Verify the complete documented flow across the running backend and UI.
Level: ui
Test file: linkshelf/tests/e2e.spec.js
Bead ID:
Phase: e2e-testing
Scenarios:
- Start the application and execute the principal end-to-end flow in a clean environment.
- Repeat the flow with invalid input and an unavailable or failing backend response.
Assertions:
- The browser can complete the happy path against the real server.
- UI and API data agree after each mutation.
- Failure states are deterministic and actionable.

### smoke-test
Requirement: Confirm the delivered application starts and its documented critical path remains operational.
Level: integration
Test file: linkshelf/tests/e2e.spec.js
Bead ID:
Phase: smoke-test
Scenarios:
- Run the delivery smoke command against a clean build.
- Exercise the critical route and one representative UI/API path.
Assertions:
- The application starts within the configured command and exits/serves as documented.
- Critical endpoints respond with the documented status and shape.
