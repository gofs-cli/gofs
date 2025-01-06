package db

import (
	"context"
	"database/sql"
	"fmt"

	"module/placeholder/internal/db/migrations"
)

type DB struct {
	conn    *sql.DB
	closeFn func() error
}

func (d *DB) Conn() *sql.DB {
	return d.conn
}

func (d *DB) Close(ctx context.Context) error {
	if d.closeFn == nil {
		return nil
	}
	return d.closeFn()
}

func LocalPG(dsn string) (DB, error) {
	sDb, err := sql.Open("pgx", dsn)
	if err != nil {
		return DB{}, fmt.Errorf("error making connection: %v", err)
	}
	err = sDb.Ping()
	if err != nil {
		return DB{}, fmt.Errorf("error pinging db: %v", err)
	}
	return DB{
		conn:    sDb,
		closeFn: sDb.Close,
	}, nil
}

func MigrateTables(db DB) error {
	if db.conn == nil {
		return nil
	}
	files, _ := migrations.Dir.ReadDir(".")
	for _, file := range files {
		data, _ := migrations.Dir.ReadFile(file.Name())
		_, err := db.Conn().Exec(string(data))
		if err != nil {
			return fmt.Errorf("error executing sql: %v", err)
		}
	}
	return nil
}
