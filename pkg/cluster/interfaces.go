package cluster

// Cluster defines the interface for cluster coordination and leader election.
// EtcdServer is the primary implementation.
type Cluster interface {
	// Start begins the cluster node. It launches the underlying coordination
	// service in a background goroutine.
	Start()

	// Stop signals the cluster node to shut down.
	Stop()

	// WaitToStart blocks until the cluster node is ready to serve.
	WaitToStart()

	// WaitToStop blocks until the cluster node has fully stopped.
	WaitToStop()

	// Leader returns the name of the current cluster leader.
	Leader() string

	// Members returns the set of nodes currently participating in the cluster.
	Members() []Node

	// CurrentCluster returns the full cluster configuration including the
	// current leader.
	CurrentCluster() Config

	// StorageDir returns the path to the local data directory used by this
	// cluster node.
	StorageDir() string

	// PauseColonyAssignments marks a colony so that process assignments are
	// temporarily suspended across the cluster.
	PauseColonyAssignments(colonyName string) error

	// ResumeColonyAssignments removes the pause marker for a colony, allowing
	// process assignments to resume.
	ResumeColonyAssignments(colonyName string) error

	// AreColonyAssignmentsPaused reports whether process assignments for the
	// given colony are currently paused.
	AreColonyAssignmentsPaused(colonyName string) (bool, error)
}

// RelayMessage contains the message data and a channel to signal processing completion.
type RelayMessage struct {
	Data []byte
	Done chan struct{}
}

// MessageHandler is a function that processes relay messages.
type MessageHandler func(data []byte)

// Relay defines the interface for broadcasting messages between cluster nodes.
// RelayServer is the primary implementation.
type Relay interface {
	// Broadcast sends a message to every other node in the cluster.
	Broadcast(msg []byte) error

	// Subscribe registers a handler that is called synchronously for each
	// incoming relay message.
	Subscribe(handler MessageHandler)

	// Receive returns a channel that yields incoming relay messages. It is a
	// convenience wrapper around Subscribe for consumers that prefer a
	// channel-based API.
	Receive() chan RelayMessage

	// Shutdown gracefully stops the relay server.
	Shutdown()
}
