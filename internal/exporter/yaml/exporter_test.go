package yaml

import (
	"os"
	"path/filepath"
	"testing"

	yamlv3 "gopkg.in/yaml.v3"

	"github.com/schema-export/schema-export/internal/exporter"
	"github.com/schema-export/schema-export/internal/model"
)

func sampleYAMLExportData() ([]model.Table, []model.View, []model.Procedure, []model.Function, []model.Trigger, []model.Sequence) {
	return []model.Table{
			{
				Name:    "users",
				Comment: "user table",
				Columns: []model.Column{{Name: "id", DataType: "INT", IsPrimaryKey: true}},
			},
		},
		[]model.View{{Name: "active_users", Comment: "active users", Definition: "SELECT id FROM users"}},
		[]model.Procedure{{Name: "refresh_stats", Comment: "refresh stats", Definition: "BEGIN SELECT 1; END"}},
		[]model.Function{{Name: "format_name", Comment: "format name", Definition: "RETURN name", ReturnType: "VARCHAR"}},
		[]model.Trigger{{Name: "users_audit", TableName: "users", Event: "INSERT", Timing: "BEFORE", Status: "ENABLED", Definition: "SET NEW.id = NEW.id"}},
		[]model.Sequence{{Name: "user_seq", MinValue: 1, MaxValue: 999, IncrementBy: 1, Cycle: true, CacheSize: 20, LastValue: 10}}
}

func readYAMLFile(t *testing.T, path string) []byte {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	return data
}

func TestYAMLExportSingleFile(t *testing.T) {
	tables, views, procedures, functions, triggers, sequences := sampleYAMLExportData()
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

	var payload map[string]any
	if err := yamlv3.Unmarshal(readYAMLFile(t, filepath.Join(outDir, "schema.yaml")), &payload); err != nil {
		t.Fatalf("failed to unmarshal YAML: %v", err)
	}

	for _, key := range []string{"tables", "views", "procedures", "functions", "triggers", "sequences"} {
		if _, ok := payload[key]; !ok {
			t.Fatalf("expected key %q in YAML payload", key)
		}
	}
	if got := len(payload["tables"].([]any)); got != 1 {
		t.Fatalf("expected 1 table, got %d", got)
	}
	if got := len(payload["views"].([]any)); got != 1 {
		t.Fatalf("expected 1 view, got %d", got)
	}
}

func TestYAMLExportSplitFiles(t *testing.T) {
	tables, views, procedures, functions, triggers, sequences := sampleYAMLExportData()
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

	var tableOut model.Table
	if err := yamlv3.Unmarshal(readYAMLFile(t, filepath.Join(outDir, "yaml", "users.yaml")), &tableOut); err != nil {
		t.Fatalf("failed to unmarshal table: %v", err)
	}
	if tableOut.Name != "users" || tableOut.Comment != "user table" {
		t.Fatalf("unexpected table payload: %#v", tableOut)
	}

	var viewOut model.View
	if err := yamlv3.Unmarshal(readYAMLFile(t, filepath.Join(outDir, "yaml", "active_users_view.yaml")), &viewOut); err != nil {
		t.Fatalf("failed to unmarshal view: %v", err)
	}
	if viewOut.Name != "active_users" {
		t.Fatalf("unexpected view payload: %#v", viewOut)
	}

	var procOut model.Procedure
	if err := yamlv3.Unmarshal(readYAMLFile(t, filepath.Join(outDir, "yaml", "refresh_stats_procedure.yaml")), &procOut); err != nil {
		t.Fatalf("failed to unmarshal procedure: %v", err)
	}
	if procOut.Name != "refresh_stats" {
		t.Fatalf("unexpected procedure payload: %#v", procOut)
	}

	var fnOut model.Function
	if err := yamlv3.Unmarshal(readYAMLFile(t, filepath.Join(outDir, "yaml", "format_name_function.yaml")), &fnOut); err != nil {
		t.Fatalf("failed to unmarshal function: %v", err)
	}
	if fnOut.Name != "format_name" {
		t.Fatalf("unexpected function payload: %#v", fnOut)
	}

	var trigOut model.Trigger
	if err := yamlv3.Unmarshal(readYAMLFile(t, filepath.Join(outDir, "yaml", "users_audit_trigger.yaml")), &trigOut); err != nil {
		t.Fatalf("failed to unmarshal trigger: %v", err)
	}
	if trigOut.Name != "users_audit" || trigOut.TableName != "users" {
		t.Fatalf("unexpected trigger payload: %#v", trigOut)
	}

	var seqOut model.Sequence
	if err := yamlv3.Unmarshal(readYAMLFile(t, filepath.Join(outDir, "yaml", "user_seq_sequence.yaml")), &seqOut); err != nil {
		t.Fatalf("failed to unmarshal sequence: %v", err)
	}
	if seqOut.Name != "user_seq" || seqOut.CacheSize != 20 {
		t.Fatalf("unexpected sequence payload: %#v", seqOut)
	}
}

func TestYAMLExportEmptyCollectionStructure(t *testing.T) {
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

	var payload map[string]any
	if err := yamlv3.Unmarshal(readYAMLFile(t, filepath.Join(outDir, "schema.yaml")), &payload); err != nil {
		t.Fatalf("failed to unmarshal YAML: %v", err)
	}

	for _, key := range []string{"tables", "views", "procedures", "functions", "triggers", "sequences"} {
		if got := len(payload[key].([]any)); got != 0 {
			t.Fatalf("expected empty array for %s, got %d", key, got)
		}
	}
}
