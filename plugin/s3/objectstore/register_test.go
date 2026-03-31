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

func TestS3FactoryReturnsErrorOnBadEndpoint(t *testing.T) {
	// Verify CreateObjectStore reaches the s3 factory and it handles bad config gracefully.
	// An empty endpoint will fail at minio client creation, proving the factory was invoked.
	_, err := fs.CreateObjectStore("s3", fs.ObjectStoreConfig{
		Endpoint:  "",
		AccessKey: "test",
		SecretKey: "test",
		Bucket:    "test",
	})
	assert.Error(t, err, "s3 factory should return error for empty endpoint")
}
