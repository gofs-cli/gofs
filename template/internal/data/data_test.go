//go:build gendata

package data

import (
	"testing"
	"time"

	"github.com/gofs-cli/gofs/template/internal/config"
	"github.com/gofs-cli/gofs/template/internal/db"
)

func TestGenData(t *testing.T) {
	conf := config.New()
	var conn db.DB
	var err error
	for range 5 { // attempts
		conn, err = db.LocalPG(conf.DSN)
		if err != nil {
			t.Log("error creating db, retrying...")
			time.Sleep(1 * time.Second)
		} else {
			t.Log("connected to db")
			break
		}
	}
	if err != nil {
		t.Fatalf("error creating db: %v", err)
	}
	err = db.MigrateTables(conn)
	if err != nil {
		t.Fatalf("error migrating tables: %v", err)
	}
}
