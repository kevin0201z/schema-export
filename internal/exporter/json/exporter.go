package json

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/schema-export/schema-export/internal/exporter"
	"github.com/schema-export/schema-export/internal/model"
)

type Exporter struct{}

func NewExporter() *Exporter {
	return &Exporter{}
}

func (e *Exporter) Export(tables []model.Table, views []model.View, procedures []model.Procedure, functions []model.Function, triggers []model.Trigger, sequences []model.Sequence, options exporter.ExportOptions) error {
	if options.SplitFiles {
		return e.exportSplitFiles(tables, views, procedures, functions, triggers, sequences, options)
	}
	return e.exportSingleFile(tables, views, procedures, functions, triggers, sequences, options)
}

func (e *Exporter) exportSingleFile(tables []model.Table, views []model.View, procedures []model.Procedure, functions []model.Function, triggers []model.Trigger, sequences []model.Sequence, options exporter.ExportOptions) error {
	outputPath := filepath.Join(options.OutputDir, options.FileName)
	if outputPath == "" || outputPath == options.OutputDir {
		outputPath = filepath.Join(options.OutputDir, "schema.json")
	}

	info, err := os.Stat(outputPath)
	if err == nil && info.IsDir() {
		outputPath = filepath.Join(outputPath, "schema.json")
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	data := map[string]interface{}{
		"tables": tables,
	}
	if options.IncludeViews {
		data["views"] = views
	}
	if options.IncludeProcedures {
		data["procedures"] = procedures
	}
	if options.IncludeFunctions {
		data["functions"] = functions
	}
	if options.IncludeTriggers {
		data["triggers"] = triggers
	}
	if options.IncludeSequences {
		data["sequences"] = sequences
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

func (e *Exporter) exportSplitFiles(tables []model.Table, views []model.View, procedures []model.Procedure, functions []model.Function, triggers []model.Trigger, sequences []model.Sequence, options exporter.ExportOptions) error {
	jsonDir := filepath.Join(options.OutputDir, "json")
	if err := os.MkdirAll(jsonDir, 0755); err != nil {
		return fmt.Errorf("failed to create json directory: %w", err)
	}

	for _, table := range tables {
		outputPath := filepath.Join(jsonDir, table.Name+".json")
		file, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", outputPath, err)
		}

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(table); err != nil {
			file.Close()
			return fmt.Errorf("failed to encode JSON: %w", err)
		}
		file.Close()
	}

	if options.IncludeViews {
		for _, view := range views {
			outputPath := filepath.Join(jsonDir, view.Name+"_view.json")
			file, err := os.Create(outputPath)
			if err != nil {
				return fmt.Errorf("failed to create file %s: %w", outputPath, err)
			}

			encoder := json.NewEncoder(file)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(view); err != nil {
				file.Close()
				return fmt.Errorf("failed to encode JSON: %w", err)
			}
			file.Close()
		}
	}

	if options.IncludeProcedures {
		for _, proc := range procedures {
			outputPath := filepath.Join(jsonDir, proc.Name+"_procedure.json")
			file, err := os.Create(outputPath)
			if err != nil {
				return fmt.Errorf("failed to create file %s: %w", outputPath, err)
			}

			encoder := json.NewEncoder(file)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(proc); err != nil {
				file.Close()
				return fmt.Errorf("failed to encode JSON: %w", err)
			}
			file.Close()
		}
	}

	if options.IncludeFunctions {
		for _, fn := range functions {
			outputPath := filepath.Join(jsonDir, fn.Name+"_function.json")
			file, err := os.Create(outputPath)
			if err != nil {
				return fmt.Errorf("failed to create file %s: %w", outputPath, err)
			}

			encoder := json.NewEncoder(file)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(fn); err != nil {
				file.Close()
				return fmt.Errorf("failed to encode JSON: %w", err)
			}
			file.Close()
		}
	}

	if options.IncludeTriggers {
		for _, tr := range triggers {
			outputPath := filepath.Join(jsonDir, tr.Name+"_trigger.json")
			file, err := os.Create(outputPath)
			if err != nil {
				return fmt.Errorf("failed to create file %s: %w", outputPath, err)
			}

			encoder := json.NewEncoder(file)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(tr); err != nil {
				file.Close()
				return fmt.Errorf("failed to encode JSON: %w", err)
			}
			file.Close()
		}
	}

	if options.IncludeSequences {
		for _, seq := range sequences {
			outputPath := filepath.Join(jsonDir, seq.Name+"_sequence.json")
			file, err := os.Create(outputPath)
			if err != nil {
				return fmt.Errorf("failed to create file %s: %w", outputPath, err)
			}

			encoder := json.NewEncoder(file)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(seq); err != nil {
				file.Close()
				return fmt.Errorf("failed to encode JSON: %w", err)
			}
			file.Close()
		}
	}

	return nil
}

func (e *Exporter) GetName() string {
	return "json"
}

func (e *Exporter) GetExtension() string {
	return ".json"
}

type Factory struct{}

func (f *Factory) Create() (exporter.Exporter, error) {
	return NewExporter(), nil
}

func (f *Factory) GetType() string {
	return "json"
}

func init() {
	exporter.Register("json", &Factory{})
}
