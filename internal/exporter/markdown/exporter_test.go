package markdown

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/schema-export/schema-export/internal/exporter"
	"github.com/schema-export/schema-export/internal/model"
)

func sampleExportData() ([]model.Table, []model.View, []model.Procedure, []model.Function, []model.Trigger, []model.Sequence) {
	tables := []model.Table{
		{
			Name:    "users",
			Comment: "user table",
			Type:    model.TableTypeTable,
			Columns: []model.Column{
				{Name: "id", DataType: "INT", IsPrimaryKey: true, IsNullable: false},
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
		[]model.Procedure{{Name: "refresh_stats", Comment: "refresh stats", Definition: "BEGIN SELECT 1; END;"}},
		[]model.Function{{Name: "format_name", Comment: "format name", ReturnType: "VARCHAR", Definition: "RETURN name;"}},
		[]model.Trigger{{Name: "users_audit", TableName: "users", Event: "INSERT", Timing: "BEFORE", Status: "ENABLED", Definition: "SET NEW.name = TRIM(NEW.name);"}},
		[]model.Sequence{{Name: "user_seq", MinValue: 1, MaxValue: 999, IncrementBy: 1, Cycle: true, CacheSize: 20, LastValue: 10}}
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	return string(data)
}

func TestMarkdownExportSingleFile(t *testing.T) {
	tables, views, procedures, functions, triggers, sequences := sampleExportData()
	outDir := t.TempDir()

	exp := NewExporter()
	err := exp.Export(tables, views, procedures, functions, triggers, sequences, exporter.ExportOptions{
		OutputDir:         outDir,
		SplitFiles:        false,
		IncludeViews:      true,
		IncludeProcedures: true,
		IncludeFunctions:  true,
		IncludeTriggers:   true,
		IncludeSequences:  true,
	})
	if err != nil {
		t.Fatalf("Export() failed: %v", err)
	}

	content := readFile(t, filepath.Join(outDir, "schema.md"))
	for _, want := range []string{
		"# 数据库结构文档",
		"**总表数:** 1",
		"**总视图数:** 1",
		"**总存储过程数:** 1",
		"**总函数数:** 1",
		"**总触发器数:** 1",
		"**总序列数:** 1",
		"## 概览",
		"### 表清单",
		"### 视图清单",
		"### 存储过程清单",
		"### 函数清单",
		"### 触发器清单",
		"### 序列清单",
		"## 表: users",
		"CURRENT_TIMESTAMP",
		"## 视图详情",
		"## 视图: active_users",
		"## 存储过程详情",
		"## 函数详情",
		"## 触发器详情",
		"## 序列详情",
		"- **users**: user table",
		"- [users](#表-users)",
		"- [active_users](#视图-active_users)",
		"**描述:** active users",
		"**返回类型:** VARCHAR",
		"**所属表:** users",
		"| 最小值 | 1 |",
		"| 当前值 | 10 |",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected markdown output to contain %q", want)
		}
	}
}

func TestMarkdownExportSplitFiles(t *testing.T) {
	tables, views, procedures, functions, triggers, sequences := sampleExportData()
	outDir := t.TempDir()

	exp := NewExporter()
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
		filepath.Join(outDir, "markdown", "users.md"):                   {"# 表: users", "user table"},
		filepath.Join(outDir, "markdown", "active_users_view.md"):       {"# 视图: active_users", "SELECT id, name FROM users"},
		filepath.Join(outDir, "markdown", "refresh_stats_procedure.md"): {"# 存储过程: refresh_stats", "BEGIN SELECT 1; END;"},
		filepath.Join(outDir, "markdown", "format_name_function.md"):    {"# 函数: format_name", "RETURN name;"},
		filepath.Join(outDir, "markdown", "users_audit_trigger.md"):     {"# 触发器: users_audit", "**所属表:** users", "SET NEW.name = TRIM(NEW.name);"},
		filepath.Join(outDir, "markdown", "user_seq_sequence.md"):       {"# 序列: user_seq", "| 当前值 | 10 |"},
	}

	for path, wants := range cases {
		content := readFile(t, path)
		for _, want := range wants {
			if !strings.Contains(content, want) {
				t.Fatalf("expected %s to contain %q", path, want)
			}
		}
	}
}

func TestMarkdownExportEmptyCollectionSummary(t *testing.T) {
	outDir := t.TempDir()

	exp := NewExporter()
	err := exp.Export([]model.Table{}, []model.View{}, []model.Procedure{}, []model.Function{}, []model.Trigger{}, []model.Sequence{}, exporter.ExportOptions{
		OutputDir:         outDir,
		SplitFiles:        false,
		IncludeViews:      true,
		IncludeProcedures: true,
		IncludeFunctions:  true,
		IncludeTriggers:   true,
		IncludeSequences:  true,
	})
	if err != nil {
		t.Fatalf("Export() failed: %v", err)
	}

	content := readFile(t, filepath.Join(outDir, "schema.md"))
	for _, want := range []string{
		"**总表数:** 0",
		"**总视图数:** 0",
		"**总存储过程数:** 0",
		"**总函数数:** 0",
		"**总触发器数:** 0",
		"**总序列数:** 0",
		"## 概览",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected empty markdown output to contain %q", want)
		}
	}
}
