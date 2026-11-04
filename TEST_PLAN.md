# Test Plan

## Requirements → tests

### backend-setup
Requirement: The backend project is initialized according to the backend setup requirements.
Level: unit
Test file: linkshelf/internal/store/schema_test.go
Bead ID: te3-6tp
Phase: backend-setup
Scenarios:
- Build and initialize the backend from a clean checkout.
- Verify the configured package/module and test command are usable.
Assertions:
- Backend setup completes without errors.
- The backend test suite is discoverable and passes.

### backend-store
Requirement: The backend store implements the specified data model and store operations.
Level: unit
Test file: linkshelf/internal/store/store_test.go
Bead ID: 
Phase: backend-store
Scenarios:
- Create, retrieve, update, and delete valid records through the store API.
- Exercise empty, missing, duplicate, and invalid inputs.
Assertions:
- Successful operations persist and return the specified values.
- Missing and invalid records return the specified errors without corrupting state.

### backend-api
Requirement: The backend API exposes every specified route with the specified methods, status codes, and response bodies.
Level: integration
Test file: linkshelf/internal/api/handlers_test.go
Bead ID: 
Phase: backend-api
Scenarios:
- Call each documented route with valid requests and decode its response.
- Exercise unsupported methods, malformed payloads, missing records, and invalid parameters.
Assertions:
- Routes return the documented status and response schema.
- Errors are stable, appropriately classified, and do not panic or mutate state.

### frontend-ui
Requirement: The frontend renders the specified user interface and connects user actions to the backend API.
Level: ui
Test file: linkshelf/tests/e2e.spec.js
Bead ID: 
Phase: frontend-ui
Scenarios:
- Load the UI and verify the documented controls, content, and initial state.
- Perform each documented user flow, including loading, success, empty, and error states.
Assertions:
- Required elements and rendered data are visible and accessible.
- User actions issue the expected API calls and render the resulting state.

### e2e-testing
Requirement: The complete application supports the specified end-to-end user flows.
Level: integration
Test file: linkshelf/tests/e2e.spec.js
Bead ID: 
Phase: e2e-testing
Scenarios:
- Start the backend and frontend and complete the primary documented flow.
- Verify persistence and behavior across a fresh page load.
Assertions:
- The flow succeeds through the real HTTP and UI boundaries.
- Data shown by the UI matches backend state and documented failure behavior.

### smoke-test
Requirement: The delivered application passes the documented smoke verification.
Level: integration
Test file: linkshelf/tests/e2e.spec.js
Bead ID: 
Phase: smoke-test
Scenarios:
- Run the documented startup and smoke commands from a clean environment.
- Probe the primary route and confirm the application remains healthy.
Assertions:
- Startup, build, and smoke verification complete successfully.
- The primary documented behavior is available without regressions.
