package sql

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/schema-export/schema-export/internal/exporter"
	"github.com/schema-export/schema-export/internal/model"
)

func sampleSQLExportData() ([]model.Table, []model.View, []model.Procedure, []model.Function, []model.Trigger, []model.Sequence) {
	tables := []model.Table{
		{
			Name:    "users",
			Comment: "user table",
			Type:    model.TableTypeTable,
			Columns: []model.Column{
				{Name: "id", DataType: "INT", IsPrimaryKey: true, IsAutoIncrement: true, IsNullable: false},
				{Name: "name", DataType: "VARCHAR", Length: 32, IsNullable: false, Comment: "user name"},
				{Name: "created_at", DataType: "TIMESTAMP", IsNullable: false, DefaultValue: "CURRENT_TIMESTAMP"},
				{Name: "role_id", DataType: "INT", IsNullable: false},
			},
			Indexes: []model.Index{
				{Name: "idx_users_name", Type: model.IndexTypeUnique, Columns: []string{"name"}, IsUnique: true},
			},
			ForeignKeys: []model.ForeignKey{
				{Name: "fk_users_role", Column: "role_id", RefTable: "roles", RefColumn: "id", OnDelete: "CASCADE"},
			},
			CheckConstraints: []model.CheckConstraint{
				{Name: "ck_users_name", Definition: "name <> ''"},
			},
		},
	}

	return tables,
		[]model.View{{Name: "active_users", Comment: "active users", Definition: "SELECT id, name FROM users"}},
		[]model.Procedure{{Name: "refresh_stats", Comment: "refresh stats", Definition: "BEGIN SELECT 1; END"}},
		[]model.Function{{Name: "format_name", Comment: "format name", ReturnType: "VARCHAR", Definition: "RETURN name"}},
		[]model.Trigger{{Name: "users_audit", TableName: "users", Event: "INSERT", Timing: "BEFORE", Status: "ENABLED", Definition: "SET NEW.name = TRIM(NEW.name)"}},
		[]model.Sequence{{Name: "user_seq", MinValue: 1, MaxValue: 999, IncrementBy: 1, Cycle: true, CacheSize: 20, LastValue: 10}}
}

func mustReadFile(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	return string(data)
}

func TestSQLExportSingleFileDialectSelection(t *testing.T) {
	tables, views, procedures, functions, triggers, sequences := sampleSQLExportData()
	outDir := t.TempDir()

	tests := []struct {
		name     string
		dbType   string
		wants    []string
		unwanted []string
	}{
		{
			name:   "mysql",
			dbType: "mysql",
			wants: []string{
				"-- Dialect: mysql",
				"CREATE TABLE `users`",
				"COMMENT 'user name'",
				"ALTER TABLE `users` COMMENT = 'user table';",
				"CREATE UNIQUE INDEX `idx_users_name`",
				"ALTER TABLE `users` ADD CONSTRAINT `fk_users_role`",
				"DEFAULT CURRENT_TIMESTAMP",
			},
		},
		{
			name:   "postgres",
			dbType: "postgres",
			wants: []string{
				"-- Dialect: postgres",
				"CREATE TABLE \"users\"",
				"COMMENT ON TABLE \"users\" IS 'user table';",
				"COMMENT ON COLUMN \"users\".\"name\" IS 'user name';",
				"CREATE UNIQUE INDEX \"idx_users_name\"",
				"ALTER TABLE \"users\" ADD CONSTRAINT \"fk_users_role\"",
			},
		},
		{
			name:   "sqlserver",
			dbType: "sqlserver",
			wants: []string{
				"-- Dialect: sqlserver",
				"CREATE TABLE [users]",
				"IDENTITY(1,1)",
				"EXEC sp_addextendedproperty 'MS_Description'",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp := NewExporter()
			err := exp.Export(tables, views, procedures, functions, triggers, sequences, exporter.ExportOptions{
				OutputDir:         outDir,
				SplitFiles:        false,
				DbType:            tt.dbType,
				IncludeViews:      true,
				IncludeProcedures: true,
				IncludeFunctions:  true,
				IncludeTriggers:   true,
				IncludeSequences:  true,
			})
			if err != nil {
				t.Fatalf("Export() failed: %v", err)
			}

			content := mustReadFile(t, filepath.Join(outDir, "schema.sql"))
			for _, want := range tt.wants {
				if !strings.Contains(content, want) {
					t.Fatalf("expected SQL output for %s to contain %q", tt.dbType, want)
				}
			}
		})
	}
}

func TestSQLExportSplitFiles(t *testing.T) {
	tables, views, procedures, functions, triggers, sequences := sampleSQLExportData()
	outDir := t.TempDir()

	exp := NewExporterWithDialect("postgres")
	err := exp.Export(tables, views, procedures, functions, triggers, sequences, exporter.ExportOptions{
		OutputDir:         outDir,
		SplitFiles:        true,
		IncludeViews:      true,
		IncludeProcedures: true,
		IncludeFunctions:  true,
		IncludeTriggers:   true,
		IncludeSequences:  true,
	})
	if err != nil {
		t.Fatalf("Export() failed: %v", err)
	}

	cases := map[string][]string{
		filepath.Join(outDir, "sql", "users.sql"):                   {"CREATE TABLE \"users\"", "COMMENT ON TABLE \"users\" IS 'user table';"},
		filepath.Join(outDir, "sql", "active_users_view.sql"):       {"-- View: active_users", "CREATE OR REPLACE VIEW \"active_users\" AS"},
		filepath.Join(outDir, "sql", "refresh_stats_procedure.sql"): {"-- Procedure: refresh_stats", "BEGIN SELECT 1; END"},
		filepath.Join(outDir, "sql", "format_name_function.sql"):    {"-- Function: format_name", "RETURN name"},
		filepath.Join(outDir, "sql", "users_audit_trigger.sql"):     {"-- Trigger: users_audit", "-- Table: users"},
		filepath.Join(outDir, "sql", "user_seq_sequence.sql"):       {"-- Sequence: user_seq", "CREATE SEQUENCE \"user_seq\""},
	}

	for path, wants := range cases {
		content := mustReadFile(t, path)
		for _, want := range wants {
			if !strings.Contains(content, want) {
				t.Fatalf("expected %s to contain %q", path, want)
			}
		}
	}
}

func TestSQLExportEmptyCollectionSummary(t *testing.T) {
	outDir := t.TempDir()

	exp := NewExporterWithDialect("mysql")
	err := exp.Export([]model.Table{}, []model.View{}, []model.Procedure{}, []model.Function{}, []model.Trigger{}, []model.Sequence{}, exporter.ExportOptions{
		OutputDir:         outDir,
		SplitFiles:        false,
		DbType:            "mysql",
		IncludeViews:      true,
		IncludeProcedures: true,
		IncludeFunctions:  true,
		IncludeTriggers:   true,
		IncludeSequences:  true,
	})
	if err != nil {
		t.Fatalf("Export() failed: %v", err)
	}

	content := mustReadFile(t, filepath.Join(outDir, "schema.sql"))
	for _, want := range []string{
		"-- Total Tables: 0",
		"-- Total Views: 0",
		"-- Total Procedures: 0",
		"-- Total Functions: 0",
		"-- Total Triggers: 0",
		"-- Total Sequences: 0",
		"-- Dialect: mysql",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected empty SQL output to contain %q", want)
		}
	}
}
