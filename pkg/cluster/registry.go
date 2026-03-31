package cluster

import (
	"fmt"
	"sort"
	"sync"
)

// ClusterFactory is a constructor function that creates a Cluster from a
// node identity, cluster configuration, and a data-storage path.
type ClusterFactory func(thisNode Node, config Config, dataPath string) Cluster

var (
	registryMu sync.RWMutex
	registry   = make(map[string]ClusterFactory)
)

// Register adds a named ClusterFactory to the global registry. It panics if
// a factory with the same name has already been registered.
func Register(name string, factory ClusterFactory) {
	if name == "" {
		panic("cluster: Register name must not be empty")
	}
	if factory == nil {
		panic("cluster: Register factory must not be nil")
	}

	registryMu.Lock()
	defer registryMu.Unlock()

	if _, exists := registry[name]; exists {
		panic(fmt.Sprintf("cluster: factory already registered for name %q", name))
	}
	registry[name] = factory
}

// Create looks up the named factory in the registry and uses it to build a
// new Cluster. It returns an error if no factory with that name has been
// registered.
func Create(name string, thisNode Node, config Config, dataPath string) (Cluster, error) {
	registryMu.RLock()
	factory, ok := registry[name]
	registryMu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("cluster: no factory registered for name %q", name)
	}
	return factory(thisNode, config, dataPath), nil
}

// RegisteredClusters returns a slice of all factory names that have been
// registered.
func RegisteredClusters() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()

	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
