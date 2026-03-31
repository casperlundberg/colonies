package database

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

type DatabaseType string

const (
	PostgreSQL DatabaseType = "postgresql"
	Embedded   DatabaseType = "embedded"
)

type DatabaseConfig struct {
	Type        DatabaseType
	Host        string
	Port        int
	User        string
	Password    string
	Name        string
	Prefix      string
	TimescaleDB bool

	DataDir string
}

func CreateDatabase(config DatabaseConfig) (Database, error) {
	log.WithFields(log.Fields{
		"DatabaseType": config.Type,
		"Host":         config.Host,
		"Port":         config.Port,
		"Name":         config.Name,
		"Prefix":       config.Prefix,
		"TimescaleDB":  config.TimescaleDB,
		"DataDir":      config.DataDir,
	}).Info("Creating database connection")

	name := string(config.Type)
	db, err := CreateFromRegistry(name, config)
	if err != nil {
		log.WithField("DatabaseType", config.Type).Error("Unsupported database type requested")
		return nil, fmt.Errorf("unsupported database type: %s: %w", config.Type, err)
	}
	return db, nil
}
