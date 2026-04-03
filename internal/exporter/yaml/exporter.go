package yaml

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/schema-export/schema-export/internal/exporter"
	"github.com/schema-export/schema-export/internal/model"
	"gopkg.in/yaml.v3"
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
		outputPath = filepath.Join(options.OutputDir, "schema.yaml")
	}

	info, err := os.Stat(outputPath)
	if err == nil && info.IsDir() {
		outputPath = filepath.Join(outputPath, "schema.yaml")
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

	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode YAML: %w", err)
	}

	return nil
}

func (e *Exporter) exportSplitFiles(tables []model.Table, views []model.View, procedures []model.Procedure, functions []model.Function, triggers []model.Trigger, sequences []model.Sequence, options exporter.ExportOptions) error {
	yamlDir := filepath.Join(options.OutputDir, "yaml")
	if err := os.MkdirAll(yamlDir, 0755); err != nil {
		return fmt.Errorf("failed to create yaml directory: %w", err)
	}

	for _, table := range tables {
		outputPath := filepath.Join(yamlDir, table.Name+".yaml")
		file, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", outputPath, err)
		}

		encoder := yaml.NewEncoder(file)
		encoder.SetIndent(2)
		if err := encoder.Encode(table); err != nil {
			file.Close()
			return fmt.Errorf("failed to encode YAML: %w", err)
		}
		file.Close()
	}

	if options.IncludeViews {
		for _, view := range views {
			outputPath := filepath.Join(yamlDir, view.Name+"_view.yaml")
			file, err := os.Create(outputPath)
			if err != nil {
				return fmt.Errorf("failed to create file %s: %w", outputPath, err)
			}

			encoder := yaml.NewEncoder(file)
			encoder.SetIndent(2)
			if err := encoder.Encode(view); err != nil {
				file.Close()
				return fmt.Errorf("failed to encode YAML: %w", err)
			}
			file.Close()
		}
	}

	if options.IncludeProcedures {
		for _, proc := range procedures {
			outputPath := filepath.Join(yamlDir, proc.Name+"_procedure.yaml")
			file, err := os.Create(outputPath)
			if err != nil {
				return fmt.Errorf("failed to create file %s: %w", outputPath, err)
			}

			encoder := yaml.NewEncoder(file)
			encoder.SetIndent(2)
			if err := encoder.Encode(proc); err != nil {
				file.Close()
				return fmt.Errorf("failed to encode YAML: %w", err)
			}
			file.Close()
		}
	}

	if options.IncludeFunctions {
		for _, fn := range functions {
			outputPath := filepath.Join(yamlDir, fn.Name+"_function.yaml")
			file, err := os.Create(outputPath)
			if err != nil {
				return fmt.Errorf("failed to create file %s: %w", outputPath, err)
			}

			encoder := yaml.NewEncoder(file)
			encoder.SetIndent(2)
			if err := encoder.Encode(fn); err != nil {
				file.Close()
				return fmt.Errorf("failed to encode YAML: %w", err)
			}
			file.Close()
		}
	}

	if options.IncludeTriggers {
		for _, tr := range triggers {
			outputPath := filepath.Join(yamlDir, tr.Name+"_trigger.yaml")
			file, err := os.Create(outputPath)
			if err != nil {
				return fmt.Errorf("failed to create file %s: %w", outputPath, err)
			}

			encoder := yaml.NewEncoder(file)
			encoder.SetIndent(2)
			if err := encoder.Encode(tr); err != nil {
				file.Close()
				return fmt.Errorf("failed to encode YAML: %w", err)
			}
			file.Close()
		}
	}

	if options.IncludeSequences {
		for _, seq := range sequences {
			outputPath := filepath.Join(yamlDir, seq.Name+"_sequence.yaml")
			file, err := os.Create(outputPath)
			if err != nil {
				return fmt.Errorf("failed to create file %s: %w", outputPath, err)
			}

			encoder := yaml.NewEncoder(file)
			encoder.SetIndent(2)
			if err := encoder.Encode(seq); err != nil {
				file.Close()
				return fmt.Errorf("failed to encode YAML: %w", err)
			}
			file.Close()
		}
	}

	return nil
}

func (e *Exporter) GetName() string {
	return "yaml"
}

func (e *Exporter) GetExtension() string {
	return ".yaml"
}

type Factory struct{}

func (f *Factory) Create() (exporter.Exporter, error) {
	return NewExporter(), nil
}

func (f *Factory) GetType() string {
	return "yaml"
}

func init() {
	exporter.Register("yaml", &Factory{})
}
