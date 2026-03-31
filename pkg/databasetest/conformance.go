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
