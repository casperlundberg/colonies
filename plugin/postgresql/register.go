package postgresql

import "github.com/colonyos/colonies/pkg/database"

func init() {
	database.Register("postgresql", func(config database.DatabaseConfig) (database.Database, error) {
		db := CreatePQDatabase(config.Host, config.Port, config.User, config.Password, config.Name, config.Prefix, config.TimescaleDB)
		return db, nil
	})
}
