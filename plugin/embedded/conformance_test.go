package embedded_test

import (
	"os"
	"testing"

	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/databasetest"
	"github.com/colonyos/colonies/plugin/embedded"
)

func TestDatabaseConformance(t *testing.T) {
	databasetest.RunConformanceTests(t, func(t *testing.T) (database.Database, func()) {
		t.Helper()
		dir, err := os.MkdirTemp("", "conformance-embedded-*")
		if err != nil {
			t.Fatal(err)
		}
		db := embedded.CreateEmbeddedDatabase(dir)
		if err := db.Initialize(); err != nil {
			t.Fatal(err)
		}
		return db, func() {
			db.Close()
			os.RemoveAll(dir)
		}
	})
}
