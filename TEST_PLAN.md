# Test Plan — Link Shelf MVP

## Purpose

This plan maps every requirement extracted from SPEC.md and architecture.md to
one or more unit, integration, or UI tests. Section headings use the exact
delivery_phases IDs from .gastown/workflow-profile.json:
backend-setup, backend-store, backend-api, frontend-ui, e2e-testing.

Source-of-truth mapping:

- SPEC.md: HTTP routes table, store API signatures, Link struct, validation rules.
- architecture.md: file ownership table, symbol signatures, e2e selectors.
- plan.md: bead IDs te-4s8 (go.mod) and te-7y3 (schema.go). Later phase beads
  do not yet exist in bd list --limit=0 and are reported as plan_gap so the Planner can
  spawn them.

The SPEC marks backend tests as optional but recommended; the e2e Playwright
suite is required. The plan assigns unit to pure logic (validation rules, SQL
via in-memory database), integration to HTTP handler wiring with httptest,
and ui only to the Playwright spec that exercises the running browser.

---

## Requirements → tests

### backend-setup
Requirement: linkshelf/go.mod declares module linkshelf on go 1.22 and pulls
github.com/mattn/go-sqlite3 v1.14.22; linkshelf/internal/store/schema.go
defines the Link struct (ID, Title, URL, CreatedAt with JSON tags) and
exports InitSchema(*sql.DB) error containing the CREATE TABLE IF NOT EXISTS
links (...) DDL string. The schema file is the only file allowed to contain
CREATE TABLE.
Level: unit
Test file: linkshelf/internal/store/schema_test.go
Bead ID: te-7y3
Phase: backend-setup
Scenarios:
- Compile the package: go build ./... succeeds.
- go.mod parses and declares module linkshelf with go 1.22.
- Link JSON tags round-trip through encoding/json to the keys id, title, url, created_at.
- InitSchema on a fresh sql.Open("sqlite3", ":memory:") creates the links table without error.
- InitSchema called twice is idempotent (no error on second call).
- After InitSchema, a SELECT against the links table returns 0 rows.
Assertions:
- strings.Contains on goMod text finds "module linkshelf" and "go 1.22".
- goMod contains the github.com/mattn/go-sqlite3 require line.
- json.Marshal of a Link value produces an object with keys id, title, url, created_at.
- InitSchema(db) returns nil; a follow-up SELECT returns sql.ErrNoRows.
- Second InitSchema(db) call returns nil (idempotent).

### backend-store
Requirement: linkshelf/internal/store/store.go exposes the package-level
variable DB *sql.DB and the functions List(ctx) ([]Link, error),
Create(ctx, title, url) (Link, error), and Delete(ctx, id) error. List returns
rows ordered by id DESC and never returns a nil slice. Create validates title
(non-empty, max 200 runes) and URL (non-empty, must start with http:// or
https://) and rejects anything else. Delete returns a non-nil error if the id
does not exist. The file must not contain any CREATE TABLE statements.
Level: unit
Test file: linkshelf/internal/store/store_test.go
Bead ID: plan_gap
Phase: backend-store
Scenarios:
- After InitSchema and store.DB = db, List on an empty DB returns an empty non-nil slice.
- Create("Example", "https://example.com") inserts a row and returns a populated Link with ID > 0 and a non-empty RFC3339 CreatedAt.
- After two Create calls, List returns them in id DESC order.
- Create with empty title rejects the input.
- Create with a 201-rune title rejects the input.
- Create with empty URL rejects the input.
- Create with ftp:// scheme rejects the input.
- Create with URL missing scheme prefix rejects the input.
- Delete on an existing id returns nil and removes the row.
- Delete on a missing id returns a non-nil error.
- File scan: store.go source does not contain the literal CREATE TABLE.
Assertions:
- len(list) == 0 and list != nil on empty DB.
- link.ID > 0, link.Title equals Example, link.URL equals https://example.com, link.CreatedAt is non-empty and parses as RFC3339.
- After two inserts, list[0].ID > list[1].ID.
- All four validation failures return a non-nil error and do not insert a row.
- Delete on an existing id returns nil; a follow-up List does not include the row.
- Delete on a missing id returns a non-nil error.
- strings.Contains(storeSource, "CREATE TABLE") is false.

### backend-api
Requirement: linkshelf/internal/api/handlers.go registers the routes in the
SPEC table on http.DefaultServeMux via RegisterRoutes(mux).
cmd/server/main.go opens linkshelf.db, calls InitSchema, assigns store.DB,
registers the routes, and calls http.ListenAndServe(":8080", nil) after
logging "listening on :8080". Static paths containing .. are rejected with 404.
Level: integration
Test file: linkshelf/internal/api/handlers_test.go
Bead ID: plan_gap
Phase: backend-api
Scenarios:
- GET / returns 200 and serves linkshelf/web/index.html.
- GET /api/links on an empty DB returns 200 and a JSON array [] (not null).
- GET /api/links after two Create calls returns 200 and a JSON array of length 2.
- POST /api/links with valid title and URL returns 201 and a JSON object containing id, title, url, created_at.
- POST /api/links with empty title returns 400 and an error JSON object.
- POST /api/links with non-http URL returns 400 and an error JSON object.
- DELETE /api/links/{id} on an existing id returns 204 and an empty body.
- DELETE /api/links/{id} on a missing id returns 404 and an error JSON object.
- GET /static/style.css returns 200 when the file exists.
- GET /static/app.js returns 200 when the file exists.
- GET /static/missing.txt returns 404.
- GET /static/../go.mod returns 404 (path traversal rejected).
- cmd/server/main.go source contains the literal string "listening on :8080" and calls http.ListenAndServe(":8080", nil).
- cmd/server/main.go source does not redefine the Link struct or re-declare CREATE TABLE.
Assertions:
- httptest.NewRecorder + http.DefaultServeMux round-trips each row of the SPEC route table to the documented status code and body shape.
- 201 response decodes into a struct with non-zero ID and a parseable CreatedAt.
- All 400/404 responses decode to a struct with a non-empty Error field.
- GET /static/../go.mod is rejected (404) even though the file exists on disk.
- main.go source scan finds "listening on :8080" and http.ListenAndServe(":8080", nil).

### frontend-ui
Requirement: linkshelf/web/index.html, linkshelf/web/app.js, and
linkshelf/web/style.css together provide a title input (#title), a URL input
(#url), an Add button, and a list with id links. On load, app.js calls
GET /api/links and renders the list; submitting the form POSTs to /api/links;
clicking a delete control calls DELETE /api/links/{id} and refreshes the list.
Level: ui
Test file: linkshelf/tests/e2e.spec.js
Bead ID: plan_gap
Phase: frontend-ui
Scenarios:
- index.html contains an input with id title, an input with id url, a button labeled Add, and a list element with id links.
- app.js source issues a GET to /api/links and renders results into the #links list.
- app.js source issues a POST to /api/links triggered by the Add button or form.
- app.js source issues a DELETE to /api/links/ triggered from list items.
- style.css is non-empty and contains at least one selector (body, button, etc.).
- When the page is served at /, the DOM exposes #title, #url, the Add button, and #links.
Assertions:
- index.html text matches id="title", id="url", id="links", and the Add button.
- app.js text references GET /api/links, POST /api/links, and DELETE /api/links.
- style.css is non-empty.
- Playwright expect(page.locator("#title")).toBeVisible() and equivalents pass against the running server.

### e2e-testing
Requirement: linkshelf/playwright.config.js, linkshelf/package.json, and
linkshelf/tests/e2e.spec.js together run npx playwright test against a Go
server started on port 8080 and verify: UI loads, add link, delete link, and
static file serving. The webServer block boots "go run ./cmd/server" with the
working directory set so linkshelf.db is created locally.
Level: ui
Test file: linkshelf/tests/e2e.spec.js
Bead ID: plan_gap
Phase: e2e-testing
Scenarios:
- playwright.config.js declares testDir ./tests, baseURL http://localhost:8080, a webServer running "go run ./cmd/server" on port 8080, and at least one browser project (chromium).
- package.json declares @playwright/test as a devDependency and a test script invoking playwright test.
- tests/e2e.spec.js contains four named tests: UI loads, add a link, delete a link, static file serving.
- Test "UI loads" navigates to / and asserts #title, #url, the Add button, and #links are visible.
- Test "add a link" types into #title and #url, clicks Add, and asserts a new list item containing the title appears in #links.
- Test "delete a link" creates a link, clicks the per-item delete control, and asserts the list item is removed from #links.
- Test "static file serving" issues page.request.get("/static/style.css") and asserts 200.
- npx playwright test --list reports the four tests without error (the QA qa_verify_command for the e2e phase).
Assertions:
- playwright.config.js parses as a module exporting an object with testDir, use.baseURL, and webServer.
- package.json has devDependencies entry for @playwright/test and scripts.test equal to "playwright test".
- All four Playwright tests are discoverable and each ends with expect(...) assertions on the documented selectors.
- tests/e2e.spec.js source contains the literal substrings for the #title, #url, #links selectors and the /static/ path.

---

## Coverage summary

- backend-setup: 1 unit (schema + go.mod compilation).
- backend-store: 1 unit (store package via in-memory sqlite).
- backend-api: 1 integration (httptest against http.DefaultServeMux plus main.go source scan).
- frontend-ui: 1 UI (Playwright DOM visibility plus source scan).
- e2e-testing: 1 UI (Playwright spec running against the live server).

Total: 5 requirement blocks across 5 delivery phases. Future-phase beads for
backend-store, backend-api, frontend-ui, and e2e-testing are reported as
plan_gap so the Planner can spawn the missing test-owner beads. The
active-phase bead te-7y3 covers the schema and go.mod test (te-4s8 is verified
by the same compile-and-parse test).
