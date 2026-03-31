package monitoring

import (
	"testing"

	monitoringpkg "github.com/colonyos/colonies/pkg/monitoring"
	"github.com/stretchr/testify/assert"
)

func TestPrometheusRegistersOnImport(t *testing.T) {
	names := monitoringpkg.RegisteredMonitors()
	found := false
	for _, name := range names {
		if name == "prometheus" {
			found = true
			break
		}
	}
	assert.True(t, found, "prometheus monitor should be registered via init()")
}
