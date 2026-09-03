package exporter_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/schema-export/schema-export/internal/exporter"
	jsonexport "github.com/schema-export/schema-export/internal/exporter/json"
	"github.com/schema-export/schema-export/internal/exporter/markdown"
	sqlexport "github.com/schema-export/schema-export/internal/exporter/sql"
	"github.com/schema-export/schema-export/internal/exporter/yaml"
	"github.com/schema-export/schema-export/internal/model"
)

func splitExporters() []exporter.Exporter {
	return []exporter.Exporter{jsonexport.NewExporter(), markdown.NewExporter(), sqlexport.NewExporter(), yaml.NewExporter()}
}

func TestSplitExportRejectsCollisionsBeforeWriting(t *testing.T) {
	for _, exp := range splitExporters() {
		t.Run(exp.GetName(), func(t *testing.T) {
			for _, pair := range [][2]string{{"Users", "users"}, {"caf\u00e9", "cafe\u0301"}, {"same", "same"}} {
				t.Run(pair[0]+"_"+pair[1], func(t *testing.T) {
					out := t.TempDir()
					dir := filepath.Join(out, exp.GetName())
					if err := os.Mkdir(dir, 0755); err != nil {
						t.Fatal(err)
					}
					// An earlier, valid object must not be truncated before a later conflict.
					preserved := filepath.Join(dir, "existing"+exp.GetExtension())
					if err := os.WriteFile(preserved, []byte("previous export"), 0644); err != nil {
						t.Fatal(err)
					}
					err := exp.Export([]model.Table{{Name: "existing"}, {Name: pair[0]}, {Name: pair[1]}}, nil, nil, nil, nil, nil, exporter.ExportOptions{OutputDir: out, SplitFiles: true})
					if err == nil || !strings.Contains(err.Error(), "collision") || !strings.Contains(err.Error(), "--split") {
						t.Fatalf("expected actionable collision error, got %v", err)
					}
					data, err := os.ReadFile(preserved)
					if err != nil || string(data) != "previous export" {
						t.Fatalf("existing file changed: %q, %v", data, err)
					}
					files, err := os.ReadDir(dir)
					if err != nil || len(files) != 1 {
						t.Fatalf("unexpected files after rejected export: %v, %v", files, err)
					}
				})
			}
		})
	}
}

func TestSplitExportChecksEnabledObjectSuffixes(t *testing.T) {
	for _, exp := range splitExporters() {
		for _, suffix := range []string{"view", "procedure", "function", "trigger", "sequence"} {
			t.Run(exp.GetName()+"/"+suffix, func(t *testing.T) {
				out := filepath.Join(t.TempDir(), "not-created")
				options := exporter.ExportOptions{OutputDir: out, SplitFiles: true}
				export := func() error {
					return exp.Export([]model.Table{{Name: "item_" + suffix}}, []model.View{{Name: "item"}}, []model.Procedure{{Name: "item"}}, []model.Function{{Name: "item"}}, []model.Trigger{{Name: "item"}}, []model.Sequence{{Name: "item"}}, options)
				}
				options.IncludeViews = suffix == "view"
				options.IncludeProcedures = suffix == "procedure"
				options.IncludeFunctions = suffix == "function"
				options.IncludeTriggers = suffix == "trigger"
				options.IncludeSequences = suffix == "sequence"
				if err := export(); err == nil || !strings.Contains(err.Error(), "collision") {
					t.Fatalf("expected cross-object collision, got %v", err)
				}
				if _, err := os.Stat(out); !os.IsNotExist(err) {
					t.Fatalf("output directory created before validation: %v", err)
				}
				options = exporter.ExportOptions{OutputDir: out, SplitFiles: true}
				if err := export(); err != nil {
					t.Fatalf("excluded objects must not cause conflicts: %v", err)
				}
			})
		}
	}
}

func TestSplitExportNamesAndRepeatedExports(t *testing.T) {
	for _, exp := range splitExporters() {
		t.Run(exp.GetName(), func(t *testing.T) {
			out := filepath.Join(t.TempDir(), "中文目录 with spaces")
			options := exporter.ExportOptions{OutputDir: out, SplitFiles: true}
			for _, name := range []string{"", "a/b", "a\\b", "nul\x00name"} {
				if err := exp.Export([]model.Table{{Name: name}}, nil, nil, nil, nil, nil, options); err == nil {
					t.Fatalf("unsafe filename %q accepted", name)
				}
			}
			for _, comment := range []string{"first export", "updated export"} {
				if err := exp.Export([]model.Table{{Name: "订单 history", Comment: comment}}, nil, nil, nil, nil, nil, options); err != nil {
					t.Fatal(err)
				}
				data, err := os.ReadFile(filepath.Join(out, exp.GetName(), "订单 history"+exp.GetExtension()))
				if err != nil || !strings.Contains(string(data), comment) {
					t.Fatalf("export content not updated: %s, %v", data, err)
				}
			}
			options.SplitFiles = false
			if err := exp.Export([]model.Table{{Name: "caf\u00e9"}, {Name: "cafe\u0301"}}, nil, nil, nil, nil, nil, options); err != nil {
				t.Fatalf("single-file export should allow distinct identifiers: %v", err)
			}
		})
	}
}
