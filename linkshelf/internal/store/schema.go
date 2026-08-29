// Package store contains the data model and persistence layer for LinkShelf.
//
// schema.go owns the Link domain type and the database DDL. It is the only
// file in the package that contains CREATE TABLE statements.
package store

import (
	"database/sql"
)

// Link is the canonical representation of a bookmarked link as it appears in
// both the database and JSON responses.
type Link struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	CreatedAt string `json:"created_at"` // RFC3339 UTC
}

// schemaDDL is the canonical CREATE TABLE statement for the links table.
// It is intentionally unexported so that no caller outside this file can
// inline the DDL — callers must go through InitSchema.
const schemaDDL = `
CREATE TABLE IF NOT EXISTS links (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    title      TEXT    NOT NULL,
    url        TEXT    NOT NULL,
    created_at TEXT    NOT NULL
);
`

// InitSchema creates the links table in the provided database if it does
// not already exist. It is safe to call on every startup.
//
// This is the only place in the codebase that issues CREATE TABLE for the
// links table. main.go and tests both call this helper rather than
// duplicating the DDL string.
func InitSchema(db *sql.DB) error {
	if _, err := db.Exec(schemaDDL); err != nil {
		return err
	}
	return nil
}
