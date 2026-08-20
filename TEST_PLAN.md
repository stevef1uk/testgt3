# Test Plan — module-phase

## Requirements → tests

### module-1
Requirement: The Go module linkshelf must be declared at linkshelf/go.mod with Go 1.22 and the github.com/mattn/go-sqlite3 dependency, supporting `go mod tidy` and `go test ./...` from the linkshelf directory.
Level: unit
Test file: linkshelf/go.mod (verification via `go mod tidy` and `go mod download`)
Bead ID: te-vbs
Scenarios:
- Module declaration matches module linkshelf
- Go version directive is 1.22
- Dependency github.com/mattn/go-sqlite3 is present with a resolvable version
- `go mod tidy` completes without error
- `go mod download` completes without error (as specified in plan.md Verify)
- `go test ./...` compiles all packages (even those without test files)
Assertions:
- go.mod file exists at linkshelf/go.mod
- The module directive reads `module linkshelf`
- The go directive reads `go 1.22`
- The require block includes github.com/mattn/go-sqlite3
- Running `go mod tidy` from linkshelf/ succeeds and produces no errors
- Running `go mod download` from linkshelf/ succeeds and produces no errors
