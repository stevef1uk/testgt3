# Implementation plan

## Phase
Active phase: `e2e-testing`. Implement only the three required paths below. The exact phase verification for every bead is `cd linkshelf && npx playwright test`. Backend, server, and UI production files belong to other phases.

## Bead map

### te3-0i8: linkshelf/package.json
- Scope: Define the Playwright package metadata, Chromium test script, and `@playwright/test` development dependency.
- Architecture: `architecture.md` e2e-testing requirement and E2E integration section; this file owns E2E dependencies only.
- Interfaces: Provides the package dependency and npm test script consumed by `linkshelf/playwright.config.js` and `linkshelf/tests/e2e.spec.js`.
- Depends on: `<none>` within this phase; it relies on the application delivered by earlier phases.
- Consumed by: `te3-3am` and `te3-wmm`, plus the phase runner.
- Acceptance:
  - `linkshelf/package.json` is valid JSON, names the Playwright test package, and declares `@playwright/test`.
  - Its test script invokes Playwright for the Chromium project and introduces no backend or UI feature.
  - Verify: `cd linkshelf && npx playwright test`.

### te3-3am: linkshelf/playwright.config.js
- Scope: Configure browser discovery, Chromium, server startup/readiness, base URL, retries, reporting, and local server reuse.
- Architecture: `architecture.md` E2E / integration testing section; target `http://127.0.0.1:8080`, discover `./tests`, and start the real server with `go run ./cmd/server`.
- Interfaces: Exports the Playwright configuration; readiness endpoint is `/health`, base URL is `http://127.0.0.1:8080`, and the web-server command is `go run ./cmd/server`.
- Depends on: `te3-0i8` (`linkshelf/package.json`) for the Playwright runner; the server and UI are delivered by earlier phases.
- Consumed by: `te3-wmm` through the Playwright runner.
- Acceptance:
  - Configuration discovers `linkshelf/tests`, defines a Chromium project, and uses the specified base URL.
  - `webServer` starts `go run ./cmd/server`, waits for `http://127.0.0.1:8080/health`, and uses the architecture's CI retry/worker behavior.
  - Verification exercises the real application rather than a mock.
  - Verify: `cd linkshelf && npx playwright test`.

### te3-wmm: linkshelf/tests/e2e.spec.js
- Scope: Provide browser assertions for UI loading, static asset serving, link creation/persistence, and deletion.
- Architecture: `architecture.md` E2E / integration testing section and SPEC.md frontend and definition-of-done requirements.
- Interfaces: Uses Playwright `page` and `request`; targets `/`, `/static/app.js`, `#links`, the title and URL inputs, Add, rendered link anchors, list items, and Delete controls.
- Depends on: `te3-0i8` and `te3-3am`, plus the server, API, and frontend files delivered by prerequisite phases.
- Consumed by: the phase verification command and QA review; no production package imports this test.
- Acceptance:
  - Load coverage visits `/`, asserts `#links` is attached, requests `/static/app.js`, asserts success, and confirms the asset contains `fetch`.
  - Add coverage fills title and URL, clicks Add, asserts the created anchor is visible with the supplied title and href, reloads, and asserts persistence.
  - Delete coverage creates a uniquely named link, locates its list item, clicks Delete, and asserts that item is removed.
  - Tests use stable selectors, Playwright assertions, and unique data to avoid collisions.
  - Verify: `cd linkshelf && npx playwright test`.

## Verification and boundaries
The active phase has no `linkshelf/cmd/.../main.go` required file, so an entrypoint integration contract is not applicable. The polecat must implement only the three mapped paths and run the exact phase verification after implementation. Acceptance is limited to SPEC behavior: loading the UI, serving the JavaScript asset, adding and persisting a link, and deleting a link. No extra routes, configuration features, or application behavior should be introduced.

## Future considerations
Server wiring, backend routes, and frontend production assets remain owned by their designated delivery phases and are not added to these beads.
