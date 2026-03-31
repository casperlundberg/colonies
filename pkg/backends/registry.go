package backends

import (
	"fmt"
	"sort"
	"sync"
)

// BackendFactory creates backend components for a specific HTTP framework implementation.
type BackendFactory interface {
	CreateEngine() Engine
	CreateServer(port int, engine Engine) Server
	CreateRealtimeHandler() RealtimeEventHandler
	CreateTestableRealtimeHandler() TestableRealtimeEventHandler
	ConfigureSilentMode()
	CORS() HandlerFunc
}

var (
	registryMu sync.RWMutex
	registry   = make(map[string]BackendFactory)
)

// RegisterBackend registers a BackendFactory under the given name.
// It panics if name is empty or a factory with the same name is already registered.
func RegisterBackend(name string, factory BackendFactory) {
	if name == "" {
		panic("backends: RegisterBackend name must not be empty")
	}
	if factory == nil {
		panic("backends: RegisterBackend factory must not be nil")
	}

	registryMu.Lock()
	defer registryMu.Unlock()

	if _, exists := registry[name]; exists {
		panic(fmt.Sprintf("backends: backend %q already registered", name))
	}
	registry[name] = factory
}

// GetBackend returns the BackendFactory registered under the given name.
func GetBackend(name string) (BackendFactory, error) {
	registryMu.RLock()
	defer registryMu.RUnlock()

	factory, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("backends: unknown backend %q", name)
	}
	return factory, nil
}

// RegisteredBackends returns a sorted list of registered backend names.
func RegisteredBackends() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()

	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
