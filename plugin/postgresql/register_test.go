package postgresql

import (
	"testing"

	"github.com/colonyos/colonies/pkg/database"
	"github.com/stretchr/testify/assert"
)

func TestPostgresqlRegistersOnImport(t *testing.T) {
	drivers := database.RegisteredDrivers()
	found := false
	for _, name := range drivers {
		if name == "postgresql" {
			found = true
			break
		}
	}
	assert.True(t, found, "postgresql driver should be registered via init()")
}
