package store

import (
	"database/sql"
	"database/sql/driver"
	"testing"
)

type schemaTestDriver struct{}

type schemaTestConn struct{}

func (schemaTestDriver) Open(string) (driver.Conn, error) {
	return schemaTestConn{}, nil
}

func (schemaTestConn) Prepare(string) (driver.Stmt, error) {
	return schemaTestStmt{}, nil
}

func (schemaTestConn) Close() error { return nil }

func (schemaTestConn) Begin() (driver.Tx, error) { return schemaTestTx{}, nil }

type schemaTestStmt struct{}

func (schemaTestStmt) Close() error  { return nil }
func (schemaTestStmt) NumInput() int { return 0 }
func (schemaTestStmt) Exec([]driver.Value) (driver.Result, error) {
	return driver.RowsAffected(0), nil
}

type schemaTestTx struct{}

func (schemaTestTx) Commit() error   { return nil }
func (schemaTestTx) Rollback() error { return nil }

func TestLinkModel(t *testing.T) {
	link := Link{
		ID:        7,
		Title:     "Example",
		URL:       "https://example.com",
		CreatedAt: "2024-01-02T03:04:05Z",
	}
	if link.ID != 7 || link.Title != "Example" || link.URL != "https://example.com" {
		t.Fatalf("unexpected link: %+v", link)
	}
}

func TestInitSchemaRejectsNilDatabase(t *testing.T) {
	if err := InitSchema(nil); err == nil {
		t.Fatal("InitSchema(nil) error = nil, want error")
	}
}

func TestInitSchemaExecutesDDL(t *testing.T) {
	sql.Register("schema-test", schemaTestDriver{})
	db, err := sql.Open("schema-test", "")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := InitSchema(db); err != nil {
		t.Fatalf("InitSchema() error = %v", err)
	}
}
