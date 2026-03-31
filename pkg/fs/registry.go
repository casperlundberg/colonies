package fs

import (
	"fmt"
	"sort"
	"sync"
)

// ObjectStoreConfig holds configuration for creating an ObjectStore.
type ObjectStoreConfig struct {
	Type      string
	Dir       string
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	TLS       bool
}

// ObjectStoreFactory is a constructor function for an ObjectStore.
type ObjectStoreFactory func(config ObjectStoreConfig) (ObjectStore, error)

var (
	mu        sync.RWMutex
	factories = make(map[string]ObjectStoreFactory)
)

// RegisterObjectStore registers a factory under the given name.
func RegisterObjectStore(name string, factory ObjectStoreFactory) {
	mu.Lock()
	defer mu.Unlock()
	factories[name] = factory
}

// CreateObjectStore creates an ObjectStore using the factory registered under name.
func CreateObjectStore(name string, config ObjectStoreConfig) (ObjectStore, error) {
	mu.RLock()
	factory, ok := factories[name]
	mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unknown object store: %s", name)
	}
	return factory(config)
}

// RegisteredObjectStores returns a sorted list of registered store names.
func RegisteredObjectStores() []string {
	mu.RLock()
	defer mu.RUnlock()
	names := make([]string, 0, len(factories))
	for name := range factories {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
