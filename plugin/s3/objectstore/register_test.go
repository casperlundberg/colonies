package objectstore

import (
	"testing"

	"github.com/colonyos/colonies/pkg/fs"
	"github.com/stretchr/testify/assert"
)

func TestS3RegistersOnImport(t *testing.T) {
	names := fs.RegisteredObjectStores()
	found := false
	for _, name := range names {
		if name == "s3" {
			found = true
			break
		}
	}
	assert.True(t, found, "s3 should be registered via init()")
}
