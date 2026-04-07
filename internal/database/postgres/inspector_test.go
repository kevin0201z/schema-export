package postgres

import (
	"context"
	"errors"
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	"github.com/schema-export/schema-export/internal/inspector"
	"github.com/schema-export/schema-export/internal/model"
)

func newMockInspector(t *testing.T) (*Inspector, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	ins := NewInspector(inspector.ConnectionConfig{
		Type:     "postgres",
		Host:     "localhost",
		Port:     5432,
		Database: "app",
		Username: "user",
		Password: "pass",
	})
	ins.SetDB(db)

	return ins, mock, func() {
		_ = db.Close()
	}
}

func TestBuildDSN(t *testing.T) {
	tests := []struct {
		name string
		cfg  inspector.ConnectionConfig
		want string
	}{
		{
			name: "passthrough postgres uri",
			cfg:  inspector.ConnectionConfig{DSN: "postgres://u:p@h:5432/db"},
			want: "postgres://u:p@h:5432/db",
		},
		{
			name: "passthrough postgresql uri",
			cfg:  inspector.ConnectionConfig{DSN: "postgresql://u:p@h:5432/db"},
			want: "postgresql://u:p@h:5432/db",
		},
		{
			name: "prefix bare dsn",
			cfg:  inspector.ConnectionConfig{DSN: "u:p@h:5432/db"},
			want: "postgres://u:p@h:5432/db",
		},
		{
			name: "build from parts with default ssl",
			cfg: inspector.ConnectionConfig{
				Username: "u",
				Password: "p",
				Host:     "h",
				Port:     5432,
				Database: "db",
			},
			want: "postgres://u:p@h:5432/db?sslmode=disable",
		},
		{
			name: "build from parts with ssl mode",
			cfg: inspector.ConnectionConfig{
				Username: "u",
				Password: "p",
				Host:     "h",
				Port:     5432,
				Database: "db",
				SSLMode:  "require",
			},
			want: "postgres://u:p@h:5432/db?sslmode=require",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ins := NewInspector(tt.cfg)
			if got := ins.BuildDSN(); got != tt.want {
				t.Fatalf("BuildDSN() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetTables(t *testing.T) {
	ins, mock, cleanup := newMockInspector(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT table_name, 
			   COALESCE(obj_description((table_schema || '.' || table_name)::regclass, 'pg_class'), '')
		FROM information_schema.tables 
		WHERE table_schema = current_schema() AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`)).WillReturnRows(sqlmock.NewRows([]string{"table_name", "comment"}).
		AddRow("orders", "orders table").
		AddRow("users", "users table"))

	tables, err := ins.GetTables(context.Background())
	if err != nil {
		t.Fatalf("GetTables() failed: %v", err)
	}
	if len(tables) != 2 {
		t.Fatalf("expected 2 tables, got %d", len(tables))
	}
	if tables[0].Name != "orders" || tables[0].Comment != "orders table" || tables[0].Type != model.TableTypeTable {
		t.Fatalf("unexpected first table: %#v", tables[0])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetTablesReturnsRowsErr(t *testing.T) {
	ins, mock, cleanup := newMockInspector(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"table_name", "comment"}).
		AddRow("users", "users table").
		AddRow("orders", "orders table").
		RowError(1, errors.New("rows boom"))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT table_name")).WillReturnRows(rows)

	tables, err := ins.GetTables(context.Background())
	if err == nil || err.Error() != "rows boom" {
		t.Fatalf("expected rows error, got tables=%v err=%v", tables, err)
	}
}

func TestGetTableAggregatesRelatedMetadata(t *testing.T) {
	ins, mock, cleanup := newMockInspector(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT table_name,
			   COALESCE(obj_description((table_schema || '.' || table_name)::regclass, 'pg_class'), '')
		FROM information_schema.tables 
		WHERE table_schema = current_schema() AND table_name = $1
	`)).WithArgs("users").WillReturnRows(sqlmock.NewRows([]string{"table_name", "comment"}).
		AddRow("users", "users table"))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT column_name,
			   data_type,
			   COALESCE(character_maximum_length, 0),
			   COALESCE(numeric_precision, 0),
			   COALESCE(numeric_scale, 0),
			   CASE WHEN is_nullable = 'YES' THEN false ELSE true END,
			   COALESCE(column_default, ''),
			   COALESCE(col_description((table_schema || '.' || table_name)::regclass, ordinal_position), '')
		FROM information_schema.columns 
		WHERE table_schema = current_schema() AND table_name = $1
		ORDER BY ordinal_position
	`)).WithArgs("users").WillReturnRows(sqlmock.NewRows([]string{
		"column_name", "data_type", "char_len", "precision", "scale", "is_nullable", "column_default", "comment",
	}).
		AddRow("id", "integer", 0, 0, 0, false, "nextval('users_id_seq'::regclass)", "").
		AddRow("name", "character varying", 32, 0, 0, true, nil, "user name"))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT i.relname AS index_name,
			   am.amname AS index_type,
			   ix.indisunique AS is_unique,
			   ix.indisprimary AS is_primary,
			   array_agg(a.attname ORDER BY array_position(ix.indkey, a.attnum)) AS columns
		FROM pg_index ix
		JOIN pg_class t ON t.oid = ix.indrelid
		JOIN pg_class i ON i.oid = ix.indexrelid
		JOIN pg_am am ON am.oid = i.relam
		JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = ANY(ix.indkey)
		JOIN pg_namespace n ON n.oid = t.relnamespace
		WHERE t.relname = $1 AND n.nspname = current_schema()
		GROUP BY i.relname, am.amname, ix.indisunique, ix.indisprimary
		ORDER BY i.relname
	`)).WithArgs("users").WillReturnRows(sqlmock.NewRows([]string{"index_name", "index_type", "is_unique", "is_primary", "columns"}).
		AddRow("idx_users_name", "btree", true, false, "{name}").
		AddRow("users_pkey", "btree", true, true, "{id}"))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT tc.constraint_name,
			   kcu.column_name,
			   ccu.table_name AS foreign_table,
			   ccu.column_name AS foreign_column,
			   rc.delete_rule
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name
		JOIN information_schema.constraint_column_usage ccu ON ccu.constraint_name = tc.constraint_name
		JOIN information_schema.referential_constraints rc ON rc.constraint_name = tc.constraint_name
		WHERE tc.constraint_type = 'FOREIGN KEY' 
		  AND tc.table_schema = current_schema() 
		  AND tc.table_name = $1
		ORDER BY tc.constraint_name
	`)).WithArgs("users").WillReturnRows(sqlmock.NewRows([]string{"constraint_name", "column_name", "foreign_table", "foreign_column", "delete_rule"}).
		AddRow("fk_users_role", "role_id", "roles", "id", "CASCADE"))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT con.conname AS constraint_name,
			   pg_get_constraintdef(con.oid) AS definition,
			   array_agg(att.attname ORDER BY array_position(con.conkey, att.attnum)) FILTER (WHERE att.attname IS NOT NULL) AS columns
		FROM pg_constraint con
		JOIN pg_class rel ON rel.oid = con.conrelid
		JOIN pg_namespace nsp ON nsp.oid = rel.relnamespace
		LEFT JOIN pg_attribute att ON att.attrelid = con.conrelid AND att.attnum = ANY(con.conkey)
		WHERE con.contype = 'c' 
		  AND rel.relname = $1 
		  AND nsp.nspname = current_schema()
		GROUP BY con.conname, pg_get_constraintdef(con.oid)
		ORDER BY con.conname
	`)).WithArgs("users").WillReturnRows(sqlmock.NewRows([]string{"constraint_name", "definition", "columns"}).
		AddRow("ck_users_name", "CHECK ((name <> ''::text))", "{name}"))

	table, err := ins.GetTable(context.Background(), "users")
	if err != nil {
		t.Fatalf("GetTable() failed: %v", err)
	}
	if table.Name != "users" || table.Comment != "users table" || table.Type != model.TableTypeTable {
		t.Fatalf("unexpected table metadata: %#v", table)
	}
	if len(table.Columns) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(table.Columns))
	}
	if table.Columns[0].Name != "id" || !table.Columns[0].IsPrimaryKey || !table.Columns[0].IsAutoIncrement {
		t.Fatalf("unexpected first column: %#v", table.Columns[0])
	}
	if table.Columns[1].Name != "name" || table.Columns[1].Comment != "user name" {
		t.Fatalf("unexpected second column: %#v", table.Columns[1])
	}
	if len(table.Indexes) != 2 {
		t.Fatalf("expected 2 indexes, got %d", len(table.Indexes))
	}
	if table.Indexes[0].Name != "idx_users_name" || table.Indexes[1].Name != "users_pkey" {
		t.Fatalf("expected indexes sorted by name, got %#v", table.Indexes)
	}
	if len(table.ForeignKeys) != 1 || table.ForeignKeys[0].Name != "fk_users_role" {
		t.Fatalf("unexpected foreign keys: %#v", table.ForeignKeys)
	}
	if len(table.CheckConstraints) != 1 || table.CheckConstraints[0].Name != "ck_users_name" {
		t.Fatalf("unexpected check constraints: %#v", table.CheckConstraints)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetColumnsQueryFailure(t *testing.T) {
	ins, mock, cleanup := newMockInspector(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("FROM information_schema.columns")).
		WithArgs("users").
		WillReturnError(errors.New("columns boom"))

	columns, err := ins.GetColumns(context.Background(), "users")
	if err == nil || err.Error() != "failed to query columns for table users: columns boom" {
		t.Fatalf("expected query failure, got columns=%v err=%v", columns, err)
	}
}

func TestGetViewsAndRowsErr(t *testing.T) {
	ins, mock, cleanup := newMockInspector(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"table_name", "comment", "view_definition"}).
		AddRow("active_users", "active users", "SELECT id FROM users").
		AddRow("archived_users", "archived users", "SELECT id FROM archived_users").
		RowError(1, errors.New("views rows boom"))
	mock.ExpectQuery(regexp.QuoteMeta("FROM information_schema.views")).WillReturnRows(rows)

	views, err := ins.GetViews(context.Background())
	if err == nil || err.Error() != "views rows boom" {
		t.Fatalf("expected rows error, got views=%v err=%v", views, err)
	}
	if len(views) != 1 || views[0].Name != "active_users" {
		t.Fatalf("unexpected views: %#v", views)
	}
}

func TestGetProceduresFunctionsTriggersAndSequences(t *testing.T) {
	ins, mock, cleanup := newMockInspector(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("p.prokind = 'p'")).WillReturnRows(sqlmock.NewRows([]string{"proname", "comment", "definition"}).
		AddRow("refresh_stats", "refresh stats", "BEGIN SELECT 1; END"))
	mock.ExpectQuery(regexp.QuoteMeta("p.prokind = 'f'")).WillReturnRows(sqlmock.NewRows([]string{"proname", "comment", "definition", "return_type"}).
		AddRow("format_name", "format name", "RETURN name", "character varying"))
	mock.ExpectQuery(regexp.QuoteMeta("FROM pg_trigger t")).WithArgs("users").WillReturnRows(sqlmock.NewRows([]string{"tgname", "event_manipulation", "action_timing", "status", "definition"}).
		AddRow("users_audit", "INSERT", "BEFORE", "ENABLED", "CREATE TRIGGER ..."))
	mock.ExpectQuery(regexp.QuoteMeta("FROM information_schema.sequences")).WillReturnRows(sqlmock.NewRows([]string{"sequence_name", "minimum_value", "maximum_value", "increment", "cycle_flag", "cache_size", "last_value"}).
		AddRow("user_seq", 1, 999, 1, "YES", 20, 10))

	procedures, err := ins.GetProcedures(context.Background())
	if err != nil {
		t.Fatalf("GetProcedures() failed: %v", err)
	}
	if len(procedures) != 1 || procedures[0].Name != "refresh_stats" {
		t.Fatalf("unexpected procedures: %#v", procedures)
	}

	functions, err := ins.GetFunctions(context.Background())
	if err != nil {
		t.Fatalf("GetFunctions() failed: %v", err)
	}
	if len(functions) != 1 || functions[0].Name != "format_name" || functions[0].ReturnType != "character varying" {
		t.Fatalf("unexpected functions: %#v", functions)
	}

	triggers, err := ins.GetTriggers(context.Background(), "users")
	if err != nil {
		t.Fatalf("GetTriggers() failed: %v", err)
	}
	if len(triggers) != 1 || triggers[0].Name != "users_audit" || triggers[0].TableName != "users" {
		t.Fatalf("unexpected triggers: %#v", triggers)
	}

	sequences, err := ins.GetSequences(context.Background())
	if err != nil {
		t.Fatalf("GetSequences() failed: %v", err)
	}
	if len(sequences) != 1 || sequences[0].Name != "user_seq" || !sequences[0].Cycle {
		t.Fatalf("unexpected sequences: %#v", sequences)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetForeignKeysAndCheckConstraintsQueryFailures(t *testing.T) {
	ins, mock, cleanup := newMockInspector(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("FROM information_schema.table_constraints")).
		WithArgs("users").
		WillReturnError(errors.New("fk boom"))
	mock.ExpectQuery(regexp.QuoteMeta("FROM pg_constraint con")).
		WithArgs("users").
		WillReturnError(errors.New("check boom"))

	if _, err := ins.GetForeignKeys(context.Background(), "users"); err == nil || err.Error() != "failed to query foreign keys for table users: fk boom" {
		t.Fatalf("expected foreign key query failure, got %v", err)
	}
	if _, err := ins.GetCheckConstraints(context.Background(), "users"); err == nil || err.Error() != "failed to query check constraints for table users: check boom" {
		t.Fatalf("expected check constraint query failure, got %v", err)
	}
}

func TestGetIndexesQueryFailure(t *testing.T) {
	ins, mock, cleanup := newMockInspector(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("FROM pg_index ix")).
		WithArgs("users").
		WillReturnError(errors.New("index boom"))

	indexes, err := ins.GetIndexes(context.Background(), "users")
	if err == nil || err.Error() != "failed to query indexes for table users: index boom" {
		t.Fatalf("expected index query failure, got indexes=%v err=%v", indexes, err)
	}
}
