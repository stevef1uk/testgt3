# Architecture for testgt3

## Overview

This design implements the Link Shelf MVP from the SPEC as a small Go
application with SQLite persistence, a package-level store API, HTTP JSON
handlers, and a plain browser interface. The layout root is linkshelf. The
application supports exactly the required operations: list links, create a
validated link, and delete a link. It avoids extra abstractions and follows
the exact exported names and function shapes in the SPEC.

The server opens linkshelf.db in its current working directory, initializes the
database schema before registering request handling, assigns the connection to
the store package variable, registers routes on http.DefaultServeMux, and
starts on port 8080. API requests are kept under /api/links. The root route
serves the HTML document, and static requests use the SPEC's
/static/{file} route while restricting files to the web directory and rejecting
any path containing ... The frontend uses only the JSON API and refreshes its
list after creation and deletion.

## Planned file layout

These are all implementation files planned by the SPEC. No directory
placeholder, test file, or additional production source file is required.

- `linkshelf/go.mod` — declares module linkshelf, Go 1.22, and the
  github.com/mattn/go-sqlite3 v1.14.22 dependency.
- `linkshelf/internal/store/schema.go` — defines Link and InitSchema and owns
  all links-table DDL.
- `linkshelf/internal/store/store.go` — defines DB and the package-level List,
  Create, and Delete persistence functions.
- `linkshelf/internal/api/handlers.go` — defines JSON API handlers and the
  root/static route registration used by main.
- `linkshelf/cmd/server/main.go` — opens SQLite, initializes schema, assigns
  store.DB, registers handlers, logs startup, and starts the server.
- `linkshelf/web/index.html` — contains the title input, URL input, Add button,
  and the required ul element with id links.
- `linkshelf/web/app.js` — loads, renders, creates, deletes, and refreshes
  links through the API.
- `linkshelf/web/style.css` — supplies a simple readable layout without a
  framework.

The static handler maps only the file portion of /static/{file} into
linkshelf/web. It must reject traversal input containing ... It must not serve
API paths as assets. The root handler serves linkshelf/web/index.html, while
the browser references its JavaScript and stylesheet through the static route
shape specified by the SPEC.

## Delivery phases

1. Module phase: create `linkshelf/go.mod` with module linkshelf, Go 1.22,
   and the required sqlite3 dependency. The module must support the commands
   go mod tidy and go test ./... from the linkshelf directory.
2. Schema phase: create `linkshelf/internal/store/schema.go`. Define the
   exact Link fields and JSON tags from the SPEC. Define
   InitSchema(db *sql.DB) error and put the CREATE TABLE IF NOT EXISTS links
   statement in this file only.
3. Store phase: create `linkshelf/internal/store/store.go`. Define
   var DB *sql.DB and the exact functions
   List(ctx context.Context) ([]Link, error),
   Create(ctx context.Context, title, url string) (Link, error), and
   Delete(ctx context.Context, id int64) error. List must use ORDER BY id DESC
   and return [] rather than a nil slice when there are no rows. Create must
   validate title and URL exactly as required. Delete must return an error if
   no matching id exists.
4. API phase: create `linkshelf/internal/api/handlers.go`. Implement the
   exact GET, POST, and DELETE endpoints, JSON response bodies, status codes,
   malformed-input handling, missing-id handling, and safe static behavior.
   Handlers call store.List, store.Create, and store.Delete only.
5. Entrypoint phase: create `linkshelf/cmd/server/main.go`. Open
   linkshelf.db in the current working directory, call store.InitSchema(db),
   assign store.DB = db, register routes on http.DefaultServeMux, log
   listening on :8080, and call http.ListenAndServe(":8080", nil).
6. Frontend phase: create `linkshelf/web/index.html`,
   `linkshelf/web/app.js`, and `linkshelf/web/style.css`. The HTML supplies
   the required controls and list. The script performs initial GET, POST on
   Add, DELETE on removal, and a fresh GET after each change. The stylesheet
   provides readable basic presentation.
7. Verification phase: from linkshelf run go mod tidy and go test ./...;
   then run go run ./cmd/server and exercise the UI list, add, and delete
   flow on port 8080.

## Go package / bead ownership

The store package deliberately has no Store struct and no NewStore function.
The model and schema belong exclusively to schema.go. Persistence uses the
package variable required by the SPEC.

| File | Owns (exported) | Must not define |
|---|---|---|
| `linkshelf/internal/store/schema.go` | `Link` (struct); `InitSchema(db *sql.DB) error` | List, Create, Delete, HTTP logic, or duplicate schema statements |
| `linkshelf/internal/store/store.go` | `DB *sql.DB`; `List(ctx context.Context) ([]Link, error)`; `Create(ctx context.Context, title, url string) (Link, error)`; `Delete(ctx context.Context, id int64) error` | Link, CREATE TABLE, InitSchema, route registration, or response writing |
| `linkshelf/internal/api/handlers.go` | HTTP handler functions and route registration helpers | SQL statements, database opening, Store/NewStore types, or schema logic |
| `linkshelf/cmd/server/main.go` | `main()` | Store queries, JSON request decoding, or duplicated handler behavior |

Create rejects an empty title, a title longer than 200 runes, an empty URL,
and any URL that does not begin with http:// or https://. It inserts the UTC
creation time in RFC3339 form and returns the inserted Link. Store operations
use context-aware database calls and propagate failures to their callers.

## HTTP + entrypoint integration

The exact HTTP contract is:

| Method | Path | Success | Error |
|---|---|---|---|
| GET | / | 200 and linkshelf/web/index.html | not applicable |
| GET | /static/{file} | 200 and a file beneath linkshelf/web | 404; reject any static path containing .. |
| GET | /api/links | 200 and a JSON array, including [] when empty | JSON server error if listing fails |
| POST | /api/links | 201 and the created Link as JSON | 400 with {"error":"..."} |
| DELETE | /api/links/{id} | 204 with no body | 404 with {"error":"..."} |

The POST body is exactly an object containing title and url strings. The
handler decodes it, rejects malformed JSON and validation failures with the
required JSON error shape, invokes store.Create, and encodes the returned
Link with status 201. GET invokes store.List and sets a JSON content type
before encoding the result. DELETE parses the id path, invokes store.Delete,
returns 204 on success, and maps an invalid or missing record to 404.

main opens the SQLite database file linkshelf.db, calls
store.InitSchema(db), then assigns store.DB = db. It registers the API and
static routes on http.DefaultServeMux, logs the literal startup message
listening on :8080, and starts http.ListenAndServe(":8080", nil). The same
database connection is used for all requests. No handler creates a second
connection or performs schema initialization.

## Unit tests

The SPEC does not require test files, so the required implementation list
contains no test source. If optional tests are added, they must use concrete
files under linkshelf and must initialize a temporary or :memory: SQLite
database with store.InitSchema before assigning store.DB.

Store tests should cover idempotent schema setup, a non-nil empty result,
descending id order, successful creation, title rune limits, URL scheme
validation, successful deletion, and an error for a missing id. Handler tests
may use httptest to cover the exact API paths, 201 creation, 204 deletion,
400 validation errors, 404 deletion errors, JSON content types, empty arrays,
root serving, and traversal rejection.

## Integration and testing

The pipeline check is run from linkshelf:

    go mod tidy
    go test ./...

The runtime check is run from linkshelf with:

    go run ./cmd/server

Verify that startup logs listening on :8080, the root returns the HTML
document, an allowed static asset can be fetched using the /static/{file}
route shape, and GET /api/links returns [] on a fresh database. Use the page
to submit a valid title and HTTP or HTTPS URL, confirm the link appears, then
remove it and confirm the refreshed list is empty. Also verify invalid title
and URL submissions display the server's error without adding a row.

## Docker & Deployment

The SPEC does not request a Dockerfile or compose configuration, so no Docker
artifact is planned. Deployment is a normal Go process with the web assets
present at linkshelf/web and write access for linkshelf.db in the process's
current working directory. The service listens on port 8080 and serves both
the API and browser resources from one process.

## Acceptance mapping

The planned module and exact file set satisfy the requested linkshelf layout.
schema.go owns Link and DDL, while store.go exposes only the required
package-level DB, List, Create, and Delete API. SQLite initialization occurs
before requests, and list ordering, empty-array behavior, validation, and
missing-delete errors match the SPEC.

The API table maps directly to the required root, static, list, create, and
delete behavior. Static traversal is rejected and API paths cannot fall
through to file serving. The frontend includes the required inputs, Add
button, links list, loading, rendering, POST, DELETE, and refresh operations.
The specified tidy, test, and server commands provide the pipeline and
runtime definition of done.
