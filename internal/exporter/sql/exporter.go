package sql

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/schema-export/schema-export/internal/exporter"
	"github.com/schema-export/schema-export/internal/model"
)

// Exporter SQL DDL 导出器
type Exporter struct {
	template *template.Template
}

// NewExporter 创建 SQL 导出器
func NewExporter() *Exporter {
	tmpl := template.Must(template.New("sql").Funcs(template.FuncMap{
		"join": strings.Join,
		"add":  func(a, b int) int { return a + b },
	}).Parse(tableTemplate))
	
	return &Exporter{
		template: tmpl,
	}
}

// Export 导出表结构
func (e *Exporter) Export(tables []model.Table, options exporter.ExportOptions) error {
	if options.SplitFiles {
		return e.exportSplitFiles(tables, options)
	}
	return e.exportSingleFile(tables, options)
}

// exportSingleFile 导出到单个文件
func (e *Exporter) exportSingleFile(tables []model.Table, options exporter.ExportOptions) error {
	outputPath := filepath.Join(options.OutputDir, options.FileName)
	if outputPath == "" || outputPath == options.OutputDir {
		outputPath = filepath.Join(options.OutputDir, "schema.sql")
	}

	// 检查输出路径是否为已存在的目录
	info, err := os.Stat(outputPath)
	if err == nil && info.IsDir() {
		outputPath = filepath.Join(outputPath, "schema.sql")
	}
	// 文件已存在时直接覆盖，不再重命名

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()
	
	// 写入文件头
	fmt.Fprintln(file, "-- Database Schema DDL")
	fmt.Fprintf(file, "-- Total Tables: %d\n", len(tables))
	fmt.Fprintln(file, "")
	
	// 写入每个表的 DDL
	for _, table := range tables {
		if err := e.template.Execute(file, table); err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}
		fmt.Fprintln(file, "")
	}
	
	return nil
}

// exportSplitFiles 分文件导出
func (e *Exporter) exportSplitFiles(tables []model.Table, options exporter.ExportOptions) error {
	sqlDir := filepath.Join(options.OutputDir, "sql")
	if err := os.MkdirAll(sqlDir, 0755); err != nil {
		return fmt.Errorf("failed to create sql directory: %w", err)
	}
	
	for _, table := range tables {
		outputPath := filepath.Join(sqlDir, table.Name+".sql")
		file, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", outputPath, err)
		}
		
		if err := e.template.Execute(file, table); err != nil {
			file.Close()
			return fmt.Errorf("failed to execute template: %w", err)
		}
		file.Close()
	}
	
	return nil
}

// GetName 获取导出器名称
func (e *Exporter) GetName() string {
	return "sql"
}

// GetExtension 获取文件扩展名
func (e *Exporter) GetExtension() string {
	return ".sql"
}

// tableTemplate SQL DDL 表模板
const tableTemplate = `-- Table: {{.Name}}
{{if .Comment}}-- {{.Comment}}{{end}}

CREATE TABLE {{.Name}} (
{{- range $i, $col := .Columns }}
    {{$col.Name}} {{$col.GetFullDataType}}{{if not $col.IsNullable}} NOT NULL{{end}}{{if $col.DefaultValue}} DEFAULT {{$col.DefaultValue}}{{end}}{{if $col.IsPrimaryKey}} PRIMARY KEY{{end}}{{if $col.IsAutoIncrement}} AUTO_INCREMENT{{end}}{{if $col.Comment}} -- {{$col.Comment}}{{end}}{{if lt (add $i 1) (len $.Columns)}},{{end}}
{{- end }}
);

{{if .Indexes -}}
-- Indexes for {{.Name}}
{{- range .Indexes }}
{{if not .IsPrimary -}}
CREATE {{if .IsUnique}}UNIQUE {{end}}INDEX {{.Name}} ON {{$.Name}} ({{.GetColumnsString}});
{{end -}}
{{- end }}
{{end -}}

{{if .ForeignKeys -}}
-- Foreign Keys for {{.Name}}
{{- range .ForeignKeys }}
ALTER TABLE {{$.Name}} ADD CONSTRAINT {{.Name}} FOREIGN KEY ({{.Column}}) REFERENCES {{.RefTable}}({{.RefColumn}}){{if .OnDelete}} ON DELETE {{.OnDelete}}{{end}};
{{- end }}
{{end -}}
`

// Factory SQL Exporter 工厂
type Factory struct{}

// Create 创建 Exporter 实例
func (f *Factory) Create() (exporter.Exporter, error) {
	return NewExporter(), nil
}

// GetType 获取导出器类型
func (f *Factory) GetType() string {
	return "sql"
}

func init() {
	exporter.Register("sql", &Factory{})
}
