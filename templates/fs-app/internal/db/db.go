package db

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func New() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "file::memory:?cache=shared")
	if err != nil {
		return nil, err
	}
	return db, nil
}

func MigrateTables(db *sql.DB) error {
	return fs.WalkDir(migrationsFS, "migrations", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || filepath.Ext(path) != ".sql" {
			return nil
		}

		content, err := migrationsFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", path, err)
		}

		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", path, err)
		}

		return nil
	})
}
