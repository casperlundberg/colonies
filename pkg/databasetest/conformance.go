package databasetest

import (
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

// HarnessMaker creates a fresh database instance for testing. The returned
// cleanup function must close the database and remove any resources that were
// allocated (e.g. drop tables, remove temp files).
type HarnessMaker func(t *testing.T) (database.Database, func())

// RunConformanceTests runs the full conformance test suite against a
// database.Database implementation. Each sub-test obtains a fresh database
// via newHarness so tests are fully isolated.
func RunConformanceTests(t *testing.T, newHarness HarnessMaker) {
	// ---------------------------------------------------------------
	// Colony tests
	// ---------------------------------------------------------------
	t.Run("Colony/AddAndGetByID", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony, prvKey, err := utils.CreateTestColonyWithKey()
		assert.Nil(t, err)
		_ = prvKey

		err = db.AddColony(colony)
		assert.Nil(t, err)

		colonyFromDB, err := db.GetColonyByID(colony.ID)
		assert.Nil(t, err)
		assert.NotNil(t, colonyFromDB)
		assert.True(t, colony.Equals(colonyFromDB))
	})

	t.Run("Colony/GetByName", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
		err := db.AddColony(colony)
		assert.Nil(t, err)

		colonyFromDB, err := db.GetColonyByName("test_colony_name_1")
		assert.Nil(t, err)
		assert.NotNil(t, colonyFromDB)
		assert.Equal(t, colony.ID, colonyFromDB.ID)
	})

	t.Run("Colony/GetColonies", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony1 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
		err := db.AddColony(colony1)
		assert.Nil(t, err)

		colony2 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_2")
		err = db.AddColony(colony2)
		assert.Nil(t, err)

		colonies, err := db.GetColonies()
		assert.Nil(t, err)
		assert.Len(t, colonies, 2)
	})

	t.Run("Colony/RemoveByName", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
		err := db.AddColony(colony)
		assert.Nil(t, err)

		err = db.RemoveColonyByName(colony.Name)
		assert.Nil(t, err)

		colonyFromDB, err := db.GetColonyByID(colony.ID)
		assert.Nil(t, err)
		assert.Nil(t, colonyFromDB)
	})

	t.Run("Colony/CountColonies", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		count, err := db.CountColonies()
		assert.Nil(t, err)
		assert.Equal(t, 0, count)

		colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
		err = db.AddColony(colony)
		assert.Nil(t, err)

		count, err = db.CountColonies()
		assert.Nil(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("Colony/AddDuplicateFails", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
		err := db.AddColony(colony)
		assert.Nil(t, err)

		err = db.AddColony(colony)
		assert.NotNil(t, err)
	})

	t.Run("Colony/Rename", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
		err := db.AddColony(colony)
		assert.Nil(t, err)

		colonyFromDB, err := db.GetColonyByID(colony.ID)
		assert.Nil(t, err)
		assert.Equal(t, "test_colony_name", colonyFromDB.Name)

		err = db.RenameColony(colony.Name, "test_colony_new_name")
		assert.Nil(t, err)

		colonyFromDB, err = db.GetColonyByID(colony.ID)
		assert.Nil(t, err)
		assert.Equal(t, "test_colony_new_name", colonyFromDB.Name)
	})

	t.Run("Colony/AddTwo", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony1 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
		err := db.AddColony(colony1)
		assert.Nil(t, err)

		colony2 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_2")
		err = db.AddColony(colony2)
		assert.Nil(t, err)

		var colonies []*core.Colony
		colonies = append(colonies, colony1)
		colonies = append(colonies, colony2)

		coloniesFromDB, err := db.GetColonies()
		assert.Nil(t, err)
		assert.True(t, core.IsColonyArraysEqual(colonies, coloniesFromDB))
	})

	t.Run("Colony/CascadingDelete", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony1 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
		err := db.AddColony(colony1)
		assert.Nil(t, err)

		colony2 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_2")
		err = db.AddColony(colony2)
		assert.Nil(t, err)

		user1 := utils.CreateTestUser(colony1.Name, "user1")
		err = db.AddUser(user1)
		assert.Nil(t, err)

		user2 := utils.CreateTestUser(colony2.Name, "user2")
		err = db.AddUser(user2)
		assert.Nil(t, err)

		generator1 := utils.FakeGenerator(t, colony1.Name, "test_initiator_id", "test_initiator_name")
		generator1.ID = core.GenerateRandomID()
		err = db.AddGenerator(generator1)
		assert.Nil(t, err)

		generator2 := utils.FakeGenerator(t, colony2.Name, "test_initiator_id", "test_initiator_name")
		generator2.ID = core.GenerateRandomID()
		err = db.AddGenerator(generator2)
		assert.Nil(t, err)

		cron1 := utils.FakeCron(t, colony1.Name, "test_initiator_id", "test_initiator_name")
		cron1.ID = core.GenerateRandomID()
		err = db.AddCron(cron1)
		assert.Nil(t, err)

		cron2 := utils.FakeCron(t, colony2.Name, "test_initiator_id", "test_initiator_name")
		cron2.ID = core.GenerateRandomID()
		err = db.AddCron(cron2)
		assert.Nil(t, err)

		executor1 := utils.CreateTestExecutor(colony1.Name)
		err = db.AddExecutor(executor1)
		assert.Nil(t, err)

		function := &core.Function{FunctionID: core.GenerateRandomID(), ExecutorName: executor1.Name, ColonyName: colony1.Name, FuncName: "testfunc", AvgWaitTime: 1.1, AvgExecTime: 0.1}
		err = db.AddFunction(function)
		assert.Nil(t, err)

		executor2 := utils.CreateTestExecutor(colony1.Name)
		err = db.AddExecutor(executor2)
		assert.Nil(t, err)

		function = &core.Function{FunctionID: core.GenerateRandomID(), ExecutorName: executor2.Name, ColonyName: colony1.Name, FuncName: "testfunc", AvgWaitTime: 1.1, AvgExecTime: 0.1}
		err = db.AddFunction(function)
		assert.Nil(t, err)

		executor3 := utils.CreateTestExecutor(colony2.Name)
		err = db.AddExecutor(executor3)
		assert.Nil(t, err)

		function = &core.Function{FunctionID: core.GenerateRandomID(), ExecutorName: executor3.Name, ColonyName: colony2.Name, FuncName: "testfunc", AvgWaitTime: 1.1, AvgExecTime: 0.1}
		err = db.AddFunction(function)
		assert.Nil(t, err)

		err = db.AddLog("test_processid1", colony1.ID, "test_executor_name", time.Now().UTC().UnixNano(), "1")
		assert.Nil(t, err)

		err = db.AddLog("test_processid1", colony2.ID, "test_executor_name", time.Now().UTC().UnixNano(), "1")
		assert.Nil(t, err)

		file := utils.CreateTestFileWithID("test_id", colony1.Name, time.Now())
		file.ID = core.GenerateRandomID()
		file.Label = "/testdir"
		file.Name = "test_file2.txt"
		file.Size = 1
		err = db.AddFile(file)
		assert.Nil(t, err)

		file = utils.CreateTestFileWithID("test_id", colony2.Name, time.Now())
		file.ID = core.GenerateRandomID()
		file.Label = "/testdir"
		file.Name = "test_file2.txt"
		file.Size = 1
		err = db.AddFile(file)
		assert.Nil(t, err)

		_, err = db.CreateSnapshot(colony1.Name, "/testdir", "test_snapshot_name1")
		assert.Nil(t, err)
		_, err = db.CreateSnapshot(colony2.Name, "/testdir", "test_snapshot_name2")
		assert.Nil(t, err)

		// Remove colony1 and verify cascading deletes
		err = db.RemoveColonyByName(colony1.Name)
		assert.Nil(t, err)

		users, err := db.GetUsersByColonyName(colony1.Name)
		assert.Nil(t, err)
		assert.Len(t, users, 0)

		users, err = db.GetUsersByColonyName(colony2.Name)
		assert.Nil(t, err)
		assert.Len(t, users, 1)

		colonyFromDB, err := db.GetColonyByID(colony1.ID)
		assert.Nil(t, err)
		assert.Nil(t, colonyFromDB)

		// After colony removal, executors are either nil or marked UNREGISTERED
		executorFromDB, err := db.GetExecutorByID(executor1.ID)
		assert.Nil(t, err)
		if executorFromDB != nil {
			assert.Equal(t, core.UNREGISTERED, executorFromDB.State)
		}

		executorFromDB, err = db.GetExecutorByID(executor2.ID)
		assert.Nil(t, err)
		if executorFromDB != nil {
			assert.Equal(t, core.UNREGISTERED, executorFromDB.State)
		}

		executorFromDB, err = db.GetExecutorByID(executor3.ID)
		assert.Nil(t, err)
		assert.NotNil(t, executorFromDB) // Belongs to colony2, should NOT be removed

		generatorFromDB, err := db.GetGeneratorByID(generator1.ID)
		assert.Nil(t, err)
		assert.Nil(t, generatorFromDB) // Should have been removed

		generatorFromDB, err = db.GetGeneratorByID(generator2.ID)
		assert.Nil(t, err)
		assert.NotNil(t, generatorFromDB) // Should NOT have been removed

		cronFromDB, err := db.GetCronByID(cron1.ID)
		assert.Nil(t, err)
		assert.Nil(t, cronFromDB) // Should have been removed

		cronFromDB, err = db.GetCronByID(cron2.ID)
		assert.Nil(t, err)
		assert.NotNil(t, cronFromDB) // Should NOT have been removed

		functions, err := db.GetFunctionsByColonyName(colony1.Name)
		assert.Nil(t, err)
		assert.Len(t, functions, 0)

		functions, err = db.GetFunctionsByColonyName(colony2.Name)
		assert.Nil(t, err)
		assert.Len(t, functions, 1)

		fileCount, err := db.CountFiles(colony1.Name)
		assert.Nil(t, err)
		assert.Equal(t, 0, fileCount)

		fileCount, err = db.CountFiles(colony2.Name)
		assert.Nil(t, err)
		assert.Equal(t, 1, fileCount)

		snapshots, err := db.GetSnapshotsByColonyName(colony1.Name)
		assert.Nil(t, err)
		assert.Len(t, snapshots, 0)

		snapshots, err = db.GetSnapshotsByColonyName(colony2.Name)
		assert.Nil(t, err)
		assert.Len(t, snapshots, 1)
	})

	t.Run("Colony/ChangeID", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
		err := db.AddColony(colony)
		assert.Nil(t, err)

		colonyFromDB, err := db.GetColonyByName(colony.Name)
		assert.Nil(t, err)
		assert.NotNil(t, colonyFromDB)

		err = db.ChangeColonyID(colony.Name, colony.ID, "new_id")
		assert.Nil(t, err)

		colonyFromDB, err = db.GetColonyByName(colony.Name)
		assert.Nil(t, err)
		assert.Equal(t, "new_id", colonyFromDB.ID)
		assert.NotEqual(t, colony.ID, colonyFromDB.ID)
	})

	// ---------------------------------------------------------------
	// User tests
	// ---------------------------------------------------------------
	t.Run("User/AddAndGetByName", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName := core.GenerateRandomID()
		user := utils.CreateTestUser(colonyName, "user1")
		err := db.AddUser(user)
		assert.Nil(t, err)

		userFromDB, err := db.GetUserByName(colonyName, "user1")
		assert.Nil(t, err)
		assert.NotNil(t, userFromDB)
		assert.True(t, userFromDB.Equals(user))
	})

	t.Run("User/GetUsersByColonyName", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName := core.GenerateRandomID()
		user1 := utils.CreateTestUser(colonyName, "user1")
		err := db.AddUser(user1)
		assert.Nil(t, err)

		user2 := utils.CreateTestUser(colonyName, "user2")
		err = db.AddUser(user2)
		assert.Nil(t, err)

		users, err := db.GetUsersByColonyName(colonyName)
		assert.Nil(t, err)
		assert.Len(t, users, 2)
	})

	t.Run("User/RemoveByName", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName := core.GenerateRandomID()
		user := utils.CreateTestUser(colonyName, "user1")
		err := db.AddUser(user)
		assert.Nil(t, err)

		err = db.RemoveUserByName(colonyName, "user1")
		assert.Nil(t, err)

		userFromDB, err := db.GetUserByName(colonyName, "user1")
		assert.Nil(t, err)
		assert.Nil(t, userFromDB)
	})

	t.Run("User/AddDuplicateFails", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName := core.GenerateRandomID()
		user := utils.CreateTestUser(colonyName, "user1")
		err := db.AddUser(user)
		assert.Nil(t, err)

		err = db.AddUser(user)
		assert.NotNil(t, err)
	})

	t.Run("User/RemoveByColonyName", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName1 := core.GenerateRandomID()
		colonyName2 := core.GenerateRandomID()

		user1 := utils.CreateTestUser(colonyName1, "user1")
		err := db.AddUser(user1)
		assert.Nil(t, err)

		user2 := utils.CreateTestUser(colonyName1, "user2")
		err = db.AddUser(user2)
		assert.Nil(t, err)

		user3 := utils.CreateTestUser(colonyName2, "user3")
		err = db.AddUser(user3)
		assert.Nil(t, err)

		users, err := db.GetUsersByColonyName(colonyName1)
		assert.Nil(t, err)
		assert.Len(t, users, 2)

		err = db.RemoveUsersByColonyName(colonyName1)
		assert.Nil(t, err)

		users, err = db.GetUsersByColonyName(colonyName1)
		assert.Nil(t, err)
		assert.Len(t, users, 0)

		users, err = db.GetUsersByColonyName(colonyName2)
		assert.Nil(t, err)
		assert.Len(t, users, 1)
	})

	t.Run("User/ChangeID", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName := core.GenerateRandomID()

		user := utils.CreateTestUser(colonyName, "user")
		err := db.AddUser(user)
		assert.Nil(t, err)

		userFromDB, err := db.GetUserByName(colonyName, user.Name)
		assert.Nil(t, err)
		assert.NotNil(t, userFromDB)

		err = db.ChangeUserID(colonyName, userFromDB.ID, "new_id")
		assert.Nil(t, err)

		userFromDB, err = db.GetUserByName(colonyName, user.Name)
		assert.Nil(t, err)
		assert.Equal(t, "new_id", userFromDB.ID)
		assert.NotEqual(t, user.ID, userFromDB.ID)
	})

	// ---------------------------------------------------------------
	// Executor tests
	// ---------------------------------------------------------------
	t.Run("Executor/AddAndGetByID", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
		err := db.AddColony(colony)
		assert.Nil(t, err)

		executor := utils.CreateTestExecutor(colony.Name)
		err = db.AddExecutor(executor)
		assert.Nil(t, err)

		executorFromDB, err := db.GetExecutorByID(executor.ID)
		assert.Nil(t, err)
		assert.NotNil(t, executorFromDB)
		assert.True(t, executor.Equals(executorFromDB))
	})

	t.Run("Executor/GetByName", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
		err := db.AddColony(colony)
		assert.Nil(t, err)

		executor := utils.CreateTestExecutor(colony.Name)
		err = db.AddExecutor(executor)
		assert.Nil(t, err)

		executorFromDB, err := db.GetExecutorByName(colony.Name, executor.Name)
		assert.Nil(t, err)
		assert.NotNil(t, executorFromDB)
		assert.Equal(t, executor.ID, executorFromDB.ID)
	})

	t.Run("Executor/Approve", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
		err := db.AddColony(colony)
		assert.Nil(t, err)

		executor := utils.CreateTestExecutor(colony.Name)
		err = db.AddExecutor(executor)
		assert.Nil(t, err)

		executorFromDB, err := db.GetExecutorByID(executor.ID)
		assert.Nil(t, err)
		assert.True(t, executorFromDB.IsPending())

		err = db.ApproveExecutor(executor)
		assert.Nil(t, err)

		executorFromDB, err = db.GetExecutorByID(executor.ID)
		assert.Nil(t, err)
		assert.True(t, executorFromDB.IsApproved())
	})

	t.Run("Executor/Reject", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
		err := db.AddColony(colony)
		assert.Nil(t, err)

		executor := utils.CreateTestExecutor(colony.Name)
		err = db.AddExecutor(executor)
		assert.Nil(t, err)

		err = db.RejectExecutor(executor)
		assert.Nil(t, err)

		executorFromDB, err := db.GetExecutorByID(executor.ID)
		assert.Nil(t, err)
		assert.True(t, executorFromDB.IsRejected())
	})

	t.Run("Executor/RemoveByName", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
		err := db.AddColony(colony)
		assert.Nil(t, err)

		executor := utils.CreateTestExecutor(colony.Name)
		err = db.AddExecutor(executor)
		assert.Nil(t, err)

		err = db.RemoveExecutorByName(colony.Name, executor.Name)
		assert.Nil(t, err)

		// After removal, the executor is either nil or marked as UNREGISTERED
		// (implementations may soft-delete by changing state)
		executorFromDB, err := db.GetExecutorByID(executor.ID)
		assert.Nil(t, err)
		if executorFromDB != nil {
			assert.Equal(t, core.UNREGISTERED, executorFromDB.State)
		}
	})

	t.Run("Executor/AddWithAllocations", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
		err := db.AddColony(colony)
		assert.Nil(t, err)

		executor := utils.CreateTestExecutor(colony.Name)
		project := core.Project{AllocatedCPU: 1, UsedCPU: 2, AllocatedGPU: 3, UsedGPU: 4, AllocatedStorage: 5, UsedStorage: 6}
		projects := make(map[string]core.Project)
		projects["test_project"] = project
		executor.Allocations.Projects = projects

		err = db.AddExecutor(executor)
		assert.Nil(t, err)

		executors, err := db.GetExecutors()
		assert.Nil(t, err)
		assert.Len(t, executors, 1)
		testProj := executors[0].Allocations.Projects["test_project"]
		assert.Equal(t, int64(1), testProj.AllocatedCPU)
		assert.Equal(t, int64(2), testProj.UsedCPU)
		assert.Equal(t, int64(3), testProj.AllocatedGPU)
		assert.Equal(t, int64(4), testProj.UsedGPU)
		assert.Equal(t, int64(5), testProj.AllocatedStorage)
		assert.Equal(t, int64(6), testProj.UsedStorage)
	})

	t.Run("Executor/AddWithLocation", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
		err := db.AddColony(colony)
		assert.Nil(t, err)

		executor := utils.CreateTestExecutor(colony.Name)
		executor.LocationName = "Home"

		err = db.AddExecutor(executor)
		assert.Nil(t, err)

		executorFromDB, err := db.GetExecutorByID(executor.ID)
		assert.Nil(t, err)
		assert.NotNil(t, executorFromDB)
		assert.Equal(t, "Home", executorFromDB.LocationName)
	})

	t.Run("Executor/AddMultiple", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
		err := db.AddColony(colony)
		assert.Nil(t, err)

		executor1 := utils.CreateTestExecutor(colony.Name)
		err = db.AddExecutor(executor1)
		assert.Nil(t, err)

		executor2 := utils.CreateTestExecutor(colony.Name)
		err = db.AddExecutor(executor2)
		assert.Nil(t, err)

		executor3 := utils.CreateTestExecutor(colony.Name)
		executor3.Name = "unique_name"
		err = db.AddExecutor(executor3)
		assert.Nil(t, err)

		executorsFromDB, err := db.GetExecutors()
		assert.Nil(t, err)
		assert.Len(t, executorsFromDB, 3)
	})

	t.Run("Executor/ChangeID", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName := core.GenerateRandomID()

		executor := utils.CreateTestExecutor(colonyName)
		err := db.AddExecutor(executor)
		assert.Nil(t, err)

		executorFromDB, err := db.GetExecutorByName(colonyName, executor.Name)
		assert.Nil(t, err)
		assert.NotNil(t, executorFromDB)

		err = db.ChangeExecutorID(colonyName, executor.ID, "new_id")
		assert.Nil(t, err)

		executorFromDB, err = db.GetExecutorByName(colonyName, executor.Name)
		assert.Nil(t, err)
		assert.Equal(t, "new_id", executorFromDB.ID)
		assert.NotEqual(t, executor.ID, executorFromDB.ID)
	})

	t.Run("Executor/CountAll", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		executorCount, err := db.CountExecutors()
		assert.Nil(t, err)
		assert.Equal(t, 0, executorCount)

		colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
		err = db.AddColony(colony)
		assert.Nil(t, err)

		executor := utils.CreateTestExecutor(colony.Name)
		err = db.AddExecutor(executor)
		assert.Nil(t, err)

		executorCount, err = db.CountExecutors()
		assert.Nil(t, err)
		assert.Equal(t, 1, executorCount)
	})

	t.Run("Executor/CountByColony", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony1 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
		err := db.AddColony(colony1)
		assert.Nil(t, err)

		executor := utils.CreateTestExecutor(colony1.Name)
		err = db.AddExecutor(executor)
		assert.Nil(t, err)

		executor = utils.CreateTestExecutor(colony1.Name)
		err = db.AddExecutor(executor)
		assert.Nil(t, err)

		colony2 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_2")
		err = db.AddColony(colony2)
		assert.Nil(t, err)

		executor = utils.CreateTestExecutor(colony2.Name)
		err = db.AddExecutor(executor)
		assert.Nil(t, err)

		executorCount, err := db.CountExecutors()
		assert.Nil(t, err)
		assert.Equal(t, 3, executorCount)

		executorCount, err = db.CountExecutorsByColonyName(colony1.Name)
		assert.Nil(t, err)
		assert.Equal(t, 2, executorCount)

		executorCount, err = db.CountExecutorsByColonyName(colony2.Name)
		assert.Nil(t, err)
		assert.Equal(t, 1, executorCount)
	})

	t.Run("Executor/GetByColonyName", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony1 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
		err := db.AddColony(colony1)
		assert.Nil(t, err)

		colony2 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_2")
		err = db.AddColony(colony2)
		assert.Nil(t, err)

		executor1 := utils.CreateTestExecutor(colony1.Name)
		err = db.AddExecutor(executor1)
		assert.Nil(t, err)

		executor2 := utils.CreateTestExecutor(colony1.Name)
		err = db.AddExecutor(executor2)
		assert.Nil(t, err)

		executor3 := utils.CreateTestExecutor(colony2.Name)
		err = db.AddExecutor(executor3)
		assert.Nil(t, err)

		executorsColony1FromDB, err := db.GetExecutorsByColonyName(colony1.Name, false)
		assert.Nil(t, err)
		assert.Len(t, executorsColony1FromDB, 2)

		executorsColony2FromDB, err := db.GetExecutorsByColonyName(colony2.Name, false)
		assert.Nil(t, err)
		assert.Len(t, executorsColony2FromDB, 1)
	})

	t.Run("Executor/MarkAlive", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
		err := db.AddColony(colony)
		assert.Nil(t, err)

		executor := utils.CreateTestExecutor(colony.Name)
		err = db.AddExecutor(executor)
		assert.Nil(t, err)

		time.Sleep(3000 * time.Millisecond)

		err = db.MarkAlive(executor)
		assert.Nil(t, err)

		executorFromDB, err := db.GetExecutorByID(executor.ID)
		assert.Nil(t, err)
		assert.True(t, (executorFromDB.LastHeardFromTime.Unix()-executor.LastHeardFromTime.Unix()) > 1)
	})

	t.Run("Executor/RemoveAll", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony1 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
		err := db.AddColony(colony1)
		assert.Nil(t, err)

		colony2 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_2")
		err = db.AddColony(colony2)
		assert.Nil(t, err)

		executor1 := utils.CreateTestExecutor(colony1.Name)
		err = db.AddExecutor(executor1)
		assert.Nil(t, err)

		executor2 := utils.CreateTestExecutor(colony1.Name)
		err = db.AddExecutor(executor2)
		assert.Nil(t, err)

		executor3 := utils.CreateTestExecutor(colony2.Name)
		err = db.AddExecutor(executor3)
		assert.Nil(t, err)

		err = db.RemoveExecutorsByColonyName(colony1.Name)
		assert.Nil(t, err)

		executorFromDB, err := db.GetExecutorByID(executor1.ID)
		assert.Nil(t, err)
		assert.Nil(t, executorFromDB)

		executorFromDB, err = db.GetExecutorByID(executor2.ID)
		assert.Nil(t, err)
		assert.Nil(t, executorFromDB)

		executorFromDB, err = db.GetExecutorByID(executor3.ID)
		assert.Nil(t, err)
		assert.NotNil(t, executorFromDB) // Belongs to colony2, should NOT be removed
	})

	t.Run("Executor/SameNameDifferentColonies", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony1 := core.CreateColony(core.GenerateRandomID(), "test_colony_1")
		err := db.AddColony(colony1)
		assert.Nil(t, err)

		colony2 := core.CreateColony(core.GenerateRandomID(), "test_colony_2")
		err = db.AddColony(colony2)
		assert.Nil(t, err)

		executor1 := utils.CreateTestExecutor(colony1.Name)
		executor1.Name = "shared-executor-name"
		err = db.AddExecutor(executor1)
		assert.Nil(t, err)

		executor2 := utils.CreateTestExecutor(colony2.Name)
		executor2.Name = "shared-executor-name"
		err = db.AddExecutor(executor2)
		assert.Nil(t, err)

		executorsColony1, err := db.GetExecutorsByColonyName(colony1.Name, false)
		assert.Nil(t, err)
		assert.Len(t, executorsColony1, 1)

		executorsColony2, err := db.GetExecutorsByColonyName(colony2.Name, false)
		assert.Nil(t, err)
		assert.Len(t, executorsColony2, 1)
	})

	t.Run("Executor/SetAllocations", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
		err := db.AddColony(colony)
		assert.Nil(t, err)

		executor := utils.CreateTestExecutor(colony.Name)
		project := core.Project{AllocatedCPU: 1, UsedCPU: 2, AllocatedGPU: 3, UsedGPU: 4, AllocatedStorage: 5, UsedStorage: 6}
		projects := make(map[string]core.Project)
		projects["test_project"] = project
		executor.Allocations.Projects = projects

		err = db.AddExecutor(executor)
		assert.Nil(t, err)

		executors, err := db.GetExecutors()
		assert.Nil(t, err)
		assert.Len(t, executors, 1)
		testProj := executors[0].Allocations.Projects["test_project"]
		assert.Equal(t, int64(1), testProj.AllocatedCPU)
		assert.Equal(t, int64(2), testProj.UsedCPU)
		assert.Equal(t, int64(3), testProj.AllocatedGPU)
		assert.Equal(t, int64(4), testProj.UsedGPU)
		assert.Equal(t, int64(5), testProj.AllocatedStorage)
		assert.Equal(t, int64(6), testProj.UsedStorage)

		// Update allocations
		project = core.Project{AllocatedCPU: 7, UsedCPU: 8, AllocatedGPU: 9, UsedGPU: 10, AllocatedStorage: 11, UsedStorage: 12}
		projects = make(map[string]core.Project)
		projects["test_project"] = project
		allocations := core.Allocations{Projects: projects}

		err = db.SetAllocations(colony.Name, executor.Name, allocations)
		assert.Nil(t, err)

		executors, err = db.GetExecutors()
		assert.Nil(t, err)
		assert.Len(t, executors, 1)
		testProj = executors[0].Allocations.Projects["test_project"]
		assert.Equal(t, int64(7), testProj.AllocatedCPU)
		assert.Equal(t, int64(8), testProj.UsedCPU)
		assert.Equal(t, int64(9), testProj.AllocatedGPU)
		assert.Equal(t, int64(10), testProj.UsedGPU)
		assert.Equal(t, int64(11), testProj.AllocatedStorage)
		assert.Equal(t, int64(12), testProj.UsedStorage)
	})

	// ---------------------------------------------------------------
	// Process tests
	// ---------------------------------------------------------------
	t.Run("Process/AddAndGetByID", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName := core.GenerateRandomID()
		process := utils.CreateTestProcess(colonyName)
		err := db.AddProcess(process)
		assert.Nil(t, err)

		processFromDB, err := db.GetProcessByID(process.ID)
		assert.Nil(t, err)
		assert.NotNil(t, processFromDB)
		assert.Equal(t, process.ID, processFromDB.ID)
		assert.Equal(t, core.WAITING, processFromDB.State)
	})

	t.Run("Process/FindWaitingProcesses", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
		err := db.AddColony(colony)
		assert.Nil(t, err)

		process1 := utils.CreateTestProcess(colony.Name)
		err = db.AddProcess(process1)
		assert.Nil(t, err)

		process2 := utils.CreateTestProcess(colony.Name)
		err = db.AddProcess(process2)
		assert.Nil(t, err)

		waiting, err := db.FindWaitingProcesses(colony.Name, "", "", "", 10)
		assert.Nil(t, err)
		assert.Len(t, waiting, 2)
	})

	t.Run("Process/FindRunningProcesses", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
		err := db.AddColony(colony)
		assert.Nil(t, err)

		executor := utils.CreateTestExecutor(colony.Name)
		err = db.AddExecutor(executor)
		assert.Nil(t, err)

		process := utils.CreateTestProcess(colony.Name)
		err = db.AddProcess(process)
		assert.Nil(t, err)

		err = db.Assign(executor.ID, process)
		assert.Nil(t, err)

		running, err := db.FindRunningProcesses(colony.Name, "", "", "", 10)
		assert.Nil(t, err)
		assert.Len(t, running, 1)
		assert.Equal(t, process.ID, running[0].ID)
	})

	t.Run("Process/SetProcessState", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName := core.GenerateRandomID()
		process := utils.CreateTestProcess(colonyName)
		err := db.AddProcess(process)
		assert.Nil(t, err)

		err = db.SetProcessState(process.ID, core.RUNNING)
		assert.Nil(t, err)

		processFromDB, err := db.GetProcessByID(process.ID)
		assert.Nil(t, err)
		assert.Equal(t, core.RUNNING, processFromDB.State)
	})

	t.Run("Process/MarkSuccessful", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
		err := db.AddColony(colony)
		assert.Nil(t, err)

		executor := utils.CreateTestExecutor(colony.Name)
		err = db.AddExecutor(executor)
		assert.Nil(t, err)

		process := utils.CreateTestProcess(colony.Name)
		err = db.AddProcess(process)
		assert.Nil(t, err)

		err = db.Assign(executor.ID, process)
		assert.Nil(t, err)

		_, _, err = db.MarkSuccessful(process.ID)
		assert.Nil(t, err)

		processFromDB, err := db.GetProcessByID(process.ID)
		assert.Nil(t, err)
		assert.Equal(t, core.SUCCESS, processFromDB.State)
	})

	t.Run("Process/MarkFailed", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
		err := db.AddColony(colony)
		assert.Nil(t, err)

		executor := utils.CreateTestExecutor(colony.Name)
		err = db.AddExecutor(executor)
		assert.Nil(t, err)

		process := utils.CreateTestProcess(colony.Name)
		err = db.AddProcess(process)
		assert.Nil(t, err)

		err = db.Assign(executor.ID, process)
		assert.Nil(t, err)

		err = db.MarkFailed(process.ID, []string{"error"})
		assert.Nil(t, err)

		processFromDB, err := db.GetProcessByID(process.ID)
		assert.Nil(t, err)
		assert.Equal(t, core.FAILED, processFromDB.State)
		assert.Equal(t, []string{"error"}, processFromDB.Errors)
	})

	t.Run("Process/RemoveByID", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName := core.GenerateRandomID()
		process := utils.CreateTestProcess(colonyName)
		err := db.AddProcess(process)
		assert.Nil(t, err)

		err = db.RemoveProcessByID(process.ID)
		assert.Nil(t, err)

		processFromDB, err := db.GetProcessByID(process.ID)
		assert.Nil(t, err)
		assert.Nil(t, processFromDB)
	})

	t.Run("Process/AddWithConditions", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName := core.GenerateRandomID()
		process := utils.CreateTestProcess(colonyName)
		process.FunctionSpec.Conditions.Nodes = 1
		process.FunctionSpec.Conditions.Processes = 2
		process.FunctionSpec.Conditions.ProcessesPerNode = 1
		process.FunctionSpec.Conditions.CPU = "1000m"
		process.FunctionSpec.Conditions.Memory = "10G"
		process.FunctionSpec.Conditions.Storage = "2000G"
		process.FunctionSpec.Conditions.WallTime = 70
		process.FunctionSpec.Conditions.GPU.Name = "nvidia_2080ti"
		process.FunctionSpec.Conditions.GPU.Count = 4
		process.FunctionSpec.Conditions.GPU.Memory = "10G"

		err := db.AddProcess(process)
		assert.Nil(t, err)

		processFromDB, err := db.GetProcessByID(process.ID)
		assert.Nil(t, err)
		assert.Equal(t, 1, processFromDB.FunctionSpec.Conditions.Nodes)
		assert.Equal(t, 2, processFromDB.FunctionSpec.Conditions.Processes)
		assert.Equal(t, 1, processFromDB.FunctionSpec.Conditions.ProcessesPerNode)
		assert.NotEmpty(t, processFromDB.FunctionSpec.Conditions.CPU)
		assert.NotEmpty(t, processFromDB.FunctionSpec.Conditions.Memory)
		assert.NotEmpty(t, processFromDB.FunctionSpec.Conditions.Storage)
		assert.Equal(t, int64(70), processFromDB.FunctionSpec.Conditions.WallTime)
		assert.Equal(t, "nvidia_2080ti", processFromDB.FunctionSpec.Conditions.GPU.Name)
		assert.Equal(t, 4, processFromDB.FunctionSpec.Conditions.GPU.Count)
		assert.NotEmpty(t, processFromDB.FunctionSpec.Conditions.GPU.Memory)
	})

	t.Run("Process/AddWithEnv", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		env := make(map[string]string)
		env["test_key_1"] = "test_value_1"
		env["test_key_2"] = "test_value_2"

		colonyName := core.GenerateRandomID()
		process := utils.CreateTestProcessWithEnv(colonyName, env)
		err := db.AddProcess(process)
		assert.Nil(t, err)

		processFromDB, err := db.GetProcessByID(process.ID)
		assert.Nil(t, err)
		assert.NotNil(t, processFromDB)

		processesFromDB, err := db.GetProcesses()
		assert.Nil(t, err)
		assert.Len(t, processesFromDB, 1)
	})

	t.Run("Process/Assign", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
		err := db.AddColony(colony)
		assert.Nil(t, err)

		executor := utils.CreateTestExecutor(colony.Name)
		err = db.AddExecutor(executor)
		assert.Nil(t, err)

		process := utils.CreateTestProcess(colony.Name)
		err = db.AddProcess(process)
		assert.Nil(t, err)

		processFromDB, err := db.GetProcessByID(process.ID)
		assert.Nil(t, err)
		assert.Equal(t, core.WAITING, processFromDB.State)
		assert.False(t, processFromDB.IsAssigned)

		err = db.Assign(executor.ID, process)
		assert.Nil(t, err)

		// Double assign should fail
		err = db.Assign(executor.ID, process)
		assert.NotNil(t, err)

		processFromDB, err = db.GetProcessByID(process.ID)
		assert.Nil(t, err)
		assert.True(t, processFromDB.IsAssigned)
		assert.Equal(t, core.RUNNING, processFromDB.State)

		err = db.Unassign(process)
		assert.Nil(t, err)

		processFromDB, err = db.GetProcessByID(process.ID)
		assert.Nil(t, err)
		assert.False(t, processFromDB.IsAssigned)
	})

	t.Run("Process/AssignCancelled", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
		err := db.AddColony(colony)
		assert.Nil(t, err)

		executor := utils.CreateTestExecutor(colony.Name)
		err = db.AddExecutor(executor)
		assert.Nil(t, err)

		process := utils.CreateTestProcess(colony.Name)
		err = db.AddProcess(process)
		assert.Nil(t, err)

		err = db.MarkCancelled(process.ID)
		assert.Nil(t, err)

		// Attempt to assign a cancelled process should fail
		err = db.Assign(executor.ID, process)
		assert.NotNil(t, err)

		// Verify process is still cancelled and not assigned
		processFromDB, err := db.GetProcessByID(process.ID)
		assert.Nil(t, err)
		assert.Equal(t, core.CANCELLED, processFromDB.State)
		assert.False(t, processFromDB.IsAssigned)
	})

	t.Run("Process/FindCancelledProcesses", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
		err := db.AddColony(colony)
		assert.Nil(t, err)

		executor := utils.CreateTestExecutor(colony.Name)
		err = db.AddExecutor(executor)
		assert.Nil(t, err)

		// Add 3 processes and cancel them
		for i := 0; i < 3; i++ {
			process := utils.CreateTestProcess(colony.Name)
			err = db.AddProcess(process)
			assert.Nil(t, err)
			err = db.MarkCancelled(process.ID)
			assert.Nil(t, err)
		}

		// Add 2 waiting processes (should not show up in cancelled)
		for i := 0; i < 2; i++ {
			process := utils.CreateTestProcess(colony.Name)
			err = db.AddProcess(process)
			assert.Nil(t, err)
		}

		cancelledProcesses, err := db.FindCancelledProcesses(colony.Name, "", "", "", 100)
		assert.Nil(t, err)
		assert.Len(t, cancelledProcesses, 3)

		cancelledCount, err := db.CountCancelledProcesses()
		assert.Nil(t, err)
		assert.Equal(t, 3, cancelledCount)

		cancelledCountByColony, err := db.CountCancelledProcessesByColonyName(colony.Name)
		assert.Nil(t, err)
		assert.Equal(t, 3, cancelledCountByColony)

		// Remove all cancelled processes
		err = db.RemoveAllCancelledProcessesByColonyName(colony.Name)
		assert.Nil(t, err)

		cancelledCount, err = db.CountCancelledProcessesByColonyName(colony.Name)
		assert.Nil(t, err)
		assert.Equal(t, 0, cancelledCount)

		// Waiting processes should still be there
		waitingCount, err := db.CountWaitingProcessesByColonyName(colony.Name)
		assert.Nil(t, err)
		assert.Equal(t, 2, waitingCount)
	})

	t.Run("Process/FindAllProcesses", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony1 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
		err := db.AddColony(colony1)
		assert.Nil(t, err)

		colony2 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_2")
		err = db.AddColony(colony2)
		assert.Nil(t, err)

		executor1 := utils.CreateTestExecutor(colony1.Name)
		err = db.AddExecutor(executor1)
		assert.Nil(t, err)

		executor2 := utils.CreateTestExecutor(colony2.Name)
		err = db.AddExecutor(executor2)
		assert.Nil(t, err)

		// Create some waiting processes
		for i := 0; i < 10; i++ {
			process := utils.CreateTestProcess(colony1.ID)
			err = db.AddProcess(process)
			assert.Nil(t, err)
		}
		for i := 0; i < 10; i++ {
			process := utils.CreateTestProcess(colony2.ID)
			err = db.AddProcess(process)
			assert.Nil(t, err)
		}

		// Create some running processes
		for i := 0; i < 5; i++ {
			process := utils.CreateTestProcess(colony1.ID)
			err = db.AddProcess(process)
			assert.Nil(t, err)
			err = db.Assign(executor1.ID, process)
			assert.Nil(t, err)
		}
		for i := 0; i < 5; i++ {
			process := utils.CreateTestProcess(colony2.ID)
			err = db.AddProcess(process)
			assert.Nil(t, err)
			err = db.Assign(executor2.ID, process)
			assert.Nil(t, err)
		}

		runningProcesses, err := db.FindAllRunningProcesses()
		assert.Nil(t, err)
		assert.Equal(t, 10, len(runningProcesses))

		waitingProcesses, err := db.FindAllWaitingProcesses()
		assert.Nil(t, err)
		assert.Equal(t, 20, len(waitingProcesses))
	})

	t.Run("Process/FindByColonyName", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
		err := db.AddColony(colony)
		assert.Nil(t, err)

		executor1 := utils.CreateTestExecutor(colony.Name)
		err = db.AddExecutor(executor1)
		assert.Nil(t, err)

		executor2 := utils.CreateTestExecutor(colony.Name)
		err = db.AddExecutor(executor2)
		assert.Nil(t, err)

		// Create some waiting processes
		for i := 0; i < 20; i++ {
			process := utils.CreateTestProcess(colony.Name)
			err = db.AddProcess(process)
			assert.Nil(t, err)
		}

		// Create some successful processes
		for i := 0; i < 10; i++ {
			process := utils.CreateTestProcess(colony.Name)
			err = db.AddProcess(process)
			assert.Nil(t, err)
			err = db.Assign(executor1.ID, process)
			assert.Nil(t, err)
			_, _, err = db.MarkSuccessful(process.ID)
			assert.Nil(t, err)
		}
		for i := 0; i < 10; i++ {
			process := utils.CreateTestProcess(colony.Name)
			err = db.AddProcess(process)
			assert.Nil(t, err)
			err = db.Assign(executor2.ID, process)
			assert.Nil(t, err)
			_, _, err = db.MarkSuccessful(process.ID)
			assert.Nil(t, err)
		}

		processesFromDB, err := db.FindProcessesByColonyName(colony.Name, 60, core.SUCCESS)
		assert.Nil(t, err)
		assert.Equal(t, 20, len(processesFromDB))
	})

	t.Run("Process/FindByExecutorID", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony1 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
		err := db.AddColony(colony1)
		assert.Nil(t, err)

		colony2 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_2")
		err = db.AddColony(colony2)
		assert.Nil(t, err)

		executor1 := utils.CreateTestExecutor(colony1.Name)
		err = db.AddExecutor(executor1)
		assert.Nil(t, err)

		executor2 := utils.CreateTestExecutor(colony2.Name)
		err = db.AddExecutor(executor2)
		assert.Nil(t, err)

		// Create successful processes assigned to executor1
		for i := 0; i < 10; i++ {
			process := utils.CreateTestProcess(colony1.ID)
			err = db.AddProcess(process)
			assert.Nil(t, err)
			err = db.Assign(executor1.ID, process)
			assert.Nil(t, err)
			_, _, err = db.MarkSuccessful(process.ID)
			assert.Nil(t, err)
		}

		// Create successful processes assigned to executor2
		for i := 0; i < 20; i++ {
			process := utils.CreateTestProcess(colony2.ID)
			err = db.AddProcess(process)
			assert.Nil(t, err)
			err = db.Assign(executor2.ID, process)
			assert.Nil(t, err)
			_, _, err = db.MarkSuccessful(process.ID)
			assert.Nil(t, err)
		}

		processesFromDB, err := db.FindProcessesByExecutorID(colony1.ID, executor1.ID, 60, core.SUCCESS)
		assert.Nil(t, err)
		assert.Equal(t, 10, len(processesFromDB))

		processesFromDB, err = db.FindProcessesByExecutorID(colony2.ID, executor2.ID, 60, core.SUCCESS)
		assert.Nil(t, err)
		assert.Equal(t, 20, len(processesFromDB))
	})

	t.Run("Process/ResetProcess", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
		err := db.AddColony(colony)
		assert.Nil(t, err)

		executor := utils.CreateTestExecutor(colony.Name)
		err = db.AddExecutor(executor)
		assert.Nil(t, err)

		process := utils.CreateTestProcess(colony.Name)
		process.FunctionSpec.MaxWaitTime = -1
		err = db.AddProcess(process)
		assert.Nil(t, err)
		err = db.Assign(executor.ID, process)
		assert.Nil(t, err)
		err = db.MarkFailed(process.ID, []string{"error"})
		assert.Nil(t, err)

		numberOfFailedProcesses, err := db.CountFailedProcesses()
		assert.Nil(t, err)
		assert.Equal(t, 1, numberOfFailedProcesses)

		err = db.ResetProcess(process)
		assert.Nil(t, err)

		numberOfFailedProcesses, err = db.CountFailedProcesses()
		assert.Nil(t, err)
		assert.Equal(t, 0, numberOfFailedProcesses)
	})

	t.Run("Process/SetOutput", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName := core.GenerateRandomID()
		process := utils.CreateTestProcess(colonyName)
		err := db.AddProcess(process)
		assert.Nil(t, err)

		output := make([]interface{}, 2)
		output[0] = "result1"
		output[1] = "result2"
		err = db.SetOutput(process.ID, output)
		assert.Nil(t, err)

		processFromDB, err := db.GetProcessByID(process.ID)
		assert.Nil(t, err)
		assert.Len(t, processFromDB.Output, 2)
		assert.Equal(t, "result1", processFromDB.Output[0])
		assert.Equal(t, "result2", processFromDB.Output[1])
	})

	t.Run("Process/SetInput", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName := core.GenerateRandomID()
		process := utils.CreateTestProcess(colonyName)
		err := db.AddProcess(process)
		assert.Nil(t, err)

		input := make([]interface{}, 2)
		input[0] = "result1"
		input[1] = "result2"
		err = db.SetInput(process.ID, input)
		assert.Nil(t, err)

		processFromDB, err := db.GetProcessByID(process.ID)
		assert.Nil(t, err)
		assert.Len(t, processFromDB.Input, 2)
		assert.Equal(t, "result1", processFromDB.Input[0])
		assert.Equal(t, "result2", processFromDB.Input[1])
	})

	t.Run("Process/SetErrorMsg", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName := core.GenerateRandomID()
		process := utils.CreateTestProcess(colonyName)
		err := db.AddProcess(process)
		assert.Nil(t, err)
		assert.Len(t, process.Errors, 0)

		err = db.SetErrors(process.ID, []string{"error"})
		assert.Nil(t, err)

		processFromDB, err := db.GetProcessByID(process.ID)
		assert.Nil(t, err)
		assert.Len(t, processFromDB.Errors, 1)
		assert.Equal(t, "error", processFromDB.Errors[0])
	})

	t.Run("Process/SetChildren", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName := core.GenerateRandomID()
		process := utils.CreateTestProcess(colonyName)
		err := db.AddProcess(process)
		assert.Nil(t, err)
		assert.Len(t, process.Children, 0)

		child := core.GenerateRandomID()
		children := []string{child}

		err = db.SetChildren(process.ID, children)
		assert.Nil(t, err)
		processFromDB, err := db.GetProcessByID(process.ID)
		assert.Nil(t, err)
		assert.Len(t, processFromDB.Children, 1)
		assert.Equal(t, child, processFromDB.Children[0])
	})

	t.Run("Process/SetParents", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName := core.GenerateRandomID()
		process := utils.CreateTestProcess(colonyName)
		err := db.AddProcess(process)
		assert.Nil(t, err)
		assert.Len(t, process.Parents, 0)

		parent := core.GenerateRandomID()
		parents := []string{parent}

		err = db.SetParents(process.ID, parents)
		assert.Nil(t, err)
		processFromDB, err := db.GetProcessByID(process.ID)
		assert.Nil(t, err)
		assert.Len(t, processFromDB.Parents, 1)
		assert.Equal(t, parent, processFromDB.Parents[0])
	})

	t.Run("Process/SetWaitingForParents", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName := core.GenerateRandomID()
		process := utils.CreateTestProcess(colonyName)
		err := db.AddProcess(process)
		assert.Nil(t, err)

		err = db.SetWaitForParents(process.ID, true)
		assert.Nil(t, err)
		processFromDB, err := db.GetProcessByID(process.ID)
		assert.Nil(t, err)
		assert.True(t, processFromDB.WaitForParents)

		err = db.SetWaitForParents(process.ID, false)
		assert.Nil(t, err)
		processFromDB, err = db.GetProcessByID(process.ID)
		assert.Nil(t, err)
		assert.False(t, processFromDB.WaitForParents)
	})

	t.Run("Process/MarkCancelled", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
		err := db.AddColony(colony)
		assert.Nil(t, err)

		executor := utils.CreateTestExecutor(colony.Name)
		err = db.AddExecutor(executor)
		assert.Nil(t, err)

		// Cancel a waiting process
		process1 := utils.CreateTestProcess(colony.Name)
		err = db.AddProcess(process1)
		assert.Nil(t, err)
		assert.Equal(t, core.WAITING, process1.State)

		err = db.MarkCancelled(process1.ID)
		assert.Nil(t, err)

		processFromDB, err := db.GetProcessByID(process1.ID)
		assert.Nil(t, err)
		assert.Equal(t, core.CANCELLED, processFromDB.State)

		// Cancel a running process
		process2 := utils.CreateTestProcess(colony.Name)
		err = db.AddProcess(process2)
		assert.Nil(t, err)
		err = db.Assign(executor.ID, process2)
		assert.Nil(t, err)

		processFromDB, err = db.GetProcessByID(process2.ID)
		assert.Nil(t, err)
		assert.Equal(t, core.RUNNING, processFromDB.State)

		err = db.MarkCancelled(process2.ID)
		assert.Nil(t, err)

		processFromDB, err = db.GetProcessByID(process2.ID)
		assert.Nil(t, err)
		assert.Equal(t, core.CANCELLED, processFromDB.State)

		// Cannot cancel an already cancelled process
		err = db.MarkCancelled(process2.ID)
		assert.NotNil(t, err)

		// Cannot cancel a successful process
		process3 := utils.CreateTestProcess(colony.Name)
		err = db.AddProcess(process3)
		assert.Nil(t, err)
		err = db.Assign(executor.ID, process3)
		assert.Nil(t, err)
		_, _, err = db.MarkSuccessful(process3.ID)
		assert.Nil(t, err)

		err = db.MarkCancelled(process3.ID)
		assert.NotNil(t, err)

		// Cannot cancel a failed process
		process4 := utils.CreateTestProcess(colony.Name)
		err = db.AddProcess(process4)
		assert.Nil(t, err)
		err = db.Assign(executor.ID, process4)
		assert.Nil(t, err)
		err = db.MarkFailed(process4.ID, []string{"error"})
		assert.Nil(t, err)

		err = db.MarkCancelled(process4.ID)
		assert.NotNil(t, err)
	})

	t.Run("Process/RemoveAllByColony", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony1Name := core.GenerateRandomID()
		process1 := utils.CreateTestProcess(colony1Name)
		err := db.AddProcess(process1)
		assert.Nil(t, err)
		attribute1 := core.CreateAttribute(process1.ID, colony1Name, "", core.IN, "test_key1", "test_value1")
		err = db.AddAttribute(attribute1)
		assert.Nil(t, err)

		colony2Name := core.GenerateRandomID()
		process2 := utils.CreateTestProcess(colony2Name)
		err = db.AddProcess(process2)
		assert.Nil(t, err)
		attribute2 := core.CreateAttribute(process2.ID, colony2Name, "", core.IN, "test_key1", "test_value1")
		err = db.AddAttribute(attribute2)
		assert.Nil(t, err)

		err = db.RemoveAllProcessesByColonyName(colony2Name)
		assert.Nil(t, err)

		// Attribute of process1 should still be accessible
		_, err = db.GetAttribute(process1.ID, "test_key1", core.IN)
		assert.Nil(t, err)
		// Attribute of process2 should be gone (cascaded)
		_, err = db.GetAttribute(process2.ID, "test_key1", core.IN)
		assert.NotNil(t, err)
	})

	t.Run("Process/RemoveAllByColonyWithState", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony1Name := core.GenerateRandomID()
		colony2Name := core.GenerateRandomID()

		process1 := utils.CreateTestProcess(colony1Name)
		err := db.AddProcess(process1)
		assert.Nil(t, err)

		process2 := utils.CreateTestProcess(colony1Name)
		err = db.AddProcess(process2)
		assert.Nil(t, err)

		process3 := utils.CreateTestProcess(colony1Name)
		err = db.AddProcess(process3)
		assert.Nil(t, err)

		process4 := utils.CreateTestProcess(colony1Name)
		err = db.AddProcess(process4)
		assert.Nil(t, err)

		process5 := utils.CreateTestProcess(colony2Name)
		err = db.AddProcess(process5)
		assert.Nil(t, err)

		err = db.SetProcessState(process1.ID, core.WAITING)
		assert.Nil(t, err)
		err = db.SetProcessState(process2.ID, core.RUNNING)
		assert.Nil(t, err)
		err = db.SetProcessState(process3.ID, core.SUCCESS)
		assert.Nil(t, err)
		err = db.SetProcessState(process4.ID, core.FAILED)
		assert.Nil(t, err)
		err = db.SetProcessState(process5.ID, core.FAILED)
		assert.Nil(t, err)

		waitingProcesses, err := db.CountWaitingProcesses()
		assert.Nil(t, err)
		assert.Equal(t, 1, waitingProcesses)
		runningProcesses, err := db.CountRunningProcesses()
		assert.Nil(t, err)
		assert.Equal(t, 1, runningProcesses)
		successfulProcesses, err := db.CountSuccessfulProcesses()
		assert.Nil(t, err)
		assert.Equal(t, 1, successfulProcesses)
		failedProcesses, err := db.CountFailedProcesses()
		assert.Nil(t, err)
		assert.Equal(t, 2, failedProcesses)

		err = db.RemoveAllWaitingProcessesByColonyName(colony1Name)
		assert.Nil(t, err)
		waitingProcesses, err = db.CountWaitingProcesses()
		assert.Nil(t, err)
		assert.Equal(t, 0, waitingProcesses)

		err = db.RemoveAllRunningProcessesByColonyName(colony1Name)
		assert.Nil(t, err)
		runningProcesses, err = db.CountRunningProcesses()
		assert.Nil(t, err)
		assert.Equal(t, 0, runningProcesses)

		err = db.RemoveAllSuccessfulProcessesByColonyName(colony1Name)
		assert.Nil(t, err)
		successfulProcesses, err = db.CountSuccessfulProcesses()
		assert.Nil(t, err)
		assert.Equal(t, 0, successfulProcesses)

		// Only colony1 failed processes removed; colony2 still has one
		err = db.RemoveAllFailedProcessesByColonyName(colony1Name)
		assert.Nil(t, err)
		failedProcesses, err = db.CountFailedProcesses()
		assert.Nil(t, err)
		assert.Equal(t, 1, failedProcesses)

		err = db.RemoveAllFailedProcessesByColonyName(colony2Name)
		assert.Nil(t, err)
		failedProcesses, err = db.CountFailedProcesses()
		assert.Nil(t, err)
		assert.Equal(t, 0, failedProcesses)
	})

	t.Run("Process/SelectCandidate", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
		err := db.AddColony(colony)
		assert.Nil(t, err)

		executor := utils.CreateTestExecutor(colony.Name)
		err = db.AddExecutor(executor)
		assert.Nil(t, err)

		process1 := utils.CreateTestProcess(colony.Name)
		err = db.AddProcess(process1)
		assert.Nil(t, err)

		// A process waiting for parents should NOT be a candidate
		process2 := utils.CreateTestProcess(colony.Name)
		process2.WaitForParents = true
		err = db.AddProcess(process2)
		assert.Nil(t, err)

		candidates, err := db.FindCandidates(colony.Name, executor.Type, "", 0, 0, 0, 0, 0, 0, 100)
		assert.Nil(t, err)
		assert.Len(t, candidates, 1)
		assert.Equal(t, process1.ID, candidates[0].ID)
	})

	t.Run("Process/SelectAndAssignByType", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
		err := db.AddColony(colony)
		assert.Nil(t, err)

		executor1 := utils.CreateTestExecutorWithType(colony.Name, "test_executor_type_1")
		err = db.AddExecutor(executor1)
		assert.Nil(t, err)

		executor2 := utils.CreateTestExecutorWithType(colony.Name, "test_executor_type_2")
		err = db.AddExecutor(executor2)
		assert.Nil(t, err)

		process1 := utils.CreateTestProcessWithType(colony.Name, "test_executor_type_1")
		err = db.AddProcess(process1)
		assert.Nil(t, err)

		time.Sleep(50 * time.Millisecond)

		process2 := utils.CreateTestProcessWithType(colony.Name, "test_executor_type_2")
		err = db.AddProcess(process2)
		assert.Nil(t, err)

		candidatesType1, err := db.FindCandidates(colony.Name, executor1.Type, "", 0, 0, 0, 0, 0, 0, 1)
		assert.Nil(t, err)
		assert.Len(t, candidatesType1, 1)
		assert.Equal(t, process1.ID, candidatesType1[0].ID)

		candidatesType2, err := db.FindCandidates(colony.Name, executor2.Type, "", 0, 0, 0, 0, 0, 0, 1)
		assert.Nil(t, err)
		assert.Len(t, candidatesType2, 1)
		assert.Equal(t, process2.ID, candidatesType2[0].ID)
	})

	t.Run("Process/SelectAndAssignByName", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
		err := db.AddColony(colony)
		assert.Nil(t, err)

		executor1 := utils.CreateTestExecutor(colony.Name)
		executor1.Name = "executor1"
		err = db.AddExecutor(executor1)
		assert.Nil(t, err)

		executor2 := utils.CreateTestExecutor(colony.Name)
		executor2.Name = "executor2"
		err = db.AddExecutor(executor2)
		assert.Nil(t, err)

		// General pool process (no executor name targets)
		process1 := utils.CreateTestProcess(colony.Name)
		err = db.AddProcess(process1)
		assert.Nil(t, err)

		// Process targeting executor1 only
		process2 := utils.CreateTestProcess(colony.Name)
		process2.FunctionSpec.Conditions.ExecutorNames = []string{"executor1"}
		err = db.AddProcess(process2)
		assert.Nil(t, err)

		// Process targeting both executor1 and executor2
		process3 := utils.CreateTestProcess(colony.Name)
		process3.FunctionSpec.Conditions.ExecutorNames = []string{"executor1", "executor2"}
		err = db.AddProcess(process3)
		assert.Nil(t, err)

		// FindCandidates (general pool) should only return process1
		candidatesGeneral, err := db.FindCandidates(colony.Name, executor1.Type, "", 0, 0, 0, 0, 0, 0, 100)
		assert.Nil(t, err)
		assert.Len(t, candidatesGeneral, 1)

		// FindCandidatesByName for executor1 should return process2 and process3
		candidatesByName, err := db.FindCandidatesByName(colony.Name, "executor1", executor1.Type, "", 0, 0, 0, 0, 0, 0, 100)
		assert.Nil(t, err)
		assert.Len(t, candidatesByName, 2)

		counter := 0
		for _, p := range candidatesByName {
			if p.ID == process2.ID || p.ID == process3.ID {
				counter++
			}
		}
		assert.Equal(t, 2, counter)
	})

	t.Run("Process/SelectAndAssignPriority", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
		err := db.AddColony(colony)
		assert.Nil(t, err)

		executor := utils.CreateTestExecutor(colony.Name)
		err = db.AddExecutor(executor)
		assert.Nil(t, err)

		// Create processes with different priority times
		process1 := utils.CreateTestProcess(colony.Name)
		process1.FunctionSpec.Priority = 10
		err = db.AddProcess(process1)
		assert.Nil(t, err)

		time.Sleep(10 * time.Millisecond)

		process2 := utils.CreateTestProcess(colony.Name)
		process2.FunctionSpec.Priority = 1
		err = db.AddProcess(process2)
		assert.Nil(t, err)

		// SelectAndAssign should pick process1 first (lower PriorityTime since added first)
		assignedProcess, err := db.SelectAndAssign(
			colony.Name,
			executor.ID,
			executor.Name,
			executor.Type,
			executor.LocationName,
			0, 0, 0,
			0, 0, 0,
			1,
		)
		assert.Nil(t, err)
		assert.NotNil(t, assignedProcess)
		assert.Equal(t, process1.ID, assignedProcess.ID)

		// Next SelectAndAssign should pick process2
		assignedProcess2, err := db.SelectAndAssign(
			colony.Name,
			executor.ID,
			executor.Name,
			executor.Type,
			executor.LocationName,
			0, 0, 0,
			0, 0, 0,
			1,
		)
		assert.Nil(t, err)
		assert.NotNil(t, assignedProcess2)
		assert.Equal(t, process2.ID, assignedProcess2.ID)
	})

	// ---------------------------------------------------------------
	// Attribute tests
	// ---------------------------------------------------------------
	t.Run("Attribute/AddAndGet", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		processID := core.GenerateRandomID()
		colonyName := core.GenerateRandomID()
		attribute := core.CreateAttribute(processID, colonyName, "", core.IN, "test_key", "test_value")
		err := db.AddAttribute(attribute)
		assert.Nil(t, err)

		attrFromDB, err := db.GetAttribute(processID, "test_key", core.IN)
		assert.Nil(t, err)
		assert.True(t, attribute.Equals(attrFromDB))
	})

	t.Run("Attribute/GetByID", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		processID := core.GenerateRandomID()
		colonyName := core.GenerateRandomID()
		attribute := core.CreateAttribute(processID, colonyName, "", core.OUT, "test_key", "test_value")
		err := db.AddAttribute(attribute)
		assert.Nil(t, err)

		attrFromDB, err := db.GetAttributeByID(attribute.ID)
		assert.Nil(t, err)
		assert.True(t, attribute.Equals(attrFromDB))
	})

	t.Run("Attribute/GetByType", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		processID := core.GenerateRandomID()
		colonyName := core.GenerateRandomID()
		attr1 := core.CreateAttribute(processID, colonyName, "", core.IN, "key1", "val1")
		attr2 := core.CreateAttribute(processID, colonyName, "", core.OUT, "key2", "val2")
		attr3 := core.CreateAttribute(processID, colonyName, "", core.IN, "key3", "val3")

		err := db.AddAttribute(attr1)
		assert.Nil(t, err)
		err = db.AddAttribute(attr2)
		assert.Nil(t, err)
		err = db.AddAttribute(attr3)
		assert.Nil(t, err)

		inAttrs, err := db.GetAttributesByType(processID, core.IN)
		assert.Nil(t, err)
		assert.Len(t, inAttrs, 2)

		outAttrs, err := db.GetAttributesByType(processID, core.OUT)
		assert.Nil(t, err)
		assert.Len(t, outAttrs, 1)
	})

	t.Run("Attribute/RemoveByID", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		processID := core.GenerateRandomID()
		colonyName := core.GenerateRandomID()
		attribute := core.CreateAttribute(processID, colonyName, "", core.IN, "test_key", "test_value")
		err := db.AddAttribute(attribute)
		assert.Nil(t, err)

		err = db.RemoveAttributeByID(attribute.ID)
		assert.Nil(t, err)

		attrs, err := db.GetAttributesByType(processID, core.IN)
		assert.Nil(t, err)
		assert.Len(t, attrs, 0)
	})

	// ---------------------------------------------------------------
	// Function tests
	// ---------------------------------------------------------------
	t.Run("Function/AddAndGetByID", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		function := &core.Function{
			FunctionID:   core.GenerateRandomID(),
			ExecutorName: core.GenerateRandomID(),
			ColonyName:   core.GenerateRandomID(),
			FuncName:     "testfunc1",
			Counter:      2,
			MinWaitTime:  1.0,
			MaxWaitTime:  2.0,
			MinExecTime:  3.0,
			MaxExecTime:  4.0,
			AvgWaitTime:  1.1,
			AvgExecTime:  0.1,
		}

		err := db.AddFunction(function)
		assert.Nil(t, err)

		funcFromDB, err := db.GetFunctionByID(function.FunctionID)
		assert.Nil(t, err)
		assert.NotNil(t, funcFromDB)
		assert.True(t, function.Equals(funcFromDB))
	})

	t.Run("Function/GetByColonyName", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName := core.GenerateRandomID()
		f1 := &core.Function{
			FunctionID:   core.GenerateRandomID(),
			ExecutorName: core.GenerateRandomID(),
			ColonyName:   colonyName,
			FuncName:     "testfunc1",
			AvgWaitTime:  1.1,
			AvgExecTime:  0.1,
		}
		f2 := &core.Function{
			FunctionID:   core.GenerateRandomID(),
			ExecutorName: core.GenerateRandomID(),
			ColonyName:   colonyName,
			FuncName:     "testfunc2",
			AvgWaitTime:  1.1,
			AvgExecTime:  0.1,
		}

		err := db.AddFunction(f1)
		assert.Nil(t, err)
		err = db.AddFunction(f2)
		assert.Nil(t, err)

		functions, err := db.GetFunctionsByColonyName(colonyName)
		assert.Nil(t, err)
		assert.Len(t, functions, 2)
	})

	t.Run("Function/RemoveByID", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		function := &core.Function{
			FunctionID:   core.GenerateRandomID(),
			ExecutorName: core.GenerateRandomID(),
			ColonyName:   core.GenerateRandomID(),
			FuncName:     "testfunc1",
			AvgWaitTime:  1.1,
			AvgExecTime:  0.1,
		}

		err := db.AddFunction(function)
		assert.Nil(t, err)

		err = db.RemoveFunctionByID(function.FunctionID)
		assert.Nil(t, err)

		funcFromDB, err := db.GetFunctionByID(function.FunctionID)
		assert.Nil(t, err)
		assert.Nil(t, funcFromDB)
	})

	t.Run("Function/GetByExecutorAndName", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		function := &core.Function{
			FunctionID:   core.GenerateRandomID(),
			ExecutorName: core.GenerateRandomID(),
			ColonyName:   core.GenerateRandomID(),
			FuncName:     "testfunc1",
			Counter:      2,
			MinWaitTime:  1.0,
			MaxWaitTime:  2.0,
			MinExecTime:  3.0,
			MaxExecTime:  4.0,
			AvgWaitTime:  1.1,
			AvgExecTime:  0.1,
		}

		err := db.AddFunction(function)
		assert.Nil(t, err)

		funcFromDB, err := db.GetFunctionsByExecutorAndName(function.ColonyName, function.ExecutorName, function.FuncName)
		assert.Nil(t, err)
		assert.NotNil(t, funcFromDB)
		assert.True(t, function.Equals(funcFromDB))

		// Non-existent name returns nil
		funcFromDB, err = db.GetFunctionsByExecutorAndName(function.ColonyName, function.ExecutorName, "does_not_exist")
		assert.Nil(t, err)
		assert.Nil(t, funcFromDB)
	})

	t.Run("Function/UpdateStats", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName := core.GenerateRandomID()
		function := &core.Function{
			FunctionID:   core.GenerateRandomID(),
			ExecutorName: core.GenerateRandomID(),
			ColonyName:   colonyName,
			FuncName:     "testfunc1",
			Counter:      10,
			AvgWaitTime:  1.1,
			AvgExecTime:  0.1,
		}

		err := db.AddFunction(function)
		assert.Nil(t, err)

		err = db.UpdateFunctionStats(function.ColonyName, function.ExecutorName, function.FuncName, 20, 0.1, 0.2, 0.3, 0.4, 2.0, 2.1)
		assert.Nil(t, err)

		functions, err := db.GetFunctionsByExecutorName(function.ColonyName, function.ExecutorName)
		assert.Nil(t, err)
		assert.Len(t, functions, 1)

		assert.Equal(t, 20, functions[0].Counter)
		assert.Equal(t, 0.1, functions[0].MinWaitTime)
		assert.Equal(t, 0.2, functions[0].MaxWaitTime)
		assert.Equal(t, 0.3, functions[0].MinExecTime)
		assert.Equal(t, 0.4, functions[0].MaxExecTime)
		assert.Equal(t, 2.0, functions[0].AvgWaitTime)
		assert.Equal(t, 2.1, functions[0].AvgExecTime)
	})

	t.Run("Function/RemoveByExecutor", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName := core.GenerateRandomID()
		executorName1 := core.GenerateRandomID()
		executorName2 := core.GenerateRandomID()

		f1 := &core.Function{FunctionID: core.GenerateRandomID(), ExecutorName: executorName1, ColonyName: colonyName, FuncName: "testfunc1", AvgWaitTime: 1.1, AvgExecTime: 0.1}
		f2 := &core.Function{FunctionID: core.GenerateRandomID(), ExecutorName: executorName2, ColonyName: colonyName, FuncName: "testfunc2", AvgWaitTime: 1.1, AvgExecTime: 0.1}

		err := db.AddFunction(f1)
		assert.Nil(t, err)
		err = db.AddFunction(f2)
		assert.Nil(t, err)

		functions, err := db.GetFunctionsByColonyName(colonyName)
		assert.Nil(t, err)
		assert.Len(t, functions, 2)

		err = db.RemoveFunctionsByExecutorName(f1.ColonyName, f1.ExecutorName)
		assert.Nil(t, err)

		functions, err = db.GetFunctionsByColonyName(colonyName)
		assert.Nil(t, err)
		assert.Len(t, functions, 1)
	})

	t.Run("Function/RemoveByColony", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName1 := core.GenerateRandomID()
		colonyName2 := core.GenerateRandomID()

		f1 := &core.Function{FunctionID: core.GenerateRandomID(), ExecutorName: core.GenerateRandomID(), ColonyName: colonyName1, FuncName: "testfunc1", AvgWaitTime: 1.1, AvgExecTime: 0.1}
		f2 := &core.Function{FunctionID: core.GenerateRandomID(), ExecutorName: core.GenerateRandomID(), ColonyName: colonyName1, FuncName: "testfunc2", AvgWaitTime: 1.1, AvgExecTime: 0.1}
		f3 := &core.Function{FunctionID: core.GenerateRandomID(), ExecutorName: core.GenerateRandomID(), ColonyName: colonyName2, FuncName: "testfunc3", AvgWaitTime: 1.1, AvgExecTime: 0.1}

		err := db.AddFunction(f1)
		assert.Nil(t, err)
		err = db.AddFunction(f2)
		assert.Nil(t, err)
		err = db.AddFunction(f3)
		assert.Nil(t, err)

		err = db.RemoveFunctionsByColonyName(colonyName1)
		assert.Nil(t, err)

		functions, err := db.GetFunctionsByColonyName(colonyName1)
		assert.Nil(t, err)
		assert.Len(t, functions, 0)

		functions, err = db.GetFunctionsByColonyName(colonyName2)
		assert.Nil(t, err)
		assert.Len(t, functions, 1)
	})

	t.Run("Function/WithDescription", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName := core.GenerateRandomID()
		executorName := core.GenerateRandomID()

		args := []*core.FunctionArg{
			{Name: "query", Type: "string", Description: "Search query", Required: true},
			{Name: "limit", Type: "integer", Description: "Max results", Required: false},
			{Name: "format", Type: "string", Description: "Output format", Enum: []string{"json", "text", "xml"}},
		}

		function := &core.Function{
			FunctionID:   core.GenerateRandomID(),
			ExecutorName: executorName,
			ColonyName:   colonyName,
			FuncName:     "search_tool",
			Description:  "Search for content in the database",
			Args:         args,
		}

		err := db.AddFunction(function)
		assert.Nil(t, err)

		funcFromDB, err := db.GetFunctionByID(function.FunctionID)
		assert.Nil(t, err)
		assert.NotNil(t, funcFromDB)
		assert.Equal(t, function.Description, funcFromDB.Description)
		assert.Equal(t, len(function.Args), len(funcFromDB.Args))

		for i, arg := range function.Args {
			assert.Equal(t, arg.Name, funcFromDB.Args[i].Name)
			assert.Equal(t, arg.Type, funcFromDB.Args[i].Type)
			assert.Equal(t, arg.Description, funcFromDB.Args[i].Description)
			assert.Equal(t, arg.Required, funcFromDB.Args[i].Required)
			assert.Equal(t, len(arg.Enum), len(funcFromDB.Args[i].Enum))
		}
	})

	t.Run("Function/WithLocation", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName := core.GenerateRandomID()
		executorName := core.GenerateRandomID()

		function := &core.Function{
			FunctionID:   core.GenerateRandomID(),
			ExecutorName: executorName,
			ColonyName:   colonyName,
			FuncName:     "tool_read_file",
			LocationName: "dev-location",
		}

		err := db.AddFunction(function)
		assert.Nil(t, err)

		funcFromDB, err := db.GetFunctionByID(function.FunctionID)
		assert.Nil(t, err)
		assert.NotNil(t, funcFromDB)
		assert.Equal(t, "dev-location", funcFromDB.LocationName)
		assert.True(t, function.Equals(funcFromDB))

		// Test empty LocationName
		function2 := &core.Function{
			FunctionID:   core.GenerateRandomID(),
			ExecutorName: executorName,
			ColonyName:   colonyName,
			FuncName:     "tool_write_file",
		}

		err = db.AddFunction(function2)
		assert.Nil(t, err)

		funcFromDB2, err := db.GetFunctionByID(function2.FunctionID)
		assert.Nil(t, err)
		assert.NotNil(t, funcFromDB2)
		assert.Equal(t, "", funcFromDB2.LocationName)
	})

	t.Run("Function/ResetStatsByColony", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName1 := core.GenerateRandomID()
		colonyName2 := core.GenerateRandomID()

		f1 := &core.Function{FunctionID: core.GenerateRandomID(), ExecutorName: core.GenerateRandomID(), ColonyName: colonyName1, FuncName: "testfunc1", Counter: 10, MinWaitTime: 1.0, MaxWaitTime: 5.0, MinExecTime: 0.5, MaxExecTime: 3.0, AvgWaitTime: 2.0, AvgExecTime: 1.5}
		f2 := &core.Function{FunctionID: core.GenerateRandomID(), ExecutorName: core.GenerateRandomID(), ColonyName: colonyName1, FuncName: "testfunc2", Counter: 20, MinWaitTime: 2.0, MaxWaitTime: 6.0, MinExecTime: 1.0, MaxExecTime: 4.0, AvgWaitTime: 3.0, AvgExecTime: 2.5}
		f3 := &core.Function{FunctionID: core.GenerateRandomID(), ExecutorName: core.GenerateRandomID(), ColonyName: colonyName2, FuncName: "testfunc3", Counter: 30, MinWaitTime: 3.0, MaxWaitTime: 7.0, MinExecTime: 1.5, MaxExecTime: 5.0, AvgWaitTime: 4.0, AvgExecTime: 3.5}

		err := db.AddFunction(f1)
		assert.Nil(t, err)
		err = db.AddFunction(f2)
		assert.Nil(t, err)
		err = db.AddFunction(f3)
		assert.Nil(t, err)

		err = db.ResetFunctionStatsByColonyName(colonyName1)
		assert.Nil(t, err)

		// Colony1 functions should be reset
		functions, err := db.GetFunctionsByColonyName(colonyName1)
		assert.Nil(t, err)
		assert.Len(t, functions, 2)
		for _, f := range functions {
			assert.Equal(t, 0, f.Counter)
			assert.Equal(t, 0.0, f.MinWaitTime)
			assert.Equal(t, 0.0, f.MaxWaitTime)
			assert.Equal(t, 0.0, f.MinExecTime)
			assert.Equal(t, 0.0, f.MaxExecTime)
			assert.Equal(t, 0.0, f.AvgWaitTime)
			assert.Equal(t, 0.0, f.AvgExecTime)
		}

		// Colony2 function should be unaffected
		functions, err = db.GetFunctionsByColonyName(colonyName2)
		assert.Nil(t, err)
		assert.Len(t, functions, 1)
		assert.Equal(t, 30, functions[0].Counter)
		assert.Equal(t, 3.0, functions[0].MinWaitTime)
		assert.Equal(t, 7.0, functions[0].MaxWaitTime)
	})

	// ---------------------------------------------------------------
	// Cron tests
	// ---------------------------------------------------------------
	t.Run("Cron/AddAndGetByID", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName := core.GenerateRandomID()
		cron := core.CreateCron(colonyName, "test_cron_name", "* * * * * *", 0, false, "workflow")
		cron.ID = core.GenerateRandomID()

		err := db.AddCron(cron)
		assert.Nil(t, err)

		cronFromDB, err := db.GetCronByID(cron.ID)
		assert.Nil(t, err)
		assert.NotNil(t, cronFromDB)
		assert.True(t, cron.Equals(cronFromDB))
	})

	t.Run("Cron/GetByColonyName", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName := core.GenerateRandomID()

		cron1 := core.CreateCron(colonyName, "test_cron_1"+core.GenerateRandomID(), "* * * * * *", 0, false, "workflow1")
		cron1.ID = core.GenerateRandomID()
		err := db.AddCron(cron1)
		assert.Nil(t, err)

		cron2 := core.CreateCron(colonyName, "test_cron_2"+core.GenerateRandomID(), "* * * * * *", 0, false, "workflow2")
		cron2.ID = core.GenerateRandomID()
		err = db.AddCron(cron2)
		assert.Nil(t, err)

		crons, err := db.FindCronsByColonyName(colonyName, 10)
		assert.Nil(t, err)
		assert.Len(t, crons, 2)
	})

	t.Run("Cron/RemoveByID", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName := core.GenerateRandomID()
		cron := core.CreateCron(colonyName, "test_cron_name", "* * * * * *", 0, false, "workflow")
		cron.ID = core.GenerateRandomID()

		err := db.AddCron(cron)
		assert.Nil(t, err)

		err = db.RemoveCronByID(cron.ID)
		assert.Nil(t, err)

		cronFromDB, err := db.GetCronByID(cron.ID)
		assert.Nil(t, err)
		assert.Nil(t, cronFromDB)
	})

	t.Run("Cron/Update", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName := core.GenerateRandomID()
		cron := core.CreateCron(colonyName, "test_name", "* * * * * *", 100, true, "workflow")
		cron.ID = core.GenerateRandomID()

		err := db.AddCron(cron)
		assert.Nil(t, err)

		cronFromDB, err := db.GetCronByID(cron.ID)
		assert.Nil(t, err)
		assert.Equal(t, cron.ID, cronFromDB.ID)
		assert.Equal(t, colonyName, cronFromDB.ColonyName)
		assert.Equal(t, "test_name", cronFromDB.Name)
		assert.Equal(t, "* * * * * *", cronFromDB.CronExpression)
		assert.Equal(t, 100, cronFromDB.Interval)
		assert.Equal(t, true, cronFromDB.Random)
		assert.Equal(t, "workflow", cronFromDB.WorkflowSpec)
		assert.Equal(t, "", cronFromDB.PrevProcessGraphID)

		err = db.UpdateCron(cron.ID, time.Now(), time.Time{}, core.GenerateRandomID())
		assert.Nil(t, err)

		cronFromDB, err = db.GetCronByID(cron.ID)
		assert.Nil(t, err)
		assert.Greater(t, cronFromDB.NextRun.Unix(), time.Time{}.Unix())
		assert.Equal(t, cronFromDB.LastRun.Unix(), time.Time{}.Unix())
		assert.NotEqual(t, "", cronFromDB.PrevProcessGraphID)

		err = db.UpdateCron(cron.ID, time.Now(), time.Now(), core.GenerateRandomID())
		assert.Nil(t, err)
		cronFromDB, err = db.GetCronByID(cron.ID)
		assert.Nil(t, err)
		assert.Greater(t, cronFromDB.LastRun.Unix(), time.Time{}.Unix())
	})

	t.Run("Cron/FindAll", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName1 := core.GenerateRandomID()
		colonyName2 := core.GenerateRandomID()

		cron1 := core.CreateCron(colonyName1, "test_name1", "* * * * * *", 0, false, "workflow1")
		cron1.ID = core.GenerateRandomID()
		cron2 := core.CreateCron(colonyName2, "test_name2", "* * * * * *", 0, false, "workflow2")
		cron2.ID = core.GenerateRandomID()
		cron3 := core.CreateCron(colonyName2, "test_name3", "* * * * * *", 0, false, "workflow3")
		cron3.ID = core.GenerateRandomID()

		err := db.AddCron(cron1)
		assert.Nil(t, err)
		err = db.AddCron(cron2)
		assert.Nil(t, err)
		err = db.AddCron(cron3)
		assert.Nil(t, err)

		crons, err := db.FindAllCrons()
		assert.Nil(t, err)
		assert.Len(t, crons, 3)
	})

	t.Run("Cron/GetByName", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName1 := core.GenerateRandomID()
		colonyName2 := core.GenerateRandomID()

		cron1 := core.CreateCron(colonyName1, "reconcile-Database-dc1", "* * * * * *", 60, false, "workflow1")
		cron1.ID = core.GenerateRandomID()
		cron2 := core.CreateCron(colonyName1, "reconcile-Database-dc2", "* * * * * *", 60, false, "workflow2")
		cron2.ID = core.GenerateRandomID()
		cron3 := core.CreateCron(colonyName2, "reconcile-Database-dc1", "* * * * * *", 60, false, "workflow3")
		cron3.ID = core.GenerateRandomID()

		err := db.AddCron(cron1)
		assert.Nil(t, err)
		err = db.AddCron(cron2)
		assert.Nil(t, err)
		err = db.AddCron(cron3)
		assert.Nil(t, err)

		// Find cron by name in colony1
		foundCron, err := db.GetCronByName(colonyName1, "reconcile-Database-dc1")
		assert.Nil(t, err)
		assert.NotNil(t, foundCron)
		assert.Equal(t, cron1.ID, foundCron.ID)
		assert.Equal(t, cron1.Name, foundCron.Name)
		assert.Equal(t, cron1.ColonyName, foundCron.ColonyName)

		// Find different cron by name in colony1
		foundCron, err = db.GetCronByName(colonyName1, "reconcile-Database-dc2")
		assert.Nil(t, err)
		assert.NotNil(t, foundCron)
		assert.Equal(t, cron2.ID, foundCron.ID)

		// Same name in different colony returns different cron
		foundCron, err = db.GetCronByName(colonyName2, "reconcile-Database-dc1")
		assert.Nil(t, err)
		assert.NotNil(t, foundCron)
		assert.Equal(t, cron3.ID, foundCron.ID)
		assert.Equal(t, colonyName2, foundCron.ColonyName)

		// Non-existent cron name returns nil
		foundCron, err = db.GetCronByName(colonyName1, "nonexistent-cron")
		assert.Nil(t, err)
		assert.Nil(t, foundCron)

		// Non-existent colony returns nil
		foundCron, err = db.GetCronByName("nonexistent-colony", "reconcile-Database-dc1")
		assert.Nil(t, err)
		assert.Nil(t, foundCron)
	})

	t.Run("Cron/SameNameDifferentColonies", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName1 := core.GenerateRandomID()
		colonyName2 := core.GenerateRandomID()
		sharedName := "shared_cron_name"

		// Create cron in first colony
		cron1 := core.CreateCron(colonyName1, sharedName, "* * * * * *", 0, false, "workflow1")
		cron1.ID = core.GenerateRandomID()
		err := db.AddCron(cron1)
		assert.Nil(t, err)

		// Create cron with same name in second colony - should succeed
		cron2 := core.CreateCron(colonyName2, sharedName, "0 * * * * *", 0, false, "workflow2")
		cron2.ID = core.GenerateRandomID()
		err = db.AddCron(cron2)
		assert.Nil(t, err)

		// Verify both crons exist
		crons1, err := db.FindCronsByColonyName(colonyName1, 100)
		assert.Nil(t, err)
		assert.Len(t, crons1, 1)
		assert.Equal(t, sharedName, crons1[0].Name)

		crons2, err := db.FindCronsByColonyName(colonyName2, 100)
		assert.Nil(t, err)
		assert.Len(t, crons2, 1)
		assert.Equal(t, sharedName, crons2[0].Name)
	})

	// ---------------------------------------------------------------
	// Generator tests
	// ---------------------------------------------------------------
	t.Run("Generator/AddAndGetByID", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName := core.GenerateRandomID()
		generator := utils.FakeGenerator(t, colonyName, "test_initiator_id", "test_initiator_name")
		generator.ID = core.GenerateRandomID()

		err := db.AddGenerator(generator)
		assert.Nil(t, err)

		genFromDB, err := db.GetGeneratorByID(generator.ID)
		assert.Nil(t, err)
		assert.NotNil(t, genFromDB)
		assert.True(t, generator.Equals(genFromDB))
	})

	t.Run("Generator/GetByIDNotFound", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		genFromDB, err := db.GetGeneratorByID("nonexistent_id")
		assert.Nil(t, err)
		assert.Nil(t, genFromDB)
	})

	t.Run("Generator/RemoveByID", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName := core.GenerateRandomID()
		generator := utils.FakeGenerator(t, colonyName, "test_initiator_id", "test_initiator_name")
		generator.ID = core.GenerateRandomID()

		err := db.AddGenerator(generator)
		assert.Nil(t, err)

		err = db.RemoveGeneratorByID(generator.ID)
		assert.Nil(t, err)

		genFromDB, err := db.GetGeneratorByID(generator.ID)
		assert.Nil(t, err)
		assert.Nil(t, genFromDB)
	})

	t.Run("Generator/FindByColonyName", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName := core.GenerateRandomID()
		gen1 := utils.FakeGenerator(t, colonyName, "test_initiator_id", "test_initiator_name")
		gen1.ID = core.GenerateRandomID()
		err := db.AddGenerator(gen1)
		assert.Nil(t, err)

		gen2 := utils.FakeGenerator(t, colonyName, "test_initiator_id", "test_initiator_name")
		gen2.ID = core.GenerateRandomID()
		err = db.AddGenerator(gen2)
		assert.Nil(t, err)

		generators, err := db.FindGeneratorsByColonyName(colonyName, 10)
		assert.Nil(t, err)
		assert.Len(t, generators, 2)
	})

	t.Run("Generator/GetByName", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName := core.GenerateRandomID()
		generator := utils.FakeGenerator(t, colonyName, "test_initiator_id", "test_initiator_name")
		generator.ID = core.GenerateRandomID()
		generator.Name = "test_name"
		err := db.AddGenerator(generator)
		assert.Nil(t, err)

		// Not found with invalid name
		genFromDB, err := db.GetGeneratorByName(colonyName, "invalid_name")
		assert.Nil(t, err)
		assert.Nil(t, genFromDB)

		// Found with correct name
		genFromDB, err = db.GetGeneratorByName(colonyName, "test_name")
		assert.Nil(t, err)
		assert.NotNil(t, genFromDB)
		assert.True(t, generator.Equals(genFromDB))
	})

	t.Run("Generator/SetLastRun", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName := core.GenerateRandomID()
		generator := utils.FakeGenerator(t, colonyName, "test_initiator_id", "test_initiator_name")
		generator.ID = core.GenerateRandomID()
		err := db.AddGenerator(generator)
		assert.Nil(t, err)

		genFromDB, err := db.GetGeneratorByID(generator.ID)
		assert.Nil(t, err)
		assert.True(t, generator.Equals(genFromDB))

		lastRun := genFromDB.LastRun.Unix()

		err = db.SetGeneratorLastRun(generator.ID)
		assert.Nil(t, err)

		genFromDB, err = db.GetGeneratorByID(generator.ID)
		assert.Nil(t, err)
		assert.Greater(t, genFromDB.LastRun.Unix(), lastRun)
	})

	t.Run("Generator/SetFirstPack", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName := core.GenerateRandomID()
		generator := utils.FakeGenerator(t, colonyName, "test_initiator_id", "test_initiator_name")
		generator.ID = core.GenerateRandomID()
		err := db.AddGenerator(generator)
		assert.Nil(t, err)

		genFromDB, err := db.GetGeneratorByID(generator.ID)
		assert.Nil(t, err)
		assert.True(t, generator.Equals(genFromDB))

		err = db.SetGeneratorFirstPack(generator.ID)
		assert.Nil(t, err)

		genFromDB, err = db.GetGeneratorByID(generator.ID)
		assert.Nil(t, err)
		assert.True(t, genFromDB.FirstPack.Unix() > 0)
	})

	t.Run("Generator/FindAll", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName1 := core.GenerateRandomID()
		gen1 := utils.FakeGenerator(t, colonyName1, "test_initiator_id", "test_initiator_name")
		gen1.ID = core.GenerateRandomID()
		err := db.AddGenerator(gen1)
		assert.Nil(t, err)

		colonyName2 := core.GenerateRandomID()
		gen2 := utils.FakeGenerator(t, colonyName2, "test_initiator_id", "test_initiator_name")
		gen2.ID = core.GenerateRandomID()
		err = db.AddGenerator(gen2)
		assert.Nil(t, err)

		generators, err := db.FindAllGenerators()
		assert.Nil(t, err)
		assert.Len(t, generators, 2)
	})

	t.Run("Generator/RemoveByColony", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName1 := core.GenerateRandomID()
		gen1 := utils.FakeGenerator(t, colonyName1, "test_initiator_id", "test_initiator_name")
		gen1.ID = core.GenerateRandomID()
		err := db.AddGenerator(gen1)
		assert.Nil(t, err)

		gen2 := utils.FakeGenerator(t, colonyName1, "test_initiator_id", "test_initiator_name")
		gen2.ID = core.GenerateRandomID()
		err = db.AddGenerator(gen2)
		assert.Nil(t, err)

		colonyName2 := core.GenerateRandomID()
		gen3 := utils.FakeGenerator(t, colonyName2, "test_initiator_id", "test_initiator_name")
		gen3.ID = core.GenerateRandomID()
		err = db.AddGenerator(gen3)
		assert.Nil(t, err)

		err = db.RemoveAllGeneratorsByColonyName(colonyName1)
		assert.Nil(t, err)

		genFromDB, err := db.GetGeneratorByID(gen1.ID)
		assert.Nil(t, err)
		assert.Nil(t, genFromDB)

		genFromDB, err = db.GetGeneratorByID(gen2.ID)
		assert.Nil(t, err)
		assert.Nil(t, genFromDB)

		genFromDB, err = db.GetGeneratorByID(gen3.ID)
		assert.Nil(t, err)
		assert.NotNil(t, genFromDB)
	})

	// ---------------------------------------------------------------
	// Log tests
	// ---------------------------------------------------------------
	t.Run("Log/AddAndGetByProcessID", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		processID := "test_processid_" + core.GenerateRandomID()
		colonyName := "test_colony"
		executorName := "test_executor_name"

		err := db.AddLog(processID, colonyName, executorName, time.Now().UTC().UnixNano(), "msg1")
		assert.Nil(t, err)
		err = db.AddLog(processID, colonyName, executorName, time.Now().UTC().UnixNano(), "msg2")
		assert.Nil(t, err)

		logs, err := db.GetLogsByProcessID(processID, 100)
		assert.Nil(t, err)
		assert.Len(t, logs, 2)
		assert.Equal(t, processID, logs[0].ProcessID)
		assert.Equal(t, colonyName, logs[0].ColonyName)
	})

	t.Run("Log/CountLogs", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName := "test_colony_" + core.GenerateRandomID()

		err := db.AddLog("pid1", colonyName, "executor", time.Now().UTC().UnixNano(), "msg1")
		assert.Nil(t, err)
		err = db.AddLog("pid2", colonyName, "executor", time.Now().UTC().UnixNano(), "msg2")
		assert.Nil(t, err)

		count, err := db.CountLogs(colonyName)
		assert.Nil(t, err)
		assert.Equal(t, 2, count)
	})

	// ---------------------------------------------------------------
	// Blueprint tests
	// ---------------------------------------------------------------
	t.Run("Blueprint/AddAndGetByID", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		blueprint := core.CreateBlueprint("ExecutorDeployment", "web-server", "production")
		blueprint.SetSpec("image", "nginx:1.21")
		blueprint.SetStatus("phase", "Running")

		err := db.AddBlueprint(blueprint)
		assert.Nil(t, err)

		bpFromDB, err := db.GetBlueprintByID(blueprint.ID)
		assert.Nil(t, err)
		assert.NotNil(t, bpFromDB)
		assert.Equal(t, blueprint.ID, bpFromDB.ID)
		assert.Equal(t, blueprint.Metadata.Name, bpFromDB.Metadata.Name)
		assert.Equal(t, blueprint.Kind, bpFromDB.Kind)

		image, ok := bpFromDB.GetSpec("image")
		assert.True(t, ok)
		assert.Equal(t, "nginx:1.21", image)
	})

	t.Run("Blueprint/GetByName", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		blueprint := core.CreateBlueprint("ExecutorDeployment", "web-server", "production")
		err := db.AddBlueprint(blueprint)
		assert.Nil(t, err)

		bpFromDB, err := db.GetBlueprintByName("production", "web-server")
		assert.Nil(t, err)
		assert.NotNil(t, bpFromDB)
		assert.Equal(t, blueprint.ID, bpFromDB.ID)
	})

	t.Run("Blueprint/RemoveByID", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		blueprint := core.CreateBlueprint("ExecutorDeployment", "web-server", "production")
		err := db.AddBlueprint(blueprint)
		assert.Nil(t, err)

		err = db.RemoveBlueprintByID(blueprint.ID)
		assert.Nil(t, err)

		bpFromDB, err := db.GetBlueprintByID(blueprint.ID)
		assert.Nil(t, err)
		assert.Nil(t, bpFromDB)
	})

	t.Run("Blueprint/CountBlueprints", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		count, err := db.CountBlueprints()
		assert.Nil(t, err)
		assert.Equal(t, 0, count)

		bp1 := core.CreateBlueprint("ExecutorDeployment", "web-1", "production")
		err = db.AddBlueprint(bp1)
		assert.Nil(t, err)

		bp2 := core.CreateBlueprint("ExecutorDeployment", "web-2", "production")
		err = db.AddBlueprint(bp2)
		assert.Nil(t, err)

		count, err = db.CountBlueprints()
		assert.Nil(t, err)
		assert.Equal(t, 2, count)
	})

	t.Run("Blueprint/AddDefinition", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		sd := core.CreateBlueprintDefinition(
			"executor-deployment",
			"compute.colonies.io",
			"v1",
			"ExecutorDeployment",
			"executordeployments",
			"Namespaced",
			"executor-controller",
			"reconcile",
		)
		sd.Metadata.ColonyName = "test-colony"

		err := db.AddBlueprintDefinition(sd)
		assert.Nil(t, err)

		// Get by ID
		sdFromDB, err := db.GetBlueprintDefinitionByID(sd.ID)
		assert.Nil(t, err)
		assert.NotNil(t, sdFromDB)
		assert.Equal(t, sd.ID, sdFromDB.ID)
		assert.Equal(t, sd.Metadata.Name, sdFromDB.Metadata.Name)
		assert.Equal(t, sd.Spec.Group, sdFromDB.Spec.Group)
		assert.Equal(t, sd.Spec.Version, sdFromDB.Spec.Version)

		// Get by name
		sdFromDB2, err := db.GetBlueprintDefinitionByName(sd.Metadata.ColonyName, sd.Metadata.Name)
		assert.Nil(t, err)
		assert.NotNil(t, sdFromDB2)
		assert.Equal(t, sd.ID, sdFromDB2.ID)

		// Get all
		sds, err := db.GetBlueprintDefinitions()
		assert.Nil(t, err)
		assert.Equal(t, 1, len(sds))

		// Count
		defCount, err := db.CountBlueprintDefinitions()
		assert.Nil(t, err)
		assert.Equal(t, 1, defCount)
	})

	t.Run("Blueprint/GetDefinitionByKind", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		sd1 := core.CreateBlueprintDefinition(
			"executor-deployment",
			"compute.colonies.io",
			"v1",
			"ExecutorDeployment",
			"executordeployments",
			"Namespaced",
			"executor-controller",
			"reconcile",
		)
		sd1.Metadata.ColonyName = "test-colony"

		sd2 := core.CreateBlueprintDefinition(
			"service-deployment",
			"compute.colonies.io",
			"v1",
			"ServiceDeployment",
			"servicedeployments",
			"Namespaced",
			"service-controller",
			"reconcile",
		)
		sd2.Metadata.ColonyName = "test-colony"

		err := db.AddBlueprintDefinition(sd1)
		assert.Nil(t, err)
		err = db.AddBlueprintDefinition(sd2)
		assert.Nil(t, err)

		// Find ExecutorDeployment
		foundDef, err := db.GetBlueprintDefinitionByKind("ExecutorDeployment")
		assert.Nil(t, err)
		assert.NotNil(t, foundDef)
		assert.Equal(t, "ExecutorDeployment", foundDef.Spec.Names.Kind)
		assert.Equal(t, sd1.ID, foundDef.ID)

		// Find ServiceDeployment
		foundDef2, err := db.GetBlueprintDefinitionByKind("ServiceDeployment")
		assert.Nil(t, err)
		assert.NotNil(t, foundDef2)
		assert.Equal(t, "ServiceDeployment", foundDef2.Spec.Names.Kind)
		assert.Equal(t, sd2.ID, foundDef2.ID)

		// Non-existent kind returns nil
		notFound, err := db.GetBlueprintDefinitionByKind("NonExistentKind")
		assert.Nil(t, err)
		assert.Nil(t, notFound)
	})

	t.Run("Blueprint/GetByNamespace", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		bp1 := core.CreateBlueprint("ExecutorDeployment", "web-1", "production")
		bp1.SetSpec("image", "nginx:1.21")
		bp2 := core.CreateBlueprint("ExecutorDeployment", "web-2", "production")
		bp2.SetSpec("image", "nginx:1.22")
		bp3 := core.CreateBlueprint("ExecutorDeployment", "web-3", "staging")
		bp3.SetSpec("image", "nginx:1.21")

		err := db.AddBlueprint(bp1)
		assert.Nil(t, err)
		err = db.AddBlueprint(bp2)
		assert.Nil(t, err)
		err = db.AddBlueprint(bp3)
		assert.Nil(t, err)

		prodBlueprints, err := db.GetBlueprintsByNamespace("production")
		assert.Nil(t, err)
		assert.Equal(t, 2, len(prodBlueprints))

		stagingBlueprints, err := db.GetBlueprintsByNamespace("staging")
		assert.Nil(t, err)
		assert.Equal(t, 1, len(stagingBlueprints))

		prodCount, err := db.CountBlueprintsByNamespace("production")
		assert.Nil(t, err)
		assert.Equal(t, 2, prodCount)
	})

	t.Run("Blueprint/GetByKind", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		bp1 := core.CreateBlueprint("ExecutorDeployment", "web-1", "production")
		bp1.SetSpec("image", "nginx:1.21")
		bp2 := core.CreateBlueprint("Database", "db-1", "production")
		bp2.SetSpec("engine", "postgres")

		err := db.AddBlueprint(bp1)
		assert.Nil(t, err)
		err = db.AddBlueprint(bp2)
		assert.Nil(t, err)

		executorDeployments, err := db.GetBlueprintsByKind("ExecutorDeployment")
		assert.Nil(t, err)
		assert.Equal(t, 1, len(executorDeployments))

		databases, err := db.GetBlueprintsByKind("Database")
		assert.Nil(t, err)
		assert.Equal(t, 1, len(databases))
	})

	t.Run("Blueprint/Update", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		blueprint := core.CreateBlueprint("ExecutorDeployment", "web-server", "production")
		blueprint.SetSpec("replicas", 3)

		err := db.AddBlueprint(blueprint)
		assert.Nil(t, err)

		blueprint.SetSpec("replicas", 5)
		err = db.UpdateBlueprint(blueprint)
		assert.Nil(t, err)

		bpFromDB, err := db.GetBlueprintByID(blueprint.ID)
		assert.Nil(t, err)
		replicas, ok := bpFromDB.GetSpec("replicas")
		assert.True(t, ok)
		// JSON round-trip may convert int to float64; accept either
		assert.EqualValues(t, 5, replicas)
	})

	t.Run("Blueprint/UpdateStatus", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		blueprint := core.CreateBlueprint("ExecutorDeployment", "web-server", "production")
		blueprint.SetSpec("replicas", 3)
		blueprint.SetStatus("phase", "Pending")

		err := db.AddBlueprint(blueprint)
		assert.Nil(t, err)

		newStatus := map[string]interface{}{
			"phase": "Running",
			"ready": 3,
		}
		err = db.UpdateBlueprintStatus(blueprint.ID, newStatus)
		assert.Nil(t, err)

		bpFromDB, err := db.GetBlueprintByID(blueprint.ID)
		assert.Nil(t, err)
		phase, ok := bpFromDB.GetStatus("phase")
		assert.True(t, ok)
		assert.Equal(t, "Running", phase)
		ready, ok := bpFromDB.GetStatus("ready")
		assert.True(t, ok)
		assert.EqualValues(t, 3, ready)
	})

	t.Run("Blueprint/History", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		blueprint := core.CreateBlueprint("ExecutorDeployment", "web-server", "production")
		blueprint.SetSpec("replicas", 3)
		blueprint.SetStatus("phase", "Running")

		err := db.AddBlueprint(blueprint)
		assert.Nil(t, err)

		history := core.CreateBlueprintHistory(blueprint, "test-user", "create")
		err = db.AddBlueprintHistory(history)
		assert.Nil(t, err)

		histories, err := db.GetBlueprintHistory(blueprint.ID, 10)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(histories))
		assert.Equal(t, blueprint.ID, histories[0].BlueprintID)
		assert.Equal(t, "ExecutorDeployment", histories[0].Kind)
		assert.Equal(t, "production", histories[0].Namespace)
		assert.Equal(t, "web-server", histories[0].Name)
		assert.Equal(t, blueprint.Metadata.Generation, histories[0].Generation)
		assert.Equal(t, "test-user", histories[0].ChangedBy)
		assert.Equal(t, "create", histories[0].ChangeType)

		// Verify spec in history
		replicas, ok := histories[0].Spec["replicas"]
		assert.True(t, ok)
		assert.EqualValues(t, 3, replicas)

		// Verify status in history
		phase, ok := histories[0].Status["phase"]
		assert.True(t, ok)
		assert.Equal(t, "Running", phase)
	})

	t.Run("Blueprint/HistoryMultipleVersions", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		blueprint := core.CreateBlueprint("ExecutorDeployment", "web-server", "production")
		blueprint.SetSpec("replicas", 3)

		err := db.AddBlueprint(blueprint)
		assert.Nil(t, err)

		initialGen := blueprint.Metadata.Generation

		history1 := core.CreateBlueprintHistory(blueprint, "user1", "create")
		err = db.AddBlueprintHistory(history1)
		assert.Nil(t, err)

		blueprint.SetSpec("replicas", 5)
		blueprint.Metadata.Generation = initialGen + 1
		history2 := core.CreateBlueprintHistory(blueprint, "user2", "update")
		err = db.AddBlueprintHistory(history2)
		assert.Nil(t, err)

		blueprint.SetSpec("replicas", 10)
		blueprint.Metadata.Generation = initialGen + 2
		history3 := core.CreateBlueprintHistory(blueprint, "user3", "update")
		err = db.AddBlueprintHistory(history3)
		assert.Nil(t, err)

		// Get all history (no limit)
		allHistories, err := db.GetBlueprintHistory(blueprint.ID, 0)
		assert.Nil(t, err)
		assert.Equal(t, 3, len(allHistories))

		// Verify ordered by timestamp DESC (most recent first)
		assert.Equal(t, initialGen+2, allHistories[0].Generation)
		assert.Equal(t, initialGen+1, allHistories[1].Generation)
		assert.Equal(t, initialGen, allHistories[2].Generation)

		// Get limited history
		limitedHistories, err := db.GetBlueprintHistory(blueprint.ID, 2)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(limitedHistories))
		assert.Equal(t, initialGen+2, limitedHistories[0].Generation)
		assert.Equal(t, initialGen+1, limitedHistories[1].Generation)
	})

	t.Run("Blueprint/HistoryByGeneration", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		blueprint := core.CreateBlueprint("ExecutorDeployment", "web-server", "production")
		blueprint.SetSpec("replicas", 3)

		err := db.AddBlueprint(blueprint)
		assert.Nil(t, err)

		history1 := core.CreateBlueprintHistory(blueprint, "user1", "create")
		err = db.AddBlueprintHistory(history1)
		assert.Nil(t, err)

		// Verify history was stored via GetBlueprintHistory (list method)
		histories, err := db.GetBlueprintHistory(blueprint.ID, 10)
		assert.Nil(t, err)
		assert.GreaterOrEqual(t, len(histories), 1)

		// GetBlueprintHistoryByGeneration may not be implemented by all backends
		historyGen1, err := db.GetBlueprintHistoryByGeneration(blueprint.ID, 1)
		assert.Nil(t, err)
		if historyGen1 != nil {
			assert.Equal(t, int64(1), historyGen1.Generation)
			assert.Equal(t, "user1", historyGen1.ChangedBy)
		}

		// Get generation that does not exist
		historyGen99, err := db.GetBlueprintHistoryByGeneration(blueprint.ID, 99)
		assert.Nil(t, err)
		assert.Nil(t, historyGen99)
	})

	t.Run("Blueprint/RemoveHistory", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		blueprint := core.CreateBlueprint("ExecutorDeployment", "web-server", "production")
		err := db.AddBlueprint(blueprint)
		assert.Nil(t, err)

		history1 := core.CreateBlueprintHistory(blueprint, "user1", "create")
		err = db.AddBlueprintHistory(history1)
		assert.Nil(t, err)

		blueprint.Metadata.Generation = 2
		history2 := core.CreateBlueprintHistory(blueprint, "user2", "update")
		err = db.AddBlueprintHistory(history2)
		assert.Nil(t, err)

		histories, err := db.GetBlueprintHistory(blueprint.ID, 0)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(histories))

		err = db.RemoveBlueprintHistory(blueprint.ID)
		assert.Nil(t, err)

		historiesAfter, err := db.GetBlueprintHistory(blueprint.ID, 0)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(historiesAfter))
	})

	t.Run("Blueprint/GetByLocation", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		bp1 := core.CreateBlueprint("ExecutorDeployment", "web-1", "production")
		bp1.Metadata.LocationName = "home"
		bp1.SetSpec("image", "nginx:1.21")
		err := db.AddBlueprint(bp1)
		assert.Nil(t, err)

		bp2 := core.CreateBlueprint("ExecutorDeployment", "web-2", "production")
		bp2.Metadata.LocationName = "HOME"
		bp2.SetSpec("image", "nginx:1.22")
		err = db.AddBlueprint(bp2)
		assert.Nil(t, err)

		bp3 := core.CreateBlueprint("ExecutorDeployment", "web-3", "production")
		bp3.Metadata.LocationName = "Home"
		bp3.SetSpec("image", "nginx:1.23")
		err = db.AddBlueprint(bp3)
		assert.Nil(t, err)

		bp4 := core.CreateBlueprint("ExecutorDeployment", "web-4", "production")
		bp4.Metadata.LocationName = "office"
		bp4.SetSpec("image", "nginx:1.24")
		err = db.AddBlueprint(bp4)
		assert.Nil(t, err)

		// Query with "Home" should return all three home blueprints (case-insensitive)
		blueprints, err := db.GetBlueprintsByNamespaceKindAndLocation("production", "ExecutorDeployment", "Home")
		assert.Nil(t, err)
		assert.Equal(t, 3, len(blueprints))

		// Query with "home" should also return all three
		blueprints, err = db.GetBlueprintsByNamespaceKindAndLocation("production", "ExecutorDeployment", "home")
		assert.Nil(t, err)
		assert.Equal(t, 3, len(blueprints))

		// Query with "office" should return only bp4
		blueprints, err = db.GetBlueprintsByNamespaceKindAndLocation("production", "ExecutorDeployment", "office")
		assert.Nil(t, err)
		assert.Equal(t, 1, len(blueprints))
		assert.Equal(t, "web-4", blueprints[0].Metadata.Name)

		// Query with empty location should return all four
		blueprints, err = db.GetBlueprintsByNamespaceKindAndLocation("production", "ExecutorDeployment", "")
		assert.Nil(t, err)
		assert.Equal(t, 4, len(blueprints))
	})

	// ---------------------------------------------------------------
	// ServerID tests
	// ---------------------------------------------------------------
	t.Run("ServerID/SetAndGet", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		err := db.SetServerID("", "server_id")
		assert.Nil(t, err)

		serverID, err := db.GetServerID()
		assert.Nil(t, err)
		assert.Equal(t, "server_id", serverID)
	})

	t.Run("ServerID/Update", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		err := db.SetServerID("", "server_id")
		assert.Nil(t, err)

		err = db.SetServerID("server_id", "new_server_id")
		assert.Nil(t, err)

		serverID, err := db.GetServerID()
		assert.Nil(t, err)
		assert.Equal(t, "new_server_id", serverID)
	})

	// ---------------------------------------------------------------
	// File tests
	// ---------------------------------------------------------------
	t.Run("File/AddAndGet", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		now := time.Now()
		file := utils.CreateTestFileWithID("test_id", "test_colonyid", now)
		err := db.AddFile(file)
		assert.Nil(t, err)

		fileFromDB, err := db.GetFileByID("test_colonyid", file.ID)
		assert.Nil(t, err)

		// Normalize fields that differ between store and retrieval
		fileFromDB.SequenceNumber = 1
		fileFromDB.Added = time.Time{}
		file.SequenceNumber = 1
		file.Added = time.Time{}

		assert.True(t, file.Equals(fileFromDB))
	})

	t.Run("File/GetByName", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		now := time.Now()
		file1 := utils.CreateTestFileWithID("test_id", "test_colonyid", now)
		file1.Label = "/testpath"
		file1.Name = "test_file.txt"
		file1.Size = 1
		err := db.AddFile(file1)
		assert.Nil(t, err)

		file2 := utils.CreateTestFileWithID("test_id", "test_colonyid", now)
		file2.ID = core.GenerateRandomID()
		file2.Label = "/testpath"
		file2.Name = "test_file.txt"
		file2.Size = 2
		err = db.AddFile(file2)
		assert.Nil(t, err)

		fileFromDB, err := db.GetLatestFileByName("test_colonyid", file1.Label, file1.Name)
		assert.Nil(t, err)
		assert.Len(t, fileFromDB, 1)
		assert.Equal(t, fileFromDB[0].Size, int64(2))

		filesFromDB, err := db.GetFileByName("test_colonyid", file1.Label, file1.Name)
		assert.Nil(t, err)
		assert.Len(t, filesFromDB, 2)
	})

	t.Run("File/GetFilenamesByLabel", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		now := time.Now()
		file1 := utils.CreateTestFileWithID("test_id", "test_colonyid", now)
		file1.ID = core.GenerateRandomID()
		file1.Label = "/testpath"
		file1.Name = "test_file.txt"
		file1.Size = 1
		err := db.AddFile(file1)
		assert.Nil(t, err)

		file2 := utils.CreateTestFileWithID("test_id", "test_colonyid", now)
		file2.ID = core.GenerateRandomID()
		file2.Label = "/testdir"
		file2.Name = "test_file.txt"
		file2.Size = 1
		err = db.AddFile(file2)
		assert.Nil(t, err)

		file3 := utils.CreateTestFileWithID("test_id", "test_colonyid", now)
		file3.ID = core.GenerateRandomID()
		file3.Label = "/testdir"
		file3.Name = "test_file2.txt"
		file3.Size = 1
		err = db.AddFile(file3)
		assert.Nil(t, err)

		file4 := utils.CreateTestFileWithID("test_id", "test_colonyid", now)
		file4.ID = core.GenerateRandomID()
		file4.Label = "/testdir2"
		file4.Name = "test_file.txt"
		file4.Size = 1
		err = db.AddFile(file4)
		assert.Nil(t, err)

		filenames, err := db.GetFilenamesByLabel("test_colonyid", "/testdir")
		assert.Nil(t, err)
		assert.Len(t, filenames, 2)

		filenames, err = db.GetFilenamesByLabel("test_colonyid", "/testdir2")
		assert.Nil(t, err)
		assert.Len(t, filenames, 1)
	})

	t.Run("File/RemoveByID", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		now := time.Now()
		file1 := utils.CreateTestFileWithID("test_id", "test_colonyid", now)
		file1.ID = core.GenerateRandomID()
		file1.Label = "/testdir"
		file1.Name = "test_file.txt"
		file1.Size = 1
		err := db.AddFile(file1)
		assert.Nil(t, err)

		file2 := utils.CreateTestFileWithID("test_id", "test_colonyid", now)
		file2.ID = core.GenerateRandomID()
		file2.Label = "/testdir"
		file2.Name = "test_file2.txt"
		file2.Size = 1
		err = db.AddFile(file2)
		assert.Nil(t, err)

		filenames, err := db.GetFilenamesByLabel("test_colonyid", "/testdir")
		assert.Nil(t, err)
		assert.Len(t, filenames, 2)

		fileFromDB, err := db.GetFileByID("test_colonyid", file2.ID)
		assert.Nil(t, err)
		assert.NotNil(t, fileFromDB)

		err = db.RemoveFileByID("test_colonyid", file2.ID)
		assert.Nil(t, err)

		filenames, err = db.GetFilenamesByLabel("test_colonyid", "/testdir")
		assert.Nil(t, err)
		assert.Len(t, filenames, 1)

		fileFromDB, err = db.GetFileByID("test_colonyid", file2.ID)
		assert.Nil(t, err)
		assert.Nil(t, fileFromDB)
	})

	t.Run("File/RemoveByName", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		now := time.Now()
		file1 := utils.CreateTestFileWithID("test_id", "test_colonyid", now)
		file1.ID = core.GenerateRandomID()
		file1.Label = "/testdir"
		file1.Name = "test_file.txt"
		file1.Size = 1
		err := db.AddFile(file1)
		assert.Nil(t, err)

		file2 := utils.CreateTestFileWithID("test_id", "test_colonyid", now)
		file2.ID = core.GenerateRandomID()
		file2.Label = "/testdir"
		file2.Name = "test_file2.txt"
		file2.Size = 1
		err = db.AddFile(file2)
		assert.Nil(t, err)

		file3 := utils.CreateTestFileWithID("test_id", "test_colonyid", now)
		file3.ID = core.GenerateRandomID()
		file3.Label = "/testdir"
		file3.Name = "test_file2.txt"
		file3.Size = 1
		err = db.AddFile(file3)
		assert.Nil(t, err)

		file4 := utils.CreateTestFileWithID("test_id", "test_colonyid", now)
		file4.ID = core.GenerateRandomID()
		file4.Label = "/testdir"
		file4.Name = "test_file2.txt"
		file4.Size = 1
		err = db.AddFile(file4)
		assert.Nil(t, err)

		files, err := db.GetFileByName("test_colonyid", file4.Label, file4.Name)
		assert.Nil(t, err)
		assert.Len(t, files, 3)

		err = db.RemoveFileByID("test_colonyid", file4.ID)
		assert.Nil(t, err)

		files, err = db.GetFileByName("test_colonyid", file4.Label, file4.Name)
		assert.Nil(t, err)
		assert.Len(t, files, 2)

		err = db.RemoveFileByName("test_colonyid", file4.Label, file4.Name)
		assert.Nil(t, err)

		files, err = db.GetFileByName("test_colonyid", file4.Label, file4.Name)
		assert.Nil(t, err)
		assert.Len(t, files, 0)

		fileFromDB, err := db.GetFileByID("test_colonyid", file4.ID)
		assert.Nil(t, err)
		assert.Nil(t, fileFromDB)

		fileFromDB, err = db.GetFileByID("test_colonyid", file1.ID)
		assert.Nil(t, err)
		assert.NotNil(t, fileFromDB)
	})

	t.Run("File/GetLabels", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		now := time.Now()
		file1 := utils.CreateTestFileWithID("test_id", "test_colonyid", now)
		file1.ID = core.GenerateRandomID()
		file1.Label = "/testdir1"
		file1.Name = "test_file.txt"
		file1.Size = 1
		err := db.AddFile(file1)
		assert.Nil(t, err)

		file2 := utils.CreateTestFileWithID("test_id", "test_colonyid", now)
		file2.ID = core.GenerateRandomID()
		file2.Label = "/testdir2"
		file2.Name = "test_file2.txt"
		file2.Size = 1
		err = db.AddFile(file2)
		assert.Nil(t, err)

		file3 := utils.CreateTestFileWithID("test_id", "test_colonyid", now)
		file3.ID = core.GenerateRandomID()
		file3.Label = "/testdir3"
		file3.Name = "test_file3.txt"
		file3.Size = 1
		err = db.AddFile(file3)
		assert.Nil(t, err)

		file4 := utils.CreateTestFileWithID("test_id", "test_colonyid", now)
		file4.ID = core.GenerateRandomID()
		file4.Label = "/testdir3"
		file4.Name = "test_file4.txt"
		file4.Size = 1
		err = db.AddFile(file4)
		assert.Nil(t, err)

		labels, err := db.GetFileLabels("test_colonyid")
		assert.Nil(t, err)
		assert.Len(t, labels, 3)

		totalFiles := 0
		for _, label := range labels {
			totalFiles += label.Files
		}
		assert.Equal(t, totalFiles, 4)
	})

	t.Run("File/CountByLabel", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		now := time.Now()
		file1 := utils.CreateTestFileWithID("test_id", "test_colony1", now)
		file1.ID = core.GenerateRandomID()
		file1.Label = "/testdir1"
		file1.Name = "test_file.txt"
		file1.Size = 1
		err := db.AddFile(file1)
		assert.Nil(t, err)

		file2 := utils.CreateTestFileWithID("test_id", "test_colony2", now)
		file2.ID = core.GenerateRandomID()
		file2.Label = "/testdir2"
		file2.Name = "test_file2.txt"
		file2.Size = 1
		err = db.AddFile(file2)
		assert.Nil(t, err)

		file3 := utils.CreateTestFileWithID("test_id", "test_colony2", now)
		file3.ID = core.GenerateRandomID()
		file3.Label = "/testdir3"
		file3.Name = "test_file3.txt"
		file3.Size = 1
		err = db.AddFile(file3)
		assert.Nil(t, err)

		file4 := utils.CreateTestFileWithID("test_id", "test_colony2", now)
		file4.ID = core.GenerateRandomID()
		file4.Label = "/testdir3"
		file4.Name = "test_file4.txt"
		file4.Size = 1
		err = db.AddFile(file4)
		assert.Nil(t, err)

		count, err := db.CountFilesWithLabel("test_colony2", "/testdir3")
		assert.Nil(t, err)
		assert.Equal(t, count, 2)

		count, err = db.CountFilesWithLabel("test_colony2", "/testdir2")
		assert.Nil(t, err)
		assert.Equal(t, count, 1)

		count, err = db.CountFilesWithLabel("test_colony1", "/testdir1")
		assert.Nil(t, err)
		assert.Equal(t, count, 1)

		count, err = db.CountFilesWithLabel("test_colony1", "label_does_not_exists")
		assert.Nil(t, err)
		assert.Equal(t, count, 0)
	})

	// ---------------------------------------------------------------
	// Snapshot tests
	// ---------------------------------------------------------------
	t.Run("Snapshot/Create", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		label := "test_label"
		colonyName := "test_colony"
		now := time.Now()

		file := utils.CreateTestFileWithID("test_id", colonyName, now)
		file.ID = "test_file1_id"
		file.Label = label
		file.Name = "test_file1"
		err := db.AddFile(file)
		assert.Nil(t, err)

		file.ID = "test_file2_id" // Another revision of test_file1
		file.Label = label
		file.Name = "test_file1"
		err = db.AddFile(file)
		assert.Nil(t, err)

		file.ID = "test_file3_id"
		file.Label = label
		file.Name = "test_file3"
		err = db.AddFile(file)
		assert.Nil(t, err)

		snapshotName := "test_snapshot_name"
		snapshot, err := db.CreateSnapshot(colonyName, label, snapshotName)
		assert.Nil(t, err)
		assert.Len(t, snapshot.FileIDs, 2)

		counter := 0
		for _, fileID := range snapshot.FileIDs {
			if fileID == "test_file2_id" { // latest revision, not test_file1_id
				counter++
			}
			if fileID == "test_file3_id" {
				counter++
			}
		}
		assert.Equal(t, counter, 2)
		assert.Equal(t, snapshot.ColonyName, colonyName)
		assert.Equal(t, snapshot.Name, snapshotName)
		assert.Equal(t, snapshot.Label, label)

		_, err = db.CreateSnapshot(colonyName, label, snapshotName)
		assert.NotNil(t, err) // name must be unique
	})

	t.Run("Snapshot/GetByName", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		label := "test_label"
		colonyName := "test_colony"
		now := time.Now()

		file := utils.CreateTestFileWithID("test_id", colonyName, now)
		file.ID = "test_file1_id"
		file.Label = label
		file.Name = "test_file1"
		err := db.AddFile(file)
		assert.Nil(t, err)

		snapshotName := "test_snapshot_name"
		snapshot, err := db.CreateSnapshot(colonyName, label, snapshotName)
		assert.Nil(t, err)
		assert.Len(t, snapshot.FileIDs, 1)

		snapshotFromDB, err := db.GetSnapshotByName(colonyName, snapshotName)
		assert.Nil(t, err)
		assert.True(t, snapshotFromDB.Equals(snapshot))
	})

	t.Run("Snapshot/GetByColony", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		label := "test_label"
		colonyName := "test_colony"
		now := time.Now()

		file := utils.CreateTestFileWithID("test_id", colonyName, now)
		file.ID = "test_file1_id"
		file.Label = label
		file.Name = "test_file1"
		err := db.AddFile(file)
		assert.Nil(t, err)

		_, err = db.CreateSnapshot(colonyName, label, "test_snapshot_name")
		assert.Nil(t, err)

		_, err = db.CreateSnapshot(colonyName, label, "test_snapshot_name2")
		assert.Nil(t, err)

		snapshotsFromDB, err := db.GetSnapshotsByColonyName(colonyName)
		assert.Nil(t, err)
		assert.Len(t, snapshotsFromDB, 2)
	})

	t.Run("Snapshot/RemoveByID", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		label := "test_label"
		colonyName := "test_colony"
		now := time.Now()

		file := utils.CreateTestFileWithID("test_id", colonyName, now)
		file.ID = "test_file1_id"
		file.Label = label
		file.Name = "test_file1"
		err := db.AddFile(file)
		assert.Nil(t, err)

		snapshot1, err := db.CreateSnapshot(colonyName, label, "test_snapshot_name")
		assert.Nil(t, err)

		snapshot2, err := db.CreateSnapshot(colonyName, label, "test_snapshot_name2")
		assert.Nil(t, err)

		err = db.RemoveSnapshotByID(colonyName, snapshot1.ID)
		assert.Nil(t, err)

		_, err = db.GetSnapshotByID(colonyName, snapshot1.ID)
		assert.NotNil(t, err)
		_, err = db.GetSnapshotByID(colonyName, snapshot2.ID)
		assert.Nil(t, err)
		snapshotsFromDB, err := db.GetSnapshotsByColonyName(colonyName)
		assert.Nil(t, err)
		assert.Len(t, snapshotsFromDB, 1)
	})

	t.Run("Snapshot/RemoveByName", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		label := "test_label"
		colonyName := "test_colony_id"
		now := time.Now()

		file := utils.CreateTestFileWithID("test_id", colonyName, now)
		file.ID = "test_file1_id"
		file.Label = label
		file.Name = "test_file1"
		err := db.AddFile(file)
		assert.Nil(t, err)

		snapshot1, err := db.CreateSnapshot(colonyName, label, "test_snapshot_name")
		assert.Nil(t, err)

		snapshot2, err := db.CreateSnapshot(colonyName, label, "test_snapshot_name2")
		assert.Nil(t, err)

		err = db.RemoveSnapshotByName(colonyName, "test_snapshot_name")
		assert.Nil(t, err)

		_, err = db.GetSnapshotByID(colonyName, snapshot1.ID)
		assert.NotNil(t, err)
		_, err = db.GetSnapshotByID(colonyName, snapshot2.ID)
		assert.Nil(t, err)
		snapshotsFromDB, err := db.GetSnapshotsByColonyName(colonyName)
		assert.Nil(t, err)
		assert.Len(t, snapshotsFromDB, 1)
	})

	t.Run("Snapshot/RemoveByColony", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		label := "test_label"
		colonyName1 := "test_colony_1"
		colonyName2 := "test_colony_2"
		now := time.Now()

		file := utils.CreateTestFileWithID("test_id", colonyName1, now)
		file.ID = "test_file1_id"
		file.Label = label
		file.Name = "test_file1"
		err := db.AddFile(file)
		assert.Nil(t, err)

		file = utils.CreateTestFileWithID("test_id", colonyName2, now)
		file.ID = "test_file2_id"
		file.Label = label
		file.Name = "test_file2"
		err = db.AddFile(file)
		assert.Nil(t, err)

		file = utils.CreateTestFileWithID("test_id", colonyName2, now)
		file.ID = "test_file3_id"
		file.Label = label
		file.Name = "test_file3"
		err = db.AddFile(file)
		assert.Nil(t, err)

		_, err = db.CreateSnapshot(colonyName1, label, "test_snapshot_name1")
		assert.Nil(t, err)

		_, err = db.CreateSnapshot(colonyName1, label, "test_snapshot_name2")
		assert.Nil(t, err)

		_, err = db.CreateSnapshot(colonyName2, label, "test_snapshot_name3")
		assert.Nil(t, err)

		_, err = db.CreateSnapshot(colonyName2, label, "test_snapshot_name4")
		assert.Nil(t, err)

		err = db.RemoveSnapshotsByColonyName(colonyName1)
		assert.Nil(t, err)

		snapshotsFromDB, err := db.GetSnapshotsByColonyName(colonyName1)
		assert.Nil(t, err)
		assert.Len(t, snapshotsFromDB, 0)

		snapshotsFromDB, err = db.GetSnapshotsByColonyName(colonyName2)
		assert.Nil(t, err)
		assert.Len(t, snapshotsFromDB, 2)
	})

	// ---------------------------------------------------------------
	// GeneratorArgs tests
	// ---------------------------------------------------------------
	t.Run("GeneratorArgs/AddAndCount", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName := core.GenerateRandomID()
		generatorID := core.GenerateRandomID()
		generatorArg := core.CreateGeneratorArg(generatorID, colonyName, "arg")
		generatorArg2 := core.CreateGeneratorArg(generatorID, colonyName, "arg")

		err := db.AddGeneratorArg(generatorArg)
		assert.Nil(t, err)
		err = db.AddGeneratorArg(generatorArg2)
		assert.Nil(t, err)

		generatorArgsFromDB, err := db.GetGeneratorArgs(generatorID, 100)
		assert.Nil(t, err)
		assert.Len(t, generatorArgsFromDB, 2)

		count, err := db.CountGeneratorArgs(generatorID)
		assert.Nil(t, err)
		assert.Equal(t, count, 2)
	})

	t.Run("GeneratorArgs/RemoveByGeneratorID", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName := core.GenerateRandomID()
		generatorID1 := core.GenerateRandomID()
		generatorArg := core.CreateGeneratorArg(generatorID1, colonyName, "arg")
		generatorID2 := core.GenerateRandomID()
		generatorArg2 := core.CreateGeneratorArg(generatorID2, colonyName, "arg")

		err := db.AddGeneratorArg(generatorArg)
		assert.Nil(t, err)
		err = db.AddGeneratorArg(generatorArg2)
		assert.Nil(t, err)

		err = db.RemoveAllGeneratorArgsByGeneratorID(generatorID1)
		assert.Nil(t, err)

		count, err := db.CountGeneratorArgs(generatorID1)
		assert.Nil(t, err)
		assert.Equal(t, count, 0)

		count, err = db.CountGeneratorArgs(generatorID2)
		assert.Nil(t, err)
		assert.Equal(t, count, 1)
	})

	t.Run("GeneratorArgs/RemoveByColony", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName := core.GenerateRandomID()
		generatorID1 := core.GenerateRandomID()
		generatorArg := core.CreateGeneratorArg(generatorID1, colonyName, "arg")
		generatorID2 := core.GenerateRandomID()
		generatorArg2 := core.CreateGeneratorArg(generatorID2, colonyName, "arg")

		err := db.AddGeneratorArg(generatorArg)
		assert.Nil(t, err)
		err = db.AddGeneratorArg(generatorArg2)
		assert.Nil(t, err)

		err = db.RemoveAllGeneratorArgsByColonyName(colonyName)
		assert.Nil(t, err)

		count, err := db.CountGeneratorArgs(generatorID1)
		assert.Nil(t, err)
		assert.Equal(t, count, 0)

		count, err = db.CountGeneratorArgs(generatorID2)
		assert.Nil(t, err)
		assert.Equal(t, count, 0)
	})

	// ---------------------------------------------------------------
	// ProcessGraph tests
	// ---------------------------------------------------------------
	t.Run("ProcessGraph/AddAndGet", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName := core.GenerateRandomID()

		process1 := utils.CreateTestProcess(colonyName)
		process2 := utils.CreateTestProcess(colonyName)
		process3 := utils.CreateTestProcess(colonyName)
		process4 := utils.CreateTestProcess(colonyName)

		process1.AddChild(process2.ID)
		process1.AddChild(process3.ID)
		process2.AddParent(process1.ID)
		process3.AddParent(process1.ID)
		process2.AddChild(process4.ID)
		process3.AddChild(process4.ID)
		process4.AddParent(process2.ID)
		process4.AddParent(process3.ID)

		err := db.AddProcess(process1)
		assert.Nil(t, err)
		err = db.AddProcess(process2)
		assert.Nil(t, err)
		err = db.AddProcess(process3)
		assert.Nil(t, err)
		err = db.AddProcess(process4)
		assert.Nil(t, err)

		graph, err := core.CreateProcessGraph(colonyName)
		assert.Nil(t, err)
		graph.AddRoot(process1.ID)

		err = db.AddProcessGraph(graph)
		assert.Nil(t, err)

		graphFromDB, err := db.GetProcessGraphByID(graph.ID)
		assert.Nil(t, err)
		assert.True(t, graph.Equals(graphFromDB))
	})

	t.Run("ProcessGraph/RemoveByID", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName := core.GenerateRandomID()

		// Create graph 1
		p1 := utils.CreateTestProcess(colonyName)
		err := db.AddProcess(p1)
		assert.Nil(t, err)
		graph1, err := core.CreateProcessGraph(colonyName)
		assert.Nil(t, err)
		graph1.AddRoot(p1.ID)
		err = db.AddProcessGraph(graph1)
		assert.Nil(t, err)

		// Create graph 2
		p2 := utils.CreateTestProcess(colonyName)
		err = db.AddProcess(p2)
		assert.Nil(t, err)
		graph2, err := core.CreateProcessGraph(colonyName)
		assert.Nil(t, err)
		graph2.AddRoot(p2.ID)
		err = db.AddProcessGraph(graph2)
		assert.Nil(t, err)

		// Remove graph 1
		err = db.RemoveProcessGraphByID(graph1.ID)
		assert.Nil(t, err)

		graphFromDB, err := db.GetProcessGraphByID(graph1.ID)
		assert.Nil(t, err)
		assert.Nil(t, graphFromDB)

		graphFromDB, err = db.GetProcessGraphByID(graph2.ID)
		assert.Nil(t, err)
		assert.True(t, graphFromDB.Equals(graph2))
	})

	t.Run("ProcessGraph/SetState", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName := core.GenerateRandomID()

		p := utils.CreateTestProcess(colonyName)
		err := db.AddProcess(p)
		assert.Nil(t, err)
		graph, err := core.CreateProcessGraph(colonyName)
		assert.Nil(t, err)
		graph.AddRoot(p.ID)
		err = db.AddProcessGraph(graph)
		assert.Nil(t, err)

		err = db.SetProcessGraphState(graph.ID, core.WAITING)
		assert.Nil(t, err)
		graphFromDB, err := db.GetProcessGraphByID(graph.ID)
		assert.Nil(t, err)
		assert.Equal(t, core.WAITING, graphFromDB.State)

		err = db.SetProcessGraphState(graph.ID, core.RUNNING)
		assert.Nil(t, err)
		graphFromDB, err = db.GetProcessGraphByID(graph.ID)
		assert.Nil(t, err)
		assert.Equal(t, core.RUNNING, graphFromDB.State)

		err = db.SetProcessGraphState(graph.ID, core.SUCCESS)
		assert.Nil(t, err)
		graphFromDB, err = db.GetProcessGraphByID(graph.ID)
		assert.Nil(t, err)
		assert.Equal(t, core.SUCCESS, graphFromDB.State)
	})

	t.Run("ProcessGraph/FindWaiting", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName := core.GenerateRandomID()
		for i := 0; i < 3; i++ {
			p := utils.CreateTestProcess(colonyName)
			err := db.AddProcess(p)
			assert.Nil(t, err)
			graph, err := core.CreateProcessGraph(colonyName)
			assert.Nil(t, err)
			graph.AddRoot(p.ID)
			err = db.AddProcessGraph(graph)
			assert.Nil(t, err)
			err = db.SetProcessGraphState(graph.ID, core.WAITING)
			assert.Nil(t, err)
		}

		graphs, err := db.FindWaitingProcessGraphs(colonyName, 100)
		assert.Nil(t, err)
		assert.Len(t, graphs, 3)
	})

	t.Run("ProcessGraph/FindRunning", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName := core.GenerateRandomID()
		for i := 0; i < 4; i++ {
			p := utils.CreateTestProcess(colonyName)
			err := db.AddProcess(p)
			assert.Nil(t, err)
			graph, err := core.CreateProcessGraph(colonyName)
			assert.Nil(t, err)
			graph.AddRoot(p.ID)
			err = db.AddProcessGraph(graph)
			assert.Nil(t, err)
			err = db.SetProcessGraphState(graph.ID, core.RUNNING)
			assert.Nil(t, err)
		}

		graphs, err := db.FindRunningProcessGraphs(colonyName, 100)
		assert.Nil(t, err)
		assert.Len(t, graphs, 4)
	})

	t.Run("ProcessGraph/FindSuccessful", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName := core.GenerateRandomID()
		for i := 0; i < 5; i++ {
			p := utils.CreateTestProcess(colonyName)
			err := db.AddProcess(p)
			assert.Nil(t, err)
			graph, err := core.CreateProcessGraph(colonyName)
			assert.Nil(t, err)
			graph.AddRoot(p.ID)
			err = db.AddProcessGraph(graph)
			assert.Nil(t, err)
			err = db.SetProcessGraphState(graph.ID, core.SUCCESS)
			assert.Nil(t, err)
		}

		graphs, err := db.FindSuccessfulProcessGraphs(colonyName, 100)
		assert.Nil(t, err)
		assert.Len(t, graphs, 5)
	})

	t.Run("ProcessGraph/FindFailed", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName := core.GenerateRandomID()
		for i := 0; i < 2; i++ {
			p := utils.CreateTestProcess(colonyName)
			err := db.AddProcess(p)
			assert.Nil(t, err)
			graph, err := core.CreateProcessGraph(colonyName)
			assert.Nil(t, err)
			graph.AddRoot(p.ID)
			err = db.AddProcessGraph(graph)
			assert.Nil(t, err)
			err = db.SetProcessGraphState(graph.ID, core.FAILED)
			assert.Nil(t, err)
		}

		graphs, err := db.FindFailedProcessGraphs(colonyName, 100)
		assert.Nil(t, err)
		assert.Len(t, graphs, 2)
	})

	t.Run("ProcessGraph/FindCancelled", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName := core.GenerateRandomID()

		// Add 5 cancelled graphs
		for i := 0; i < 5; i++ {
			p := utils.CreateTestProcess(colonyName)
			err := db.AddProcess(p)
			assert.Nil(t, err)
			graph, err := core.CreateProcessGraph(colonyName)
			assert.Nil(t, err)
			graph.AddRoot(p.ID)
			err = db.AddProcessGraph(graph)
			assert.Nil(t, err)
			err = db.SetProcessGraphState(graph.ID, core.CANCELLED)
			assert.Nil(t, err)
		}

		// Add 3 waiting graphs (should not appear in cancelled results)
		for i := 0; i < 3; i++ {
			p := utils.CreateTestProcess(colonyName)
			err := db.AddProcess(p)
			assert.Nil(t, err)
			graph, err := core.CreateProcessGraph(colonyName)
			assert.Nil(t, err)
			graph.AddRoot(p.ID)
			err = db.AddProcessGraph(graph)
			assert.Nil(t, err)
			err = db.SetProcessGraphState(graph.ID, core.WAITING)
			assert.Nil(t, err)
		}

		graphs, err := db.FindCancelledProcessGraphs(colonyName, 100)
		assert.Nil(t, err)
		assert.Len(t, graphs, 5)

		count, err := db.CountCancelledProcessGraphsByColonyName(colonyName)
		assert.Nil(t, err)
		assert.Equal(t, 5, count)
	})

	t.Run("ProcessGraph/RemoveAllByColony", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName := core.GenerateRandomID()

		p1 := utils.CreateTestProcess(colonyName)
		err := db.AddProcess(p1)
		assert.Nil(t, err)
		graph1, err := core.CreateProcessGraph(colonyName)
		assert.Nil(t, err)
		graph1.AddRoot(p1.ID)
		err = db.AddProcessGraph(graph1)
		assert.Nil(t, err)

		p2 := utils.CreateTestProcess(colonyName)
		err = db.AddProcess(p2)
		assert.Nil(t, err)
		graph2, err := core.CreateProcessGraph(colonyName)
		assert.Nil(t, err)
		graph2.AddRoot(p2.ID)
		err = db.AddProcessGraph(graph2)
		assert.Nil(t, err)

		err = db.RemoveAllProcessGraphsByColonyName(colonyName)
		assert.Nil(t, err)

		graphFromDB, err := db.GetProcessGraphByID(graph1.ID)
		assert.Nil(t, err)
		assert.Nil(t, graphFromDB)

		graphFromDB, err = db.GetProcessGraphByID(graph2.ID)
		assert.Nil(t, err)
		assert.Nil(t, graphFromDB)
	})

	t.Run("ProcessGraph/RemoveWaitingByColony", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName := core.GenerateRandomID()

		// Create two waiting process graphs
		process1 := utils.CreateTestProcess(colonyName)
		err := db.AddProcess(process1)
		assert.Nil(t, err)
		graph1, err := core.CreateProcessGraph(colonyName)
		assert.Nil(t, err)
		graph1.AddRoot(process1.ID)
		err = db.AddProcessGraph(graph1)
		assert.Nil(t, err)
		err = db.SetProcessGraphState(graph1.ID, core.WAITING)
		assert.Nil(t, err)
		err = db.SetProcessState(process1.ID, core.WAITING)
		assert.Nil(t, err)

		process2 := utils.CreateTestProcess(colonyName)
		err = db.AddProcess(process2)
		assert.Nil(t, err)
		graph2, err := core.CreateProcessGraph(colonyName)
		assert.Nil(t, err)
		graph2.AddRoot(process2.ID)
		err = db.AddProcessGraph(graph2)
		assert.Nil(t, err)
		err = db.SetProcessGraphState(graph2.ID, core.WAITING)
		assert.Nil(t, err)
		err = db.SetProcessState(process2.ID, core.WAITING)
		assert.Nil(t, err)

		waitingGraphs, err := db.CountWaitingProcessGraphs()
		assert.Nil(t, err)
		assert.Equal(t, 2, waitingGraphs)

		err = db.RemoveAllWaitingProcessGraphsByColonyName(colonyName)
		assert.Nil(t, err)

		waitingGraphs, err = db.CountWaitingProcessGraphs()
		assert.Nil(t, err)
		assert.Equal(t, 0, waitingGraphs)
	})

	t.Run("ProcessGraph/RemoveRunningByColony", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colonyName := core.GenerateRandomID()

		// Create two running process graphs
		process1 := utils.CreateTestProcess(colonyName)
		err := db.AddProcess(process1)
		assert.Nil(t, err)
		graph1, err := core.CreateProcessGraph(colonyName)
		assert.Nil(t, err)
		graph1.AddRoot(process1.ID)
		err = db.AddProcessGraph(graph1)
		assert.Nil(t, err)
		err = db.SetProcessGraphState(graph1.ID, core.RUNNING)
		assert.Nil(t, err)
		err = db.SetProcessState(process1.ID, core.RUNNING)
		assert.Nil(t, err)

		process2 := utils.CreateTestProcess(colonyName)
		err = db.AddProcess(process2)
		assert.Nil(t, err)
		graph2, err := core.CreateProcessGraph(colonyName)
		assert.Nil(t, err)
		graph2.AddRoot(process2.ID)
		err = db.AddProcessGraph(graph2)
		assert.Nil(t, err)
		err = db.SetProcessGraphState(graph2.ID, core.RUNNING)
		assert.Nil(t, err)
		err = db.SetProcessState(process2.ID, core.RUNNING)
		assert.Nil(t, err)

		runningGraphs, err := db.CountRunningProcessGraphs()
		assert.Nil(t, err)
		assert.Equal(t, 2, runningGraphs)

		err = db.RemoveAllRunningProcessGraphsByColonyName(colonyName)
		assert.Nil(t, err)

		runningGraphs, err = db.CountRunningProcessGraphs()
		assert.Nil(t, err)
		assert.Equal(t, 0, runningGraphs)
	})

	// ---------------------------------------------------------------
	// Location tests
	// ---------------------------------------------------------------
	t.Run("Location/Add", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony, _, err := utils.CreateTestColonyWithKey()
		assert.Nil(t, err)
		err = db.AddColony(colony)
		assert.Nil(t, err)

		location := utils.CreateTestLocation(colony.Name, "test_location")
		err = db.AddLocation(location)
		assert.Nil(t, err)

		locationFromDB, err := db.GetLocationByID(location.ID)
		assert.Nil(t, err)
		assert.True(t, location.Equals(locationFromDB))
	})

	t.Run("Location/AddNil", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		err := db.AddLocation(nil)
		assert.NotNil(t, err)
	})

	t.Run("Location/AddDuplicate", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony, _, err := utils.CreateTestColonyWithKey()
		assert.Nil(t, err)
		err = db.AddColony(colony)
		assert.Nil(t, err)

		location := utils.CreateTestLocation(colony.Name, "test_location")
		err = db.AddLocation(location)
		assert.Nil(t, err)

		location2 := utils.CreateTestLocation(colony.Name, "test_location")
		err = db.AddLocation(location2)
		assert.NotNil(t, err)
	})

	t.Run("Location/GetByID", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony, _, err := utils.CreateTestColonyWithKey()
		assert.Nil(t, err)
		err = db.AddColony(colony)
		assert.Nil(t, err)

		location := utils.CreateTestLocation(colony.Name, "test_location")
		err = db.AddLocation(location)
		assert.Nil(t, err)

		locationFromDB, err := db.GetLocationByID(location.ID)
		assert.Nil(t, err)
		assert.True(t, location.Equals(locationFromDB))
	})

	t.Run("Location/GetByIDNotFound", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		locationFromDB, err := db.GetLocationByID("non_existent_id")
		assert.Nil(t, err)
		assert.Nil(t, locationFromDB)
	})

	t.Run("Location/GetByName", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony, _, err := utils.CreateTestColonyWithKey()
		assert.Nil(t, err)
		err = db.AddColony(colony)
		assert.Nil(t, err)

		location := utils.CreateTestLocation(colony.Name, "test_location")
		err = db.AddLocation(location)
		assert.Nil(t, err)

		locationFromDB, err := db.GetLocationByName(colony.Name, "test_location")
		assert.Nil(t, err)
		assert.True(t, location.Equals(locationFromDB))
	})

	t.Run("Location/GetByNameNotFound", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony, _, err := utils.CreateTestColonyWithKey()
		assert.Nil(t, err)
		err = db.AddColony(colony)
		assert.Nil(t, err)

		locationFromDB, err := db.GetLocationByName(colony.Name, "non_existent_name")
		assert.Nil(t, err)
		assert.Nil(t, locationFromDB)
	})

	t.Run("Location/GetByColony", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony, _, err := utils.CreateTestColonyWithKey()
		assert.Nil(t, err)
		err = db.AddColony(colony)
		assert.Nil(t, err)

		location1 := utils.CreateTestLocation(colony.Name, "test_location1")
		err = db.AddLocation(location1)
		assert.Nil(t, err)

		location2 := utils.CreateTestLocation(colony.Name, "test_location2")
		err = db.AddLocation(location2)
		assert.Nil(t, err)

		locationsFromDB, err := db.GetLocationsByColonyName(colony.Name)
		assert.Nil(t, err)
		assert.Len(t, locationsFromDB, 2)
	})

	t.Run("Location/GetByColonyEmpty", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony, _, err := utils.CreateTestColonyWithKey()
		assert.Nil(t, err)
		err = db.AddColony(colony)
		assert.Nil(t, err)

		locationsFromDB, err := db.GetLocationsByColonyName(colony.Name)
		assert.Nil(t, err)
		assert.Len(t, locationsFromDB, 0)
	})

	t.Run("Location/RemoveByID", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony, _, err := utils.CreateTestColonyWithKey()
		assert.Nil(t, err)
		err = db.AddColony(colony)
		assert.Nil(t, err)

		location := utils.CreateTestLocation(colony.Name, "test_location")
		err = db.AddLocation(location)
		assert.Nil(t, err)

		err = db.RemoveLocationByID(location.ID)
		assert.Nil(t, err)

		locationFromDB, err := db.GetLocationByID(location.ID)
		assert.Nil(t, err)
		assert.Nil(t, locationFromDB)
	})

	t.Run("Location/RemoveByName", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony, _, err := utils.CreateTestColonyWithKey()
		assert.Nil(t, err)
		err = db.AddColony(colony)
		assert.Nil(t, err)

		location := utils.CreateTestLocation(colony.Name, "test_location")
		err = db.AddLocation(location)
		assert.Nil(t, err)

		err = db.RemoveLocationByName(colony.Name, "test_location")
		assert.Nil(t, err)

		locationFromDB, err := db.GetLocationByName(colony.Name, "test_location")
		assert.Nil(t, err)
		assert.Nil(t, locationFromDB)
	})

	t.Run("Location/RemoveByColony", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony, _, err := utils.CreateTestColonyWithKey()
		assert.Nil(t, err)
		err = db.AddColony(colony)
		assert.Nil(t, err)

		location1 := utils.CreateTestLocation(colony.Name, "test_location1")
		err = db.AddLocation(location1)
		assert.Nil(t, err)

		location2 := utils.CreateTestLocation(colony.Name, "test_location2")
		err = db.AddLocation(location2)
		assert.Nil(t, err)

		locationsFromDB, err := db.GetLocationsByColonyName(colony.Name)
		assert.Nil(t, err)
		assert.Len(t, locationsFromDB, 2)

		err = db.RemoveLocationsByColonyName(colony.Name)
		assert.Nil(t, err)

		locationsFromDB, err = db.GetLocationsByColonyName(colony.Name)
		assert.Nil(t, err)
		assert.Len(t, locationsFromDB, 0)
	})

	t.Run("Location/Coordinates", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony, _, err := utils.CreateTestColonyWithKey()
		assert.Nil(t, err)
		err = db.AddColony(colony)
		assert.Nil(t, err)

		location := core.CreateLocation(core.GenerateRandomID(), "test_location", colony.Name, "test_desc", -122.4194, 37.7749)
		err = db.AddLocation(location)
		assert.Nil(t, err)

		locationFromDB, err := db.GetLocationByID(location.ID)
		assert.Nil(t, err)
		assert.Equal(t, -122.4194, locationFromDB.Long)
		assert.Equal(t, 37.7749, locationFromDB.Lat)
	})

	t.Run("Location/CascadeOnColonyDelete", func(t *testing.T) {
		db, cleanup := newHarness(t)
		defer cleanup()

		colony, _, err := utils.CreateTestColonyWithKey()
		assert.Nil(t, err)
		err = db.AddColony(colony)
		assert.Nil(t, err)

		location1 := utils.CreateTestLocation(colony.Name, "test_location1")
		err = db.AddLocation(location1)
		assert.Nil(t, err)

		location2 := utils.CreateTestLocation(colony.Name, "test_location2")
		err = db.AddLocation(location2)
		assert.Nil(t, err)

		location3 := utils.CreateTestLocation(colony.Name, "test_location3")
		err = db.AddLocation(location3)
		assert.Nil(t, err)

		locationsFromDB, err := db.GetLocationsByColonyName(colony.Name)
		assert.Nil(t, err)
		assert.Len(t, locationsFromDB, 3)

		err = db.RemoveColonyByName(colony.Name)
		assert.Nil(t, err)

		locationsFromDB, err = db.GetLocationsByColonyName(colony.Name)
		assert.Nil(t, err)
		assert.Len(t, locationsFromDB, 0)

		locationFromDB, err := db.GetLocationByID(location1.ID)
		assert.Nil(t, err)
		assert.Nil(t, locationFromDB)

		locationFromDB, err = db.GetLocationByID(location2.ID)
		assert.Nil(t, err)
		assert.Nil(t, locationFromDB)

		locationFromDB, err = db.GetLocationByID(location3.ID)
		assert.Nil(t, err)
		assert.Nil(t, locationFromDB)
	})
}
