# Architecture for testgt3

## Overview
The **Linkshelf** MVP is a minimal Go‑based bookmark manager that stores links in a SQLite database and exposes a tiny HTTP UI together with a JSON REST API.  
The system is deliberately small: a single Go module (`linkshelf`), three core Go packages (`linkshelf/cmd/server`, `linkshelf/internal/store`, `linkshelf/internal/api`), a static front‑end under `linkshelf/web/`, and a single SQLite file `linkshelf.db`. All components are wired at start‑up and communicate via well‑defined function contracts. The design follows the SPEC verbatim, guaranteeing that route paths, function names, and data‑model fields match the expectations of the automated test harness.

Key constraints:

* Go version 1.22, module path `linkshelf`.
* No additional abstractions; the store uses package‑level functions and a global `DB` variable.
* All HTTP routes must be registered on `http.DefaultServeMux`.
* Static assets are served under `/static/{file}` with strict path traversal protection.
* Validation rules for incoming link data are enforced in the `store.Create` function.
* The server listens on `:8080` and serves the UI and API concurrently.

## Planned file layout
All implementation files live under the **layout root** `linkshelf/`. The architecture references each file but does **not** create them – the polecat will implement later.

- `linkshelf/go.mod` – module declaration, Go version, dependency on `github.com/mattn/go-sqlite3`.
- `linkshelf/cmd/server/main.go` – application entry point: opens/creates `linkshelf.db`, calls `schema.InitSchema`, wires handlers, and starts `http.ListenAndServe(":8080", nil)`.
- `linkshelf/internal/store/schema.go` – defines the `Link` struct, the `InitSchema` function, and the DDL for the `links` table.
- `linkshelf/internal/store/store.go` – provides package‑level API: `List`, `Create`, `Delete`, and the global `DB *sql.DB`.
- `linkshelf/internal/api/handlers.go` – HTTP handler functions for UI routing (`/`, `/static/{file}`) and the JSON API (`/api/links`, `/api/links/{id}`).
- `linkshelf/web/index.html` – static HTML page containing inputs for title/url, an Add button, and an unordered list with id `links`.
- `linkshelf/web/app.js` – front‑end JavaScript that fetches the list of links, renders them, handles form submissions, and performs DELETE requests.
- `linkshelf/web/style.css` – minimal stylesheet for readability.

### Persistence file
All database schema definitions reside in `linkshelf/internal/store/schema.go`. The `InitSchema` function creates the `links` table if it does not exist. No other file contains DDL; the store package only performs CRUD operations against the already‑initialized schema.

## Go package / bead ownership
Multiple source files share the same Go package (`linkshelf/internal/store`). To avoid symbol duplication and to keep the architecture explicit, ownership of exported symbols is documented per file.

| File                                         | Owns (exported)                                 | Must not define |
|----------------------------------------------|-------------------------------------------------|-----------------|
| `linkshelf/internal/store/schema.go`         | ``Link`` struct, ``InitSchema`` function, ``sqlite`` imports | ``List``, ``Create``, ``Delete`` |
| `linkshelf/internal/store/store.go`          | ``DB`` variable, ``List``, ``Create``, ``Delete`` functions | ``Link`` struct, ``InitSchema`` |
| `linkshelf/internal/api/handlers.go`         | ``handleRoot``, ``handleStatic``, ``handleList``, ``handleCreate``, ``handleDelete`` (handler functions) | ``Link`` struct, ``InitSchema`` |

The `linkshelf/cmd/server/main.go` file does not own exported symbols; it only composes them.

## HTTP + entrypoint integration
The SPEC defines the exact HTTP contract. The table below reproduces it verbatim:

| Method | Path               | Success Response                              | Error Response |
|--------|--------------------|-----------------------------------------------|----------------|
| GET    | `/`                | 200, serves `linkshelf/web/index.html`        | — |
| GET    | `/static/{file}`   | 200, file under `linkshelf/web/`                | 404 if not found or path traversal |
| GET    | `/api/links`       | 200, JSON array `[]` (empty when no links)    | — |
| POST   | `/api/links`       | 201, JSON representation of created `Link`    | 400 `{"error":"..."}` |
| DELETE | `/api/links/{id}`  | 204 No Content                                 | 404 `{"error":"..."}` |

**Wiring story**  
`linkshelf/cmd/server/main.go` performs the following steps in order:

1. Open (or create) the SQLite file `linkshelf.db` using `sql.Open("sqlite3", "linkshelf.db")`.
2. Call `schema.InitSchema(db)` to ensure the `links` table exists.
3. Assign the opened DB to the store package: `store.DB = db`.
4. Register HTTP handlers on `http.DefaultServeMux`:
   - `http.HandleFunc("/", handlers.handleRoot)` – serves the UI.
   - `http.HandleFunc("/static/", handlers.handleStatic)` – serves static assets with security checks.
   - `http.HandleFunc("/api/links", handlers.handleListOrCreate)` – dispatches GET to `store.List` and POST to `store.Create`.
   - `http.HandleFunc("/api/links/", handlers.handleDelete)` – extracts `{id}` from the URL path and calls `store.Delete`.
5. Call `log.Printf("listening on :8080")` then `http.ListenAndServe(":8080", nil)`.

All handler functions are thin adapters: they decode JSON bodies, invoke the store functions, and encode JSON responses or appropriate status codes. Errors from validation or database operations are translated to the error response format defined in the SPEC.

## Unit tests
Even though the SPEC marks unit tests as optional, the architecture anticipates a conventional Go test layout to guarantee correctness of the core packages.

- **Package `linkshelf/internal/store`** – `linkshelf/internal/store/store_test.go` (not required but recommended):
  - Uses an in‑memory SQLite DB (`:memory:`), calls `schema.InitSchema`, assigns `store.DB`, then exercises `List`, `Create`, and `Delete`.
  - Tests validation rules for `Create` (empty title, overly long title, missing scheme in URL, etc.).
  - Confirms ordering (`List` returns links in descending `id` order).

- **Package `linkshelf/internal/api`** – `linkshelf/internal/api/handlers_test.go`:
  - Creates an HTTP test server with the same handler registration.
  - Sends GET `/api/links`, expects empty array.
  - Sends POST `/api/links` with valid JSON, expects 201 and correct body.
  - Sends DELETE `/api/links/{id}` for both existing and non‑existing IDs, checks status codes.

- **Package `linkshelf/cmd/server`** – no explicit test file; the integration test will launch the server.

All tests are run with:

cd linkshelf && go test ./...

which must succeed even if no `*_test.go` files exist (the compilation succeeds).

## Integration and testing
The full system is exercised by two primary verification steps performed by the pipeline:

1. **Unit/compile verification**  
   cd linkshelf && go test ./...
   This compiles every package, runs any existing tests, and verifies that the module builds without errors.

2. **Runtime verification**  
   cd linkshelf && go run ./cmd/server
   The server must start, listen on `:8080`, and serve the UI. Manual verification (performed by the pipeline) will request `/` and `/api/links` to confirm that the HTTP layer is correctly wired.

The integration phase (if later added) could leverage `curl` or a simple Go integration test that spawns the server in a goroutine, performs a sequence of API calls (list → create → delete → list), and asserts the expected JSON payloads and status codes.

## Docker & Deployment
The SPEC does **not** list a Dockerfile or docker‑compose configuration, so this section is omitted. If a future phase adds containerisation, the architecture would be extended accordingly.

## E2E / integration testing
No end‑to‑end test files are listed in the SPEC, thus this section is omitted as well. Should an e2e suite be introduced, the architecture would detail how the server is started (via `docker compose` or locally) and the exact Playwright/Playwright‑like commands used.

## Acceptance mapping
| SPEC Goal | Architecture Decision | Verification |
|-----------|----------------------|--------------|
| Go module builds (`go mod tidy`) | `linkshelf/go.mod` declares `module linkshelf` and `go 1.22` | `go mod tidy` succeeds |
| `go test ./...` passes | All packages compile; optional test files follow the contract described | `go test ./...` returns success |
| Server starts on `:8080` and serves UI | `linkshelf/cmd/server/main.go` calls `http.ListenAndServe(":8080", nil)` and registers `handleRoot` for `/` | Manual `curl http://localhost:8080/` returns HTML |
| CRUD API matches contract | Handlers in `linkshelf/internal/api/handlers.go` implement routes exactly as defined in the SPEC table; store functions follow specified signatures | `curl`/`http` client tests validate status codes and JSON structures |
| Validation rules enforced | `store.Create` validates title length and URL scheme; errors are propagated as `400` JSON responses | Test cases cover each validation branch |
| Static asset protection | `handleStatic` rejects any path containing `..` before serving files | Attempted path traversal yields `404` |
| Data model matches SPEC | `Link` struct defined in `schema.go` with fields `ID`, `Title`, `URL`, `CreatedAt` and JSON tags | JSON marshalling produces exact field names |

By adhering to the above architecture, the implementation will satisfy every functional and non‑functional requirement of the Linkshelf MVP, allowing the automated pipeline to verify compile‑time correctness and runtime behavior without any drift from the specification.
