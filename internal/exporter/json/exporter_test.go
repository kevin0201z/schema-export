package json

import (
	stdjson "encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/schema-export/schema-export/internal/exporter"
	"github.com/schema-export/schema-export/internal/model"
)

func sampleJSONExportData() ([]model.Table, []model.View, []model.Procedure, []model.Function, []model.Trigger, []model.Sequence) {
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

func readJSONFile(t *testing.T, path string) []byte {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	return data
}

func TestJSONExportSingleFile(t *testing.T) {
	tables, views, procedures, functions, triggers, sequences := sampleJSONExportData()
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

	var payload map[string]stdjson.RawMessage
	if err := stdjson.Unmarshal(readJSONFile(t, filepath.Join(outDir, "schema.json")), &payload); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	for _, key := range []string{"tables", "views", "procedures", "functions", "triggers", "sequences"} {
		if _, ok := payload[key]; !ok {
			t.Fatalf("expected key %q in JSON payload", key)
		}
	}

	var tablesOut []model.Table
	if err := stdjson.Unmarshal(payload["tables"], &tablesOut); err != nil {
		t.Fatalf("failed to decode tables: %v", err)
	}
	if got := len(tablesOut); got != 1 {
		t.Fatalf("expected 1 table, got %d", got)
	}
	var viewsOut []model.View
	if err := stdjson.Unmarshal(payload["views"], &viewsOut); err != nil {
		t.Fatalf("failed to decode views: %v", err)
	}
	if got := len(viewsOut); got != 1 {
		t.Fatalf("expected 1 view, got %d", got)
	}
	var proceduresOut []model.Procedure
	if err := stdjson.Unmarshal(payload["procedures"], &proceduresOut); err != nil {
		t.Fatalf("failed to decode procedures: %v", err)
	}
	if got := len(proceduresOut); got != 1 {
		t.Fatalf("expected 1 procedure, got %d", got)
	}
	var functionsOut []model.Function
	if err := stdjson.Unmarshal(payload["functions"], &functionsOut); err != nil {
		t.Fatalf("failed to decode functions: %v", err)
	}
	if got := len(functionsOut); got != 1 {
		t.Fatalf("expected 1 function, got %d", got)
	}
	var triggersOut []model.Trigger
	if err := stdjson.Unmarshal(payload["triggers"], &triggersOut); err != nil {
		t.Fatalf("failed to decode triggers: %v", err)
	}
	if got := len(triggersOut); got != 1 {
		t.Fatalf("expected 1 trigger, got %d", got)
	}
	var sequencesOut []model.Sequence
	if err := stdjson.Unmarshal(payload["sequences"], &sequencesOut); err != nil {
		t.Fatalf("failed to decode sequences: %v", err)
	}
	if got := len(sequencesOut); got != 1 {
		t.Fatalf("expected 1 sequence, got %d", got)
	}
}

func TestJSONExportSplitFiles(t *testing.T) {
	tables, views, procedures, functions, triggers, sequences := sampleJSONExportData()
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
		filepath.Join(outDir, "json", "users.json"):                   {"\"Name\": \"users\"", "\"Comment\": \"user table\""},
		filepath.Join(outDir, "json", "active_users_view.json"):       {"\"Name\": \"active_users\""},
		filepath.Join(outDir, "json", "refresh_stats_procedure.json"): {"\"Name\": \"refresh_stats\""},
		filepath.Join(outDir, "json", "format_name_function.json"):    {"\"Name\": \"format_name\""},
		filepath.Join(outDir, "json", "users_audit_trigger.json"):     {"\"Name\": \"users_audit\""},
		filepath.Join(outDir, "json", "user_seq_sequence.json"):       {"\"Name\": \"user_seq\""},
	}

	for path, wants := range cases {
		content := string(readJSONFile(t, path))
		for _, want := range wants {
			if !strings.Contains(content, want) {
				t.Fatalf("expected %s to contain %q", path, want)
			}
		}
	}
}

func TestJSONExportEmptyCollectionStructure(t *testing.T) {
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

	var payload map[string]stdjson.RawMessage
	if err := stdjson.Unmarshal(readJSONFile(t, filepath.Join(outDir, "schema.json")), &payload); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	for _, key := range []string{"tables", "views", "procedures", "functions", "triggers", "sequences"} {
		if string(payload[key]) != "[]" {
			t.Fatalf("expected empty array for %s, got %s", key, string(payload[key]))
		}
	}
}
