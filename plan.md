# Implementation plan

## Phase
Active phase: `e2e-testing`. Implement only the three required Playwright artifacts. Phase verification:

```sh
cd linkshelf && npx playwright test
```

## Bead map

### te3-fq3: linkshelf/package.json
- Scope: Define the Playwright E2E package manifest and Chromium test command.
- Architecture: `linkshelf/package.json` owns the E2E dependency and npm script.
- Interfaces: Provides pinned `@playwright/test` version `1.62.1` and script `npx playwright test --project=chromium`.
- Depends on: `<none>`.
- Consumed by: `te3-4ph` and `te3-aio` through the Playwright runner.
- Acceptance:
  - Valid JSON declares `@playwright/test` in `devDependencies`, pinned to `1.62.1`.
  - The npm `test` script runs `npx playwright test --project=chromium`.
  - The manifest supports `cd linkshelf && npx playwright test`.
- Verify: `cd linkshelf && npx playwright test`

### te3-4ph: linkshelf/playwright.config.js
- Scope: Configure test discovery, Chromium, application base URL, and server readiness.
- Architecture: `linkshelf/playwright.config.js` owns Playwright browser orchestration for the server on port 8080.
- Interfaces: Exports the CommonJS Playwright configuration with `testDir: './tests'`, base URL `http://127.0.0.1:8080`, Chromium project, and web server command `go run ./cmd/server`.
- Depends on: `te3-fq3` (`linkshelf/package.json`).
- Consumed by: `te3-aio` (`linkshelf/tests/e2e.spec.js`).
- Acceptance:
  - `testDir` is exactly `./tests`.
  - `use.baseURL` is exactly `http://127.0.0.1:8080`.
  - A project named `chromium` uses Desktop Chrome.
  - `webServer.command` is exactly `go run ./cmd/server`.
  - Web-server readiness is checked at `http://127.0.0.1:8080/health`.
  - The verification command starts or reuses the server and runs the Chromium suite.
- Verify: `cd linkshelf && npx playwright test`

### te3-aio: linkshelf/tests/e2e.spec.js
- Scope: Verify the user-visible Link Shelf UI, static asset serving, creation, persistence, and deletion.
- Architecture: Tests use `/`, the `#links` list, title and URL inputs, the Add button, rendered link items, and Delete buttons.
- Interfaces: Uses Playwright `page`, `request`, `page.goto('/')`, and the SPEC static route `/static/{file}` with `app.js` as the requested file.
- Depends on: `te3-fq3` (`linkshelf/package.json`) and `te3-4ph` (`linkshelf/playwright.config.js`).
- Consumed by: Phase verification `cd linkshelf && npx playwright test`.
- Acceptance:
  - The home-page test loads `/` and asserts that `#links` is attached.
  - The static-file test performs `GET /static/{file}` for `app.js`, asserts success, and asserts the response body contains `fetch`.
  - The add test uses unique data, fills the title and URL inputs, clicks Add, asserts the new link appears in `#links` with the expected URL, reloads, and asserts it remains visible with that URL.
  - The delete test uses independent unique data, creates a link, asserts its list item is visible, clicks its Delete button, and asserts that the item is removed.
- Verify: `cd linkshelf && npx playwright test`

## Completion criteria
All three required files are implemented according to SPEC.md and architecture.md, with no backend or later-phase scope added. The exact phase verification command passes.
