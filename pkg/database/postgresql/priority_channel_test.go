package postgresql

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

// The priority channel: a bulk, bounded update of the priority of processes
// that are still WAITING.
//
// ColonyOS stores priority as a derived ordering key,
//
//	PRIORITYTIME = PRIORITY * dt + SUBMISSION_TIME.UnixNano(),  dt = -86_400e9
//
// and the assign path selects with ORDER BY PRIORITYTIME. So changing a
// priority means rewriting both PRIORITY and PRIORITYTIME together; writing
// only one leaves the queue order disagreeing with the reported priority.
//
// Semantics under test:
//   - WAITING only. Assigned/running processes are never reordered.
//   - Bounded by [floor, ceiling]; the ceiling defaults to the priority the
//     process was submitted with, so nothing escalates above where it started.
//   - Bulk, in one transaction, with a per-process outcome returned so the
//     caller can tell how many writes landed.

func TestSetProcessPrioritiesReordersWaiting(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	high := utils.CreateTestProcess(colony.Name)
	high.FunctionSpec.Priority = 100
	assert.Nil(t, db.AddProcess(high))

	low := utils.CreateTestProcess(colony.Name)
	low.FunctionSpec.Priority = 50
	assert.Nil(t, db.AddProcess(low))

	before, err := db.GetProcessByID(high.ID)
	assert.Nil(t, err)

	results, err := db.SetProcessPriorities(colony.Name, []core.PriorityUpdate{
		{ProcessID: high.ID, Priority: 0},
	})
	assert.Nil(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, core.PriorityUpdated, results[0].Outcome)

	after, err := db.GetProcessByID(high.ID)
	assert.Nil(t, err)

	// Both the reported priority and the derived ordering key must move.
	assert.Equal(t, 0, after.FunctionSpec.Priority)
	assert.True(t, after.PriorityTime > before.PriorityTime,
		"PRIORITYTIME must increase when priority drops, or the queue order "+
			"will disagree with the reported priority")
}

func TestSetProcessPrioritiesEscalates(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	process := utils.CreateTestProcess(colony.Name)
	process.FunctionSpec.Priority = 25
	assert.Nil(t, db.AddProcess(process))

	// Decay first, then bring it back up: the channel is bidirectional but
	// bounded by the submission priority, so 25 must be reachable again.
	_, err = db.SetProcessPriorities(colony.Name, []core.PriorityUpdate{{ProcessID: process.ID, Priority: 0}})
	assert.Nil(t, err)

	results, err := db.SetProcessPriorities(colony.Name, []core.PriorityUpdate{
		{ProcessID: process.ID, Priority: 25},
	})
	assert.Nil(t, err)
	assert.Equal(t, core.PriorityUpdated, results[0].Outcome)

	after, err := db.GetProcessByID(process.ID)
	assert.Nil(t, err)
	assert.Equal(t, 25, after.FunctionSpec.Priority)
}

func TestSetProcessPrioritiesRejectsAboveCeiling(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	process := utils.CreateTestProcess(colony.Name)
	process.FunctionSpec.Priority = 25
	assert.Nil(t, db.AddProcess(process))

	// Nothing may be escalated above the priority it was submitted with.
	results, err := db.SetProcessPriorities(colony.Name, []core.PriorityUpdate{
		{ProcessID: process.ID, Priority: 400},
	})
	assert.Nil(t, err)
	assert.Equal(t, core.PriorityOutOfBounds, results[0].Outcome)

	after, err := db.GetProcessByID(process.ID)
	assert.Nil(t, err)
	assert.Equal(t, 25, after.FunctionSpec.Priority, "rejected write must change nothing")
}

func TestSetProcessPrioritiesIgnoresRunning(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	process := utils.CreateTestProcess(colony.Name)
	process.FunctionSpec.Priority = 100
	assert.Nil(t, db.AddProcess(process))
	assert.Nil(t, db.SetProcessState(process.ID, core.RUNNING))

	before, err := db.GetProcessByID(process.ID)
	assert.Nil(t, err)

	results, err := db.SetProcessPriorities(colony.Name, []core.PriorityUpdate{
		{ProcessID: process.ID, Priority: 0},
	})
	assert.Nil(t, err)
	assert.Equal(t, core.PriorityNotWaiting, results[0].Outcome)

	after, err := db.GetProcessByID(process.ID)
	assert.Nil(t, err)
	assert.Equal(t, before.FunctionSpec.Priority, after.FunctionSpec.Priority)
	assert.Equal(t, before.PriorityTime, after.PriorityTime,
		"a running process must never be reordered")
}

func TestSetProcessPrioritiesUnknownID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	results, err := db.SetProcessPriorities("test_colony_name", []core.PriorityUpdate{
		{ProcessID: core.GenerateRandomID(), Priority: 0},
	})
	assert.Nil(t, err, "an unknown id is a reported outcome, not an error")
	assert.Equal(t, core.PriorityNotFound, results[0].Outcome)
}

func TestSetProcessPrioritiesBulkPreservesFIFOWithinTier(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	var ids []string
	for i := 0; i < 3; i++ {
		p := utils.CreateTestProcess(colony.Name)
		p.FunctionSpec.Priority = 50
		assert.Nil(t, db.AddProcess(p))
		ids = append(ids, p.ID)
	}

	updates := make([]core.PriorityUpdate, 0, len(ids))
	for _, id := range ids {
		updates = append(updates, core.PriorityUpdate{ProcessID: id, Priority: 0})
	}
	results, err := db.SetProcessPriorities(colony.Name, updates)
	assert.Nil(t, err)
	assert.Len(t, results, 3)

	// All landed on the same rung, so submission order must still decide.
	var times []int64
	for _, id := range ids {
		p, err := db.GetProcessByID(id)
		assert.Nil(t, err)
		assert.Equal(t, 0, p.FunctionSpec.Priority)
		times = append(times, p.PriorityTime)
	}
	assert.True(t, times[0] < times[1] && times[1] < times[2],
		"equal priority must keep submission order (FIFO within tier)")
}
