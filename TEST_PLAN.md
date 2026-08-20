# Test Plan — module-and-persistence-bead

## Requirements → tests

### R1: Module declaration
Requirement: linkshelf/go.mod must declare module linkshelf with Go 1.22 and github.com/mattn/go-sqlite3 v1.14.22 dependency
Level: unit
Test file: linkshelf/go.mod (verified by `go mod download` and `go build ./...`)
Bead ID: te-22l
Scenarios:
- Module file compiles: `go mod download` succeeds
- Module file matches expected module path and Go version
Assertions:
- `go mod download` returns exit code 0
- Module path is `linkshelf`, Go version is 1.22

### R2: Link struct and InitSchema
Requirement: linkshelf/internal/store/schema.go must export Link struct with ID, Title, URL, CreatedAt fields and InitSchema(*sql.DB) error creating links table
Level: unit
Test file: linkshelf/internal/store/schema_test.go
Bead ID: te-xcs
Scenarios:
- InitSchema creates links table when it does not exist
- InitSchema does not error when called twice (idempotent/repeat-safe)
- Link struct has correct JSON tags and field types
- Table DDL matches SPEC exactly (id integer PK, title, url, created_at)
Assertions:
- `InitSchema(db)` returns nil on first call
- `InitSchema(db)` returns nil on second call
- Table introspection confirms id, title, url, created_at columns exist

### R3: Store List function
Requirement: store.List(ctx) returns ([]Link, error) ordering by id DESC, returns non-nil empty slice when no rows
Level: unit
Test file: linkshelf/internal/store/store_test.go
Bead ID: te-19f
Scenarios:
- List returns empty []Link (not nil) after InitSchema with no inserts
- List returns links in descending ID order after multiple creates
- List returns correct field values after a single create
Assertions:
- len(result) > 0 and result[0].ID > result[1].ID for two inserts
- Empty database returns `[]Link{}` (not nil, len == 0)

### R4: Store Create with validation
Requirement: store.Create(ctx, title, url) validates non-empty title (max 200 runes), non-empty URL starting with http:// or https://, returns full Link with RFC3339 CreatedAt
Level: unit
Test file: linkshelf/internal/store/store_test.go
Bead ID: te-19f
Scenarios:
- Create with valid title and URL returns Link with ID > 0 and RFC3339 CreatedAt
- Create with empty title returns validation error
- Create with title exceeding 200 runes returns validation error
- Create with empty URL returns validation error
- Create with URL not starting with http:// or https:// returns validation error
- CreatedAt timestamp is valid RFC3339 UTC
Assertions:
- Successful Create returns Link.ID != 0
- Validation errors are non-nil and descriptive
- Returned Link.CreatedAt matches RFC3339 format

### R5: Store Delete
Requirement: store.Delete(ctx, id) deletes by ID, returns error if id not found
Level: unit
Test file: linkshelf/internal/store/store_test.go
Bead ID: te-19f
Scenarios:
- Delete of existing ID succeeds (no error)
- Delete makes the link no longer listable
- Delete of non-existent ID returns error
Assertions:
- Delete(existingID) returns nil
- After Delete, List does not include the deleted link
- Delete(nonExistentID) returns non-nil error

### R6: Store integration (List/Create/Delete end-to-end)
Requirement: Store functions work together end-to-end with :memory: SQLite
Level: integration
Test file: linkshelf/internal/store/store_test.go
Bead ID: te-19f
Scenarios:
- Create a link, List returns it, Delete it, List no longer returns it
- Multiple creates followed by List returns all in descending order
Assertions:
- Full create→list→delete→list cycle returns empty list
- ID ordering is consistent across at least 3 inserts

### R7: Schema + Store wire-up (integration)
Requirement: InitSchema called before store.DB assignment; tests use :memory: DB with InitSchema
Level: integration
Test file: linkshelf/internal/store/store_test.go
Bead ID: te-xcs, te-19f
Scenarios:
- Test helper uses `db, _ := sql.Open("sqlite3", ":memory:"); InitSchema(db); store.DB = db`
- Store functions operate correctly after this setup
Assertions:
- No panic or nil pointer from store functions after InitSchema + DB assignment
- store.List returns empty slice after fresh setup

### R8: Package compilation
Requirement: All module-and-persistence-bead packages compile: `cd linkshelf && go test ./...` is green
Level: unit
Test file: linkshelf/internal/store/schema_test.go, linkshelf/internal/store/store_test.go
Bead ID: te-22l, te-xcs, te-19f
Scenarios:
- `go build ./internal/store/...` succeeds
- `go test -count=1 ./internal/store/...` passes all tests
Assertions:
- Compilation has no errors
- All tests pass with exit code 0
