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
}
