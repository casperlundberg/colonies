package main

import (
	"testing"

	"github.com/colonyos/colonies/pkg/backends"
	"github.com/colonyos/colonies/pkg/cluster"
	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/fs"
	"github.com/colonyos/colonies/pkg/monitoring"
	"github.com/stretchr/testify/assert"

	// These blank imports mirror cmd/main.go and trigger init() registration.
	_ "github.com/colonyos/colonies/plugin/embedded"
	_ "github.com/colonyos/colonies/plugin/gin"
	_ "github.com/colonyos/colonies/plugin/localfs"
	_ "github.com/colonyos/colonies/plugin/postgresql"
	_ "github.com/colonyos/colonies/plugin/prometheus"
	_ "github.com/colonyos/colonies/plugin/s3/objectstore"
)

func TestAllPluginsRegisterOnImport(t *testing.T) {
	t.Run("database/postgresql", func(t *testing.T) {
		assert.Contains(t, database.RegisteredDrivers(), "postgresql")
	})
	t.Run("database/embedded", func(t *testing.T) {
		assert.Contains(t, database.RegisteredDrivers(), "embedded")
	})
	t.Run("cluster/etcd", func(t *testing.T) {
		assert.Contains(t, cluster.RegisteredClusters(), "etcd")
	})
	t.Run("backends/gin", func(t *testing.T) {
		assert.Contains(t, backends.RegisteredBackends(), "gin")
	})
	t.Run("fs/s3", func(t *testing.T) {
		assert.Contains(t, fs.RegisteredObjectStores(), "s3")
	})
	t.Run("fs/coloniesfs", func(t *testing.T) {
		assert.Contains(t, fs.RegisteredObjectStores(), "coloniesfs")
	})
	t.Run("monitoring/prometheus", func(t *testing.T) {
		assert.Contains(t, monitoring.RegisteredMonitors(), "prometheus")
	})
}
