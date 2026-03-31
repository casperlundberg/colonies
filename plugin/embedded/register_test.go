package embedded

import (
	"testing"

	"github.com/colonyos/colonies/pkg/database"
	"github.com/stretchr/testify/assert"
)

func TestEmbeddedRegistersOnImport(t *testing.T) {
	drivers := database.RegisteredDrivers()
	found := false
	for _, name := range drivers {
		if name == "embedded" {
			found = true
			break
		}
	}
	assert.True(t, found, "embedded driver should be registered via init()")
}
