package embedded

import (
	"fmt"

	"github.com/colonyos/colonies/pkg/database"
)

func init() {
	database.Register("embedded", func(config database.DatabaseConfig) (database.Database, error) {
		if config.DataDir == "" {
			return nil, fmt.Errorf("embedded database requires DataDir")
		}
		db := CreateEmbeddedDatabase(config.DataDir)
		if err := db.Initialize(); err != nil {
			return nil, fmt.Errorf("failed to initialize embedded database: %w", err)
		}
		return db, nil
	})
}
