package daos_test

import (
	"errors"
	"testing"
	"time"

	"github.com/pafthang/servicebase/daos"
	basemodels "github.com/pafthang/servicebase/services/base/models"

	collectionmodels "github.com/pafthang/servicebase/services/collection/models"
	teammodels "github.com/pafthang/servicebase/services/team/models"
	"github.com/pafthang/servicebase/tests"
)

func createTestTeam(t *testing.T, app *tests.TestApp, name string) *teammodels.Team {
	t.Helper()

	team := &teammodels.Team{Name: name}
	if err := app.Dao().SaveTeam(team); err != nil {
		t.Fatalf("failed to create test team %q: %v", name, err)
	}

	return team
}

func findExistingAdminTeam(t *testing.T, app *tests.TestApp) *teammodels.Team {
	t.Helper()

	team, err := app.Dao().FindTeamByName("admin")
	if err != nil {
		t.Fatalf("failed to load existing admin team: %v", err)
	}

	return team
}

func TestNew(t *testing.T) {
	testApp, _ := tests.NewTestApp()
	defer testApp.Cleanup()

	dao := daos.New(testApp.Dao().DB())

	if dao.DB() != testApp.Dao().DB() {
		t.Fatal("The 2 db instances are different")
	}
}

func TestNewMultiDB(t *testing.T) {
	testApp, _ := tests.NewTestApp()
	defer testApp.Cleanup()

	dao := daos.NewMultiDB(testApp.Dao().ConcurrentDB(), testApp.Dao().NonconcurrentDB())

	if dao.DB() != testApp.Dao().ConcurrentDB() {
		t.Fatal("[db-concurrentDB] The 2 db instances are different")
	}

	if dao.ConcurrentDB() != testApp.Dao().ConcurrentDB() {
		t.Fatal("[concurrentDB-concurrentDB] The 2 db instances are different")
	}

	if dao.NonconcurrentDB() != testApp.Dao().NonconcurrentDB() {
		t.Fatal("[nonconcurrentDB-nonconcurrentDB] The 2 db instances are different")
	}
}

func TestDaoClone(t *testing.T) {
	testApp, _ := tests.NewTestApp()
	defer testApp.Cleanup()

	hookCalls := map[string]int{}

	dao := daos.NewMultiDB(testApp.Dao().ConcurrentDB(), testApp.Dao().NonconcurrentDB())
	dao.MaxLockRetries = 1
	dao.ModelQueryTimeout = 2
	dao.BeforeDeleteFunc = func(eventDao *daos.Dao, m basemodels.Model, action func() error) error {
		hookCalls["BeforeDeleteFunc"]++
		return action()
	}
	dao.BeforeUpdateFunc = func(eventDao *daos.Dao, m basemodels.Model, action func() error) error {
		hookCalls["BeforeUpdateFunc"]++
		return action()
	}
	dao.BeforeCreateFunc = func(eventDao *daos.Dao, m basemodels.Model, action func() error) error {
		hookCalls["BeforeCreateFunc"]++
		return action()
	}
	dao.AfterDeleteFunc = func(eventDao *daos.Dao, m basemodels.Model) error {
		hookCalls["AfterDeleteFunc"]++
		return nil
	}
	dao.AfterUpdateFunc = func(eventDao *daos.Dao, m basemodels.Model) error {
		hookCalls["AfterUpdateFunc"]++
		return nil
	}
	dao.AfterCreateFunc = func(eventDao *daos.Dao, m basemodels.Model) error {
		hookCalls["AfterCreateFunc"]++
		return nil
	}

	clone := dao.Clone()
	clone.MaxLockRetries = 3
	clone.ModelQueryTimeout = 4
	clone.AfterCreateFunc = func(eventDao *daos.Dao, m basemodels.Model) error {
		hookCalls["NewAfterCreateFunc"]++
		return nil
	}

	if dao.MaxLockRetries == clone.MaxLockRetries {
		t.Fatal("Expected different MaxLockRetries")
	}

	if dao.ModelQueryTimeout == clone.ModelQueryTimeout {
		t.Fatal("Expected different ModelQueryTimeout")
	}

	emptyAction := func() error { return nil }

	// trigger hooks
	dao.BeforeDeleteFunc(nil, nil, emptyAction)
	dao.BeforeUpdateFunc(nil, nil, emptyAction)
	dao.BeforeCreateFunc(nil, nil, emptyAction)
	dao.AfterDeleteFunc(nil, nil)
	dao.AfterUpdateFunc(nil, nil)
	dao.AfterCreateFunc(nil, nil)
	clone.BeforeDeleteFunc(nil, nil, emptyAction)
	clone.BeforeUpdateFunc(nil, nil, emptyAction)
	clone.BeforeCreateFunc(nil, nil, emptyAction)
	clone.AfterDeleteFunc(nil, nil)
	clone.AfterUpdateFunc(nil, nil)
	clone.AfterCreateFunc(nil, nil)

	expectations := []struct {
		hook  string
		total int
	}{
		{"BeforeDeleteFunc", 2},
		{"BeforeUpdateFunc", 2},
		{"BeforeCreateFunc", 2},
		{"AfterDeleteFunc", 2},
		{"AfterUpdateFunc", 2},
		{"AfterCreateFunc", 1},
		{"NewAfterCreateFunc", 1},
	}

	for _, e := range expectations {
		if hookCalls[e.hook] != e.total {
			t.Errorf("Expected %s to be caleed %d", e.hook, e.total)
		}
	}
}

func TestDaoWithoutHooks(t *testing.T) {
	testApp, _ := tests.NewTestApp()
	defer testApp.Cleanup()

	hookCalls := map[string]int{}

	dao := daos.NewMultiDB(testApp.Dao().ConcurrentDB(), testApp.Dao().NonconcurrentDB())
	dao.MaxLockRetries = 1
	dao.ModelQueryTimeout = 2
	dao.BeforeDeleteFunc = func(eventDao *daos.Dao, m basemodels.Model, action func() error) error {
		hookCalls["BeforeDeleteFunc"]++
		return action()
	}
	dao.BeforeUpdateFunc = func(eventDao *daos.Dao, m basemodels.Model, action func() error) error {
		hookCalls["BeforeUpdateFunc"]++
		return action()
	}
	dao.BeforeCreateFunc = func(eventDao *daos.Dao, m basemodels.Model, action func() error) error {
		hookCalls["BeforeCreateFunc"]++
		return action()
	}
	dao.AfterDeleteFunc = func(eventDao *daos.Dao, m basemodels.Model) error {
		hookCalls["AfterDeleteFunc"]++
		return nil
	}
	dao.AfterUpdateFunc = func(eventDao *daos.Dao, m basemodels.Model) error {
		hookCalls["AfterUpdateFunc"]++
		return nil
	}
	dao.AfterCreateFunc = func(eventDao *daos.Dao, m basemodels.Model) error {
		hookCalls["AfterCreateFunc"]++
		return nil
	}

	new := dao.WithoutHooks()

	if new.MaxLockRetries != dao.MaxLockRetries {
		t.Fatalf("Expected MaxLockRetries %d, got %d", new.Clone().MaxLockRetries, dao.MaxLockRetries)
	}

	if new.ModelQueryTimeout != dao.ModelQueryTimeout {
		t.Fatalf("Expected ModelQueryTimeout %d, got %d", new.Clone().ModelQueryTimeout, dao.ModelQueryTimeout)
	}

	if new.BeforeDeleteFunc != nil {
		t.Fatal("Expected BeforeDeleteFunc to be nil")
	}

	if new.BeforeUpdateFunc != nil {
		t.Fatal("Expected BeforeUpdateFunc to be nil")
	}

	if new.BeforeCreateFunc != nil {
		t.Fatal("Expected BeforeCreateFunc to be nil")
	}

	if new.AfterDeleteFunc != nil {
		t.Fatal("Expected AfterDeleteFunc to be nil")
	}

	if new.AfterUpdateFunc != nil {
		t.Fatal("Expected AfterUpdateFunc to be nil")
	}

	if new.AfterCreateFunc != nil {
		t.Fatal("Expected AfterCreateFunc to be nil")
	}
}

func TestDaoModelQuery(t *testing.T) {
	testApp, _ := tests.NewTestApp()
	defer testApp.Cleanup()

	dao := daos.New(testApp.Dao().DB())

	scenarios := []struct {
		model    basemodels.Model
		expected string
	}{
		{
			&collectionmodels.Collection{},
			"SELECT {{_collections}}.* FROM `_collections`",
		},
		{
			&teammodels.Team{},
			"SELECT {{teams}}.* FROM `teams`",
		},
	}

	for i, scenario := range scenarios {
		sql := dao.ModelQuery(scenario.model).Build().SQL()
		if sql != scenario.expected {
			t.Errorf("(%d) Expected select %s, got %s", i, scenario.expected, sql)
		}
	}
}

func TestDaoModelQueryCancellation(t *testing.T) {
	testApp, _ := tests.NewTestApp()
	defer testApp.Cleanup()

	dao := daos.New(testApp.Dao().DB())

	m := createTestTeam(t, testApp, "query-cancel-test")

	if err := dao.ModelQuery(m).One(m); err != nil {
		t.Fatalf("Failed to execute control query: %v", err)
	}

	dao.ModelQueryTimeout = 0 * time.Millisecond
	if err := dao.ModelQuery(m).One(m); err == nil {
		t.Fatal("Expected to be cancelled, got nil")
	}
}

func TestDaoFindById(t *testing.T) {
	testApp, _ := tests.NewTestApp()
	defer testApp.Cleanup()

	scenarios := []struct {
		model       basemodels.Model
		id          string
		expectError bool
	}{
		// missing id
		{
			&collectionmodels.Collection{},
			"missing",
			true,
		},
		// existing collection id
		{
			&collectionmodels.Collection{},
			"wsmn24bux7wo113",
			false,
		},
		// existing team id
		{
			createTestTeam(t, testApp, "find-by-id-test"),
			"",
			false,
		},
	}

	scenarios[2].id = scenarios[2].model.GetId()

	for i, scenario := range scenarios {
		err := testApp.Dao().FindById(scenario.model, scenario.id)
		hasErr := err != nil
		if hasErr != scenario.expectError {
			t.Errorf("(%d) Expected %v, got %v", i, scenario.expectError, err)
		}

		if !scenario.expectError && scenario.id != scenario.model.GetId() {
			t.Errorf("(%d) Expected model with id %v, got %v", i, scenario.id, scenario.model.GetId())
		}
	}
}

func TestDaoRunInTransaction(t *testing.T) {
	testApp, _ := tests.NewTestApp()
	defer testApp.Cleanup()

	team := createTestTeam(t, testApp, "tx-team")

	// failed nested transaction
	testApp.Dao().RunInTransaction(func(txDao *daos.Dao) error {
		currentTeam, _ := txDao.FindTeamById(team.Id)

		return txDao.RunInTransaction(func(tx2Dao *daos.Dao) error {
			if err := tx2Dao.DeleteTeam(currentTeam); err != nil {
				t.Fatal(err)
			}
			return errors.New("test error")
		})
	})

	// team should still exist
	team1, _ := testApp.Dao().FindTeamById(team.Id)
	if team1 == nil {
		t.Fatal("Expected team to not be deleted")
	}

	// successful nested transaction
	testApp.Dao().RunInTransaction(func(txDao *daos.Dao) error {
		currentTeam, _ := txDao.FindTeamById(team.Id)

		return txDao.RunInTransaction(func(tx2Dao *daos.Dao) error {
			return tx2Dao.DeleteTeam(currentTeam)
		})
	})

	// team should have been deleted
	team2, _ := testApp.Dao().FindTeamById(team.Id)
	if team2 != nil {
		t.Fatalf("Expected team %s to be deleted, found %v", team.Id, team2)
	}
}

func TestDaoSaveCreate(t *testing.T) {
	testApp, _ := tests.NewTestApp()
	defer testApp.Cleanup()

	model := &teammodels.Team{}
	model.Name = "test_new_team"
	if err := testApp.Dao().Save(model); err != nil {
		t.Fatal(err)
	}

	// refresh
	model, _ = testApp.Dao().FindTeamById(model.Id)

	if model.Name != "test_new_team" {
		t.Fatalf("Expected model name %q, got %q", "test_new_team", model.Name)
	}

	expectedHooks := []string{"OnModelBeforeCreate", "OnModelAfterCreate"}
	for _, h := range expectedHooks {
		if v, ok := testApp.EventCalls[h]; !ok || v != 1 {
			t.Fatalf("Expected event %s to be called exactly one time, got %d", h, v)
		}
	}
}

func TestDaoSaveWithInsertId(t *testing.T) {
	testApp, _ := tests.NewTestApp()
	defer testApp.Cleanup()

	model := &teammodels.Team{}
	model.Id = "test"
	model.Name = "test_new_team"
	model.MarkAsNew()
	if err := testApp.Dao().Save(model); err != nil {
		t.Fatal(err)
	}

	// refresh
	model, _ = testApp.Dao().FindTeamById("test")

	if model == nil {
		t.Fatal("Failed to find team with id 'test'")
	}

	expectedHooks := []string{"OnModelBeforeCreate", "OnModelAfterCreate"}
	for _, h := range expectedHooks {
		if v, ok := testApp.EventCalls[h]; !ok || v != 1 {
			t.Fatalf("Expected event %s to be called exactly one time, got %d", h, v)
		}
	}
}

func TestDaoSaveUpdate(t *testing.T) {
	testApp, _ := tests.NewTestApp()
	defer testApp.Cleanup()

	model := createTestTeam(t, testApp, "test_update_team")

	model.Name = "test_update_team_changed"
	if err := testApp.Dao().Save(model); err != nil {
		t.Fatal(err)
	}

	// refresh
	model, _ = testApp.Dao().FindTeamById(model.Id)

	if model.Name != "test_update_team_changed" {
		t.Fatalf("Expected model name to be updated, got %q", model.Name)
	}

	expectedHooks := []string{"OnModelBeforeUpdate", "OnModelAfterUpdate"}
	for _, h := range expectedHooks {
		if v, ok := testApp.EventCalls[h]; !ok || v != 1 {
			t.Fatalf("Expected event %s to be called exactly one time, got %d", h, v)
		}
	}
}

type dummyColumnValueMapper struct {
	teammodels.Team
}

func (a *dummyColumnValueMapper) ColumnValueMap() map[string]any {
	return map[string]any{
		"name": "mapped_" + a.Name,
	}
}

func TestDaoSaveWithColumnValueMapper(t *testing.T) {
	testApp, _ := tests.NewTestApp()
	defer testApp.Cleanup()

	model := &dummyColumnValueMapper{}
	model.Id = "test_mapped_id" // explicitly set an id
	model.Name = "test_mapped_create"
	model.MarkAsNew()
	if err := testApp.Dao().Save(model); err != nil {
		t.Fatal(err)
	}

	createdModel, _ := testApp.Dao().FindTeamById("test_mapped_id")
	if createdModel == nil {
		t.Fatal("[create] Failed to find model with id 'test_mapped_id'")
	}
	if createdModel.Name != "mapped_"+model.Name {
		t.Fatalf("Expected model with mapped name %q, got %q", "mapped_"+model.Name, createdModel.Name)
	}

	model.Name = "test_mapped_update"
	if err := testApp.Dao().Save(model); err != nil {
		t.Fatal(err)
	}

	updatedModel, _ := testApp.Dao().FindTeamById("test_mapped_id")
	if updatedModel == nil {
		t.Fatal("[update] Failed to find model with id 'test_mapped_id'")
	}
	if updatedModel.Name != "mapped_"+model.Name {
		t.Fatalf("Expected model with mapped name %q, got %q", "mapped_"+model.Name, updatedModel.Name)
	}
}

func TestDaoDelete(t *testing.T) {
	testApp, _ := tests.NewTestApp()
	defer testApp.Cleanup()

	model := createTestTeam(t, testApp, "test_delete_team")

	if err := testApp.Dao().Delete(model); err != nil {
		t.Fatal(err)
	}

	model, _ = testApp.Dao().FindTeamById(model.Id)
	if model != nil {
		t.Fatalf("Expected model to be deleted, found %v", model)
	}

	expectedHooks := []string{"OnModelBeforeDelete", "OnModelAfterDelete"}
	for _, h := range expectedHooks {
		if v, ok := testApp.EventCalls[h]; !ok || v != 1 {
			t.Fatalf("Expected event %s to be called exactly one time, got %d", h, v)
		}
	}
}

func TestDaoRetryCreate(t *testing.T) {
	testApp, _ := tests.NewTestApp()
	defer testApp.Cleanup()

	// init mock retry dao
	retryBeforeCreateHookCalls := 0
	retryAfterCreateHookCalls := 0
	retryDao := daos.New(testApp.Dao().DB())
	retryDao.BeforeCreateFunc = func(eventDao *daos.Dao, m basemodels.Model, action func() error) error {
		retryBeforeCreateHookCalls++
		return errors.New("database is locked")
	}
	retryDao.AfterCreateFunc = func(eventDao *daos.Dao, m basemodels.Model) error {
		retryAfterCreateHookCalls++
		return nil
	}

	model := &teammodels.Team{Name: "retry_create_team"}
	if err := retryDao.Save(model); err != nil {
		t.Fatalf("Expected nil after retry, got error: %v", err)
	}

	// the before hook is expected to be called only once because
	// it is ignored after the first "database is locked" error
	if retryBeforeCreateHookCalls != 1 {
		t.Fatalf("Expected before hook calls to be 1, got %d", retryBeforeCreateHookCalls)
	}

	if retryAfterCreateHookCalls != 1 {
		t.Fatalf("Expected after hook calls to be 1, got %d", retryAfterCreateHookCalls)
	}

	// with non-locking error
	retryBeforeCreateHookCalls = 0
	retryAfterCreateHookCalls = 0
	retryDao.BeforeCreateFunc = func(eventDao *daos.Dao, m basemodels.Model, action func() error) error {
		retryBeforeCreateHookCalls++
		return errors.New("non-locking error")
	}

	dummy := &teammodels.Team{Name: "retry_create_error_team"}
	if err := retryDao.Save(dummy); err == nil {
		t.Fatal("Expected error, got nil")
	}

	if retryBeforeCreateHookCalls != 1 {
		t.Fatalf("Expected before hook calls to be 1, got %d", retryBeforeCreateHookCalls)
	}

	if retryAfterCreateHookCalls != 0 {
		t.Fatalf("Expected after hook calls to be 0, got %d", retryAfterCreateHookCalls)
	}
}

func TestDaoRetryUpdate(t *testing.T) {
	testApp, _ := tests.NewTestApp()
	defer testApp.Cleanup()

	model := createTestTeam(t, testApp, "retry_update_team")

	// init mock retry dao
	retryBeforeUpdateHookCalls := 0
	retryAfterUpdateHookCalls := 0
	retryDao := daos.New(testApp.Dao().DB())
	retryDao.BeforeUpdateFunc = func(eventDao *daos.Dao, m basemodels.Model, action func() error) error {
		retryBeforeUpdateHookCalls++
		return errors.New("database is locked")
	}
	retryDao.AfterUpdateFunc = func(eventDao *daos.Dao, m basemodels.Model) error {
		retryAfterUpdateHookCalls++
		return nil
	}

	if err := retryDao.Save(model); err != nil {
		t.Fatalf("Expected nil after retry, got error: %v", err)
	}

	// the before hook is expected to be called only once because
	// it is ignored after the first "database is locked" error
	if retryBeforeUpdateHookCalls != 1 {
		t.Fatalf("Expected before hook calls to be 1, got %d", retryBeforeUpdateHookCalls)
	}

	if retryAfterUpdateHookCalls != 1 {
		t.Fatalf("Expected after hook calls to be 1, got %d", retryAfterUpdateHookCalls)
	}

	// with non-locking error
	retryBeforeUpdateHookCalls = 0
	retryAfterUpdateHookCalls = 0
	retryDao.BeforeUpdateFunc = func(eventDao *daos.Dao, m basemodels.Model, action func() error) error {
		retryBeforeUpdateHookCalls++
		return errors.New("non-locking error")
	}

	if err := retryDao.Save(model); err == nil {
		t.Fatal("Expected error, got nil")
	}

	if retryBeforeUpdateHookCalls != 1 {
		t.Fatalf("Expected before hook calls to be 1, got %d", retryBeforeUpdateHookCalls)
	}

	if retryAfterUpdateHookCalls != 0 {
		t.Fatalf("Expected after hook calls to be 0, got %d", retryAfterUpdateHookCalls)
	}
}

func TestDaoRetryDelete(t *testing.T) {
	testApp, _ := tests.NewTestApp()
	defer testApp.Cleanup()

	// init mock retry dao
	retryBeforeDeleteHookCalls := 0
	retryAfterDeleteHookCalls := 0
	retryDao := daos.New(testApp.Dao().DB())
	retryDao.BeforeDeleteFunc = func(eventDao *daos.Dao, m basemodels.Model, action func() error) error {
		retryBeforeDeleteHookCalls++
		return errors.New("database is locked")
	}
	retryDao.AfterDeleteFunc = func(eventDao *daos.Dao, m basemodels.Model) error {
		retryAfterDeleteHookCalls++
		return nil
	}

	model := createTestTeam(t, testApp, "retry_delete_team")
	if err := retryDao.Delete(model); err != nil {
		t.Fatalf("Expected nil after retry, got error: %v", err)
	}

	// the before hook is expected to be called only once because
	// it is ignored after the first "database is locked" error
	if retryBeforeDeleteHookCalls != 1 {
		t.Fatalf("Expected before hook calls to be 1, got %d", retryBeforeDeleteHookCalls)
	}

	if retryAfterDeleteHookCalls != 1 {
		t.Fatalf("Expected after hook calls to be 1, got %d", retryAfterDeleteHookCalls)
	}

	// with non-locking error
	retryBeforeDeleteHookCalls = 0
	retryAfterDeleteHookCalls = 0
	retryDao.BeforeDeleteFunc = func(eventDao *daos.Dao, m basemodels.Model, action func() error) error {
		retryBeforeDeleteHookCalls++
		return errors.New("non-locking error")
	}

	dummy := &teammodels.Team{}
	dummy.RefreshId()
	dummy.MarkAsNotNew()
	if err := retryDao.Delete(dummy); err == nil {
		t.Fatal("Expected error, got nil")
	}

	if retryBeforeDeleteHookCalls != 1 {
		t.Fatalf("Expected before hook calls to be 1, got %d", retryBeforeDeleteHookCalls)
	}

	if retryAfterDeleteHookCalls != 0 {
		t.Fatalf("Expected after hook calls to be 0, got %d", retryAfterDeleteHookCalls)
	}
}

func TestDaoBeforeHooksError(t *testing.T) {
	testApp, _ := tests.NewTestApp()
	defer testApp.Cleanup()

	baseDao := testApp.Dao()

	baseDao.BeforeCreateFunc = func(eventDao *daos.Dao, m basemodels.Model, action func() error) error {
		return errors.New("before_create")
	}
	baseDao.BeforeUpdateFunc = func(eventDao *daos.Dao, m basemodels.Model, action func() error) error {
		return errors.New("before_update")
	}
	baseDao.BeforeDeleteFunc = func(eventDao *daos.Dao, m basemodels.Model, action func() error) error {
		return errors.New("before_delete")
	}

	existingModel := findExistingAdminTeam(t, testApp)

	// test create error
	// ---
	newModel := &teammodels.Team{Name: "before_hooks_new_team"}
	if err := baseDao.Save(newModel); err.Error() != "before_create" {
		t.Fatalf("Expected before_create error, got %v", err)
	}

	// test update error
	// ---
	if err := baseDao.Save(existingModel); err.Error() != "before_update" {
		t.Fatalf("Expected before_update error, got %v", err)
	}

	// test delete error
	// ---
	if err := baseDao.Delete(existingModel); err.Error() != "before_delete" {
		t.Fatalf("Expected before_delete error, got %v", err)
	}
}

func TestDaoTransactionHooksCallsOnFailure(t *testing.T) {
	testApp, _ := tests.NewTestApp()
	defer testApp.Cleanup()

	beforeCreateFuncCalls := 0
	beforeUpdateFuncCalls := 0
	beforeDeleteFuncCalls := 0
	afterCreateFuncCalls := 0
	afterUpdateFuncCalls := 0
	afterDeleteFuncCalls := 0

	baseDao := testApp.Dao()

	baseDao.BeforeCreateFunc = func(eventDao *daos.Dao, m basemodels.Model, action func() error) error {
		beforeCreateFuncCalls++
		return action()
	}
	baseDao.BeforeUpdateFunc = func(eventDao *daos.Dao, m basemodels.Model, action func() error) error {
		beforeUpdateFuncCalls++
		return action()
	}
	baseDao.BeforeDeleteFunc = func(eventDao *daos.Dao, m basemodels.Model, action func() error) error {
		beforeDeleteFuncCalls++
		return action()
	}

	baseDao.AfterCreateFunc = func(eventDao *daos.Dao, m basemodels.Model) error {
		afterCreateFuncCalls++
		return nil
	}
	baseDao.AfterUpdateFunc = func(eventDao *daos.Dao, m basemodels.Model) error {
		afterUpdateFuncCalls++
		return nil
	}
	baseDao.AfterDeleteFunc = func(eventDao *daos.Dao, m basemodels.Model) error {
		afterDeleteFuncCalls++
		return nil
	}

	existingModel := findExistingAdminTeam(t, testApp)

	baseDao.RunInTransaction(func(txDao1 *daos.Dao) error {
		return txDao1.RunInTransaction(func(txDao2 *daos.Dao) error {
			// test create
			// ---
			newModel := &teammodels.Team{Name: "tx_failure_new_team"}
			if err := txDao2.Save(newModel); err != nil {
				t.Fatal(err)
			}

			// test update (twice)
			// ---
			if err := txDao2.Save(existingModel); err != nil {
				t.Fatal(err)
			}
			if err := txDao2.Save(existingModel); err != nil {
				t.Fatal(err)
			}

			// test delete
			// ---
			if err := txDao2.Delete(existingModel); err != nil {
				t.Fatal(err)
			}

			return errors.New("test_tx_error")
		})
	})

	if beforeCreateFuncCalls != 1 {
		t.Fatalf("Expected beforeCreateFuncCalls to be called 1 times, got %d", beforeCreateFuncCalls)
	}
	if beforeUpdateFuncCalls != 2 {
		t.Fatalf("Expected beforeUpdateFuncCalls to be called 2 times, got %d", beforeUpdateFuncCalls)
	}
	if beforeDeleteFuncCalls != 1 {
		t.Fatalf("Expected beforeDeleteFuncCalls to be called 1 times, got %d", beforeDeleteFuncCalls)
	}
	if afterCreateFuncCalls != 0 {
		t.Fatalf("Expected afterCreateFuncCalls to be called 0 times, got %d", afterCreateFuncCalls)
	}
	if afterUpdateFuncCalls != 0 {
		t.Fatalf("Expected afterUpdateFuncCalls to be called 0 times, got %d", afterUpdateFuncCalls)
	}
	if afterDeleteFuncCalls != 0 {
		t.Fatalf("Expected afterDeleteFuncCalls to be called 0 times, got %d", afterDeleteFuncCalls)
	}
}

func TestDaoTransactionHooksCallsOnSuccess(t *testing.T) {
	testApp, _ := tests.NewTestApp()
	defer testApp.Cleanup()

	beforeCreateFuncCalls := 0
	beforeUpdateFuncCalls := 0
	beforeDeleteFuncCalls := 0
	afterCreateFuncCalls := 0
	afterUpdateFuncCalls := 0
	afterDeleteFuncCalls := 0

	baseDao := testApp.Dao()

	baseDao.BeforeCreateFunc = func(eventDao *daos.Dao, m basemodels.Model, action func() error) error {
		beforeCreateFuncCalls++
		return action()
	}
	baseDao.BeforeUpdateFunc = func(eventDao *daos.Dao, m basemodels.Model, action func() error) error {
		beforeUpdateFuncCalls++
		return action()
	}
	baseDao.BeforeDeleteFunc = func(eventDao *daos.Dao, m basemodels.Model, action func() error) error {
		beforeDeleteFuncCalls++
		return action()
	}

	baseDao.AfterCreateFunc = func(eventDao *daos.Dao, m basemodels.Model) error {
		afterCreateFuncCalls++
		return nil
	}
	baseDao.AfterUpdateFunc = func(eventDao *daos.Dao, m basemodels.Model) error {
		afterUpdateFuncCalls++
		return nil
	}
	baseDao.AfterDeleteFunc = func(eventDao *daos.Dao, m basemodels.Model) error {
		afterDeleteFuncCalls++
		return nil
	}

	existingModel := findExistingAdminTeam(t, testApp)

	baseDao.RunInTransaction(func(txDao1 *daos.Dao) error {
		return txDao1.RunInTransaction(func(txDao2 *daos.Dao) error {
			// test create
			// ---
			newModel := &teammodels.Team{Name: "tx_success_new_team"}
			if err := txDao2.Save(newModel); err != nil {
				t.Fatal(err)
			}

			// test update (twice)
			// ---
			if err := txDao2.Save(existingModel); err != nil {
				t.Fatal(err)
			}
			if err := txDao2.Save(existingModel); err != nil {
				t.Fatal(err)
			}

			// test delete
			// ---
			if err := txDao2.Delete(existingModel); err != nil {
				t.Fatal(err)
			}

			return nil
		})
	})

	if beforeCreateFuncCalls != 1 {
		t.Fatalf("Expected beforeCreateFuncCalls to be called 1 times, got %d", beforeCreateFuncCalls)
	}
	if beforeUpdateFuncCalls != 2 {
		t.Fatalf("Expected beforeUpdateFuncCalls to be called 2 times, got %d", beforeUpdateFuncCalls)
	}
	if beforeDeleteFuncCalls != 1 {
		t.Fatalf("Expected beforeDeleteFuncCalls to be called 1 times, got %d", beforeDeleteFuncCalls)
	}
	if afterCreateFuncCalls != 1 {
		t.Fatalf("Expected afterCreateFuncCalls to be called 1 times, got %d", afterCreateFuncCalls)
	}
	if afterUpdateFuncCalls != 2 {
		t.Fatalf("Expected afterUpdateFuncCalls to be called 2 times, got %d", afterUpdateFuncCalls)
	}
	if afterDeleteFuncCalls != 1 {
		t.Fatalf("Expected afterDeleteFuncCalls to be called 1 times, got %d", afterDeleteFuncCalls)
	}
}
