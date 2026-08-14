package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestRPCSetProcessPrioritiesMsg(t *testing.T) {
	updates := []core.PriorityUpdate{
		{ProcessID: core.GenerateRandomID(), Priority: 0},
		{ProcessID: core.GenerateRandomID(), Priority: 25},
	}
	msg := CreateSetProcessPrioritiesMsg("test_colony_name", updates)
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	_, err = CreateSetProcessPrioritiesMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err := CreateSetProcessPrioritiesMsgFromJSON(jsonString)
	assert.Nil(t, err)
	assert.True(t, msg.Equals(msg2))
}

func TestRPCSetProcessPrioritiesMsgIndent(t *testing.T) {
	msg := CreateSetProcessPrioritiesMsg("test_colony_name", []core.PriorityUpdate{{ProcessID: core.GenerateRandomID(), Priority: 0}})
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateSetProcessPrioritiesMsgFromJSON(jsonString)
	assert.Nil(t, err)
	assert.True(t, msg.Equals(msg2))
}

func TestRPCSetProcessPrioritiesMsgEquals(t *testing.T) {
	id1 := core.GenerateRandomID()
	id2 := core.GenerateRandomID()

	msg := CreateSetProcessPrioritiesMsg("test_colony_name", []core.PriorityUpdate{{ProcessID: id1, Priority: 0}, {ProcessID: id2, Priority: 25}})
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))

	assert.False(t, msg.Equals(CreateSetProcessPrioritiesMsg("another_colony_name", msg.Updates)))
	assert.False(t, msg.Equals(CreateSetProcessPrioritiesMsg("test_colony_name", []core.PriorityUpdate{{ProcessID: id1, Priority: 0}})))

	// The priority carried per process is part of the message, not just the id set.
	assert.False(t, msg.Equals(CreateSetProcessPrioritiesMsg("test_colony_name", []core.PriorityUpdate{{ProcessID: id1, Priority: 0}, {ProcessID: id2, Priority: 50}})))

	// So is the order, since the reply is positional.
	assert.False(t, msg.Equals(CreateSetProcessPrioritiesMsg("test_colony_name", []core.PriorityUpdate{{ProcessID: id2, Priority: 25}, {ProcessID: id1, Priority: 0}})))
}
