package objectstore

import "github.com/colonyos/colonies/pkg/fs"

func init() {
	fs.RegisterObjectStore("s3", func(config fs.ObjectStoreConfig) (fs.ObjectStore, error) {
		return NewS3ObjectStore(config)
	})
}
