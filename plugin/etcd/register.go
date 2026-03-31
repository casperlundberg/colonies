package etcd

import "github.com/colonyos/colonies/pkg/cluster"

func init() {
	cluster.Register("etcd", func(thisNode cluster.Node, config cluster.Config, dataPath string) cluster.Cluster {
		return CreateEtcdServer(thisNode, config, dataPath)
	})
}
