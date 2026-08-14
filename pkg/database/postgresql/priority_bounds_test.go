package postgresql

import (
	"testing"

	"github.com/colonyos/colonies/pkg/constants"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

// The bounds half of the priority channel, and the invariants the semantics rest
// on. priority_channel_test.go covers the default bounds (floor unset, ceiling =
// submission priority); this covers bounds the submitter asks for, the exact
// movement of the ordering key, and what a malformed batch does.

func addProcessWithBounds(t *testing.T, db *PQDatabase, colonyName string, priority int, floor *int, ceiling *int) *core.Process {
	process := utils.CreateTestProcess(colonyName)
	process.FunctionSpec.Priority = priority
	process.FunctionSpec.PriorityFloor = floor
	process.FunctionSpec.PriorityCeiling = ceiling
	assert.Nil(t, db.AddProcess(process))

	return process
}

func TestSetProcessPrioritiesRespectsSubmittedFloor(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	floor := 10
	process := addProcessWithBounds(t, db, colony.Name, 50, &floor, nil)

	// A floor is how the application says "this may be deferred, but not shed".
	results, err := db.SetProcessPriorities(colony.Name, []core.PriorityUpdate{{ProcessID: process.ID, Priority: 0}})
	assert.Nil(t, err)
	assert.Equal(t, core.PriorityOutOfBounds, results[0].Outcome)
	assert.Equal(t, 50, results[0].Priority, "a rejected write reports the priority still in force")

	results, err = db.SetProcessPriorities(colony.Name, []core.PriorityUpdate{{ProcessID: process.ID, Priority: floor}})
	assert.Nil(t, err)
	assert.Equal(t, core.PriorityUpdated, results[0].Outcome)

	after, err := db.GetProcessByID(process.ID)
	assert.Nil(t, err)
	assert.Equal(t, floor, after.FunctionSpec.Priority, "the floor itself is reachable")
}

func TestSetProcessPrioritiesRespectsSubmittedCeiling(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	ceiling := 400
	process := addProcessWithBounds(t, db, colony.Name, 25, nil, &ceiling)

	// Escalation above the submission priority is opt-in, and the opt-in is the
	// ceiling the process was submitted with.
	results, err := db.SetProcessPriorities(colony.Name, []core.PriorityUpdate{{ProcessID: process.ID, Priority: ceiling}})
	assert.Nil(t, err)
	assert.Equal(t, core.PriorityUpdated, results[0].Outcome)

	after, err := db.GetProcessByID(process.ID)
	assert.Nil(t, err)
	assert.Equal(t, ceiling, after.FunctionSpec.Priority)

	results, err = db.SetProcessPriorities(colony.Name, []core.PriorityUpdate{{ProcessID: process.ID, Priority: ceiling + 1}})
	assert.Nil(t, err)
	assert.Equal(t, core.PriorityOutOfBounds, results[0].Outcome)
}

func TestPriorityBoundsSurviveTheDatabaseRoundTrip(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	floor := 10
	submitted := addProcessWithBounds(t, db, colony.Name, 50, &floor, nil)

	// The effective bounds are resolved once, at submission, and are readable
	// afterwards -- the ceiling especially, since the update overwrites PRIORITY
	// and the submission priority would otherwise be unrecoverable.
	stored, err := db.GetProcessByID(submitted.ID)
	assert.Nil(t, err)
	assert.NotNil(t, stored.FunctionSpec.PriorityFloor)
	assert.NotNil(t, stored.FunctionSpec.PriorityCeiling)
	assert.Equal(t, floor, *stored.FunctionSpec.PriorityFloor)
	assert.Equal(t, 50, *stored.FunctionSpec.PriorityCeiling)

	// AddProcess resolves them onto the process it was handed, too, so the
	// in-memory copy and the row agree.
	assert.NotNil(t, submitted.FunctionSpec.PriorityFloor)
	assert.Equal(t, floor, *submitted.FunctionSpec.PriorityFloor)

	unbounded := addProcessWithBounds(t, db, colony.Name, 25, nil, nil)
	stored, err = db.GetProcessByID(unbounded.ID)
	assert.Nil(t, err)
	assert.Equal(t, constants.MIN_PRIORITY, *stored.FunctionSpec.PriorityFloor)
	assert.Equal(t, 25, *stored.FunctionSpec.PriorityCeiling)
}

func TestAddProcessRejectsUnusableBounds(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	floor := 100
	ceiling := 50
	process := utils.CreateTestProcess(colony.Name)
	process.FunctionSpec.Priority = 50
	process.FunctionSpec.PriorityFloor = &floor
	process.FunctionSpec.PriorityCeiling = &ceiling
	assert.NotNil(t, db.AddProcess(process), "a floor above the ceiling admits no write at all")

	tooHigh := constants.MAX_PRIORITY + 1
	process = utils.CreateTestProcess(colony.Name)
	process.FunctionSpec.Priority = 50
	process.FunctionSpec.PriorityCeiling = &tooHigh
	assert.NotNil(t, db.AddProcess(process), "bounds may not widen past the allowed priority range")
}

func TestSetProcessPrioritiesMovesTheOrderingKeyExactly(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	process := utils.CreateTestProcess(colony.Name)
	process.FunctionSpec.Priority = 100
	assert.Nil(t, db.AddProcess(process))

	before, err := db.GetProcessByID(process.ID)
	assert.Nil(t, err)

	_, err = db.SetProcessPriorities(colony.Name, []core.PriorityUpdate{{ProcessID: process.ID, Priority: 25}})
	assert.Nil(t, err)

	after, err := db.GetProcessByID(process.ID)
	assert.Nil(t, err)

	// The key must move by exactly the priority delta and nothing else. Anything
	// looser means the submission-time component drifted, which would reorder
	// processes inside a tier.
	assert.Equal(t, int64(25-100)*core.PriorityTimeUnit, after.PriorityTime-before.PriorityTime)

	// Equivalently: the submission-time component of the key is untouched, so the
	// key still agrees with ComputePriorityTime for the new priority -- to within
	// the sub-microsecond error the TIMESTAMPTZ column introduces. Note the error
	// is signed: SUBMISSION_TIME is ROUNDED to the nearest microsecond, not
	// truncated, so a recomputed key can land either side of the original. That is
	// exactly why the update moves the key by a delta instead of recomputing it.
	residual := after.PriorityTime - core.ComputePriorityTime(25, before.SubmissionTime)
	assert.True(t, residual > -1000 && residual < 1000,
		"the submission-time component of the ordering key must be preserved, off by %d ns", residual)
}

func TestSetProcessPrioritiesReordersTheAssignQueue(t *testing.T) {
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

	_, err = db.SetProcessPriorities(colony.Name, []core.PriorityUpdate{{ProcessID: high.ID, Priority: 0}})
	assert.Nil(t, err)

	// The point of the channel: the real assign path, not just the reported
	// priority, must follow the update.
	assigned, err := db.SelectAndAssign(colony.Name, "test_executor_id", "test_executor_name", "test_executor_type", "", 1000, 1000, 1000, 10, 10, 10, 1)
	assert.Nil(t, err)
	assert.NotNil(t, assigned)
	assert.Equal(t, low.ID, assigned.ID, "the decayed process must no longer be assigned first")
}

func TestSetProcessPrioritiesAppliesTheValidPartOfABatch(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	process := utils.CreateTestProcess(colony.Name)
	process.FunctionSpec.Priority = 50
	assert.Nil(t, db.AddProcess(process))

	unknown := core.GenerateRandomID()

	// A decay pass is built from a snapshot of the queue, so by the time it lands
	// some of its ids are gone. One stale id must not discard the rest of the
	// batch.
	results, err := db.SetProcessPriorities(colony.Name, []core.PriorityUpdate{
		{ProcessID: unknown, Priority: 0},
		{ProcessID: process.ID, Priority: 0},
	})
	assert.Nil(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, unknown, results[0].ProcessID, "results come back in request order")
	assert.Equal(t, core.PriorityNotFound, results[0].Outcome)
	assert.Equal(t, core.PriorityUpdated, results[1].Outcome)

	after, err := db.GetProcessByID(process.ID)
	assert.Nil(t, err)
	assert.Equal(t, 0, after.FunctionSpec.Priority)
}

func TestSetProcessPrioritiesRejectsAmbiguousBatches(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	process := utils.CreateTestProcess(colony.Name)
	process.FunctionSpec.Priority = 50
	assert.Nil(t, db.AddProcess(process))

	// Two writes to one process in one batch: the join would pick one
	// arbitrarily, and there would be no honest per-process outcome to report.
	_, err = db.SetProcessPriorities(colony.Name, []core.PriorityUpdate{
		{ProcessID: process.ID, Priority: 0},
		{ProcessID: process.ID, Priority: 25},
	})
	assert.NotNil(t, err)

	_, err = db.SetProcessPriorities(colony.Name, []core.PriorityUpdate{{ProcessID: "", Priority: 0}})
	assert.NotNil(t, err)

	after, err := db.GetProcessByID(process.ID)
	assert.Nil(t, err)
	assert.Equal(t, 50, after.FunctionSpec.Priority, "a rejected batch writes nothing")
}

func TestSetProcessPrioritiesIsColonyScoped(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	process := utils.CreateTestProcess(colony.Name)
	process.FunctionSpec.Priority = 50
	assert.Nil(t, db.AddProcess(process))

	// The channel is a colony-scoped bulk operation, like the RemoveAll*ByColonyName
	// family. The scope is enforced here, in the statement, because the server can
	// only check membership of one colony per call -- a batch that could reach
	// outside it would make that check meaningless. A foreign id reports not_found
	// rather than not_waiting or out_of_bounds, so the call cannot be used to probe
	// for processes in another colony.
	results, err := db.SetProcessPriorities("another_colony_name", []core.PriorityUpdate{
		{ProcessID: process.ID, Priority: 0},
	})
	assert.Nil(t, err)
	assert.Equal(t, core.PriorityNotFound, results[0].Outcome)

	after, err := db.GetProcessByID(process.ID)
	assert.Nil(t, err)
	assert.Equal(t, 50, after.FunctionSpec.Priority)
}

func TestSetProcessPrioritiesEmptyBatch(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)
	defer db.Close()

	// A decision cycle with nothing to decay is normal, not an error.
	results, err := db.SetProcessPriorities("test_colony_name", []core.PriorityUpdate{})
	assert.Nil(t, err)
	assert.Len(t, results, 0)
}
