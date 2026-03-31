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
		"join":       strings.Join,
		"toLower":    strings.ToLower,
		"hasPrefix":  strings.HasPrefix,
		"hasSuffix":  strings.HasSuffix,
		"contains":   strings.Contains,
		"trimPrefix": strings.TrimPrefix,
		"trimSuffix": strings.TrimSuffix,
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
	fmt.Fprintln(file, "# 数据库结构文档")
	fmt.Fprintln(file, "")
	fmt.Fprintf(file, "**总表数:** %d\n\n", len(tables))

	// 写入 Schema 概览
	fmt.Fprintln(file, "## 概览")
	fmt.Fprintln(file, "")
	fmt.Fprintln(file, "### 表清单")
	fmt.Fprintln(file, "")
	for _, table := range tables {
		if table.Comment != "" {
			fmt.Fprintf(file, "- **%s**: %s\n", table.Name, table.Comment)
		} else {
			fmt.Fprintf(file, "- **%s**\n", table.Name)
		}
	}
	fmt.Fprintln(file, "")

	// 写入详细目录
	fmt.Fprintln(file, "### 目录")
	fmt.Fprintln(file, "")
	for _, table := range tables {
		fmt.Fprintf(file, "- [%s](#表-%s)\n", table.Name, strings.ToLower(table.Name))
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

		// 写入单个文件的头部
		fmt.Fprintf(file, "# 表: %s\n\n", table.Name)

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

// tableTemplate Markdown 表模板 - 中文显示
const tableTemplate = `## 表: {{.Name}}

{{if .Comment}}**描述:** {{.Comment}}

{{end -}}
**表类型:** {{.Type}}

### 基本信息

| 属性 | 值 |
|------|-----|
| 表名 | {{.Name}} |
| 类型 | {{.Type}} |
{{if .Comment}}| 描述 | {{.Comment}} |{{end}}
| 字段数 | {{len .Columns}} |
| 索引数 | {{len .Indexes}} |
| 外键数 | {{len .ForeignKeys}} |

### 字段详情

| 字段名 | 数据类型 | 长度/精度 | 可空 | 默认值 | 约束 | 注释 |
|--------|----------|-----------|------|--------|------|------|
{{- range .Columns }}
| {{.Name}} | {{.DataType}} | {{if gt .Length 0}}{{.Length}}{{else if gt .Precision 0}}{{if gt .Scale 0}}{{.Precision}},{{.Scale}}{{else}}{{.Precision}}{{end}}{{else}}-{{end}} | {{if .IsNullable}}是{{else}}否{{end}} | {{if .DefaultValue}}{{.DefaultValue}}{{else}}-{{end}} | {{if .IsPrimaryKey}}主键 {{end}}{{if .IsAutoIncrement}}自增 {{end}}{{if not .IsNullable}}非空{{end}} | {{if .Comment}}{{.Comment}}{{else}}-{{end}} |
{{- end }}

### 约束

#### 主键
{{$hasPK := false}}
{{- range .Columns }}
{{- if .IsPrimaryKey }}
{{$hasPK = true}}
- **{{.Name}}**: {{.DataType}}{{if gt .Length 0}}({{.Length}}){{else if gt .Precision 0}}({{if gt .Scale 0}}{{.Precision}},{{.Scale}}{{else}}{{.Precision}}{{end}}){{end}} - {{if .Comment}}{{.Comment}}{{else}}主键{{end}}
{{- end }}
{{- end }}
{{if not $hasPK}}
*未定义主键*
{{end}}

#### 唯一约束
{{$hasUnique := false}}
{{- range .Indexes }}
{{- if .IsUnique }}
{{$hasUnique = true}}
- **{{.Name}}**: {{.GetColumnsString}}
{{- end }}
{{- end }}
{{if not $hasUnique}}
*未定义唯一约束*
{{end}}

### 索引

{{if .Indexes -}}
| 索引名 | 类型 | 字段 | 是否唯一 | 是否主键 |
|--------|------|------|----------|----------|
{{- range .Indexes }}
| {{.Name}} | {{.Type}} | {{.GetColumnsString}} | {{if .IsUnique}}是{{else}}否{{end}} | {{if .IsPrimary}}是{{else}}否{{end}} |
{{- end }}
{{else -}}
*未定义索引*
{{end}}

### 外键

{{if .ForeignKeys -}}
| 外键名 | 字段 | 引用表 | 引用字段 | 删除规则 | 更新规则 |
|--------|------|--------|----------|----------|----------|
{{- range .ForeignKeys }}
| {{.Name}} | {{.Column}} | {{.RefTable}} | {{.RefColumn}} | {{.GetOnDeleteRule}} | {{.GetOnUpdateRule}} |
{{- end }}

#### 关联关系
{{- range .ForeignKeys }}
- **{{.Name}}**: ` + "`" + `{{$.Name}}.{{.Column}}` + "`" + ` → ` + "`" + `{{.RefTable}}.{{.RefColumn}}` + "`" + ` (删除时{{.GetOnDeleteRule}})
{{- end }}
{{else -}}
*未定义外键*

此表没有与其他表的外键关联关系。
{{end}}

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
