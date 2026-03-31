package monitoring

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockMonitor struct{}

func (m *mockMonitor) Start() {}
func (m *mockMonitor) Stop()  {}

// resetFactories clears the global registry between tests.
func resetFactories() {
	mu.Lock()
	defer mu.Unlock()
	factories = make(map[string]MonitorFactory)
}

func TestMonitorRegisterAndCreate(t *testing.T) {
	resetFactories()

	called := false
	factory := func(config MonitorConfig) (Monitor, error) {
		called = true
		return &mockMonitor{}, nil
	}

	err := Register("reg-and-create", factory)
	assert.NoError(t, err)

	mon, err := Create("reg-and-create", MonitorConfig{})
	assert.NoError(t, err)
	assert.NotNil(t, mon)
	assert.True(t, called)
}

func TestMonitorRegisterDuplicateReturnsError(t *testing.T) {
	resetFactories()

	factory := func(config MonitorConfig) (Monitor, error) {
		return &mockMonitor{}, nil
	}

	err := Register("dup-test", factory)
	assert.NoError(t, err)

	err = Register("dup-test", factory)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func TestMonitorRegisterNilFactoryReturnsError(t *testing.T) {
	resetFactories()

	err := Register("nil-factory-test", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "factory must not be nil")
}

func TestMonitorRegisterEmptyNameReturnsError(t *testing.T) {
	resetFactories()

	factory := func(config MonitorConfig) (Monitor, error) {
		return &mockMonitor{}, nil
	}

	err := Register("", factory)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name must not be empty")
}

func TestMonitorCreateUnknown(t *testing.T) {
	resetFactories()

	mon, err := Create("nonexistent", MonitorConfig{})
	assert.Error(t, err)
	assert.Nil(t, mon)
	assert.Contains(t, err.Error(), "unknown monitor factory")
}

func TestRegisteredMonitors(t *testing.T) {
	resetFactories()

	factory := func(config MonitorConfig) (Monitor, error) {
		return &mockMonitor{}, nil
	}

	err := Register("zulu", factory)
	assert.NoError(t, err)
	err = Register("alpha", factory)
	assert.NoError(t, err)

	names := RegisteredMonitors()
	assert.Equal(t, []string{"alpha", "zulu"}, names)
}

func TestMonitorCreatePassesConfig(t *testing.T) {
	resetFactories()

	var received MonitorConfig
	factory := func(config MonitorConfig) (Monitor, error) {
		received = config
		return &mockMonitor{}, nil
	}

	err := Register("config-pass", factory)
	assert.NoError(t, err)

	cfg := MonitorConfig{
		Port:               9090,
		ColoniesServerHost: "example.com",
		ColoniesServerPort: 8080,
		Insecure:           true,
		SkipTLSVerify:      true,
		ServerPrvKey:       "secret-key-123",
		PullInterval:       30,
	}

	_, err = Create("config-pass", cfg)
	assert.NoError(t, err)
	assert.Equal(t, cfg, received)
}

func TestMonitorRegistryConcurrency(t *testing.T) {
	resetFactories()

	factory := func(config MonitorConfig) (Monitor, error) {
		return &mockMonitor{}, nil
	}

	var wg sync.WaitGroup
	// Register 50 unique factories concurrently.
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			name := "concurrent-" + string(rune('A'+idx))
			_ = Register(name, factory)
		}(i)
	}

	// Concurrently try to create from some of them.
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			name := "concurrent-" + string(rune('A'+idx))
			_, _ = Create(name, MonitorConfig{})
		}(i)
	}

	wg.Wait()

	names := RegisteredMonitors()
	assert.True(t, len(names) > 0)
}
