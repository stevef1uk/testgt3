# Architecture for testgt3

layout_root: linkshelf

## Overview

This design follows the Link Shelf MVP SPEC literally: a small single-process Go
bookmark application that lists, creates, and deletes links. SQLite is the
persistent source of truth, the HTTP API is served by the Go process, and the
same process serves the three browser assets. The implementation stays small,
uses the exact package-level store API required by the SPEC, and adds no
second server, route prefix, migration system, or frontend framework.

A link has an integer ID, title, URL, and RFC3339 UTC creation timestamp. The
database is opened from linkshelf.db in the process working directory. Store
operations use the package variable store.DB and parameterized SQL. Schema
creation is isolated in linkshelf/internal/store/schema.go and is safe to repeat. Handlers perform JSON
decoding, request validation, status and content-type handling, and translate
store errors into the documented responses. Browser-controlled values are
inserted with DOM text nodes rather than HTML interpolation.

## Planned file layout

Every implementation file that will be created is listed below. No directory
placeholder, test wildcard, generated asset, or unlisted source file is
planned.

- `linkshelf/go.mod` — module declaration linkshelf, Go 1.22, and the
  github.com/mattn/go-sqlite3 v1.14.22 dependency.
- `linkshelf/internal/store/schema.go` — the Link type and InitSchema plus
  links-table DDL only.
- `linkshelf/internal/store/store.go` — package variable DB and List, Create,
  and Delete operations only.
- `linkshelf/internal/api/handlers.go` — API handlers, JSON errors, route
  registration, and safe static-file serving.
- `linkshelf/cmd/server/main.go` — database opening, initialization,
  dependency wiring, and ListenAndServe startup.
- `linkshelf/web/index.html` — page title, accessible title and URL inputs,
  Add button, status/error region, and links list.
- `linkshelf/web/app.js` — GET, POST, DELETE calls and DOM rendering.
- `linkshelf/web/style.css` — readable responsive layout for the page.

The concrete required files are exactly the eight literal paths above:
`linkshelf/go.mod`, `linkshelf/internal/store/schema.go`,
`linkshelf/internal/store/store.go`, `linkshelf/internal/api/handlers.go`,
`linkshelf/cmd/server/main.go`, `linkshelf/web/index.html`,
`linkshelf/web/app.js`, and `linkshelf/web/style.css`.

## Go package / bead ownership

The store model and schema belong exclusively to the store package's schema
file. Store operations are package-level functions, not methods on a Store
struct. This is important because the SPEC explicitly requires DB, List,
Create, and Delete with the signatures below.

| File | Owns (exported) | Must not define |
|---|---|---|
| `linkshelf/internal/store/schema.go` | Link (struct); InitSchema(db *sql.DB) error | A second Link type, store operations, handlers, or routes |
| `linkshelf/internal/store/store.go` | DB (*sql.DB); List(ctx context.Context) ([]Link, error); Create(ctx context.Context, title, url string) (Link, error); Delete(ctx context.Context, id int64) error | CREATE TABLE statements, InitSchema, a Store struct, NewStore, JSON handlers, or route registration |
| `linkshelf/internal/api/handlers.go` | RegisterRoutes(mux *http.ServeMux) | SQL statements, schema initialization, database opening, or process flags |
| `linkshelf/cmd/server/main.go` | main() | Store implementation, handler implementation, or duplicate route registration |

linkshelf/internal/store/schema.go defines exactly:
Link struct {
    ID int64 json:"id"
    Title string json:"title"
    URL string json:"url"
    CreatedAt string json:"created_at"
}
and InitSchema(db *sql.DB) error runs CREATE TABLE IF NOT EXISTS links with
columns for id, title, url, and created_at. The table supplies an integer
primary key and stores creation time as text. InitSchema is the only owner of
DDL and can be called on every startup.

linkshelf/internal/store/store.go declares var DB *sql.DB. List accepts a context, queries links with
ORDER BY id DESC, scans complete Link values, and returns a non-nil empty
slice when there are no rows. Create rejects an empty title, a title longer
than 200 runes, an empty URL, or a URL that does not start with http:// or
https://; it inserts the values and returns the complete newly created Link.
Delete removes the requested ID and returns an error when no row was deleted.
All database calls honor the supplied context and return database errors to
the API layer.

## HTTP + entrypoint integration

linkshelf/internal/api/handlers.go registers the following exact SPEC contract on the mux supplied by
main. The production caller supplies http.DefaultServeMux, so the application
has one routing table.

| Method | Path | Request | Success | Error behavior |
|---|---|---|---|---|
| GET | / | none | 200 and linkshelf/web/index.html | serve the page |
| GET | /static/{file} | none | 200 and a file under linkshelf/web/ | 404 for missing files |
| GET | /api/links | none | 200 JSON array, including [] when empty | 500 for store failure |
| POST | /api/links | JSON {"title":"...","url":"..."} | 201 JSON Link | 400 JSON {"error":"..."} for malformed or invalid input; 500 for store failure |
| DELETE | /api/links/{id} | none | 204 with no body | 404 JSON {"error":"..."} for unknown or invalid IDs; 500 for other store failure |

Only GET is accepted for the page, static path, and list endpoint; only POST
is accepted for creation; only DELETE is accepted for deletion. Unsupported
methods receive the appropriate method rejection rather than accidentally
running another operation. API responses set application/json where a JSON
body exists. The delete success response has status 204 and no body. Static
path handling rejects any requested path containing .. and never permits
access outside linkshelf/web. The handlers use store.List, store.Create, and
store.Delete only, never SQL directly.

The entrypoint performs the SPEC wiring in order: it parses any configured
database or listen flags while preserving the defaults linkshelf.db and :8080,
opens SQLite with the sqlite3 driver, calls store.InitSchema(db), assigns the
opened database to store.DB, calls api.RegisterRoutes(http.DefaultServeMux),
and starts http.ListenAndServe(":8080", nil) (or the configured equivalent).
Route setup includes the page, static assets, and all three API routes. Startup
logs listening on :8080. Database closure is deferred after successful open.

## Delivery phases

1. Module and persistence bead: create `linkshelf/go.mod`, then implement
   `linkshelf/internal/store/schema.go` with Link and InitSchema. Implement
   `linkshelf/internal/store/store.go` with DB, List, Create, and Delete,
   including validation, newest-first ordering, empty-slice behavior, and
   missing-ID errors. This phase must not add DDL to linkshelf/internal/store/store.go.
2. HTTP bead: implement `linkshelf/internal/api/handlers.go`. Register the
   exact /, /static/, /api/links, and /api/links/{id} behaviors, JSON shapes,
   method rejection, validation, safe static paths, and status codes while
   calling only the package-level store functions.
3. Entrypoint bead: implement `linkshelf/cmd/server/main.go`. Open SQLite,
   initialize the schema before assigning store.DB, register the handlers on
   http.DefaultServeMux, serve assets, and listen on :8080.
4. UI bead: implement `linkshelf/web/index.html`,
   `linkshelf/web/app.js`, and `linkshelf/web/style.css`. The page has
   labeled title and URL controls, an Add button, an error/status region, and
   ul id=links. JavaScript loads GET /api/links on startup, posts the two
   fields, deletes by ID, refreshes after changes, and renders an empty state.
   CSS provides a simple readable responsive layout without changing selectors
   or API behavior.
5. Verification phase: run `cd linkshelf && go mod tidy && go test ./...`,
   then run the server from linkshelf and exercise the browser workflow and
   static assets. The required concrete implementation paths remain the eight
   paths listed above.

## Unit tests

Persistence verification uses a temporary or in-memory SQLite database,
calls InitSchema before assigning store.DB, and checks empty non-nil listing,
successful creation, complete field round trips, title rune validation, URL
scheme validation, deterministic id-descending ordering, deletion, and the
error for a missing ID. Calling InitSchema twice checks restart safety. These
checks target the exported package-level functions and do not require a
Store object.

HTTP verification uses httptest requests with a real initialized database and
checks GET, POST, and DELETE, JSON fields, 400 malformed and invalid payloads,
404 unknown IDs, method rejection, content types, the [] empty response, and
the 204 empty deletion response. Asset verification reads the three concrete
web files and confirms the form controls, labels, links list, API endpoint
usage, safe text rendering, and stylesheet link. Tests use explicit paths or
test fixtures and do not depend on an incidental current working directory.

## Integration and testing

The full required command is `cd linkshelf && go mod tidy && go test ./...`.
An integration exercise assembles an ephemeral httptest server around the
registered mux and a temporary SQLite database, creates a link, lists it,
deletes it, and confirms it is absent. This verifies the order
schema-before-store-before-routes. A startup smoke check runs the assembled
server, fetches the root page and static JavaScript/CSS, then exercises each
API operation. There must be no second API prefix, duplicate schema, or
separate frontend server.

The definition of done also includes running  cd linkshelf `cd linkshelf && go run`cd linkshelf && go run go run
./cmd/server, loading the UI on :8080, adding one link, seeing it in the
list, and deleting it. The specified Playwright command npx playwright test
can then verify UI load, add/delete behavior, and static-file serving against
that running service.

## Acceptance mapping

The persistence goal is met by Link and InitSchema in linkshelf/internal/store/schema.go plus the
parameterized package-level methods in linkshelf/internal/store/store.go, with the required validation
and ordering. The API goal is met by the exact three /api operations and
their documented statuses, JSON errors, and method handling in linkshelf/internal/api/handlers.go.
The usability goal is met by the title/URL form, accessible labels, status
region, empty-list rendering, safe DOM text insertion, and responsive CSS in
the three web files. The runnable-service goal is met by linkshelf/cmd/server/main.go opening
linkshelf.db, initializing before requests, assigning store.DB, registering
http.DefaultServeMux, serving the page and static assets, logging startup,
and listening on :8080. The delivery list, ownership table, and test plan make
all SPEC goals traceable to concrete linkshelf paths.
