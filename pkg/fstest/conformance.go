package fstest

import (
	"bytes"
	"io"
	"sort"
	"strings"
	"testing"

	"github.com/colonyos/colonies/pkg/fs"
	"github.com/stretchr/testify/assert"
)

// HarnessMaker creates a fresh ObjectStore for testing and returns a cleanup
// function that tears down any resources when the test is done.
type HarnessMaker func(t *testing.T) (fs.ObjectStore, func())

// RunConformanceTests runs the full ObjectStore conformance test suite against
// the implementation provided by newHarness. Each subtest gets its own store
// instance so tests are fully isolated.
func RunConformanceTests(t *testing.T, newHarness HarnessMaker) {
	t.Run("Put/Get", func(t *testing.T) {
		store, cleanup := newHarness(t)
		defer cleanup()

		data := []byte("hello world")
		err := store.Put("test-colony", "test-object", bytes.NewReader(data), int64(len(data)))
		assert.NoError(t, err)

		rc, size, err := store.Get("test-colony", "test-object")
		assert.NoError(t, err)
		assert.Equal(t, int64(len(data)), size)
		defer rc.Close()

		got, err := io.ReadAll(rc)
		assert.NoError(t, err)
		assert.Equal(t, data, got)
	})

	t.Run("Put/Exists", func(t *testing.T) {
		store, cleanup := newHarness(t)
		defer cleanup()

		data := []byte("hello world")
		err := store.Put("test-colony", "test-object", bytes.NewReader(data), int64(len(data)))
		assert.NoError(t, err)

		assert.True(t, store.Exists("test-colony", "test-object"))
	})

	t.Run("ExistsNonexistent", func(t *testing.T) {
		store, cleanup := newHarness(t)
		defer cleanup()

		assert.False(t, store.Exists("test-colony", "no-such-object"))
	})

	t.Run("Put/Remove", func(t *testing.T) {
		store, cleanup := newHarness(t)
		defer cleanup()

		data := []byte("hello world")
		err := store.Put("test-colony", "test-object", bytes.NewReader(data), int64(len(data)))
		assert.NoError(t, err)
		assert.True(t, store.Exists("test-colony", "test-object"))

		err = store.Remove("test-colony", "test-object")
		assert.NoError(t, err)
		assert.False(t, store.Exists("test-colony", "test-object"))
	})

	t.Run("Put/Stat", func(t *testing.T) {
		store, cleanup := newHarness(t)
		defer cleanup()

		data := []byte("hello world")
		err := store.Put("test-colony", "test-object", bytes.NewReader(data), int64(len(data)))
		assert.NoError(t, err)

		info, err := store.Stat("test-colony", "test-object")
		assert.NoError(t, err)
		assert.Equal(t, int64(len(data)), info.Size)
	})

	t.Run("Put/List", func(t *testing.T) {
		store, cleanup := newHarness(t)
		defer cleanup()

		objects := []string{"alpha", "bravo", "charlie"}
		for _, name := range objects {
			data := []byte("data-" + name)
			err := store.Put("test-colony", name, strings.NewReader("data-"+name), int64(len(data)))
			assert.NoError(t, err)
		}

		names, err := store.List("test-colony")
		assert.NoError(t, err)

		sort.Strings(names)
		sort.Strings(objects)
		assert.Equal(t, objects, names)
	})

	t.Run("Put/GetRange", func(t *testing.T) {
		store, cleanup := newHarness(t)
		defer cleanup()

		data := []byte("hello world")
		err := store.Put("test-colony", "test-object", bytes.NewReader(data), int64(len(data)))
		assert.NoError(t, err)

		// Read "world" (offset=6, length=5)
		rc, err := store.GetRange("test-colony", "test-object", 6, 5)
		assert.NoError(t, err)
		defer rc.Close()

		got, err := io.ReadAll(rc)
		assert.NoError(t, err)
		assert.Equal(t, []byte("world"), got)
	})

	t.Run("RemoveAll", func(t *testing.T) {
		store, cleanup := newHarness(t)
		defer cleanup()

		for _, name := range []string{"obj1", "obj2", "obj3"} {
			data := []byte("content")
			err := store.Put("test-colony", name, bytes.NewReader(data), int64(len(data)))
			assert.NoError(t, err)
		}

		names, err := store.List("test-colony")
		assert.NoError(t, err)
		assert.Len(t, names, 3)

		err = store.RemoveAll("test-colony")
		assert.NoError(t, err)

		names, err = store.List("test-colony")
		assert.NoError(t, err)
		assert.Empty(t, names)
	})

	t.Run("DiskUsage", func(t *testing.T) {
		store, cleanup := newHarness(t)
		defer cleanup()

		data := []byte("hello world")
		err := store.Put("test-colony", "test-object", bytes.NewReader(data), int64(len(data)))
		assert.NoError(t, err)

		usage, err := store.DiskUsage("test-colony")
		assert.NoError(t, err)
		assert.True(t, usage > 0, "expected non-zero disk usage, got %d", usage)
	})

	t.Run("ChunkedUpload", func(t *testing.T) {
		store, cleanup := newHarness(t)
		defer cleanup()

		chunk0 := []byte("hello ")
		chunk1 := []byte("world")
		chunk2 := []byte("!")

		err := store.PutChunk("test-colony", "test-object", 0, bytes.NewReader(chunk0), int64(len(chunk0)))
		assert.NoError(t, err)

		err = store.PutChunk("test-colony", "test-object", 1, bytes.NewReader(chunk1), int64(len(chunk1)))
		assert.NoError(t, err)

		err = store.PutChunk("test-colony", "test-object", 2, bytes.NewReader(chunk2), int64(len(chunk2)))
		assert.NoError(t, err)

		err = store.AssembleChunks("test-colony", "test-object", 3)
		assert.NoError(t, err)

		rc, size, err := store.Get("test-colony", "test-object")
		assert.NoError(t, err)
		defer rc.Close()

		expected := []byte("hello world!")
		assert.Equal(t, int64(len(expected)), size)

		got, err := io.ReadAll(rc)
		assert.NoError(t, err)
		assert.Equal(t, expected, got)
	})
}
