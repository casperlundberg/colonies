package localfs

import "github.com/colonyos/colonies/pkg/fs"

func init() {
	fs.RegisterObjectStore("coloniesfs", func(config fs.ObjectStoreConfig) (fs.ObjectStore, error) {
		return NewLocalObjectStore(config.Dir)
	})
}
