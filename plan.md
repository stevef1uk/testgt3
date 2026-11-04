# Implementation plan

## Active phase: e2e-testing

This phase supplies the Playwright project metadata and browser acceptance coverage for
the Link Shelf application. The browser suite must exercise the existing server entry
point and the HTTP/UI contract defined by SPEC.md; it must not introduce backend
implementation or change the server wiring. The required verification is:
`cd linkshelf && npx playwright test`.

## Bead map

### te3-fq3: linkshelf/package.json
- Scope: Define the Playwright test package metadata and the `test` script using the
  pinned `@playwright/test` dependency, as specified by architecture.md.
- Architecture reference: architecture.md, e2e-testing test tooling and package
  configuration; preserve the repository's Node package layout and Chromium project
  command.
- Interfaces / Depends on / Consumed by: Exposes the npm `test` script and
  `@playwright/test` dev dependency; depends on no phase bead; consumed by
  linkshelf/playwright.config.js and linkshelf/tests/e2e.spec.js through the
  Playwright runner.
- Acceptance: `npm test` invokes the Chromium Playwright project; the package
  declares `@playwright/test` at the architecture-specified version; no backend
  paths or extra application features are added.
- Verify: `cd linkshelf && npx playwright test`

### te3-4ph: linkshelf/playwright.config.js
- Scope: Configure Playwright to discover `./tests`, use the Chromium desktop
  project, target `http://127.0.0.1:8080`, and start `go run ./cmd/server` through
  the configured web server health URL.
- Architecture reference: architecture.md, e2e-testing runner configuration and
  the SPEC requirement that the application is served on port 8080.
- Interfaces / Depends on / Consumed by: Provides the Playwright `testDir`,
  `baseURL`, Chromium project, and webServer command/configuration; depends on
  linkshelf/package.json (te3-fq3); consumed by linkshelf/tests/e2e.spec.js
  (te3-aio).
- Acceptance: Playwright discovers tests under `tests`; the Chromium project
  launches the server with `go run ./cmd/server`; the base URL is port 8080 and
  readiness checks `/health`; CI settings retain the architecture-specified
  retries, worker, and reporter behavior.
- Verify: `cd linkshelf && npx playwright test`

### te3-aio: linkshelf/tests/e2e.spec.js
- Scope: Implement browser acceptance tests for the user-visible Link Shelf UI
  and its static JavaScript asset, using the configured base URL.
- Architecture reference: architecture.md, e2e-testing scenarios; assertions must
  reflect SPEC.md's root page, `/static/app.js`, and `/api/links` CRUD behavior.
- Interfaces / Depends on / Consumed by: Uses Playwright `test` and `expect` from
  the package configured by te3-fq3 and linkshelf/playwright.config.js (te3-4ph);
  provides no application exports and is consumed by the Playwright runner.
- Acceptance: a test verifies `/` loads and `#links` is attached; a request to
  `/static/app.js` succeeds and its body contains `fetch`; a test fills the two
  inputs, clicks Add, verifies the created title and URL, reloads, and verifies
  persistence; a test adds a link, clicks its Delete control, and verifies its
  list item is removed. Tests use unique titles/URLs so CRUD cases do not rely on
  pre-existing data.
- Verify: `cd linkshelf && npx playwright test`

## Phase verification and handoff

Run the exact command above from the rig root after all three bead files are
present. A successful run must start the configured server, load the real UI,
serve the JavaScript asset, and pass add, persistence, and delete assertions.
The phase is not complete if the suite is skipped, mocked, or pointed at a
different server. Backend compilation failures belong to the backend/server
phase and must be repaired there rather than by changing these e2e files.

## Future considerations

None for this phase; persistence implementation and server dependency wiring are
owned by their architecture phases.
