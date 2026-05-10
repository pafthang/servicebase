package daos_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/pafthang/servicebase/tests"
	"github.com/pafthang/servicebase/tools/list"
	"github.com/pocketbase/dbx"
)

func TestHasTable(t *testing.T) {
	t.Parallel()

	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	scenarios := []struct {
		tableName string
		expected  bool
	}{
		{"", false},
		{"test", false},
		{"teams", true},
		{"demo3", true},
		{"DEMO3", true}, // table names are case insensitives by default
		{"view1", true}, // view
	}

	for i, scenario := range scenarios {
		result := app.Dao().HasTable(scenario.tableName)
		if result != scenario.expected {
			t.Errorf("[%d] Expected %v, got %v", i, scenario.expected, result)
		}
	}
}

func TestTableColumns(t *testing.T) {
	t.Parallel()

	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	scenarios := []struct {
		tableName string
		expected  []string
	}{
		{"", nil},
		{"_params", []string{"id", "key", "value", "created", "updated"}},
	}

	for i, s := range scenarios {
		columns, _ := app.Dao().TableColumns(s.tableName)

		if len(columns) != len(s.expected) {
			t.Errorf("[%d] Expected columns %v, got %v", i, s.expected, columns)
			continue
		}

		for _, c := range columns {
			if !list.ExistInSlice(c, s.expected) {
				t.Errorf("[%d] Didn't expect column %s", i, c)
			}
		}
	}
}

func TestTableInfo(t *testing.T) {
	t.Parallel()

	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	rows, err := app.Dao().TableInfo("")
	if err == nil || rows != nil {
		t.Fatalf("Expected missing table error for empty table name, got rows=%v err=%v", rows, err)
	}

	rows, err = app.Dao().TableInfo("missing")
	if err == nil || rows != nil {
		t.Fatalf("Expected missing table error for invalid table name, got rows=%v err=%v", rows, err)
	}

	rows, err = app.Dao().TableInfo("_params")
	if err != nil {
		t.Fatal(err)
	}

	expected := []struct {
		pk           int
		index        int
		name         string
		fieldType    string
		notNull      bool
		defaultValue string
	}{
		{1, 0, "id", "TEXT", false, ""},
		{0, 1, "key", "TEXT", true, ""},
		{0, 2, "value", "JSON", false, "NULL"},
		{0, 3, "created", "TEXT", true, `""`},
		{0, 4, "updated", "TEXT", true, `""`},
	}

	if len(rows) != len(expected) {
		t.Fatalf("Expected %d table info rows, got %d", len(expected), len(rows))
	}

	for i, exp := range expected {
		row := rows[i]
		if row.PK != exp.pk ||
			row.Index != exp.index ||
			row.Name != exp.name ||
			row.Type != exp.fieldType ||
			row.NotNull != exp.notNull ||
			row.DefaultValue.String() != exp.defaultValue {
			t.Fatalf(
				"[%d] Expected {PK:%d Index:%d Name:%q Type:%q NotNull:%v DefaultValue:%q}, got {PK:%d Index:%d Name:%q Type:%q NotNull:%v DefaultValue:%q}",
				i,
				exp.pk,
				exp.index,
				exp.name,
				exp.fieldType,
				exp.notNull,
				exp.defaultValue,
				row.PK,
				row.Index,
				row.Name,
				row.Type,
				row.NotNull,
				row.DefaultValue.String(),
			)
		}
	}
}

func TestDeleteTable(t *testing.T) {
	t.Parallel()

	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	scenarios := []struct {
		tableName   string
		expectError bool
	}{
		{"", true},
		{"test", false}, // missing tables are ignored
		{"teams", false},
		{"demo3", false},
	}

	for i, s := range scenarios {
		err := app.Dao().DeleteTable(s.tableName)

		hasErr := err != nil
		if hasErr != s.expectError {
			t.Errorf("[%d] Expected hasErr %v, got %v", i, s.expectError, hasErr)
		}
	}
}

func TestVacuum(t *testing.T) {
	t.Parallel()

	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	db, ok := app.Dao().DB().(*dbx.DB)
	if !ok {
		t.Fatal("app dao db is not *dbx.DB")
	}

	calledQueries := []string{}
	db.QueryLogFunc = func(ctx context.Context, t time.Duration, sql string, rows *sql.Rows, err error) {
		calledQueries = append(calledQueries, sql)
	}
	db.ExecLogFunc = func(ctx context.Context, t time.Duration, sql string, result sql.Result, err error) {
		calledQueries = append(calledQueries, sql)
	}

	if err := app.Dao().Vacuum(); err != nil {
		t.Fatal(err)
	}

	if total := len(calledQueries); total != 1 {
		t.Fatalf("Expected 1 query, got %d", total)
	}

	if calledQueries[0] != "VACUUM" {
		t.Fatalf("Expected VACUUM query, got %s", calledQueries[0])
	}
}
func TestTableIndexes(t *testing.T) {
	t.Parallel()

	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	scenarios := []struct {
		table         string
		expectError   bool
		expectIndexes []string
	}{
		{
			"missing",
			false,
			nil,
		},
		{
			"demo2",
			false,
			[]string{"idx_demo2_created", "idx_unique_demo2_title", "idx_demo2_active"},
		},
	}

	for _, s := range scenarios {
		result, err := app.Dao().TableIndexes(s.table)

		hasErr := err != nil
		if hasErr != s.expectError {
			t.Errorf("[%s] Expected hasErr %v, got %v", s.table, s.expectError, hasErr)
		}

		if len(s.expectIndexes) != len(result) {
			t.Errorf("[%s] Expected %d indexes, got %d:\n%v", s.table, len(s.expectIndexes), len(result), result)
			continue
		}

		for _, name := range s.expectIndexes {
			if result[name] == "" {
				t.Errorf("[%s] Missing index %q in \n%v", s.table, name, result)
			}
		}
	}
}
