# Test Review — Linkshelf

## Summary
Phase Go verification passes, but the planned test coverage is inadequate for all listed requirements. Review fails and requires Polecat test rework.

## Per-requirement results

### backend-setup — FAIL
- Test file: `linkshelf/internal/store/schema_test.go`
- Verify result: `cd linkshelf && go test ./...` passed.
- Notes: The plan describes server construction, configurable `ADDR`, root/static serving, traversal rejection, unknown paths, and unsupported methods. `schema_test.go` only exercises the schema/model with a fake SQL driver and does not invoke `newServer`, the entrypoint wiring, or any HTTP route. Reopen the backend-setup implementation bead (the plan's Bead ID is blank) and add real assertions in the planned test path or correct the plan/test mapping.

### backend-store — FAIL
- Test file: `linkshelf/internal/store/store_test.go`
- Verify result: `cd linkshelf && go test ./...` passed.
- Notes: CRUD and validation assertions are real, but the planned scenarios claim row field round-trip, idempotent schema initialization, and post-delete data integrity. The test does not verify title/URL/created-at round-trip through `List`, does not call `InitSchema` more than once, and does not assert the remaining row after deletion. Reopen the backend-store implementation bead (the plan's Bead ID is blank) and strengthen `linkshelf/internal/store/store_test.go`.

### backend-api — FAIL
- Test file: `linkshelf/internal/api/handlers_test.go`
- Verify result: `cd linkshelf && go test ./...` passed.
- Notes: The test covers collection/create/delete happy path and a few errors, but does not cover the planned JSON content-type checks, individual GET scenario, invalid URL/title cases, mutation safety after rejected input, invalid IDs, static/root routes, or traversal rejection. The planned test says “httptest server” but directly calls the handler with a custom recorder and does not verify entrypoint wiring. Reopen the backend-api implementation bead (the plan's Bead ID is blank) and add the missing contract assertions.

### frontend-ui — FAIL
- Test file: `linkshelf/tests/e2e.spec.js`
- Verify result: Go phase verification passed; UI coverage is not established by `go test ./...`.
- Notes: The plan requires empty/loading/error states, escaped user input, no full-page navigation, and API failure behavior. These must be verified by substantive Playwright assertions, not merely file presence. The configured frontend/e2e commands are separate from Go verification and must be run for this requirement. Reopen the frontend-ui implementation bead (the plan's Bead ID is blank) if those assertions are absent or failing.

### e2e-testing — FAIL
- Test file: `linkshelf/tests/e2e.spec.js`
- Verify result: Go phase verification passed; the required Playwright command is separate and must pass.
- Notes: The plan requires temporary-database startup, API verification, restart persistence, deletion reflected in API, and invalid-request/status coverage. A UI-only add/delete test does not establish those scenarios. Reopen the e2e-testing implementation bead (the plan's Bead ID is blank) and expand/run the Playwright suite.

### smoke-test — FAIL
- Test file: `linkshelf/tests/e2e.spec.js`
- Verify result: Go phase verification passed; the smoke phase command additionally requires `go mod tidy` and Playwright execution.
- Notes: The planned smoke scenarios require launching the built server, readiness, root/API probes, CRUD, clean termination, and actionable failure output. These are not covered by the Go tests listed for this row unless the Playwright suite explicitly performs them. Reopen the smoke-test implementation bead (the plan's Bead ID is blank) and add/run the repeatable smoke coverage.

## Overall assessment
FAIL. `go test ./...` is green, but passing compilation/unit tests does not establish the planned integration, UI, E2E, and smoke requirements. The exact planned test paths need stronger real assertions and the non-Go verification commands need to be executed successfully before QA sign-off.
