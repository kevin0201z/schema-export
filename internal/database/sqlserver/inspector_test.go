package sqlserver

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/schema-export/schema-export/internal/inspector"
)

func TestSQLServerBuildDSN(t *testing.T) {
	t.Run("returns raw DSN when prefixed", func(t *testing.T) {
		ins := NewInspector(inspector.ConnectionConfig{DSN: "sqlserver://user:pass@localhost:1433?database=db"})
		if got := ins.BuildDSN(); got != "sqlserver://user:pass@localhost:1433?database=db" {
			t.Fatalf("unexpected DSN: %s", got)
		}
	})

	t.Run("adds prefix for raw DSN", func(t *testing.T) {
		ins := NewInspector(inspector.ConnectionConfig{DSN: "user:pass@localhost:1433?database=db"})
		if got := ins.BuildDSN(); got != "sqlserver://user:pass@localhost:1433?database=db" {
			t.Fatalf("unexpected DSN: %s", got)
		}
	})

	t.Run("builds from components", func(t *testing.T) {
		ins := NewInspector(inspector.ConnectionConfig{
			Username: "sa",
			Password: "secret",
			Host:     "127.0.0.1",
			Port:     1433,
			Database: "schema_export",
		})
		if got := ins.BuildDSN(); got != "sqlserver://sa:secret@127.0.0.1:1433?database=schema_export" {
			t.Fatalf("unexpected DSN: %s", got)
		}
	})
}

func TestSQLServerConnect(t *testing.T) {
	originalOpen := openDB
	defer func() { openDB = originalOpen }()

	t.Run("open failure", func(t *testing.T) {
		openDB = func(string, string) (*sql.DB, error) {
			return nil, errors.New("open failed")
		}

		ins := NewInspector(inspector.ConnectionConfig{})
		if err := ins.Connect(context.Background()); err == nil || err.Error() != "failed to open sqlserver connection: open failed" {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("ping failure", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		if err != nil {
			t.Fatalf("failed to open sqlmock: %v", err)
		}
		defer db.Close()

		openDB = func(string, string) (*sql.DB, error) {
			return db, nil
		}
		mock.ExpectPing().WillReturnError(errors.New("ping failed"))

		ins := NewInspector(inspector.ConnectionConfig{})
		if err := ins.Connect(context.Background()); err == nil || err.Error() != "failed to ping sqlserver database: ping failed" {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		if err != nil {
			t.Fatalf("failed to open sqlmock: %v", err)
		}
		defer db.Close()

		openDB = func(string, string) (*sql.DB, error) {
			return db, nil
		}
		mock.ExpectPing()

		ins := NewInspector(inspector.ConnectionConfig{Database: "schema_export"})
		if err := ins.Connect(context.Background()); err != nil {
			t.Fatalf("Connect failed: %v", err)
		}
		if ins.GetDB() == nil {
			t.Fatalf("expected db to be set")
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet expectations: %v", err)
		}
	})
}

func TestSQLServerHelpers(t *testing.T) {
	for _, tc := range []struct {
		name string
		data string
		want bool
	}{
		{name: "unicode lower", data: "nvarchar", want: true},
		{name: "unicode upper", data: "NCHAR", want: true},
		{name: "non-unicode", data: "varchar", want: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := isUnicodeType(tc.data); got != tc.want {
				t.Fatalf("isUnicodeType(%q) = %v, want %v", tc.data, got, tc.want)
			}
		})
	}

	for _, tc := range []struct {
		name  string
		input string
		want  string
	}{
		{name: "nested parens", input: "((0))", want: "0"},
		{name: "string literal", input: "('abc')", want: "'abc'"},
		{name: "trim spaces", input: "  ( (1) )  ", want: "1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := cleanDefaultValue(tc.input); got != tc.want {
				t.Fatalf("cleanDefaultValue(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestSQLServerGetIndexesReturnsSortedIndexes(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	ins := NewInspector(inspector.ConnectionConfig{})
	ins.SetDB(db)

	rows := sqlmock.NewRows([]string{"index_name", "index_type", "is_unique", "is_primary_key", "column_name"}).
		AddRow("IDX_B", "NONCLUSTERED", false, false, "col_b").
		AddRow("IDX_A", "NONCLUSTERED", true, false, "col_a1").
		AddRow("IDX_A", "NONCLUSTERED", true, false, "col_a2")

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT 
			i.name AS index_name,
			i.type_desc AS index_type,
			i.is_unique,
			i.is_primary_key,
			c.name AS column_name
		FROM sys.indexes i
		INNER JOIN sys.index_columns ic ON i.object_id = ic.object_id AND i.index_id = ic.index_id
		INNER JOIN sys.columns c ON ic.object_id = c.object_id AND ic.column_id = c.column_id
		INNER JOIN sys.tables t ON i.object_id = t.object_id
		WHERE t.name = @p1 AND i.type > 0
		ORDER BY i.name, ic.key_ordinal
	`)).WithArgs("TBL").WillReturnRows(rows)

	indexes, err := ins.GetIndexes(context.Background(), "TBL")
	if err != nil {
		t.Fatalf("GetIndexes failed: %v", err)
	}

	if len(indexes) != 2 {
		t.Fatalf("expected 2 indexes, got %d", len(indexes))
	}

	if indexes[0].Name != "IDX_A" || indexes[1].Name != "IDX_B" {
		t.Fatalf("expected sorted indexes [IDX_A IDX_B], got [%s %s]", indexes[0].Name, indexes[1].Name)
	}

	if got := indexes[0].GetColumnsString(); got != "col_a1, col_a2" {
		t.Fatalf("expected IDX_A columns to preserve query order, got %q", got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSQLServerGetColumnsHandlesUnicodeAndDefaults(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	ins := NewInspector(inspector.ConnectionConfig{})
	ins.SetDB(db)

	rows := sqlmock.NewRows([]string{
		"column_name", "data_type", "max_length", "precision", "scale",
		"is_nullable", "default_value", "column_comment", "is_primary_key", "is_auto_increment",
	}).
		AddRow("id", "nvarchar", 40, 18, 0, false, "((0))", "主键ID", 1, true).
		AddRow("name", "varchar", 100, nil, nil, true, nil, nil, 0, false)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT 
			c.name AS column_name,
			ty.name AS data_type,
			c.max_length,
			c.precision,
			c.scale,
			c.is_nullable,
			CAST(dc.definition AS NVARCHAR(MAX)) AS default_value,
			CAST(ep.value AS NVARCHAR(MAX)) AS column_comment,
			CASE WHEN pk.column_id IS NOT NULL THEN 1 ELSE 0 END AS is_primary_key,
			c.is_identity AS is_auto_increment
		FROM sys.columns c
		INNER JOIN sys.types ty ON c.user_type_id = ty.user_type_id
		INNER JOIN sys.tables t ON c.object_id = t.object_id
		LEFT JOIN sys.default_constraints dc ON c.default_object_id = dc.object_id
		LEFT JOIN sys.extended_properties ep ON c.object_id = ep.major_id AND c.column_id = ep.minor_id AND ep.name = 'MS_Description'
		LEFT JOIN (
			SELECT ic.object_id, ic.column_id
			FROM sys.indexes i
			INNER JOIN sys.index_columns ic ON i.object_id = ic.object_id AND i.index_id = ic.index_id
			WHERE i.is_primary_key = 1
		) pk ON c.object_id = pk.object_id AND c.column_id = pk.column_id
		WHERE t.name = @p1
		ORDER BY c.column_id
	`)).WithArgs("users").WillReturnRows(rows)

	columns, err := ins.GetColumns(context.Background(), "users")
	if err != nil {
		t.Fatalf("GetColumns failed: %v", err)
	}

	if len(columns) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(columns))
	}
	if got := columns[0].Length; got != 20 {
		t.Fatalf("expected unicode length 20, got %d", got)
	}
	if !columns[0].IsPrimaryKey || !columns[0].IsAutoIncrement {
		t.Fatalf("expected id to be primary key and auto increment")
	}
	if got := columns[0].DefaultValue; got != "0" {
		t.Fatalf("expected cleaned default value 0, got %q", got)
	}
	if got := columns[1].Length; got != 100 {
		t.Fatalf("expected varchar length 100, got %d", got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSQLServerGetTableComment(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	ins := NewInspector(inspector.ConnectionConfig{})
	ins.SetDB(db)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT CAST(ep.value AS NVARCHAR(MAX)) AS table_comment
		FROM sys.tables t
		LEFT JOIN sys.extended_properties ep 
			ON t.object_id = ep.major_id 
			AND ep.minor_id = 0 
			AND ep.name = 'MS_Description'
		WHERE t.name = @p1
	`)).WithArgs("users").WillReturnRows(sqlmock.NewRows([]string{"table_comment"}).AddRow("用户表"))

	comment, err := ins.getTableComment(context.Background(), "users")
	if err != nil {
		t.Fatalf("getTableComment failed: %v", err)
	}
	if comment != "用户表" {
		t.Fatalf("unexpected comment %q", comment)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSQLServerGetTables(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	ins := NewInspector(inspector.ConnectionConfig{})
	ins.SetDB(db)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT 
			t.name AS table_name,
			CAST(ep.value AS NVARCHAR(MAX)) AS table_comment
		FROM sys.tables t
		LEFT JOIN sys.extended_properties ep 
			ON t.object_id = ep.major_id 
			AND ep.minor_id = 0 
			AND ep.name = 'MS_Description'
		WHERE t.is_ms_shipped = 0
		ORDER BY t.name
	`)).WillReturnRows(sqlmock.NewRows([]string{"table_name", "table_comment"}).
		AddRow("users", "用户表").
		AddRow("orders", nil))

	tables, err := ins.GetTables(context.Background())
	if err != nil {
		t.Fatalf("GetTables failed: %v", err)
	}
	if len(tables) != 2 || tables[0].Name != "users" || tables[1].Name != "orders" {
		t.Fatalf("unexpected tables: %#v", tables)
	}
	if tables[0].Type != "TABLE" || tables[0].Comment != "用户表" {
		t.Fatalf("unexpected first table: %#v", tables[0])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSQLServerGetTableAggregatesMetadata(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	ins := NewInspector(inspector.ConnectionConfig{})
	ins.SetDB(db)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT CAST(ep.value AS NVARCHAR(MAX)) AS table_comment
		FROM sys.tables t
		LEFT JOIN sys.extended_properties ep 
			ON t.object_id = ep.major_id 
			AND ep.minor_id = 0 
			AND ep.name = 'MS_Description'
		WHERE t.name = @p1
	`)).WithArgs("users").WillReturnRows(sqlmock.NewRows([]string{"table_comment"}).AddRow("用户表"))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT 
			c.name AS column_name,
			ty.name AS data_type,
			c.max_length,
			c.precision,
			c.scale,
			c.is_nullable,
			CAST(dc.definition AS NVARCHAR(MAX)) AS default_value,
			CAST(ep.value AS NVARCHAR(MAX)) AS column_comment,
			CASE WHEN pk.column_id IS NOT NULL THEN 1 ELSE 0 END AS is_primary_key,
			c.is_identity AS is_auto_increment
		FROM sys.columns c
		INNER JOIN sys.types ty ON c.user_type_id = ty.user_type_id
		INNER JOIN sys.tables t ON c.object_id = t.object_id
		LEFT JOIN sys.default_constraints dc ON c.default_object_id = dc.object_id
		LEFT JOIN sys.extended_properties ep ON c.object_id = ep.major_id AND c.column_id = ep.minor_id AND ep.name = 'MS_Description'
		LEFT JOIN (
			SELECT ic.object_id, ic.column_id
			FROM sys.indexes i
			INNER JOIN sys.index_columns ic ON i.object_id = ic.object_id AND i.index_id = ic.index_id
			WHERE i.is_primary_key = 1
		) pk ON c.object_id = pk.object_id AND c.column_id = pk.column_id
		WHERE t.name = @p1
		ORDER BY c.column_id
	`)).WithArgs("users").WillReturnRows(sqlmock.NewRows([]string{
		"column_name", "data_type", "max_length", "precision", "scale", "is_nullable", "default_value", "column_comment", "is_primary_key", "is_auto_increment",
	}).AddRow("id", "nvarchar", 40, 18, 0, false, "((0))", "主键ID", 1, true))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT 
			i.name AS index_name,
			i.type_desc AS index_type,
			i.is_unique,
			i.is_primary_key,
			c.name AS column_name
		FROM sys.indexes i
		INNER JOIN sys.index_columns ic ON i.object_id = ic.object_id AND i.index_id = ic.index_id
		INNER JOIN sys.columns c ON ic.object_id = c.object_id AND ic.column_id = c.column_id
		INNER JOIN sys.tables t ON i.object_id = t.object_id
		WHERE t.name = @p1 AND i.type > 0
		ORDER BY i.name, ic.key_ordinal
	`)).WithArgs("users").WillReturnRows(sqlmock.NewRows([]string{"index_name", "index_type", "is_unique", "is_primary_key", "column_name"}).
		AddRow("PRIMARY", "CLUSTERED", true, true, "id").
		AddRow("idx_users_name", "NONCLUSTERED", true, false, "name"))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT 
			fk.name AS fk_name,
			pc.name AS column_name,
			rt.name AS ref_table,
			rc.name AS ref_column,
			fk.delete_referential_action_desc AS on_delete
		FROM sys.foreign_keys fk
		INNER JOIN sys.foreign_key_columns fkc ON fk.object_id = fkc.constraint_object_id
		INNER JOIN sys.tables pt ON fk.parent_object_id = pt.object_id
		INNER JOIN sys.columns pc ON fkc.parent_object_id = pc.object_id AND fkc.parent_column_id = pc.column_id
		INNER JOIN sys.tables rt ON fk.referenced_object_id = rt.object_id
		INNER JOIN sys.columns rc ON fkc.referenced_object_id = rc.object_id AND fkc.referenced_column_id = rc.column_id
		WHERE pt.name = @p1
	`)).WithArgs("users").WillReturnRows(sqlmock.NewRows([]string{"fk_name", "column_name", "ref_table", "ref_column", "on_delete"}).
		AddRow("fk_users_role", "role_id", "roles", "id", "NO_ACTION"))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT 
			cc.name AS constraint_name,
			cc.definition AS definition,
			c.name AS column_name
		FROM sys.check_constraints cc
		LEFT JOIN sys.columns c ON cc.parent_column_id = c.column_id AND cc.parent_object_id = c.object_id
		INNER JOIN sys.tables t ON cc.parent_object_id = t.object_id
		WHERE t.name = @p1
		ORDER BY cc.name
	`)).WithArgs("users").WillReturnRows(sqlmock.NewRows([]string{"constraint_name", "definition", "column_name"}).
		AddRow("ck_users_age", "[age] > 0", "age"))

	table, err := ins.GetTable(context.Background(), "users")
	if err != nil {
		t.Fatalf("GetTable failed: %v", err)
	}
	if table.Comment != "用户表" || len(table.Columns) != 1 || len(table.Indexes) != 2 || len(table.ForeignKeys) != 1 || len(table.CheckConstraints) != 1 {
		t.Fatalf("unexpected table aggregate: %#v", table)
	}
	if !table.Columns[0].IsPrimaryKey || !table.Columns[0].IsAutoIncrement {
		t.Fatalf("unexpected column flags: %#v", table.Columns[0])
	}
	if table.Indexes[0].Name != "PRIMARY" || !table.Indexes[0].IsPrimary {
		t.Fatalf("unexpected indexes: %#v", table.Indexes)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSQLServerGetTableFailureBranches(t *testing.T) {
	tests := []struct {
		name string
		set  func(sqlmock.Sqlmock, *Inspector)
	}{
		{
			name: "columns failure",
			set: func(mock sqlmock.Sqlmock, ins *Inspector) {
				mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT CAST(ep.value AS NVARCHAR(MAX)) AS table_comment
		FROM sys.tables t
		LEFT JOIN sys.extended_properties ep 
			ON t.object_id = ep.major_id 
			AND ep.minor_id = 0 
			AND ep.name = 'MS_Description'
		WHERE t.name = @p1
	`)).WithArgs("users").WillReturnRows(sqlmock.NewRows([]string{"table_comment"}).AddRow(nil))
				mock.ExpectQuery(regexp.QuoteMeta("SELECT \n\t\t\tc.name AS column_name")).WithArgs("users").WillReturnError(errors.New("columns boom"))
			},
		},
		{
			name: "indexes failure",
			set: func(mock sqlmock.Sqlmock, ins *Inspector) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT CAST(ep.value AS NVARCHAR(MAX)) AS table_comment
		FROM sys.tables t
		LEFT JOIN sys.extended_properties ep 
			ON t.object_id = ep.major_id 
			AND ep.minor_id = 0 
			AND ep.name = 'MS_Description'
		WHERE t.name = @p1`)).WithArgs("users").WillReturnRows(sqlmock.NewRows([]string{"table_comment"}).AddRow(nil))
				mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT 
			c.name AS column_name,
			ty.name AS data_type,
			c.max_length,
			c.precision,
			c.scale,
			c.is_nullable,
			CAST(dc.definition AS NVARCHAR(MAX)) AS default_value,
			CAST(ep.value AS NVARCHAR(MAX)) AS column_comment,
			CASE WHEN pk.column_id IS NOT NULL THEN 1 ELSE 0 END AS is_primary_key,
			c.is_identity AS is_auto_increment
		FROM sys.columns c
		INNER JOIN sys.types ty ON c.user_type_id = ty.user_type_id
		INNER JOIN sys.tables t ON c.object_id = t.object_id
		LEFT JOIN sys.default_constraints dc ON c.default_object_id = dc.object_id
		LEFT JOIN sys.extended_properties ep ON c.object_id = ep.major_id AND c.column_id = ep.minor_id AND ep.name = 'MS_Description'
		LEFT JOIN (
			SELECT ic.object_id, ic.column_id
			FROM sys.indexes i
			INNER JOIN sys.index_columns ic ON i.object_id = ic.object_id AND i.index_id = ic.index_id
			WHERE i.is_primary_key = 1
		) pk ON c.object_id = pk.object_id AND c.column_id = pk.column_id
		WHERE t.name = @p1
		ORDER BY c.column_id
	`)).WithArgs("users").WillReturnRows(sqlmock.NewRows([]string{
					"column_name", "data_type", "max_length", "precision", "scale", "is_nullable", "default_value", "column_comment", "is_primary_key", "is_auto_increment",
				}).AddRow("id", "int", 4, nil, nil, false, nil, nil, 1, false))
				mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT 
			i.name AS index_name,
			i.type_desc AS index_type,
			i.is_unique,
			i.is_primary_key,
			c.name AS column_name
		FROM sys.indexes i
		INNER JOIN sys.index_columns ic ON i.object_id = ic.object_id AND i.index_id = ic.index_id
		INNER JOIN sys.columns c ON ic.object_id = c.object_id AND ic.column_id = c.column_id
		INNER JOIN sys.tables t ON i.object_id = t.object_id
		WHERE t.name = @p1 AND i.type > 0
		ORDER BY i.name, ic.key_ordinal
	`)).WithArgs("users").WillReturnError(errors.New("indexes boom"))
			},
		},
		{
			name: "foreign keys failure",
			set: func(mock sqlmock.Sqlmock, ins *Inspector) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT CAST(ep.value AS NVARCHAR(MAX)) AS table_comment
		FROM sys.tables t
		LEFT JOIN sys.extended_properties ep 
			ON t.object_id = ep.major_id 
			AND ep.minor_id = 0 
			AND ep.name = 'MS_Description'
		WHERE t.name = @p1`)).WithArgs("users").WillReturnRows(sqlmock.NewRows([]string{"table_comment"}).AddRow(nil))
				mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT 
			c.name AS column_name,
			ty.name AS data_type,
			c.max_length,
			c.precision,
			c.scale,
			c.is_nullable,
			CAST(dc.definition AS NVARCHAR(MAX)) AS default_value,
			CAST(ep.value AS NVARCHAR(MAX)) AS column_comment,
			CASE WHEN pk.column_id IS NOT NULL THEN 1 ELSE 0 END AS is_primary_key,
			c.is_identity AS is_auto_increment
		FROM sys.columns c
		INNER JOIN sys.types ty ON c.user_type_id = ty.user_type_id
		INNER JOIN sys.tables t ON c.object_id = t.object_id
		LEFT JOIN sys.default_constraints dc ON c.default_object_id = dc.object_id
		LEFT JOIN sys.extended_properties ep ON c.object_id = ep.major_id AND c.column_id = ep.minor_id AND ep.name = 'MS_Description'
		LEFT JOIN (
			SELECT ic.object_id, ic.column_id
			FROM sys.indexes i
			INNER JOIN sys.index_columns ic ON i.object_id = ic.object_id AND i.index_id = ic.index_id
			WHERE i.is_primary_key = 1
		) pk ON c.object_id = pk.object_id AND c.column_id = pk.column_id
		WHERE t.name = @p1
		ORDER BY c.column_id
	`)).WithArgs("users").WillReturnRows(sqlmock.NewRows([]string{
					"column_name", "data_type", "max_length", "precision", "scale", "is_nullable", "default_value", "column_comment", "is_primary_key", "is_auto_increment",
				}).AddRow("id", "int", 4, nil, nil, false, nil, nil, 1, false))
				mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT 
			i.name AS index_name,
			i.type_desc AS index_type,
			i.is_unique,
			i.is_primary_key,
			c.name AS column_name
		FROM sys.indexes i
		INNER JOIN sys.index_columns ic ON i.object_id = ic.object_id AND i.index_id = ic.index_id
		INNER JOIN sys.columns c ON ic.object_id = c.object_id AND ic.column_id = c.column_id
		INNER JOIN sys.tables t ON i.object_id = t.object_id
		WHERE t.name = @p1 AND i.type > 0
		ORDER BY i.name, ic.key_ordinal
	`)).WithArgs("users").WillReturnRows(sqlmock.NewRows([]string{"index_name", "index_type", "is_unique", "is_primary_key", "column_name"}).
					AddRow("PRIMARY", "CLUSTERED", true, true, "id"))
				mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT 
			fk.name AS fk_name,
			pc.name AS column_name,
			rt.name AS ref_table,
			rc.name AS ref_column,
			fk.delete_referential_action_desc AS on_delete
		FROM sys.foreign_keys fk
		INNER JOIN sys.foreign_key_columns fkc ON fk.object_id = fkc.constraint_object_id
		INNER JOIN sys.tables pt ON fk.parent_object_id = pt.object_id
		INNER JOIN sys.columns pc ON fkc.parent_object_id = pc.object_id AND fkc.parent_column_id = pc.column_id
		INNER JOIN sys.tables rt ON fk.referenced_object_id = rt.object_id
		INNER JOIN sys.columns rc ON fkc.referenced_object_id = rc.object_id AND fkc.referenced_column_id = rc.column_id
		WHERE pt.name = @p1
	`)).WithArgs("users").WillReturnError(errors.New("fks boom"))
			},
		},
		{
			name: "check constraints failure",
			set: func(mock sqlmock.Sqlmock, ins *Inspector) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT CAST(ep.value AS NVARCHAR(MAX)) AS table_comment
		FROM sys.tables t
		LEFT JOIN sys.extended_properties ep 
			ON t.object_id = ep.major_id 
			AND ep.minor_id = 0 
			AND ep.name = 'MS_Description'
		WHERE t.name = @p1`)).WithArgs("users").WillReturnRows(sqlmock.NewRows([]string{"table_comment"}).AddRow(nil))
				mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT 
			c.name AS column_name,
			ty.name AS data_type,
			c.max_length,
			c.precision,
			c.scale,
			c.is_nullable,
			CAST(dc.definition AS NVARCHAR(MAX)) AS default_value,
			CAST(ep.value AS NVARCHAR(MAX)) AS column_comment,
			CASE WHEN pk.column_id IS NOT NULL THEN 1 ELSE 0 END AS is_primary_key,
			c.is_identity AS is_auto_increment
		FROM sys.columns c
		INNER JOIN sys.types ty ON c.user_type_id = ty.user_type_id
		INNER JOIN sys.tables t ON c.object_id = t.object_id
		LEFT JOIN sys.default_constraints dc ON c.default_object_id = dc.object_id
		LEFT JOIN sys.extended_properties ep ON c.object_id = ep.major_id AND c.column_id = ep.minor_id AND ep.name = 'MS_Description'
		LEFT JOIN (
			SELECT ic.object_id, ic.column_id
			FROM sys.indexes i
			INNER JOIN sys.index_columns ic ON i.object_id = ic.object_id AND i.index_id = ic.index_id
			WHERE i.is_primary_key = 1
		) pk ON c.object_id = pk.object_id AND c.column_id = pk.column_id
		WHERE t.name = @p1
		ORDER BY c.column_id
	`)).WithArgs("users").WillReturnRows(sqlmock.NewRows([]string{
					"column_name", "data_type", "max_length", "precision", "scale", "is_nullable", "default_value", "column_comment", "is_primary_key", "is_auto_increment",
				}).AddRow("id", "int", 4, nil, nil, false, nil, nil, 1, false))
				mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT 
			i.name AS index_name,
			i.type_desc AS index_type,
			i.is_unique,
			i.is_primary_key,
			c.name AS column_name
		FROM sys.indexes i
		INNER JOIN sys.index_columns ic ON i.object_id = ic.object_id AND i.index_id = ic.index_id
		INNER JOIN sys.columns c ON ic.object_id = c.object_id AND ic.column_id = c.column_id
		INNER JOIN sys.tables t ON i.object_id = t.object_id
		WHERE t.name = @p1 AND i.type > 0
		ORDER BY i.name, ic.key_ordinal
	`)).WithArgs("users").WillReturnRows(sqlmock.NewRows([]string{"index_name", "index_type", "is_unique", "is_primary_key", "column_name"}).
					AddRow("PRIMARY", "CLUSTERED", true, true, "id"))
				mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT 
			fk.name AS fk_name,
			pc.name AS column_name,
			rt.name AS ref_table,
			rc.name AS ref_column,
			fk.delete_referential_action_desc AS on_delete
		FROM sys.foreign_keys fk
		INNER JOIN sys.foreign_key_columns fkc ON fk.object_id = fkc.constraint_object_id
		INNER JOIN sys.tables pt ON fk.parent_object_id = pt.object_id
		INNER JOIN sys.columns pc ON fkc.parent_object_id = pc.object_id AND fkc.parent_column_id = pc.column_id
		INNER JOIN sys.tables rt ON fk.referenced_object_id = rt.object_id
		INNER JOIN sys.columns rc ON fkc.referenced_object_id = rc.object_id AND fkc.referenced_column_id = rc.column_id
		WHERE pt.name = @p1
	`)).WithArgs("users").WillReturnRows(sqlmock.NewRows([]string{"fk_name", "column_name", "ref_table", "ref_column", "on_delete"}).
					AddRow("fk_users_role", "role_id", "roles", "id", "NO_ACTION"))
				mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT 
			cc.name AS constraint_name,
			cc.definition AS definition,
			c.name AS column_name
		FROM sys.check_constraints cc
		LEFT JOIN sys.columns c ON cc.parent_column_id = c.column_id AND cc.parent_object_id = c.object_id
		INNER JOIN sys.tables t ON cc.parent_object_id = t.object_id
		WHERE t.name = @p1
		ORDER BY cc.name
	`)).WithArgs("users").WillReturnError(errors.New("checks boom"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to open sqlmock: %v", err)
			}
			defer db.Close()

			ins := NewInspector(inspector.ConnectionConfig{})
			ins.SetDB(db)
			tt.set(mock, ins)

			_, err = ins.GetTable(context.Background(), "users")
			if err == nil {
				t.Fatalf("expected error")
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("unmet expectations: %v", err)
			}
		})
	}
}

func TestSQLServerMetadataQueries(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	ins := NewInspector(inspector.ConnectionConfig{})
	ins.SetDB(db)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT 
			cc.name AS constraint_name,
			cc.definition AS definition,
			c.name AS column_name
		FROM sys.check_constraints cc
		LEFT JOIN sys.columns c ON cc.parent_column_id = c.column_id AND cc.parent_object_id = c.object_id
		INNER JOIN sys.tables t ON cc.parent_object_id = t.object_id
		WHERE t.name = @p1
		ORDER BY cc.name
	`)).WithArgs("users").WillReturnRows(sqlmock.NewRows([]string{"constraint_name", "definition", "column_name"}).
		AddRow("ck_users_age", "[age] > 0", "age").
		AddRow("ck_users_age", "[age] > 0", "status"))

	checks, err := ins.GetCheckConstraints(context.Background(), "users")
	if err != nil {
		t.Fatalf("GetCheckConstraints failed: %v", err)
	}
	if len(checks) != 1 {
		t.Fatalf("expected 1 check constraint, got %d", len(checks))
	}
	if got := checks[0].GetColumnsString(); got != "age, status" {
		t.Fatalf("unexpected constraint columns: %q", got)
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT 
			v.name AS view_name,
			CAST(ep.value AS NVARCHAR(MAX)) AS view_comment,
			CAST(m.definition AS NVARCHAR(MAX)) AS view_definition
		FROM sys.views v
		LEFT JOIN sys.sql_modules m ON v.object_id = m.object_id
		LEFT JOIN sys.extended_properties ep 
			ON v.object_id = ep.major_id 
			AND ep.minor_id = 0 
			AND ep.name = 'MS_Description'
		WHERE v.is_ms_shipped = 0
		ORDER BY v.name
	`)).WillReturnRows(sqlmock.NewRows([]string{"view_name", "view_comment", "view_definition"}).
		AddRow("v_users", "用户视图", "SELECT * FROM users"))

	views, err := ins.GetViews(context.Background())
	if err != nil {
		t.Fatalf("GetViews failed: %v", err)
	}
	if len(views) != 1 || views[0].Name != "v_users" {
		t.Fatalf("unexpected views: %#v", views)
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT 
			p.name AS procedure_name,
			CAST(ep.value AS NVARCHAR(MAX)) AS procedure_comment,
			CAST(m.definition AS NVARCHAR(MAX)) AS procedure_definition
		FROM sys.procedures p
		LEFT JOIN sys.sql_modules m ON p.object_id = m.object_id
		LEFT JOIN sys.extended_properties ep 
			ON p.object_id = ep.major_id 
			AND ep.minor_id = 0 
			AND ep.name = 'MS_Description'
		WHERE p.is_ms_shipped = 0
		ORDER BY p.name
	`)).WillReturnRows(sqlmock.NewRows([]string{"procedure_name", "procedure_comment", "procedure_definition"}).
		AddRow("sp_cleanup", "过程", "EXEC cleanup"))

	procedures, err := ins.GetProcedures(context.Background())
	if err != nil {
		t.Fatalf("GetProcedures failed: %v", err)
	}
	if len(procedures) != 1 || procedures[0].Definition != "EXEC cleanup" {
		t.Fatalf("unexpected procedures: %#v", procedures)
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT 
			o.name AS function_name,
			CAST(ep.value AS NVARCHAR(MAX)) AS function_comment,
			CAST(m.definition AS NVARCHAR(MAX)) AS function_definition
		FROM sys.objects o
		LEFT JOIN sys.sql_modules m ON o.object_id = m.object_id
		LEFT JOIN sys.extended_properties ep 
			ON o.object_id = ep.major_id 
			AND ep.minor_id = 0 
			AND ep.name = 'MS_Description'
		WHERE o.type IN ('FN', 'IF', 'TF') AND o.is_ms_shipped = 0
		ORDER BY o.name
	`)).WillReturnRows(sqlmock.NewRows([]string{"function_name", "function_comment", "function_definition"}).
		AddRow("fn_total", "函数", "RETURN 1"))

	functions, err := ins.GetFunctions(context.Background())
	if err != nil {
		t.Fatalf("GetFunctions failed: %v", err)
	}
	if len(functions) != 1 || functions[0].Name != "fn_total" {
		t.Fatalf("unexpected functions: %#v", functions)
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT 
			tr.name AS trigger_name,
			t.name AS table_name,
			tr.is_disabled,
			CAST(m.definition AS NVARCHAR(MAX)) AS trigger_definition
		FROM sys.triggers tr
		INNER JOIN sys.tables t ON tr.parent_id = t.object_id
		LEFT JOIN sys.sql_modules m ON tr.object_id = m.object_id
		WHERE t.name = @p1
		ORDER BY tr.name
	`)).WithArgs("users").WillReturnRows(sqlmock.NewRows([]string{"trigger_name", "table_name", "is_disabled", "trigger_definition"}).
		AddRow("tr_users", "users", false, "CREATE TRIGGER..."))

	triggers, err := ins.GetTriggers(context.Background(), "users")
	if err != nil {
		t.Fatalf("GetTriggers failed: %v", err)
	}
	if len(triggers) != 1 || triggers[0].Status != "ENABLED" {
		t.Fatalf("unexpected triggers: %#v", triggers)
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT 
			s.name AS sequence_name,
			s.minimum_value,
			s.maximum_value,
			s.increment,
			s.is_cycling,
			s.cache_size,
			s.current_value
		FROM sys.sequences s
		WHERE s.is_ms_shipped = 0
		ORDER BY s.name
	`)).WillReturnRows(sqlmock.NewRows([]string{
		"sequence_name", "minimum_value", "maximum_value", "increment", "is_cycling", "cache_size", "current_value",
	}).AddRow("seq_user", int64(1), int64(10), int64(1), true, int64(20), int64(3)))

	sequences, err := ins.GetSequences(context.Background())
	if err != nil {
		t.Fatalf("GetSequences failed: %v", err)
	}
	if len(sequences) != 1 || !sequences[0].Cycle {
		t.Fatalf("unexpected sequences: %#v", sequences)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
