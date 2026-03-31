package localfs_test

import (
	"os"
	"testing"

	"github.com/colonyos/colonies/pkg/fs"
	"github.com/colonyos/colonies/pkg/fstest"
	"github.com/colonyos/colonies/plugin/localfs"
)

func TestObjectStoreConformance(t *testing.T) {
	fstest.RunConformanceTests(t, func(t *testing.T) (fs.ObjectStore, func()) {
		t.Helper()
		dir, err := os.MkdirTemp("", "conformance-localfs-*")
		if err != nil {
			t.Fatal(err)
		}
		store, err := localfs.NewLocalObjectStore(dir)
		if err != nil {
			t.Fatal(err)
		}
		return store, func() {
			os.RemoveAll(dir)
		}
	})
}
