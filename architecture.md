# Architecture for testgt3

layout_root: linkshelf

## Overview

This architecture implements the Link Shelf SPEC as a small, single-process Go
bookmark application. It provides the three required operations: list saved
links, create a link from a title and URL, and delete a link by ID. SQLite is
the persistent source of truth, and the same HTTP process serves both the API
and the browser UI. There is no separate frontend server, authentication
layer, migration framework, or additional API prefix.

The database file is linkshelf.db in the current working directory, exactly as
required by the SPEC. The store uses a package-level DB variable and
parameterized SQL. Schema creation is repeat-safe. The browser assets are
served from linkshelf/web and use the exact /api/links endpoints. User values
are rendered as text nodes rather than interpolated as HTML.

## Planned file layout

Every implementation file is enumerated here; no directory placeholder or
unlisted source file is planned.

- `linkshelf/go.mod` — Go module declaration and the SQLite dependency.
- `linkshelf/internal/store/schema.go` — Link model and repeat-safe schema DDL.
- `linkshelf/internal/store/store.go` — package DB variable and package-level
  List, Create, and Delete persistence functions.
- `linkshelf/internal/api/handlers.go` — API handlers, route registration,
  validation/error responses, and static asset serving.
- `linkshelf/cmd/server/main.go` — fixed database opening, initialization,
  default mux wiring, and server startup.
- `linkshelf/web/index.html` — accessible title/URL form, status region, and
  links-list container.
- `linkshelf/web/app.js` — list, create, delete requests and safe DOM updates.
- `linkshelf/web/style.css` — readable, responsive presentation.

The required literal implementation paths are exactly:
`linkshelf/go.mod`, `linkshelf/internal/store/schema.go`,
`linkshelf/internal/store/store.go`, `linkshelf/internal/api/handlers.go`,
`linkshelf/cmd/server/main.go`, `linkshelf/web/index.html`,
`linkshelf/web/app.js`, and `linkshelf/web/style.css`.

## Go package / bead ownership

The store model and DDL are owned by the schema file, while persistence
queries are owned by the store file. API code depends on store but never
opens the database or issues SQL. The entrypoint only performs process
wiring.

| File | Owns (exported) | Must not define |
|---|---|---|
| `linkshelf/internal/store/schema.go` | `Link` (struct); `InitSchema(db *sql.DB) error` | A second Link type, query functions, handlers, routes, or startup logic |
| `linkshelf/internal/store/store.go` | `DB *sql.DB`; `List(ctx context.Context) ([]Link, error)`; `Create(ctx context.Context, title string, url string) (Link, error)`; `Delete(ctx context.Context, id int64) error` | DDL, `InitSchema`, a Store struct, database opening, JSON, or route registration |
| `linkshelf/internal/api/handlers.go` | `RegisterRoutes(mux *http.ServeMux)` | SQL statements, schema initialization, database opening, CLI flags, or a second route prefix |
| `linkshelf/cmd/server/main.go` | `main()` | Store implementation, handler implementation, or duplicate route registration |

`Link` has ID int64, Title string, URL string, and CreatedAt string fields
with JSON names id, title, url, and created_at. `InitSchema(db *sql.DB) error`
creates the links table only when absent, with an integer primary key, title,
url, and creation timestamp columns. It returns database errors and is safe
to invoke on every startup.

`DB` is assigned by the entrypoint after schema initialization. The store
functions use context-aware parameterized statements. `List(ctx context.Context)
([]Link, error)` returns a non-nil empty slice when there are no rows and
orders by id descending. `Create(ctx context.Context, title string, url string)
(Link, error)` validates a non-empty title of at most 100 runes and a URL with
an accepted http or https scheme, inserts the current RFC3339 UTC timestamp,
and returns the complete inserted Link. `Delete(ctx context.Context, id int64)
error` deletes by ID and returns a not-found error when no row was affected.
These functions do not duplicate schema ownership.

## HTTP + entrypoint integration

The API contract follows the SPEC exactly:

| Method and path | Request | Success response | Errors |
|---|---|---|---|
| GET /api/links | none | 200 JSON array of Link objects, newest ID first | 500 JSON error on storage failure |
| POST /api/links | JSON object with title and url | 201 JSON Link object | 400 for malformed JSON, validation failure, or unsupported fields; 500 for storage failure |
| DELETE /api/links/{id} | numeric path ID | 204 with an empty body | 400 for an invalid ID; 404 for an unknown link; 500 for storage failure |

Only GET, POST, and DELETE are accepted for the corresponding API route;
unsupported methods receive 405. JSON responses set application/json where
they have a body. The delete success response has no body. Request bodies are
decoded once, reject malformed JSON and unexpected fields, and are closed.
The handlers use store.List, store.Create, and store.Delete and map validation
and missing-record errors to the statuses above.

RegisterRoutes(mux *http.ServeMux) installs the exact /api/links route and
its DELETE-by-ID handling. It also registers the root page and static asset
paths so linkshelf/web/index.html, linkshelf/web/app.js, and
linkshelf/web/style.css are available from the same process. Static serving
must prevent traversal and must use the concrete web directory; it must not
create another API prefix or another server.

The entrypoint wiring is fixed and intentionally has no configurable
database or listen flags. `main()` opens linkshelf.db in the current working
directory, calls store.InitSchema(db), assigns store.DB = db, calls
api.RegisterRoutes(http.DefaultServeMux), and then calls
http.ListenAndServe(":8080", nil) exactly. It logs fatal errors from opening
the database, schema initialization, or server startup, and closes the
database on shutdown return. The initialization order is database open,
schema initialization, package DB assignment, route registration, then server
start. The server therefore serves the UI and API on port 8080 from one
process.

The full Go verification command is `cd linkshelf && go test ./...`. A
build-only phase may use  cd linkshelf `cd linkshelf && go build`cd linkshelf && go build go build ./...`; neither command
requires a second working directory or service.

## Delivery phases

| Bead phase | Required paths | Responsibility and dependency |
|---|---|---|
| module-and-persistence-bead | `linkshelf/go.mod`, `linkshelf/internal/store/schema.go`, `linkshelf/internal/store/store.go` | Establish the Go module, Link schema, package DB, validation, and CRUD queries. It precedes all HTTP work. |
| http-bead | `linkshelf/internal/api/handlers.go` | Depend on persistence, implement the exact /api/links contract and same-process static serving. |
| entrypoint-bead | `linkshelf/cmd/server/main.go` | Depend on HTTP and persistence; perform fixed linkshelf.db initialization and exact default-mux ListenAndServe wiring. |
| ui-bead | `linkshelf/web/index.html`, `linkshelf/web/app.js`, `linkshelf/web/style.css` | Depend on the served routes and implement the browser add/list/delete workflow. |

## Unit tests

Persistence tests should use an isolated SQLite database, call
InitSchema before assigning store.DB, and cover restart-safe schema creation,
a non-nil empty list, complete field round trips, title rune-length
validation, URL scheme validation, deterministic ID-descending ordering,
successful deletion, and the not-found deletion error. HTTP tests should use
httptest with an initialized database and cover GET, POST, and DELETE,
malformed and invalid payloads, unknown IDs, method rejection, JSON content
types, the empty array, and the empty 204 deletion response. Browser asset
checks should confirm accessible form labels, the list container, API calls,
safe text rendering, and the stylesheet link. Tests must not rely on a
persistent linkshelf.db or an incidental current working directory.

## Integration and testing

An integration exercise constructs an ephemeral HTTP server with the registered
default mux and a temporary SQLite database, then creates, lists, and deletes
a link while checking that it disappears. This verifies schema-before-store-
before-routes wiring and catches route drift. A startup smoke test runs the
actual entrypoint from the application directory, fetches the root page and
both static assets, and performs every API operation. The UI acceptance
workflow loads port 8080, submits one title and URL, observes the new link,
and deletes it. The complete automated suite remains `cd linkshelf && go test
./...`; the implementation must also build with  cd linkshelf `cd linkshelf && go build`cd linkshelf && go build go build
./...`.

## Acceptance mapping

The persistence goal is satisfied by the Link type and InitSchema in
linkshelf/internal/store/schema.go and the package-level CRUD API in
linkshelf/internal/store/store.go, including validation, timestamps, ordering,
and safe deletion. The HTTP goal is satisfied by the exact GET, POST, and
DELETE /api/links operations in linkshelf/internal/api/handlers.go with the
specified statuses and JSON behavior. The usability goal is satisfied by the
accessible form and list in linkshelf/web/index.html, fetch and safe DOM
logic in linkshelf/web/app.js, and responsive styling in
linkshelf/web/style.css. The runnable-service goal is satisfied by
linkshelf/cmd/server/main.go opening linkshelf.db in the current working
directory, initializing it before assigning store.DB, registering
http.DefaultServeMux, and calling http.ListenAndServe(":8080", nil) exactly.
All deliverables remain under the linkshelf layout root and are covered by
the four delivery phases and the full-suite command.
