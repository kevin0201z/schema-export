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

type Exporter struct {
	tableTemplate *template.Template
	viewTemplate  *template.Template
}

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

func (e *Exporter) Export(tables []model.Table, views []model.View, procedures []model.Procedure, functions []model.Function, triggers []model.Trigger, sequences []model.Sequence, options exporter.ExportOptions) error {
	if options.SplitFiles {
		return e.exportSplitFiles(tables, views, procedures, functions, triggers, sequences, options)
	}
	return e.exportSingleFile(tables, views, procedures, functions, triggers, sequences, options)
}

func (e *Exporter) exportSingleFile(tables []model.Table, views []model.View, procedures []model.Procedure, functions []model.Function, triggers []model.Trigger, sequences []model.Sequence, options exporter.ExportOptions) error {
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
	if options.IncludeProcedures {
		fmt.Fprintf(file, "**总存储过程数:** %d\n\n", len(procedures))
	}
	if options.IncludeFunctions {
		fmt.Fprintf(file, "**总函数数:** %d\n\n", len(functions))
	}
	if options.IncludeTriggers {
		fmt.Fprintf(file, "**总触发器数:** %d\n\n", len(triggers))
	}
	if options.IncludeSequences {
		fmt.Fprintf(file, "**总序列数:** %d\n\n", len(sequences))
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

	if options.IncludeProcedures && len(procedures) > 0 {
		fmt.Fprintln(file, "### 存储过程清单")
		fmt.Fprintln(file, "")
		for _, proc := range procedures {
			if proc.Comment != "" {
				fmt.Fprintf(file, "- **%s**: %s\n", proc.Name, proc.Comment)
			} else {
				fmt.Fprintf(file, "- **%s**\n", proc.Name)
			}
		}
		fmt.Fprintln(file, "")
	}

	if options.IncludeFunctions && len(functions) > 0 {
		fmt.Fprintln(file, "### 函数清单")
		fmt.Fprintln(file, "")
		for _, fn := range functions {
			if fn.Comment != "" {
				fmt.Fprintf(file, "- **%s**: %s\n", fn.Name, fn.Comment)
			} else {
				fmt.Fprintf(file, "- **%s**\n", fn.Name)
			}
		}
		fmt.Fprintln(file, "")
	}

	if options.IncludeTriggers && len(triggers) > 0 {
		fmt.Fprintln(file, "### 触发器清单")
		fmt.Fprintln(file, "")
		for _, tr := range triggers {
			fmt.Fprintf(file, "- **%s** (表: %s)\n", tr.Name, tr.TableName)
		}
		fmt.Fprintln(file, "")
	}

	if options.IncludeSequences && len(sequences) > 0 {
		fmt.Fprintln(file, "### 序列清单")
		fmt.Fprintln(file, "")
		for _, seq := range sequences {
			fmt.Fprintf(file, "- **%s**\n", seq.Name)
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

	if options.IncludeProcedures && len(procedures) > 0 {
		fmt.Fprintln(file, "## 存储过程详情")
		fmt.Fprintln(file, "")
		for _, proc := range procedures {
			e.writeProcedure(file, proc)
		}
	}

	if options.IncludeFunctions && len(functions) > 0 {
		fmt.Fprintln(file, "## 函数详情")
		fmt.Fprintln(file, "")
		for _, fn := range functions {
			e.writeFunction(file, fn)
		}
	}

	if options.IncludeTriggers && len(triggers) > 0 {
		fmt.Fprintln(file, "## 触发器详情")
		fmt.Fprintln(file, "")
		for _, tr := range triggers {
			e.writeTrigger(file, tr)
		}
	}

	if options.IncludeSequences && len(sequences) > 0 {
		fmt.Fprintln(file, "## 序列详情")
		fmt.Fprintln(file, "")
		for _, seq := range sequences {
			e.writeSequence(file, seq)
		}
	}

	return nil
}

func (e *Exporter) exportSplitFiles(tables []model.Table, views []model.View, procedures []model.Procedure, functions []model.Function, triggers []model.Trigger, sequences []model.Sequence, options exporter.ExportOptions) error {
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

	if options.IncludeProcedures {
		for _, proc := range procedures {
			outputPath := filepath.Join(markdownDir, proc.Name+"_procedure.md")
			file, err := os.Create(outputPath)
			if err != nil {
				return fmt.Errorf("failed to create file %s: %w", outputPath, err)
			}
			fmt.Fprintf(file, "# 存储过程: %s\n\n", proc.Name)
			e.writeProcedure(file, proc)
			file.Close()
		}
	}

	if options.IncludeFunctions {
		for _, fn := range functions {
			outputPath := filepath.Join(markdownDir, fn.Name+"_function.md")
			file, err := os.Create(outputPath)
			if err != nil {
				return fmt.Errorf("failed to create file %s: %w", outputPath, err)
			}
			fmt.Fprintf(file, "# 函数: %s\n\n", fn.Name)
			e.writeFunction(file, fn)
			file.Close()
		}
	}

	if options.IncludeTriggers {
		for _, tr := range triggers {
			outputPath := filepath.Join(markdownDir, tr.Name+"_trigger.md")
			file, err := os.Create(outputPath)
			if err != nil {
				return fmt.Errorf("failed to create file %s: %w", outputPath, err)
			}
			fmt.Fprintf(file, "# 触发器: %s\n\n", tr.Name)
			e.writeTrigger(file, tr)
			file.Close()
		}
	}

	if options.IncludeSequences {
		for _, seq := range sequences {
			outputPath := filepath.Join(markdownDir, seq.Name+"_sequence.md")
			file, err := os.Create(outputPath)
			if err != nil {
				return fmt.Errorf("failed to create file %s: %w", outputPath, err)
			}
			fmt.Fprintf(file, "# 序列: %s\n\n", seq.Name)
			e.writeSequence(file, seq)
			file.Close()
		}
	}

	return nil
}

func (e *Exporter) writeProcedure(file *os.File, proc model.Procedure) {
	fmt.Fprintf(file, "## 存储过程: %s\n\n", proc.Name)
	if proc.Comment != "" {
		fmt.Fprintf(file, "**描述:** %s\n\n", proc.Comment)
	}
	fmt.Fprintln(file, "### 定义")
	fmt.Fprintln(file, "")
	fmt.Fprintln(file, "```sql")
	fmt.Fprintln(file, proc.Definition)
	fmt.Fprintln(file, "```")
	fmt.Fprintln(file, "")
}

func (e *Exporter) writeFunction(file *os.File, fn model.Function) {
	fmt.Fprintf(file, "## 函数: %s\n\n", fn.Name)
	if fn.Comment != "" {
		fmt.Fprintf(file, "**描述:** %s\n\n", fn.Comment)
	}
	if fn.ReturnType != "" {
		fmt.Fprintf(file, "**返回类型:** %s\n\n", fn.ReturnType)
	}
	fmt.Fprintln(file, "### 定义")
	fmt.Fprintln(file, "")
	fmt.Fprintln(file, "```sql")
	fmt.Fprintln(file, fn.Definition)
	fmt.Fprintln(file, "```")
	fmt.Fprintln(file, "")
}

func (e *Exporter) writeTrigger(file *os.File, tr model.Trigger) {
	fmt.Fprintf(file, "## 触发器: %s\n\n", tr.Name)
	fmt.Fprintf(file, "**所属表:** %s\n\n", tr.TableName)
	if tr.Event != "" {
		fmt.Fprintf(file, "**触发事件:** %s\n\n", tr.Event)
	}
	if tr.Timing != "" {
		fmt.Fprintf(file, "**触发时机:** %s\n\n", tr.Timing)
	}
	fmt.Fprintf(file, "**状态:** %s\n\n", tr.Status)
	if tr.Definition != "" {
		fmt.Fprintln(file, "### 定义")
		fmt.Fprintln(file, "")
		fmt.Fprintln(file, "```sql")
		fmt.Fprintln(file, tr.Definition)
		fmt.Fprintln(file, "```")
		fmt.Fprintln(file, "")
	}
}

func (e *Exporter) writeSequence(file *os.File, seq model.Sequence) {
	fmt.Fprintf(file, "## 序列: %s\n\n", seq.Name)
	fmt.Fprintln(file, "### 属性")
	fmt.Fprintln(file, "")
	fmt.Fprintln(file, "| 属性 | 值 |")
	fmt.Fprintln(file, "|------|-----|")
	fmt.Fprintf(file, "| 最小值 | %d |\n", seq.MinValue)
	fmt.Fprintf(file, "| 最大值 | %d |\n", seq.MaxValue)
	fmt.Fprintf(file, "| 增量 | %d |\n", seq.IncrementBy)
	fmt.Fprintf(file, "| 是否循环 | %v |\n", seq.Cycle)
	fmt.Fprintf(file, "| 缓存大小 | %d |\n", seq.CacheSize)
	fmt.Fprintf(file, "| 当前值 | %d |\n", seq.LastValue)
	fmt.Fprintln(file, "")
}

func (e *Exporter) GetName() string {
	return "markdown"
}

func (e *Exporter) GetExtension() string {
	return ".md"
}

type Factory struct{}

func (f *Factory) Create() (exporter.Exporter, error) {
	return NewExporter(), nil
}

func (f *Factory) GetType() string {
	return "markdown"
}

func init() {
	exporter.Register("markdown", &Factory{})
}
