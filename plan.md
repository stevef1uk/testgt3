# Implementation plan

## Phase
Implement the e2e-testing phase for Link Shelf. The three files below provide the Playwright package metadata, browser/server configuration, and end-to-end behavioral coverage. Do not add backend or frontend implementation in this phase.

## Bead map

### te3-fq3: linkshelf/package.json
- Scope: Define the Playwright test package metadata and the test entry point.
- Architecture: `linkshelf/package.json`; architecture.md e2e-testing configuration.
- Interfaces: npm `test` script invoking `npx playwright test --project=chromium`; development dependency `@playwright/test`.
- Depends on: `<none>`; this metadata is the first e2e-testing dependency.
- Consumed by: `te3-4ph: linkshelf/playwright.config.js` and `te3-aio: linkshelf/tests/e2e.spec.js` through the Playwright runner.
- Acceptance:
  - The JSON is valid and declares `@playwright/test` as a development dependency.
  - The package declares a `test` script that runs the Playwright suite, targeting the chromium project.
  - The package is sufficient for the phase verification command `cd linkshelf && npx playwright test`.
- Verify: `cd linkshelf && npx playwright test`

### te3-4ph: linkshelf/playwright.config.js
- Scope: Configure Playwright discovery, the application base URL, and lifecycle startup for the Go server.
- Architecture: `linkshelf/playwright.config.js`; architecture.md e2e-testing configuration.
- Interfaces: Playwright `defineConfig` configuration with `testDir: './tests'`, base URL `http://127.0.0.1:8080`, and a `webServer` running `go run ./cmd/server`.
- Depends on: `te3-fq3: linkshelf/package.json` for the Playwright package and test runner.
- Consumed by: `te3-aio: linkshelf/tests/e2e.spec.js` when tests resolve relative URLs and run against the configured server.
- Acceptance:
  - The configuration uses `testDir: './tests'` so `tests/e2e.spec.js` is discovered.
  - The browser context base URL is exactly `http://127.0.0.1:8080`.
  - `webServer.command` starts the real application with `go run ./cmd/server`.
  - The web server readiness URL is `http://127.0.0.1:8080/health`, so tests begin only after the application is ready.
  - The phase verification command is `cd linkshelf && npx playwright test`.
- Verify: `cd linkshelf && npx playwright test`

### te3-aio: linkshelf/tests/e2e.spec.js
- Scope: Exercise the user-visible Link Shelf UI and its static JavaScript asset through Playwright.
- Architecture: `linkshelf/tests/e2e.spec.js`; architecture.md e2e-testing scenarios.
- Interfaces: Playwright tests use `/`, `/static/app.js`, the `#links` container, title and URL inputs, the `Add` button, rendered anchors, and per-item `Delete` buttons.
- Depends on: `te3-fq3: linkshelf/package.json` and `te3-4ph: linkshelf/playwright.config.js`; the running server exposes the SPEC UI and HTTP behavior.
- Consumed by: The e2e-testing phase verification and CI/browser regression coverage.
- Acceptance:
  - On loading `/`, the UI has an attached `#links` element; requesting `/static/app.js` succeeds with HTTP 200 and its response text contains `fetch`.
  - Filling the title and URL inputs, clicking `Add`, and observing `#links` renders an anchor with the entered title and exact URL `href`.
  - Reloading the page still shows the newly added link with the same title and exact URL, proving persistence through the UI.
  - Adding a link, locating its list item, clicking its `Delete` button, and asserting the item count is zero proves deletion from the list.
  - The spec uses the configured relative base URL and can run with `cd linkshelf && npx playwright test`.
- Verify: `cd linkshelf && npx playwright test`

## Verification
The polecat must run the exact phase command from the workflow profile after implementing all three files:

```sh
cd linkshelf && npx playwright test
```

The phase is complete only when the concrete UI, asset, add-and-persist, and delete assertions above pass.
