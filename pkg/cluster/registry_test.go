package cluster

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockCluster is a minimal no-op implementation of the Cluster interface.
type mockCluster struct{}

func (m *mockCluster) Start()                                            {}
func (m *mockCluster) Stop()                                             {}
func (m *mockCluster) WaitToStart()                                      {}
func (m *mockCluster) WaitToStop()                                       {}
func (m *mockCluster) Leader() string                                    { return "" }
func (m *mockCluster) Members() []Node                                   { return nil }
func (m *mockCluster) CurrentCluster() Config                            { return Config{} }
func (m *mockCluster) StorageDir() string                                { return "" }
func (m *mockCluster) PauseColonyAssignments(colonyName string) error    { return nil }
func (m *mockCluster) ResumeColonyAssignments(colonyName string) error   { return nil }
func (m *mockCluster) AreColonyAssignmentsPaused(colonyName string) (bool, error) {
	return false, nil
}

// resetRegistry clears the global registry so tests are isolated.
func resetRegistry() {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry = make(map[string]ClusterFactory)
}

func TestClusterRegisterAndCreate(t *testing.T) {
	resetRegistry()

	mock := &mockCluster{}
	Register("registerAndCreate", func(thisNode Node, config Config, dataPath string) Cluster {
		return mock
	})

	c, err := Create("registerAndCreate", Node{}, Config{}, "/tmp")
	assert.NoError(t, err)
	assert.Equal(t, mock, c)
}

func TestClusterRegisterPanicsOnDuplicate(t *testing.T) {
	resetRegistry()

	factory := func(thisNode Node, config Config, dataPath string) Cluster {
		return &mockCluster{}
	}
	Register("duplicate", factory)

	assert.Panics(t, func() {
		Register("duplicate", factory)
	})
}

func TestClusterRegisterPanicsOnNilFactory(t *testing.T) {
	resetRegistry()

	assert.Panics(t, func() {
		Register("nilFactory", nil)
	})
}

func TestClusterRegisterPanicsOnEmptyName(t *testing.T) {
	resetRegistry()

	assert.Panics(t, func() {
		Register("", func(thisNode Node, config Config, dataPath string) Cluster {
			return &mockCluster{}
		})
	})
}

func TestClusterCreateUnknown(t *testing.T) {
	resetRegistry()

	c, err := Create("nonexistent", Node{}, Config{}, "/tmp")
	assert.Error(t, err)
	assert.Nil(t, c)
}

func TestRegisteredClusters(t *testing.T) {
	resetRegistry()

	Register("zeta", func(thisNode Node, config Config, dataPath string) Cluster {
		return &mockCluster{}
	})
	Register("alpha", func(thisNode Node, config Config, dataPath string) Cluster {
		return &mockCluster{}
	})

	names := RegisteredClusters()
	assert.Equal(t, []string{"alpha", "zeta"}, names)
}

func TestClusterCreatePassesArguments(t *testing.T) {
	resetRegistry()

	var capturedNode Node
	var capturedConfig Config
	var capturedDataPath string

	Register("passArgs", func(thisNode Node, config Config, dataPath string) Cluster {
		capturedNode = thisNode
		capturedConfig = config
		capturedDataPath = dataPath
		return &mockCluster{}
	})

	node := Node{Name: "node1", Host: "localhost", EtcdClientPort: 2379}
	config := Config{Leader: node}
	dataPath := "/data/test"

	_, err := Create("passArgs", node, config, dataPath)
	assert.NoError(t, err)
	assert.Equal(t, node, capturedNode)
	assert.Equal(t, config.Leader, capturedConfig.Leader)
	assert.Equal(t, dataPath, capturedDataPath)
}

func TestClusterRegistryConcurrency(t *testing.T) {
	resetRegistry()

	var wg sync.WaitGroup
	count := 50

	// Register factories concurrently with unique names.
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			name := "concurrent-" + string(rune('A'+idx))
			Register(name, func(thisNode Node, config Config, dataPath string) Cluster {
				return &mockCluster{}
			})
		}(i)
	}
	wg.Wait()

	// Create clusters concurrently.
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			name := "concurrent-" + string(rune('A'+idx))
			c, err := Create(name, Node{}, Config{}, "/tmp")
			assert.NoError(t, err)
			assert.NotNil(t, c)
		}(i)
	}
	wg.Wait()
}
