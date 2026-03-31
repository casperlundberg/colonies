package database

import (
	"fmt"
	"sort"
	"sync"
)

// DriverFactory is a constructor function that creates a Database from a DatabaseConfig.
type DriverFactory func(config DatabaseConfig) (Database, error)

var (
	driversMu sync.RWMutex
	drivers   = make(map[string]DriverFactory)
)

// Register makes a database driver available by the provided name.
// If Register is called twice with the same name, it panics.
func Register(name string, factory DriverFactory) {
	driversMu.Lock()
	defer driversMu.Unlock()

	if factory == nil {
		panic("database: Register factory is nil")
	}
	if _, exists := drivers[name]; exists {
		panic(fmt.Sprintf("database: driver already registered: %s", name))
	}
	drivers[name] = factory
}

// CreateFromRegistry creates a new Database using the named driver from the registry.
func CreateFromRegistry(name string, config DatabaseConfig) (Database, error) {
	driversMu.RLock()
	factory, ok := drivers[name]
	driversMu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("database: unknown driver %q (have you registered it?)", name)
	}
	return factory(config)
}

// RegisteredDrivers returns a sorted list of the names of the registered drivers.
func RegisteredDrivers() []string {
	driversMu.RLock()
	defer driversMu.RUnlock()

	names := make([]string, 0, len(drivers))
	for name := range drivers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
