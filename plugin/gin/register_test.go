package gin

import (
	"testing"

	"github.com/colonyos/colonies/pkg/backends"
	"github.com/stretchr/testify/assert"
)

func TestGinRegistersOnImport(t *testing.T) {
	names := backends.RegisteredBackends()
	found := false
	for _, name := range names {
		if name == "gin" {
			found = true
			break
		}
	}
	assert.True(t, found, "gin backend should be registered via init()")
}

func TestGinBackendFactoryImplementsInterface(t *testing.T) {
	factory, err := backends.GetBackend("gin")
	assert.NoError(t, err)
	assert.NotNil(t, factory)

	// Verify CreateEngine returns a non-nil engine.
	engine := factory.CreateEngine()
	assert.NotNil(t, engine)

	// Verify CreateServer returns a non-nil server.
	server := factory.CreateServer(0, engine)
	assert.NotNil(t, server)

	// Verify CORS returns a non-nil handler.
	corsHandler := factory.CORS()
	assert.NotNil(t, corsHandler)
}
