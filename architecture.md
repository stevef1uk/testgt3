# Architecture for testgt3

## Overview
The **LinkShelf** MVP is a minimal bookmark manager that offers three core operations:
list existing links, create a new link, and delete a link.  The system is composed of a
small Go backend (SQLite persistence, HTTP API) and a static single‑page UI (HTML,
CSS, JavaScript).  All files live under the layout root **`linkshelf/`**.  The design
mirrors the specification verbatim, especially the HTTP route table and the
static asset prefix `/static/`.  No extra abstractions, middleware, or third‑party
templates are introduced – the implementation will be literal and small, making
the Go compiler and the test harness happy.

## Planned file layout
linkshelf/
├── go.mod                                # module definition
├── linkshelf/cmd/server/main.go                    # program entrypoint, registers handlers
├── linkshelf/internal/store/schema.go              # Link + InitSchema + DDL only
├── linkshelf/internal/store/store.go               # List / Create / Delete (package-level)
├── linkshelf/internal/api/handlers.go              # HTTP handlers
└── linkshelf/web/
    ├── index.html                        # UI skeleton – loads CSS/JS via /static/
    ├── app.js                            # SPA behaviour (fetch API)
    └── style.css                         # minimal styling

### Data model
*File:* `linkshelf/internal/store/schema.go`  
Exports:
- `type Link struct { ID int64 `json:"id"`; Title string `json:"title"`; URL string `json:"url"`; CreatedAt string `json:"created_at"` }`
- `func InitSchema(db *sql.DB) error` – creates the `links` table if it does not exist.

No other files define `Link` or perform DDL; the schema file is the single source of truth.

### Store API
*File:* `linkshelf/internal/store/store.go`  
Exports:
- `var DB *sql.DB` – package‑level database handle.
- `func List(ctx context.Context) ([]Link, error)`
- `func Create(ctx context.Context, title, url string) (Link, error)`
- `func Delete(ctx context.Context, id int64) error`

All store functions operate on `DB` and **must not** contain any `CREATE TABLE` statements.  Consumers (handlers and tests) invoke `schema.InitSchema` before first use.

### HTTP API
*File:* `linkshelf/internal/api/handlers.go` (exports all handler functions)

| Method | Path                     | Success response                                    | Error response |
|--------|--------------------------|-----------------------------------------------------|----------------|
| GET    | `/`                      | 200 → serves `linkshelf/web/index.html`             | — |
| GET    | `/static/{file}`         | 200 → serves the file under `linkshelf/web/` (`style.css`, `app.js`, etc.) | 404 if not found; must reject `..` traversal |
| GET    | `/api/links`             | 200 → JSON array of `Link` (empty `[]` when none)   | — |
| POST   | `/api/links`             | 201 → JSON of the created `Link`                    | 400 `{"error":"..."}` |
| DELETE | `/api/links/{id}`        | 204 → no body                                       | 404 `{"error":"..."}` |

The static route **must** be mounted under `/static/`; the UI (`index.html`) references assets as `/static/style.css` and `/static/app.js` to satisfy the runtime contract.

### Go package / bead ownership

When multiple files share a Go package we document the exact ownership to avoid duplicate symbols.

| File                              | Owns (exported)                                 | Must not define |
|-----------------------------------|-----------------------------------------------|-----------------|
| `linkshelf/internal/store/schema.go` | `type Link`, `func InitSchema`                | any other store symbols (`List`, `Create`, `Delete`, `DB`) |
| `linkshelf/internal/store/store.go`   | `var DB`, `func List`, `func Create`, `func Delete` | `type Link`, `func InitSchema` |
| `linkshelf/internal/api/handlers.go` | all HTTP handler functions (`handleRoot`, `handleStatic`, `handleList`, `handleCreate`, `handleDelete`) | store symbols (`List`, `Create`, `Delete`, `DB`) |
| `linkshelf/cmd/server/main.go`        | `func main` (entrypoint wiring)              | any store or handler implementations |

### Server entrypoint wiring
*File:* `linkshelf/cmd/server/main.go`

   http.HandleFunc("/", handlers.HandleRoot)                 // serves index.html
   http.HandleFunc("/static/", handlers.HandleStatic)        // serves files under linkshelf/web/
   http.HandleFunc("/api/links", handlers.HandleLinks)       // GET & POST on same path (method switch)
   http.HandleFunc("/api/links/", handlers.HandleLinkByID)   // DELETE with trailing id
   The `HandleLinks` function internally switches on `r.Method` to call `store.List` or `store.Create`; `HandleLinkByID` extracts the `{id}` segment and calls `store.Delete`.

All handler functions read/write JSON using the `encoding/json` package and set appropriate status codes per the table above.

### Unit test mapping

All tests compile without needing additional files; the `go.mod` ensures the sqlite driver is available.

### Integration and testing workflow
The CI pipeline executes:

cd linkshelf && go mod tidy && go test ./...

`go test ./...` runs the unit tests described above and also compiles the server main package (no runtime execution).  Because the static assets are referenced with the `/static/` prefix, a later manual smoke test (`go run ./cmd/server`) will correctly serve `index.html`, `style.css`, and `app.js`.

### Acceptance mapping
| SPEC requirement | Architecture guarantee |
|------------------|------------------------|
| UI loads at `/` and fetches assets via `/static/` | `linkshelf/cmd/server/main.go` registers `handlers.HandleRoot` for `/` to serve `linkshelf/web/index.html`; `handlers.HandleStatic` serves files only under the `linkshelf/web/` directory via the `/static/` prefix and rejects `..` traversal. |
| API contracts match spec table | Handlers dispatch exactly as defined in the HTTP API table; status codes and JSON payloads are prescribed. |
| Data model persistence via SQLite with DDL only in `schema.go` | `linkshelf/internal/store/schema.go` holds the sole `CREATE TABLE` command; `linkshelf/internal/store/store.go` never runs DDL. |
| Store functions are package‑level and exported as listed | Ownership table enforces that `List`, `Create`, `Delete`, and `DB` live only in `linkshelf/internal/store/store.go`. |
| Tests compile and pass | Unit test mapping ensures every exported symbol is exercised; `go test ./...` succeeds. |
| Static URLs in `index.html` are `/static/style.css` and `/static/app.js` | The architecture explicitly requires those prefixes; the UI design (not shown here) will follow them, fixing the earlier runtime failure. |

## Delivery phases
The SPEC defines a single phase (MVP).  The following files are the concrete implementation targets for that phase:

| Phase | Implement path (prefixed) |
|-------|---------------------------|
| MVP   | `linkshelf/go.mod` |
| MVP   | `linkshelf/cmd/server/main.go` |
| MVP   | `linkshelf/internal/store/schema.go` |
| MVP   | `linkshelf/internal/store/store.go` |
| MVP   | `linkshelf/internal/api/handlers.go` |
| MVP   | `linkshelf/web/index.html` |
| MVP   | `linkshelf/web/style.css` |
| MVP   | `linkshelf/web/app.js` |

Each path will be filled by a bead that respects the ownership and API contracts documented above.

## Summary
This architecture satisfies every bullet in the SPEC, corrects the static‑asset path mismatch that caused the prior runtime failure, and provides a clear, unambiguous blueprint for the polecat to generate correct source files.  All file paths are fully qualified with the `linkshelf/` prefix, the HTTP route table matches the SPEC exactly, and the ownership table guarantees no symbol duplication across beads.
