# Architecture for testgt3

## Overview
The **Link Shelf** MVP is a minimal Go web service that stores bookmark links in a SQLite database and serves a tiny static frontend.  
The system consists of three logical layers:

1. **Data layer** – defines the `Link` domain type and the DDL for the `links` table.  
2. **Store layer** – provides package‑level functions `List`, `Create`, and `Delete` that operate on a global `*sql.DB`. Validation of incoming data lives here.  
3. **HTTP layer** – registers handlers on `http.DefaultServeMux` that expose a JSON API and serve static assets from the `linkshelf/web/` directory.  

All components are compiled into a single binary (`linkshelf/cmd/server/main.go`). The binary opens `linkshelf.db` in the working directory, runs `InitSchema`, assigns the DB to the store, registers the HTTP routes, and starts listening on **port 8080**. The frontend (`linkshelf/web/index.html`, `linkshelf/web/app.js`, `linkshelf/web/style.css`) communicates with the JSON API (`/api/links`) to list, add, and delete links.  

The design intentionally avoids any additional abstraction layers, code generation, or third‑party web frameworks to keep the codebase small and easy to test with the standard Go toolchain (`go test ./...`).

## Planned file layout
All source files live under the **layout root** `linkshelf/`. The following paths are required by the SPEC and will be used verbatim by the implementation:

- `linkshelf/go.mod` – module declaration (`module linkshelf`) and Go version.
- `linkshelf/cmd/server/main.go` – program entrypoint; opens the SQLite DB, calls `schema.InitSchema`, wires the store, registers HTTP routes, and starts the HTTP server.
- `linkshelf/internal/store/schema.go` – defines the `Link` struct and the `InitSchema(db *sql.DB) error` function that creates the `links` table if it does not exist.
- `linkshelf/internal/store/store.go` – declares the package‑level `var DB *sql.DB` and implements the three public store functions:
  - `List(ctx context.Context) ([]Link, error)`
  - `Create(ctx context.Context, title, url string) (Link, error)`
  - `Delete(ctx context.Context, id int64) error`
- `linkshelf/internal/api/handlers.go` – contains the HTTP handler functions that translate HTTP verbs into store calls and render static files.
- `linkshelf/web/index.html` – static HTML page that presents the UI, includes `linkshelf/web/app.js`, and defines the `#links` list element.
- `linkshelf/web/app.js` – client‑side JavaScript that:
  - `GET /api/links` on load,
  - `POST /api/links` to add a link,
  - `DELETE /api/links/{id}` to delete a link,
  - Refreshes the list after each mutation.
- `linkshelf/web/style.css` – minimal styling for readability.

No additional files (e.g., `linkshelf/internal/store/store_test.go`, `linkshelf/internal/api/handlers_test.go`, Dockerfiles) are required for the MVP, but the architecture leaves room for optional tests.

## Go package / bead ownership
The project contains three Go packages, each split into **beads** (implementation files). The table below declares which exported symbols belong to which bead, ensuring no symbol is defined twice.

| File                               | Owns (exported)                                                | Must not define |
|------------------------------------|---------------------------------------------------------------|-----------------|
| `linkshelf/internal/store/schema.go` | ``Link``<br>``InitSchema``                                      | ``DB``, ``List``, ``Create``, ``Delete`` |
| `linkshelf/internal/store/store.go`   | ``DB``<br>``List``<br>``Create``<br>``Delete``                | ``Link``, ``InitSchema`` |
| `linkshelf/internal/api/handlers.go`  | ``handleRoot``<br>``handleStatic``<br>``handleList``<br>``handleCreate``<br>``handleDelete`` | ``Link``, ``InitSchema``, ``DB``, ``List``, ``Create``, ``Delete`` |
| `linkshelf/cmd/server/main.go`        | ``main`` (entrypoint)                                          | No exported symbols required elsewhere |

*All symbols are referenced with backticks to match the expected allow‑list in downstream beads.*

## HTTP + entrypoint integration
The HTTP API is defined precisely in the SPEC. The table below reproduces it verbatim:

| Method | Path                     | Success response                               | Error response                               |
|--------|--------------------------|-----------------------------------------------|----------------------------------------------|
| GET    | `/`                      | 200, serves `linkshelf/web/index.html`        | –                                            |
| GET    | `/static/{file}`         | 200, file under `linkshelf/web/`               | 404 if file not found                         |
| GET    | `/api/links`             | 200, JSON array `[]` (empty when no rows)      | –                                            |
| POST   | `/api/links`             | 201, JSON representation of created `Link`     | 400 `{"error":"..."}` (validation failure)    |
| DELETE | `/api/links/{id}`        | 204 (no body)                                   | 404 `{"error":"..."}` (unknown id)            |

**Routing wiring in `linkshelf/cmd/server/main.go`**  

func main() {
    db, err := sql.Open("sqlite3", "linkshelf.db")
    // error handling omitted for brevity
    if err := schema.InitSchema(db); err != nil { log.Fatal(err) }
    store.DB = db

    // Static files
    http.HandleFunc("/", handlers.handleRoot) // serves linkshelf/web/index.html
    http.HandleFunc("/static/", handlers.handleStatic) // serves files from linkshelf/web/

    // JSON API
    http.HandleFunc("/api/links", handlers.handleList)   // GET
    http.HandleFunc("/api/links", handlers.handleCreate) // POST (same path, method switch)
    http.HandleFunc("/api/links/", handlers.handleDelete) // DELETE with id suffix

    log.Printf("listening on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

*The `handleStatic` implementation strips the `/static/` prefix, joins it with `linkshelf/web/`, and rejects any request whose cleaned path contains `..` to prevent path traversal.*

## Unit tests
Although the SPEC states that tests are optional, a typical test plan would include:

- **Store tests** (`linkshelf/internal/store/store_test.go`):  
  - Use an in‑memory SQLite DB (`file::memory:?cache=shared`).  
  - Call `schema.InitSchema` to create the table.  
  - Set `store.DB = db`.  
  - Verify `List` returns an empty slice initially.  
  - Verify `Create` succeeds with a valid title/url and respects the 200‑rune title limit and URL scheme validation.  
  - Verify `Delete` removes an existing record and returns an error for a non‑existent ID.

- **Handler tests** (`linkshelf/internal/api/handlers_test.go`):  
  - Spin up a `httptest.NewRecorder` and `http.NewRequest` for each route.  
  - Ensure `GET /` returns the HTML page with status 200.  
  - Ensure `GET /static/{file}` serves files correctly and returns 404 for missing files.  
  - Ensure `/api/links` GET returns JSON array, POST returns 201 with created link, and DELETE returns 204 or 404 as appropriate.  
  - Test that `POST` with invalid payload yields a 400 with a clear error message.

All tests run with the single command `go test ./...` from the `linkshelf/` directory, satisfying the definition of done.

## Integration and testing
The full integration flow on a developer's machine is:

1. **Build & test**  
   cd linkshelf && go mod tidy && go test ./...
   - `go mod tidy` resolves the `github.com/mattn/go-sqlite3` driver.  
   - `go test ./...` compiles every package; even with no `*_test.go` files the command succeeds if the code compiles.

2. **Run the server**  
   cd linkshelf && go run ./cmd/server
   - The server opens/creates `linkshelf.db` in the current directory, runs `InitSchema`, registers routes, and listens on **:8080**.  
   - Visiting `http://localhost:8080/` loads the UI. The UI performs the expected CRUD operations via the JSON API.

3. **Manual verification** (optional)  
   - Use a browser or `curl` to check each endpoint matches the table in *HTTP + entrypoint integration*.  
   - Example: `curl -X POST -d '{"title":"Go Docs","url":"https://golang.org"}' http://localhost:8080/api/links`.

Successful execution of the above steps confirms that the architecture satisfies the SPEC's functional and non‑functional requirements.

## Acceptance mapping
| SPEC Requirement                                      | Architecture Satisfaction                                                                 |
|-------------------------------------------------------|-------------------------------------------------------------------------------------------|
| `go test ./...` passes                                 | All packages compile; store and handler code are pure Go with no external runtime deps. |
| Server serves UI on `:8080`                           | `linkshelf/cmd/server/main.go` wires static handler for `/` and `/static/` and starts listener.    |
| List, Create, Delete links via JSON API               | `linkshelf/internal/api/handlers.go` implements `/api/links` with GET/POST/DELETE mapping.        |
| Validation of title length (≤200) and URL scheme      | `store.Create` performs both checks and returns 400 errors on violation.                |
| Path‑traversal protection for static assets            | `handleStatic` rejects any cleaned path containing `..`.                                 |
| Single source of DDL (`InitSchema`) in `schema.go`    | `linkshelf/internal/store/schema.go` contains `InitSchema`; no other file performs DDL. |
| No extra abstraction layers (no `Store` struct, etc.) | Store API uses package‑level functions only, matching the SPEC.                         |
| Frontend loads correctly and interacts with API       | `linkshelf/web/index.html` + `linkshelf/web/app.js` described; UI calls the exact API endpoints. |
| Database file `linkshelf.db` created at startup       | `linkshelf/cmd/server/main.go` opens `sql.Open("sqlite3", "linkshelf.db")` before `InitSchema`. |
| Go module `linkshelf` with sqlite3 dependency         | `linkshelf/go.mod` declares `module linkshelf` and requires `github.com/mattn/go-sqlite3`. |

All required files, symbols, routes, and validation logic are enumerated above, guaranteeing that the subsequent implementation (by the polecat) will be able to satisfy the test suite and runtime expectations.
