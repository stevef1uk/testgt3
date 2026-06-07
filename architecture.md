# Architecture for testgt3

## Overview
The **Link Shelf MVP** is a lightweight web bookmarking application designed in Go. It operates a simple React-free single-page static frontend that interacts with a SQLite-backed backend over a clean JSON REST API. 

The purpose of this architecture document is to specify the design, directory structures, package dependencies, interface boundaries, and data model to ensure a flawless implementation by the polecat.

The key design directives and constraints are:
1. All application files must live under the root directory `linkshelf/`.
2. All persistence relies on a single SQLite file named `linkshelf.db` located in the current working directory of the process.
3. Database schemas and initialization functions must live solely in `linkshelf/internal/store/schema.go`. No double `CREATE TABLE` execution or duplicated DDL inside the server entrypoint or multiple tests.
4. The database connection is registered in a global, package-level variable `store.DB` (an instance of `*sql.DB`) defined in `linkshelf/internal/store/store.go`. No struct-based or receiver-based store instantiation will be used.
5. Strict validation is enforced when creating links: the title must be non-empty and hold a maximum of 200 runes, and the URL must be non-empty and start with `http://` or `https://`.
6. Safe file serving is enforced on static asset handling: paths containing `..` must be rejected immediately to prevent directory traversal.

---

## Planned file layout
Every single implementable path lives under the layout root `linkshelf/` and is prefixed as such:

* **`linkshelf/go.mod`**  
  Declares the module `linkshelf`, targeting Go version `1.22` or later, and requiring the SQLite driver dependency `github.com/mattn/go-sqlite3`.

* **`linkshelf/internal/store/schema.go`**  
  Houses the data model `Link` struct definition, table DDL, and the initialization function `InitSchema`. This file is the single source of truth for schema creation.

* **`linkshelf/internal/store/store.go`**  
  Exposes the package-level global database variable `DB *sql.DB` and implements the CRUD functions `List`, `Create`, and `Delete` operating on that database.

* **`linkshelf/internal/api/handlers.go`**  
  Implements the HTTP layer. Registers handlers on `http.DefaultServeMux` for API endpoints (`/api/links`, `/api/links/{id}`) and the static asset directories (`/`, `/static/{file}`).

* **`linkshelf/cmd/server/main.go`**  
  The main executable. Opens the SQLite database file `linkshelf.db`, initializes the schema, registers handlers, and starts the listener.

* **`linkshelf/web/index.html`**  
  The entry-point HTML page for the frontend.

* **`linkshelf/web/app.js`**  
  The client-side JavaScript file that interacts with the backend REST endpoints to fetch, add, and delete links.

* **`linkshelf/web/style.css`**  
  Minimal CSS stylesheet for the application's single-page layout.

---

## Go package / bead ownership
When several implementation paths live in the same Go package directory, symbol ownership per file must be documented clearly so that implementation units do not duplicate types or conflict during compilation.

For the `store` package:
- **`linkshelf/internal/store/schema.go`** owns the core domain structure and table initialization.
- **`linkshelf/internal/store/store.go`** owns the stateful SQL connection and CRUD query execution.

Below is the symbol ownership map:

| File | Owns (exported) | Must not define |
| :--- | :--- | :--- |
| **`linkshelf/internal/store/schema.go`** | `type Link struct`, `func InitSchema(db *sql.DB) error` | `var DB`, `func List`, `func Create`, `func Delete` |
| **`linkshelf/internal/store/store.go`** | `var DB *sql.DB`, `func List(ctx context.Context) ([]Link, error)`, `func Create(ctx context.Context, title, url string) (Link, error)`, `func Delete(ctx context.Context, id int64) error` | `type Link`, `func InitSchema` |

---

## HTTP + entrypoint integration

### HTTP API Table

The HTTP layer registers handlers on `http.DefaultServeMux`. These handle request routing and path safety parameters.

| Method | Path | Success | Error |
| :--- | :--- | :--- | :--- |
| **GET** | `/` | 200, serves file `linkshelf/web/index.html` | — |
| **GET** | `/static/{file}` | 200, serves file under `linkshelf/web/` | 404 Not Found (or 400 Bad Request on traversal) |
| **GET** | `/api/links` | 200, returns JSON array `[]` when empty, or array of JSON links | — |
| **POST** | `/api/links` | 201, returns JSON created link | 400 Bad Request with JSON body `{"error":"..."}` |
| **DELETE** | `/api/links/{id}` | 204 No Content | 404 Not Found with JSON body `{"error":"..."}` (if ID missing or DB deletion fails) |

### Path safety and static handler rules:
1. **No directory traversal**: The handler for `/static/{file}` must explicitly check if the requested file path (or sub-route segment) contains `..`. If any traversal indicator (`..`) is found, the handler must immediately return a `400 Bad Request` or reject the query.
2. **Path extraction**: In `linkshelf/internal/api/handlers.go`, the `{file}` segment and `{id}` segment must be extracted properly. In Go 1.22+, `http.DefaultServeMux` supports wildcards in patterns (e.g. `/static/{file}` and `/api/links/{id}`). If those standard features or manual path trimming are used, they must conform strictly to standard library standards.

### Integration sequence inside the server entrypoint (`linkshelf/cmd/server/main.go`):
1. **Initialize DB Connection**: Open the SQLite driver connection using the filename `linkshelf.db`.
2. **Schema Migration**: Call `store.InitSchema(db)` directly from `main.go`. This guarantees the DB tables are created before the app accepts incoming connections.
3. **Set Global DB**: Bind the initialized connection to `store.DB = db`.
4. **Register Handlers**: Execute the registration sequence in `linkshelf/internal/api/handlers.go` which links paths (`/`, `/static/{file}`, `/api/links`, `/api/links/{id}`) to `http.DefaultServeMux`.
5. **Listen**: Invoke `http.ListenAndServe(":8080", nil)` and print or log `"listening on :8080"`.

---

## Data model and database schema details

The SQL schema mapping resides exclusively in `linkshelf/internal/store/schema.go`.

### SQL Definition:
sql
CREATE TABLE IF NOT EXISTS links (
    id INTEGER PRIMARY KEY AUTO_INCREMENT, -- SQLite automatically handles AUTOINCREMENT for INTEGER PRIMARY KEY
    title TEXT NOT NULL,
    url TEXT NOT NULL,
    created_at TEXT NOT NULL
);

### Go Data Type:
package store

type Link struct {
    ID        int64  `json:"id"`
    Title     string `json:"title"`
    URL       string `json:"url"`
    CreatedAt string `json:"created_at"` // RFC3339 UTC representation
}

---

## Validation rules

The `Create` store operation performs strict validation before inserting any row into the database:
- **Title Validation**:
  - Must not be empty.
  - Must be a maximum of 200 runes (checked via `utf8.RuneCountInString` to avoid byte/rune misalignment).
- **URL Validation**:
  - Must not be empty.
  - Must prefix-match `http://` or `https://`.

If validation fails, `Create` returns an error. The HTTP handler wrapper in `linkshelf/internal/api/handlers.go` intercepts this validation error and transforms it into a `400 Bad Request` status code with JSON payload `{"error": "..."}` matching the specification.

---

## Unit tests

Testing will verify functional interfaces across packages to guarantee robustness:

1. **Schema and Database Core Testing**:
   - Verify `InitSchema` successfully creates tables inside an in-memory SQLite database (`"file::memory:?cache=shared"`).
   - Test package-level variables and functions in `linkshelf/internal/store/store.go` to confirm that `List`, `Create`, and `Delete` query execution functions function cleanly.
   
2. **Validation Testing**:
   - Execute table-driven tests checking various invalid inputs to `store.Create` (e.g., titles containing more than 200 runes, empty URLs, and non-http/https schemes). Ensure validation fails gracefully with errors.

3. **HTTP Route and Handler Testing**:
   - Write unit and integration tests using `net/http/httptest`.
   - Ensure `GET /` successfully retrieves files from `linkshelf/web/index.html`.
   - Ensure `GET /static/{file}` successfully retrieves stylesheets from `linkshelf/web/style.css` and scripts from `linkshelf/web/app.js`.
   - Ensure requesting `/static/../main.go` or other path manipulation containing `..` returns a HTTP code 400.
   - Verify endpoint `/api/links` responds with an empty JSON array `[]` initially.
   - Verify POST request parameters and JSON serialization outputs.
   - Verify DELETE `/api/links/{id}` deletes entries.

---

## Integration and testing
To verify the full-suite implementation, the automated verification command is:
cd linkshelf && go test ./...
To run the server locally to interact with the UI:
cd linkshelf && go run ./cmd/server

---

## Acceptance mapping

This architecture mapping ensures that every single requirement defined in the `SPEC.md` has a concrete, uniquely-prefixed path and implementation contract:

1. **Database Schema Isolation**:
   - *Requirement*: `store.go` must not contain `CREATE TABLE` - call `InitSchema` from `main` and from any tests.
   - *Design*: Schema creation is restricted entirely to `InitSchema` inside `linkshelf/internal/store/schema.go`.
   
2. **Global Package-Level Store API**:
   - *Requirement*: Use package-level functions and a package variable — not a `Store` struct or `NewStore`.
   - *Design*: Implemented explicitly via `var DB *sql.DB` and top-level functions `List`, `Create`, `Delete` inside `linkshelf/internal/store/store.go`.
   
3. **Exact HTTP Route Names & Contract**:
   - *Requirement*: GET `/`, GET `/static/{file}`, GET `/api/links`, POST `/api/links`, DELETE `/api/links/{id}`.
   - *Design*: Explicit routing implemented under `linkshelf/internal/api/handlers.go` mapping directly to the exact paths.
   
4. **Data Validation for Creation**:
   - *Requirement*: Title non-empty max 200 runes, URL prefixed with http:// or https://.
   - *Design*: The verification is implemented in `Create` in `linkshelf/internal/store/store.go`.
   
5. **Path Security**:
   - *Requirement*: Reject static paths containing `..`.
   - *Design*: The static asset handler in `linkshelf/internal/api/handlers.go` intercepts and blocks requested files containing parent traversal dots.
