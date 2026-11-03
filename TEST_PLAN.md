# Test Plan — LinkShelf

## Requirements → tests

### backend-setup
Requirement: Provide a runnable Go server that starts on the configured port and serves the LinkShelf application.
Level: integration
Test file: linkshelf/internal/store/schema.go
Bead ID:
Phase: backend-setup
Scenarios:
- Start the server through the configured command and request the root page.
- Request the declared static stylesheet and JavaScript assets.
Assertions:
- The server accepts requests on port 8080 and returns a successful application page.
- `/static/style.css` and `/static/app.js` return successful responses.

### backend-store
Requirement: Store saved links and support creating, listing, and deleting link records.
Level: unit
Test file: linkshelf/internal/store/store_test.go
Bead ID:
Phase: backend-store
Scenarios:
- Add a link with a title and URL and observe it in the saved-links list.
- Delete the created link and reload the page.
Assertions:
- The created title and URL are represented by a visible link.
- DELETE succeeds and the deleted link is no longer visible after the operation.

### backend-api
Requirement: Expose the link API used by the UI, including creation and deletion, with valid HTTP responses.
Level: integration
Test file: linkshelf/internal/api/handlers_test.go
Bead ID:
Phase: backend-api
Scenarios:
- Submit the add-link form and wait for the resulting API-driven list update.
- Delete a list item and capture the DELETE request to `/api/links/<id>`.
Assertions:
- The add operation produces a saved link in the rendered list.
- The delete request uses DELETE, reaches the link endpoint, and removes the item.



### e2e-testing
Requirement: Verify the complete LinkShelf browser flow and its served static assets against the running server.
Level: ui
Test file: linkshelf/tests/e2e.spec.js
Bead ID: te-6ca
Phase: e2e-testing
Scenarios:
- Run the Playwright Chromium suite against the configured web server.
- Exercise page load, asset requests, add, and delete in one browser session.
Assertions:
- The suite can start the server and all LinkShelf smoke and CRUD tests pass.
- Browser-visible state and HTTP responses match the expected application behavior.

### smoke-test
Requirement: Confirm the built application is launchable and its primary page is healthy.
Level: integration
Test file: linkshelf/cmd/server/main.go
Bead ID:
Phase: smoke-test
Scenarios:
- Launch `go run ./cmd/server` through Playwright's webServer configuration.
- Load `/` and verify the primary shell and list container.
Assertions:
- The process becomes ready on port 8080 without a startup failure.
- The page title, Saved links heading, and links container are present.
