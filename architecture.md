# Architecture for Link Shelf MVP

## Overview
Link Shelf is a minimal, single-page web application and bookmarking service written in Go. The application allows users to list, create, and delete links. It is designed to run with minimal overhead, avoiding unnecessary layers of abstraction, and relies on SQLite for persistent storage.

Key architectural characteristics:

---

## Planned file layout
All implement file paths are nested within the root layout folder `linkshelf/`. No source files should exist outside this directory:

- `linkshelf/go.mod`: Declares the Go module name (`linkshelf`), sets Go version `1.22`, and specifies the SQLite driver dependency.
- `linkshelf/cmd/server/main.go`: Application entrypoint. It opens the database, runs schema migrations, assigns the database handle, registers API and static endpoints, and listens on port `:8080`.
- `linkshelf/internal/store/schema.go`: Defines the core domain structure and handles table initialization DDL.
- `linkshelf/internal/store/store.go`: Implement package-level functions for core database queries (List, Create, Delete) and validates inputs.
- `linkshelf/internal/api/handlers.go`: HTTP handler logic, binding the store layer to JSON request/response formats.
- `linkshelf/web/index.html`: The user interface template with a simple input form for title and URL, and an unordered list to render the items.
- `linkshelf/web/app.js`: Client-side logic that calls backend APIs via AJAX to dynamically render and alter the state of bookmarks.
- `linkshelf/web/style.css`: Minimalist CSS layout styling for clean appearance.

---

## Go package / bead ownership
The Go files inside `linkshelf/internal/store` share the same Go package `store`. To prevent symbol duplication or linker conflicts across different beads (files) in the same package, we enforce strict ownership:

| File | Owns (exported) | Must not define |
| :--- | :--- | :--- |
| `linkshelf/internal/store/schema.go` | `Link` (struct)<br>`InitSchema(*sql.DB) error` (function) | `DB` (package variable)<br>`List` (function)<br>`Create` (function)<br>`Delete` (function) |
| `linkshelf/internal/store/store.go` | `DB` (package variable *sql.DB)<br>`List(context.Context) ([]Link, error)` (function)<br>`Create(context.Context, string, string) (Link, error)` (function)<br>`Delete(context.Context, int64) error` (function) | `Link` (struct)<br>`InitSchema(*sql.DB) error` (function) |

### Ownership & DDL Rules:

---

## HTTP & entrypoint integration

The web interface and JSON API are exposed via `http.DefaultServeMux`. The handler actions are registered from the main server routine.

### HTTP Route Table

| Method | Path | Success | Error | Description / Behavior |
| :--- | :--- | :--- | :--- | :--- |
| **GET** | `/` | 200 | — | Serves the single-page frontend from `linkshelf/web/index.html`. |
| **GET** | `/static/{file}` | 200 | 404 | Serves dynamic files under `linkshelf/web/` directory. Explicitly rejects paths containing `..` directory traversal characters to prevent security vulnerabilities. |
| **GET** | `/api/links` | 200 | — | Returns JSON array `[]` when empty, or `[{"id": 1, "title": "...", "url": "...", "created_at": "..."}]` sorted by ID descending (`ORDER BY id DESC`). |
| **POST** | `/api/links` | 201 | 400 | Accepts `{"title":"...","url":"..."}`. Runs validation checks. Returns 201 with the newly created JSON Link object, or 400 `{"error":"..."}` upon validation failure. |
| **DELETE** | `/api/links/{id}` | 204 | 404 | Deletes the bookmark with the given numeric `id` parsed from the route. Returns 404 `{"error":"..."}` if the `id` is not found or fails. |

### Server Wiring (`linkshelf/cmd/server/main.go`)

---

## Unit tests
No complex infrastructure is needed for test execution. Tests should be isolated using temporary in-memory SQLite instances:
  db, err := sql.Open("sqlite3", ":memory:")
  // Initialize table
  err = linkshelf/internal/store.InitSchema(db)
  // Set context DB
  linkshelf/internal/store.DB = db

---

## Integration and testing
The full integration verification command is:
cd linkshelf && go test ./...
Even if no `*_test.go` files are present, the project packages must successfully compile under standard build tests.

---

## Acceptance mapping
