
# Link Shelf - Bookmark Manager

## Project Overview

Link Shelf is a simple bookmark manager built with Go and SQLite. It provides a REST API for managing links and a static web frontend.

## Architecture

- **Go Module**: `linkshelf` (Go 1.22)
- **Server Entry**: `cmd/server/main.go`
- **Store Package**: `internal/store` - Link struct, schema, CRUD
- **API Package**: `internal/api` - HTTP handlers
- **Web Frontend**: `web/index.html`, `web/app.js`, `web/style.css`
- **Database**: SQLite (`linkshelf.db`)

## API Endpoints

- `GET /api/links` - List all links
- `POST /api/links` - Create link (title required, URL must be http/https)
- `DELETE /api/links/{id}` - Delete link

## Web UI

- Served at `/` (static files from `web/`)
- List links, add new, delete
- Refreshes after each operation

## Acceptance Criteria

1. `cd linkshelf && go test ./...` passes
2. Server starts on :8080 and serves UI
3. API endpoints work correctly

## Tech Stack

- Go 1.22 with sqlite3 driver
- Static web (HTML/JS/CSS) - no build step
- SQLite database (file-based)

## Testing Requirements

- Unit tests for store package
- API integration tests
- Playwright E2E tests for web UI:
  - Page loads at http://localhost:8080
  - Can add a link via UI
  - Can delete a link via UI
  - List updates correctly

## Delivery Phases

1. **go-module** - Initialize go.mod
2. **store-layer** - Schema + CRUD
3. **api-handlers** - HTTP handlers
4. **server-main** - Server entrypoint
5. **web-static** - CSS/JS assets
6. **web-shell** - HTML shell
7. **integration-test** - Full smoke test + Playwright E2E
