package database

import (
	"sort"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// resetRegistry clears the global drivers map for test isolation.
func resetRegistry() {
	driversMu.Lock()
	defer driversMu.Unlock()
	drivers = make(map[string]DriverFactory)
}

func TestDatabaseRegisterAndCreate(t *testing.T) {
	resetRegistry()
	defer resetRegistry()

	called := false
	Register("testdriver-rac", func(config DatabaseConfig) (Database, error) {
		called = true
		return nil, nil
	})

	_, err := CreateFromRegistry("testdriver-rac", DatabaseConfig{})
	assert.NoError(t, err)
	assert.True(t, called, "factory should have been called")
}

func TestDatabaseRegisterPanicsOnDuplicate(t *testing.T) {
	resetRegistry()
	defer resetRegistry()

	factory := func(config DatabaseConfig) (Database, error) { return nil, nil }
	Register("dup-driver", factory)

	assert.Panics(t, func() {
		Register("dup-driver", factory)
	}, "registering the same name twice should panic")
}

func TestDatabaseRegisterPanicsOnNilFactory(t *testing.T) {
	resetRegistry()
	defer resetRegistry()

	assert.Panics(t, func() {
		Register("nil-factory-driver", nil)
	}, "registering a nil factory should panic")
}

func TestDatabaseRegisterPanicsOnEmptyName(t *testing.T) {
	resetRegistry()
	defer resetRegistry()

	assert.Panics(t, func() {
		Register("", func(config DatabaseConfig) (Database, error) { return nil, nil })
	}, "registering with an empty name should panic")
}

func TestDatabaseCreateUnknown(t *testing.T) {
	resetRegistry()
	defer resetRegistry()

	db, err := CreateFromRegistry("nonexistent-driver", DatabaseConfig{})
	assert.Nil(t, db)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent-driver")
}

func TestRegisteredDrivers(t *testing.T) {
	resetRegistry()
	defer resetRegistry()

	factory := func(config DatabaseConfig) (Database, error) { return nil, nil }
	Register("zeta-driver", factory)
	Register("alpha-driver", factory)

	got := RegisteredDrivers()
	expected := []string{"alpha-driver", "zeta-driver"}
	sort.Strings(expected)
	assert.Equal(t, expected, got, "RegisteredDrivers should return sorted names")
}

func TestDatabaseCreatePassesConfig(t *testing.T) {
	resetRegistry()
	defer resetRegistry()

	var received DatabaseConfig
	Register("config-check-driver", func(config DatabaseConfig) (Database, error) {
		received = config
		return nil, nil
	})

	cfg := DatabaseConfig{
		Type:        "customtype",
		Host:        "myhost.example.com",
		Port:        5433,
		User:        "testuser",
		Password:    "testpass",
		Name:        "testdb",
		Prefix:      "pfx_",
		TimescaleDB: true,
		DataDir:     "/tmp/data",
	}

	_, err := CreateFromRegistry("config-check-driver", cfg)
	assert.NoError(t, err)
	assert.Equal(t, cfg, received, "factory should receive the exact DatabaseConfig passed to CreateFromRegistry")
}

func TestDatabaseRegistryConcurrency(t *testing.T) {
	resetRegistry()
	defer resetRegistry()

	var wg sync.WaitGroup
	const n = 50

	// Register n drivers concurrently.
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			name := "concurrent-driver-" + string(rune('A'+id))
			Register(name, func(config DatabaseConfig) (Database, error) { return nil, nil })
		}(i)
	}
	wg.Wait()

	// Create from each driver concurrently.
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			name := "concurrent-driver-" + string(rune('A'+id))
			_, err := CreateFromRegistry(name, DatabaseConfig{})
			assert.NoError(t, err)
		}(i)
	}
	wg.Wait()

	registered := RegisteredDrivers()
	assert.Len(t, registered, n)
}
