# LinkShelf e2e-testing implementation plan

## Phase scope
Implement only the three active e2e-testing files under `linkshelf/`. The browser suite exercises the already-delivered LinkShelf server and UI; it must not add backend routes, application behavior, or unrelated infrastructure. The phase verification for every bead and for the complete phase is:
`cd linkshelf && npx playwright test`

## Bead map

### te3-0i8: linkshelf/package.json
- Scope: Define the Playwright test package metadata, the Chromium test script, and the `@playwright/test` development dependency.
- Architecture: `architecture.md` e2e-testing section; package metadata is owned by `linkshelf/package.json` and is consumed by the Playwright configuration and browser tests.
- Interfaces: Provides the `test` npm script invoking `npx playwright test --project=chromium` and the `@playwright/test` dependency.
- Depends on: `<none>` for this phase.
- Consumed by: `te3-3am` (`linkshelf/playwright.config.js`) and `te3-wmm` (`linkshelf/tests/e2e.spec.js`).
- Acceptance:
  - The file is valid JSON with package name/version metadata, a `test` script selecting the Chromium project, and `@playwright/test` version `1.62.1` under `devDependencies`.
  - The declared dependency and script support the configuration and tests without introducing a second test framework or application feature.
  - Running `cd linkshelf && npx playwright test` uses this package setup and completes the browser suite.

### te3-3am: linkshelf/playwright.config.js
- Scope: Configure Playwright discovery, Chromium execution, the base URL, and the real Go server web server.
- Architecture: `architecture.md` e2e-testing section; this file owns the Playwright project and web-server readiness configuration.
- Interfaces: Exports the Playwright configuration via `module.exports`; discovers `./tests`; uses base URL `http://127.0.0.1:8080`; starts `go run ./cmd/server`; waits for `http://127.0.0.1:8080/health`; provides the Chromium project consumed by `te3-wmm`.
- Depends on: `te3-0i8` (`linkshelf/package.json`) for the Playwright package and script.
- Consumed by: `te3-wmm` (`linkshelf/tests/e2e.spec.js`) through the Playwright runner and by phase verification.
- Acceptance:
  - `testDir` is `./tests`, the configuration defines a Chromium project using the desktop Chromium device, and `use.baseURL` is `http://127.0.0.1:8080`.
  - `webServer.command` is `go run ./cmd/server` and `webServer.url` is `http://127.0.0.1:8080/health`, so tests exercise the real server rather than a mock.
  - Local server reuse, CI retries, CI worker behavior, and the list reporter follow the architecture configuration without adding unsupported options.
  - `cd linkshelf && npx playwright test` starts or reuses the server and discovers `tests/e2e.spec.js`.

### te3-wmm: linkshelf/tests/e2e.spec.js
- Scope: Provide browser acceptance tests for UI loading, JavaScript delivery, link creation and persistence, and deletion.
- Architecture: `architecture.md` e2e-testing section plus SPEC.md functional requirements for the minimal CRUD UI.
- Interfaces: Uses Playwright `test`, `expect`, `page`, and `request`; targets `#links`, `/static/app.js`, title and URL inputs, the `Add` button, rendered link anchors/list items, and `Delete` buttons.
- Depends on: `te3-0i8` (`linkshelf/package.json`) and `te3-3am` (`linkshelf/playwright.config.js`), plus the server and UI implemented by their designated phases.
- Consumed by: phase verification and QA; no production package imports this test file.
- Acceptance:
  - The load test visits `/`, asserts `#links` is attached, requests `/static/app.js`, asserts the response is successful, and asserts its text contains `fetch`.
  - The add test fills the first two inputs with a unique title and `https://example.com/e2e-add`, clicks the `Add` button, and asserts a visible link under `#links` has the supplied title and exact href.
  - The add test reloads the page and asserts the same titled link remains visible with the same exact href, proving persistence through the UI.
  - The delete test creates a uniquely titled link at `https://example.com/e2e-delete`, locates its `#links` list item, clicks that item's Delete button, and asserts the item count becomes zero.
  - Tests use stable selectors and Playwright assertions, isolate generated titles with `Date.now()`, and do not test features absent from SPEC.md.
  - `cd linkshelf && npx playwright test` passes all three scenarios against the configured real server.

## Verification and handoff
Implement beads in dependency order: package metadata, configuration, then the browser spec. Run the exact phase verification after each relevant file and again after all three files. QA must confirm the three required paths, the concrete assertions above, and successful Playwright output before the phase is accepted.

## Future considerations
None for this phase. Backend routes, static UI implementation, and the server entrypoint remain outside this phase and are owned by their designated delivery phases.
