package markdown

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/schema-export/schema-export/internal/exporter"
	"github.com/schema-export/schema-export/internal/model"
)

// Exporter Markdown 导出器
type Exporter struct {
	template *template.Template
}

// NewExporter 创建 Markdown 导出器
func NewExporter() *Exporter {
	tmpl := template.Must(template.New("markdown").Funcs(template.FuncMap{
		"join": strings.Join,
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
		outputPath = filepath.Join(options.OutputDir, "schema.md")
	}

	// 检查输出路径状态
	info, err := os.Stat(outputPath)
	if err == nil && info.IsDir() {
		// outputPath 是目录，在目录下创建文件
		outputPath = filepath.Join(outputPath, "schema.md")
	}
	// 文件已存在时直接覆盖，不再报错

	// 确保父目录存在
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// 检查父目录是否确实是目录
	info, err = os.Stat(dir)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("output path %s is not a valid directory", dir)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()
	
	// 写入文件头
	fmt.Fprintln(file, "# Database Schema Documentation")
	fmt.Fprintln(file, "")
	fmt.Fprintf(file, "Total Tables: %d\n\n", len(tables))
	
	// 写入目录
	fmt.Fprintln(file, "## Table of Contents")
	for _, table := range tables {
		fmt.Fprintf(file, "- [%s](#table-%s)\n", table.Name, strings.ToLower(table.Name))
	}
	fmt.Fprintln(file, "")
	
	// 写入每个表的详细信息
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
	markdownDir := filepath.Join(options.OutputDir, "markdown")
	if err := os.MkdirAll(markdownDir, 0755); err != nil {
		return fmt.Errorf("failed to create markdown directory: %w", err)
	}
	
	for _, table := range tables {
		outputPath := filepath.Join(markdownDir, table.Name+".md")
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
	return "markdown"
}

// GetExtension 获取文件扩展名
func (e *Exporter) GetExtension() string {
	return ".md"
}

// tableTemplate Markdown 表模板
const tableTemplate = `## Table: {{.Name}}

{{if .Comment}}**Description:** {{.Comment}}

{{end -}}
### Columns

| Column | Type | Nullable | Default | Comment |
|--------|------|----------|---------|---------|
{{- range .Columns }}
| {{.Name}} | {{.GetFullDataType}}{{if .IsPrimaryKey}} PK{{end}}{{if .IsAutoIncrement}} AI{{end}} | {{if .IsNullable}}YES{{else}}NO{{end}} | {{.DefaultValue}} | {{.Comment}} |
{{- end }}

### Indexes

{{if .Indexes -}}
| Index | Type | Columns |
|-------|------|---------|
{{- range .Indexes }}
| {{.Name}} | {{.Type}} | {{.GetColumnsString}} |
{{- end }}
{{else -}}
*No indexes defined*
{{- end }}

### Foreign Keys

{{if .ForeignKeys -}}
| Foreign Key | Column | Reference | On Delete |
|-------------|--------|-----------|-----------|
{{- range .ForeignKeys }}
| {{.Name}} | {{.Column}} | {{.GetReferenceString}} | {{.GetOnDeleteRule}} |
{{- end }}
{{else -}}
*No foreign keys defined*
{{- end }}

---
`

// Factory Markdown Exporter 工厂
type Factory struct{}

// Create 创建 Exporter 实例
func (f *Factory) Create() (exporter.Exporter, error) {
	return NewExporter(), nil
}

// GetType 获取导出器类型
func (f *Factory) GetType() string {
	return "markdown"
}

func init() {
	exporter.Register("markdown", &Factory{})
}
