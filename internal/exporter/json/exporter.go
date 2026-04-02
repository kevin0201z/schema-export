package json

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/schema-export/schema-export/internal/exporter"
	"github.com/schema-export/schema-export/internal/model"
)

// Exporter JSON 导出器
type Exporter struct{}

// NewExporter 创建 JSON 导出器
func NewExporter() *Exporter {
	return &Exporter{}
}

// Export 导出表结构和视图
func (e *Exporter) Export(tables []model.Table, views []model.View, options exporter.ExportOptions) error {
	if options.SplitFiles {
		return e.exportSplitFiles(tables, views, options)
	}
	return e.exportSingleFile(tables, views, options)
}

// exportSingleFile 导出到单个文件
func (e *Exporter) exportSingleFile(tables []model.Table, views []model.View, options exporter.ExportOptions) error {
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

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

// exportSplitFiles 分文件导出
func (e *Exporter) exportSplitFiles(tables []model.Table, views []model.View, options exporter.ExportOptions) error {
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

	return nil
}

// GetName 获取导出器名称
func (e *Exporter) GetName() string {
	return "json"
}

// GetExtension 获取文件扩展名
func (e *Exporter) GetExtension() string {
	return ".json"
}

// Factory JSON Exporter 工厂
type Factory struct{}

// Create 创建 Exporter 实例
func (f *Factory) Create() (exporter.Exporter, error) {
	return NewExporter(), nil
}

// GetType 获取导出器类型
func (f *Factory) GetType() string {
	return "json"
}

func init() {
	exporter.Register("json", &Factory{})
}
