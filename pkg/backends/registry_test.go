package backends

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

// mockBackendFactory is a minimal implementation of BackendFactory for testing.
type mockBackendFactory struct{}

func (m *mockBackendFactory) CreateEngine() Engine                                    { return nil }
func (m *mockBackendFactory) CreateServer(port int, engine Engine) Server             { return nil }
func (m *mockBackendFactory) CreateRealtimeHandler() RealtimeEventHandler             { return nil }
func (m *mockBackendFactory) CreateTestableRealtimeHandler() TestableRealtimeEventHandler { return nil }
func (m *mockBackendFactory) ConfigureSilentMode()                                    {}
func (m *mockBackendFactory) CORS() HandlerFunc                                       { return nil }

// Ensure the mock satisfies the interface at compile time.
var _ BackendFactory = (*mockBackendFactory)(nil)

// altMockBackendFactory is a second distinct factory type used to verify correct dispatch.
type altMockBackendFactory struct {
	tag string
}

func (a *altMockBackendFactory) CreateEngine() Engine                                    { return nil }
func (a *altMockBackendFactory) CreateServer(port int, engine Engine) Server             { return nil }
func (a *altMockBackendFactory) CreateRealtimeHandler() RealtimeEventHandler             { return nil }
func (a *altMockBackendFactory) CreateTestableRealtimeHandler() TestableRealtimeEventHandler { return nil }
func (a *altMockBackendFactory) ConfigureSilentMode()                                    {}
func (a *altMockBackendFactory) CORS() HandlerFunc                                       { return nil }

// Ensure the alt mock satisfies the interface at compile time.
var _ BackendFactory = (*altMockBackendFactory)(nil)

// mockRealtimeEventHandler satisfies RealtimeEventHandler for compile-time checks only.
type mockRealtimeEventHandler struct{}

func (m *mockRealtimeEventHandler) Signal(process *core.Process) {}
func (m *mockRealtimeEventHandler) Subscribe(executorType string, state int, processID string, location string, ctx context.Context) (chan *core.Process, chan error) {
	return nil, nil
}
func (m *mockRealtimeEventHandler) WaitForProcess(executorType string, state int, processID string, location string, ctx context.Context) (*core.Process, error) {
	return nil, nil
}
func (m *mockRealtimeEventHandler) Stop() {}

func TestRegisterAndGetBackend(t *testing.T) {
	name := "TestRegisterAndGetBackend_factory"
	factory := &mockBackendFactory{}
	RegisterBackend(name, factory)

	got, err := GetBackend(name)
	assert.NoError(t, err)
	assert.Same(t, factory, got)
}

func TestRegisterBackendPanicsOnDuplicate(t *testing.T) {
	name := "TestRegisterBackendPanicsOnDuplicate_factory"
	RegisterBackend(name, &mockBackendFactory{})

	assert.Panics(t, func() {
		RegisterBackend(name, &mockBackendFactory{})
	})
}

func TestRegisterBackendPanicsOnNilFactory(t *testing.T) {
	assert.Panics(t, func() {
		RegisterBackend("TestRegisterBackendPanicsOnNilFactory_factory", nil)
	})
}

func TestRegisterBackendPanicsOnEmptyName(t *testing.T) {
	assert.Panics(t, func() {
		RegisterBackend("", &mockBackendFactory{})
	})
}

func TestGetBackendUnknown(t *testing.T) {
	got, err := GetBackend("totally_nonexistent_backend_xyz")
	assert.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "unknown backend")
}

func TestRegisteredBackends(t *testing.T) {
	name1 := "TestRegisteredBackends_alpha"
	name2 := "TestRegisteredBackends_beta"
	RegisterBackend(name1, &mockBackendFactory{})
	RegisterBackend(name2, &mockBackendFactory{})

	names := RegisteredBackends()
	assert.Contains(t, names, name1)
	assert.Contains(t, names, name2)

	// Verify the list is sorted.
	for i := 1; i < len(names); i++ {
		assert.True(t, names[i-1] <= names[i], "RegisteredBackends should return a sorted list")
	}
}

func TestGetBackendReturnsCorrectFactory(t *testing.T) {
	name1 := "TestGetBackendReturnsCorrectFactory_one"
	name2 := "TestGetBackendReturnsCorrectFactory_two"
	f1 := &altMockBackendFactory{tag: "one"}
	f2 := &altMockBackendFactory{tag: "two"}

	RegisterBackend(name1, f1)
	RegisterBackend(name2, f2)

	got1, err1 := GetBackend(name1)
	assert.NoError(t, err1)
	assert.Same(t, f1, got1)

	got2, err2 := GetBackend(name2)
	assert.NoError(t, err2)
	assert.Same(t, f2, got2)
}

func TestBackendRegistryConcurrency(t *testing.T) {
	const n = 50
	var wg sync.WaitGroup

	// N goroutines registering unique names.
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			name := fmt.Sprintf("TestBackendRegistryConcurrency_%d", idx)
			RegisterBackend(name, &mockBackendFactory{})
		}(i)
	}

	// N goroutines looking up (some may exist, some may not yet).
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			name := fmt.Sprintf("TestBackendRegistryConcurrency_%d", idx)
			// We don't care about the result; we just verify no data race or panic.
			_, _ = GetBackend(name)
		}(i)
	}

	// Also read the full list concurrently.
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = RegisteredBackends()
		}()
	}

	wg.Wait()

	// After all goroutines finish, every name should be present.
	for i := 0; i < n; i++ {
		name := fmt.Sprintf("TestBackendRegistryConcurrency_%d", i)
		got, err := GetBackend(name)
		assert.NoError(t, err, "expected %s to be registered", name)
		assert.NotNil(t, got)
	}
}
