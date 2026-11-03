# Test Plan — LinkShelf

## Requirements → tests

### backend-setup
Requirement: Establish the Go module and schema foundation required to build LinkShelf.
Level: integration
Test file: linkshelf/internal/store/schema.go
Bead ID: 
Phase: backend-setup
Scenarios:
- Run the phase build verification with the declared module and SQLite dependency.
- Initialize the schema on a fresh database and repeat initialization.
Assertions:
- The linkshelf module builds successfully with the required Go and SQLite declarations.
- Schema initialization is idempotent and provides the Link table required by later phases.

### backend-store
Requirement: Persist links with the domain operations defined by the specification.
Level: unit
Test file: linkshelf/internal/store/store_test.go
Bead ID: 
Phase: backend-store
Scenarios:
- Create, list, and delete a link through the store API.
- Exercise empty storage and missing-link behavior.
Assertions:
- Stored fields round-trip without mutation.
- Missing records return the specified not-found/error result.
- List results are deterministic and deletion removes the record.

### backend-api
Requirement: Expose the specified link API routes and response contracts.
Level: integration
Test file: linkshelf/internal/api/handlers_test.go
Bead ID: 
Phase: backend-api
Scenarios:
- Call every documented route with valid requests.
- Send malformed JSON, missing fields, and unknown IDs.
Assertions:
- Each route uses the documented method, path, status, and JSON shape.
- Invalid input is rejected without changing stored data.
- Unknown resources return the documented error status.





### e2e-testing
Requirement: Cover the complete link-management flow against the running server.
Level: ui
Test file: linkshelf/tests/e2e.spec.js
Bead ID: te-6ca
Phase: e2e-testing
Scenarios:
- Run the Playwright Chromium suite against the configured web server.
- Execute the complete create, read, update, and delete flow.
- Reload between operations to verify server-backed persistence.
Assertions:
- The suite starts the Go server using the Playwright webServer configuration.
- The end-to-end flow passes in Chromium.
- Browser-visible state agrees with the server-backed state.

### smoke-test
Requirement: Verify the assembled server remains buildable and its primary UI/API entrypoint is reachable.
Level: integration
Test file: linkshelf/cmd/server/main.go
Bead ID: 
Phase: smoke-test
Scenarios:
- Build the completed server with the documented command.
- Start it and request the root page and primary links API endpoint.
Assertions:
- The server builds and starts without schema or dependency errors.
- The root serves the UI and the API returns the documented JSON response.


