package monitoring

// Monitor defines the interface for a monitoring service that collects
// and exposes metrics from a Colonies server.
type Monitor interface {
	Start()
	Stop()
}
