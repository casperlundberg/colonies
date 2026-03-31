package monitoring

import (
	"fmt"
	"sort"
	"sync"
)

// MonitorConfig holds the configuration needed to create a Monitor.
type MonitorConfig struct {
	Port               int
	ColoniesServerHost string
	ColoniesServerPort int
	Insecure           bool
	SkipTLSVerify      bool
	ServerPrvKey       string
	PullInterval       int
}

// MonitorFactory is a function that creates a Monitor from a MonitorConfig.
type MonitorFactory func(config MonitorConfig) (Monitor, error)

var (
	mu        sync.RWMutex
	factories = make(map[string]MonitorFactory)
)

// Register adds a MonitorFactory under the given name.
// It returns an error if a factory with that name is already registered.
func Register(name string, factory MonitorFactory) error {
	if name == "" {
		return fmt.Errorf("monitoring: Register name must not be empty")
	}
	if factory == nil {
		return fmt.Errorf("monitoring: Register factory must not be nil")
	}

	mu.Lock()
	defer mu.Unlock()

	if _, exists := factories[name]; exists {
		return fmt.Errorf("monitor factory already registered: %s", name)
	}

	factories[name] = factory
	return nil
}

// Create instantiates a Monitor using the factory registered under the given name.
func Create(name string, config MonitorConfig) (Monitor, error) {
	mu.RLock()
	factory, exists := factories[name]
	mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("unknown monitor factory: %s", name)
	}

	return factory(config)
}

// RegisteredMonitors returns the names of all registered monitor factories.
func RegisteredMonitors() []string {
	mu.RLock()
	defer mu.RUnlock()

	names := make([]string, 0, len(factories))
	for name := range factories {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
