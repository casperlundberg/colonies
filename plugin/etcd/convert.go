package etcd

import "github.com/colonyos/colonies/pkg/cluster"

// NodeFromCluster converts a cluster.Node to an etcd.Node.
// Since etcd.Node is a type alias for cluster.Node, this is a direct
// assignment, but the helper makes the intent explicit at call sites.
func NodeFromCluster(n cluster.Node) Node {
	return Node{
		Name:           n.Name,
		Host:           n.Host,
		EtcdClientPort: n.EtcdClientPort,
		EtcdPeerPort:   n.EtcdPeerPort,
		RelayPort:      n.RelayPort,
		APIPort:        n.APIPort,
	}
}

// ConfigFromCluster converts a cluster.Config to an etcd.Config by
// converting each node in the config.
func ConfigFromCluster(c cluster.Config) Config {
	cfg := Config{}
	for _, n := range c.Nodes {
		cfg.AddNode(NodeFromCluster(n))
	}
	return cfg
}
