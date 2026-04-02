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
	tableTemplate *template.Template
	viewTemplate  *template.Template
}

// NewExporter 创建 Markdown 导出器
func NewExporter() *Exporter {
	funcMap := template.FuncMap{
		"join":       strings.Join,
		"toLower":    strings.ToLower,
		"hasPrefix":  strings.HasPrefix,
		"hasSuffix":  strings.HasSuffix,
		"contains":   strings.Contains,
		"trimPrefix": strings.TrimPrefix,
		"trimSuffix": strings.TrimSuffix,
	}

	return &Exporter{
		tableTemplate: template.Must(template.New("table").Funcs(funcMap).Parse(tableTemplate)),
		viewTemplate:  template.Must(template.New("view").Funcs(funcMap).Parse(viewTemplate)),
	}
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
		outputPath = filepath.Join(options.OutputDir, "schema.md")
	}

	info, err := os.Stat(outputPath)
	if err == nil && info.IsDir() {
		outputPath = filepath.Join(outputPath, "schema.md")
	}

	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	info, err = os.Stat(dir)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("output path %s is not a valid directory", dir)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	fmt.Fprintln(file, "# 数据库结构文档")
	fmt.Fprintln(file, "")
	fmt.Fprintf(file, "**总表数:** %d\n\n", len(tables))
	if options.IncludeViews {
		fmt.Fprintf(file, "**总视图数:** %d\n\n", len(views))
	}

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

	if options.IncludeViews && len(views) > 0 {
		fmt.Fprintln(file, "### 视图清单")
		fmt.Fprintln(file, "")
		for _, view := range views {
			if view.Comment != "" {
				fmt.Fprintf(file, "- **%s**: %s\n", view.Name, view.Comment)
			} else {
				fmt.Fprintf(file, "- **%s**\n", view.Name)
			}
		}
		fmt.Fprintln(file, "")
	}

	fmt.Fprintln(file, "### 目录")
	fmt.Fprintln(file, "")
	fmt.Fprintln(file, "#### 表")
	fmt.Fprintln(file, "")
	for _, table := range tables {
		fmt.Fprintf(file, "- [%s](#表-%s)\n", table.Name, strings.ToLower(table.Name))
	}
	fmt.Fprintln(file, "")

	if options.IncludeViews && len(views) > 0 {
		fmt.Fprintln(file, "#### 视图")
		fmt.Fprintln(file, "")
		for _, view := range views {
			fmt.Fprintf(file, "- [%s](#视图-%s)\n", view.Name, strings.ToLower(view.Name))
		}
		fmt.Fprintln(file, "")
	}

	fmt.Fprintln(file, "## 表详情")
	fmt.Fprintln(file, "")
	for _, table := range tables {
		if err := e.tableTemplate.Execute(file, table); err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}
		fmt.Fprintln(file, "")
	}

	if options.IncludeViews && len(views) > 0 {
		fmt.Fprintln(file, "## 视图详情")
		fmt.Fprintln(file, "")
		for _, view := range views {
			if err := e.viewTemplate.Execute(file, view); err != nil {
				return fmt.Errorf("failed to execute view template: %w", err)
			}
			fmt.Fprintln(file, "")
		}
	}

	return nil
}

// exportSplitFiles 分文件导出
func (e *Exporter) exportSplitFiles(tables []model.Table, views []model.View, options exporter.ExportOptions) error {
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

		fmt.Fprintf(file, "# 表: %s\n\n", table.Name)

		if err := e.tableTemplate.Execute(file, table); err != nil {
			file.Close()
			return fmt.Errorf("failed to execute template: %w", err)
		}
		file.Close()
	}

	if options.IncludeViews {
		for _, view := range views {
			outputPath := filepath.Join(markdownDir, view.Name+"_view.md")
			file, err := os.Create(outputPath)
			if err != nil {
				return fmt.Errorf("failed to create file %s: %w", outputPath, err)
			}

			fmt.Fprintf(file, "# 视图: %s\n\n", view.Name)

			if err := e.viewTemplate.Execute(file, view); err != nil {
				file.Close()
				return fmt.Errorf("failed to execute view template: %w", err)
			}
			file.Close()
		}
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
