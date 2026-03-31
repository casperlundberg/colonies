package etcd

import (
	"testing"

	"github.com/colonyos/colonies/pkg/cluster"
	"github.com/stretchr/testify/assert"
)

func TestEtcdRegistersOnImport(t *testing.T) {
	names := cluster.RegisteredClusters()
	found := false
	for _, name := range names {
		if name == "etcd" {
			found = true
			break
		}
	}
	assert.True(t, found, "etcd should be registered via init()")
}
