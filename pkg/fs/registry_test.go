package fs

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// resetFactories clears the global registry so tests are isolated.
func resetFactories() {
	mu.Lock()
	defer mu.Unlock()
	factories = make(map[string]ObjectStoreFactory)
}

func dummyFactory(config ObjectStoreConfig) (ObjectStore, error) {
	return nil, nil
}

func TestObjectStoreRegisterAndCreate(t *testing.T) {
	resetFactories()

	called := false
	factory := func(config ObjectStoreConfig) (ObjectStore, error) {
		called = true
		return nil, nil
	}

	RegisterObjectStore("test-register-create", factory)
	store, err := CreateObjectStore("test-register-create", ObjectStoreConfig{})
	assert.NoError(t, err)
	assert.Nil(t, store)
	assert.True(t, called)
}

func TestObjectStoreRegisterPanicsOnDuplicate(t *testing.T) {
	resetFactories()

	RegisterObjectStore("test-dup", dummyFactory)
	assert.Panics(t, func() {
		RegisterObjectStore("test-dup", dummyFactory)
	})
}

func TestObjectStoreRegisterPanicsOnNilFactory(t *testing.T) {
	resetFactories()

	assert.Panics(t, func() {
		RegisterObjectStore("test-nil-factory", nil)
	})
}

func TestObjectStoreRegisterPanicsOnEmptyName(t *testing.T) {
	resetFactories()

	assert.Panics(t, func() {
		RegisterObjectStore("", dummyFactory)
	})
}

func TestObjectStoreCreateUnknown(t *testing.T) {
	resetFactories()

	store, err := CreateObjectStore("nonexistent", ObjectStoreConfig{})
	assert.Nil(t, store)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown object store")
}

func TestRegisteredObjectStores(t *testing.T) {
	resetFactories()

	RegisterObjectStore("beta-store", dummyFactory)
	RegisterObjectStore("alpha-store", dummyFactory)

	names := RegisteredObjectStores()
	assert.Equal(t, []string{"alpha-store", "beta-store"}, names)
}

func TestObjectStoreCreatePassesConfig(t *testing.T) {
	resetFactories()

	var received ObjectStoreConfig
	factory := func(config ObjectStoreConfig) (ObjectStore, error) {
		received = config
		return nil, nil
	}

	RegisterObjectStore("test-config-pass", factory)

	cfg := ObjectStoreConfig{
		Type:      "s3",
		Dir:       "/data",
		Endpoint:  "localhost:9000",
		AccessKey: "AKID",
		SecretKey: "SECRET",
		Bucket:    "my-bucket",
		TLS:       true,
	}

	_, err := CreateObjectStore("test-config-pass", cfg)
	assert.NoError(t, err)

	assert.Equal(t, "s3", received.Type)
	assert.Equal(t, "/data", received.Dir)
	assert.Equal(t, "localhost:9000", received.Endpoint)
	assert.Equal(t, "AKID", received.AccessKey)
	assert.Equal(t, "SECRET", received.SecretKey)
	assert.Equal(t, "my-bucket", received.Bucket)
	assert.True(t, received.TLS)
}

func TestObjectStoreRegistryConcurrency(t *testing.T) {
	resetFactories()

	const n = 50
	var wg sync.WaitGroup
	wg.Add(n * 2)

	// Parallel registers with unique names.
	for i := 0; i < n; i++ {
		name := "concurrent-" + string(rune('A'+i))
		go func() {
			defer wg.Done()
			RegisterObjectStore(name, dummyFactory)
		}()
	}

	// Parallel creates (may or may not find the name yet -- we just check no race).
	for i := 0; i < n; i++ {
		name := "concurrent-" + string(rune('A'+i))
		go func() {
			defer wg.Done()
			_, _ = CreateObjectStore(name, ObjectStoreConfig{})
		}()
	}

	wg.Wait()

	names := RegisteredObjectStores()
	assert.Len(t, names, n)
}
