package postgresql_test

import (
	"testing"

	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/databasetest"
	"github.com/colonyos/colonies/plugin/postgresql"
)

func TestDatabaseConformance(t *testing.T) {
	db, err := postgresql.PrepareTests()
	if err != nil {
		t.Skipf("PostgreSQL not available: %v", err)
	}

	databasetest.RunConformanceTests(t, func(t *testing.T) (database.Database, func()) {
		t.Helper()
		// PostgreSQL tests share a connection, just drop and reinitialize
		if err := db.Drop(); err != nil {
			t.Fatal(err)
		}
		if err := db.Initialize(); err != nil {
			t.Fatal(err)
		}
		return db, func() {} // shared connection, don't close
	})

	db.Close()
}
