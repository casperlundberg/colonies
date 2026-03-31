package monitoring

import (
	monitoringpkg "github.com/colonyos/colonies/pkg/monitoring"
)

// Start implements the monitoring.Monitor interface by being a no-op.
// MonitoringServer starts its HTTP metrics server and polling goroutine
// inside CreateMonitoringServer, so no additional startup step is needed.
func (server *MonitoringServer) Start() {}

// Stop implements the monitoring.Monitor interface by being a no-op.
// TODO: MonitoringServer currently has no shutdown mechanism. The HTTP
// server and polling goroutine started in CreateMonitoringServer run
// until the process exits. To fully implement Stop, the constructor
// should store the *http.Server and use a cancel context for the
// polling loop so they can be torn down gracefully here.
func (server *MonitoringServer) Stop() {}

// Compile-time check that MonitoringServer satisfies Monitor.
var _ monitoringpkg.Monitor = (*MonitoringServer)(nil)

func init() {
	_ = monitoringpkg.Register("prometheus", func(config monitoringpkg.MonitorConfig) (monitoringpkg.Monitor, error) {
		server := CreateMonitoringServer(
			config.Port,
			config.ColoniesServerHost,
			config.ColoniesServerPort,
			config.Insecure,
			config.SkipTLSVerify,
			config.ServerPrvKey,
			config.PullInterval,
		)
		return server, nil
	})
}
