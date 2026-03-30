package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/schema-export/schema-export/internal/config"
	"github.com/schema-export/schema-export/internal/exporter"
	"github.com/schema-export/schema-export/internal/inspector"
	"github.com/schema-export/schema-export/internal/model"
)

// ExportCommand 导出命令
type ExportCommand struct {
	Config *config.Config
}

// NewExportCommand 创建导出命令
func NewExportCommand() *ExportCommand {
	return &ExportCommand{
		Config: config.DefaultConfig(),
	}
}

// Run 执行导出
func (c *ExportCommand) Run() error {
	// 加载环境变量配置
	c.Config.LoadFromEnv()
	
	// 验证配置
	if err := c.Config.Validate(); err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}
	
	// 获取 Inspector 工厂
	factory, ok := inspector.GetFactory(c.Config.Database.Type)
	if !ok {
		return fmt.Errorf("unsupported database type: %s", c.Config.Database.Type)
	}
	
	// 创建 Inspector
	connConfig := c.Config.Database.ToConnectionConfig()
	ins, err := factory.Create(connConfig)
	if err != nil {
		return fmt.Errorf("failed to create inspector: %w", err)
	}
	
	// 连接数据库
	ctx := context.Background()
	if err := ins.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer ins.Close()
	
	// 测试连接
	if err := ins.TestConnection(ctx); err != nil {
		return fmt.Errorf("database connection test failed: %w", err)
	}
	
	fmt.Printf("Connected to %s database\n", c.Config.Database.Type)
	
	// 获取所有表
	tables, err := ins.GetTables(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tables: %w", err)
	}
	
	fmt.Printf("Found %d tables\n", len(tables))
	
	// 应用表过滤器
	filter, err := config.NewTableFilter(
		c.Config.Export.Tables,
		c.Config.Export.Exclude,
		c.Config.Export.Patterns,
	)
	if err != nil {
		return fmt.Errorf("invalid table filter: %w", err)
	}
	
	tables = filter.FilterTables(tables)
	fmt.Printf("Filtered to %d tables\n", len(tables))
	
	// 获取完整表元数据
	var fullTables []model.Table
	for _, table := range tables {
		fullTable, err := ins.GetTable(ctx, table.Name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to get table %s: %v\n", table.Name, err)
			continue
		}
		fullTables = append(fullTables, *fullTable)
	}
	
	// 导出
	for _, format := range c.Config.Export.Formats {
		if err := c.exportFormat(fullTables, format); err != nil {
			fmt.Fprintf(os.Stderr, "Error exporting to %s: %v\n", format, err)
			continue
		}
	}
	
	fmt.Println("Export completed successfully!")
	return nil
}

// exportFormat 导出指定格式
func (c *ExportCommand) exportFormat(tables []model.Table, format string) error {
	factory, ok := exporter.GetFactory(format)
	if !ok {
		return fmt.Errorf("unsupported export format: %s", format)
	}
	
	exp, err := factory.Create()
	if err != nil {
		return fmt.Errorf("failed to create exporter: %w", err)
	}
	
	// 解析输出路径
	outputPath := c.Config.Export.OutputDir
	outputDir, fileName := parseOutputPath(outputPath, format)
	
	options := exporter.ExportOptions{
		OutputDir:  outputDir,
		FileName:   fileName,
		SplitFiles: c.Config.Export.SplitFiles,
	}
	
	if err := exp.Export(tables, options); err != nil {
		return fmt.Errorf("export failed: %w", err)
	}
	
	fmt.Printf("Exported to %s format\n", format)
	return nil
}

// parseOutputPath 解析输出路径，返回目录和文件名
func parseOutputPath(outputPath string, format string) (dir string, fileName string) {
	if outputPath == "" {
		return "./output", ""
	}
	
	// 检查是否有文件扩展名
	ext := filepath.Ext(outputPath)
	if ext != "" {
		// 用户指定了文件名，提取目录和文件名
		dir = filepath.Dir(outputPath)
		fileName = filepath.Base(outputPath)
		// 根据格式调整扩展名
		if format == "sql" && ext != ".sql" {
			fileName = fileName[:len(fileName)-len(ext)] + ".sql"
		} else if format == "markdown" && ext != ".md" {
			fileName = fileName[:len(fileName)-len(ext)] + ".md"
		}
	} else {
		// 用户只指定了目录
		dir = outputPath
		fileName = ""
	}
	
	return dir, fileName
}

// SetDatabaseType 设置数据库类型
func (c *ExportCommand) SetDatabaseType(dbType string) {
	c.Config.Database.Type = dbType
}

// SetDatabaseHost 设置数据库主机
func (c *ExportCommand) SetDatabaseHost(host string) {
	c.Config.Database.Host = host
}

// SetDatabasePort 设置数据库端口
func (c *ExportCommand) SetDatabasePort(port int) {
	c.Config.Database.Port = port
}

// SetDatabaseName 设置数据库名
func (c *ExportCommand) SetDatabaseName(name string) {
	c.Config.Database.Database = name
}

// SetDatabaseUsername 设置数据库用户名
func (c *ExportCommand) SetDatabaseUsername(username string) {
	c.Config.Database.Username = username
}

// SetDatabasePassword 设置数据库密码
func (c *ExportCommand) SetDatabasePassword(password string) {
	c.Config.Database.Password = password
}

// SetDatabaseDSN 设置数据库 DSN
func (c *ExportCommand) SetDatabaseDSN(dsn string) {
	c.Config.Database.DSN = dsn
}

// SetDatabaseSchema 设置数据库 Schema
func (c *ExportCommand) SetDatabaseSchema(schema string) {
	c.Config.Database.Schema = schema
}

// SetOutputDir 设置输出目录
func (c *ExportCommand) SetOutputDir(dir string) {
	c.Config.Export.OutputDir = dir
}

// SetFormats 设置导出格式
func (c *ExportCommand) SetFormats(formats []string) {
	c.Config.Export.Formats = formats
}

// SetSplitFiles 设置是否分文件导出
func (c *ExportCommand) SetSplitFiles(split bool) {
	c.Config.Export.SplitFiles = split
}

// SetTables 设置要导出的表
func (c *ExportCommand) SetTables(tables []string) {
	c.Config.Export.Tables = tables
}

// SetExclude 设置要排除的表
func (c *ExportCommand) SetExclude(exclude []string) {
	c.Config.Export.Exclude = exclude
}

// SetPatterns 设置表名匹配模式
func (c *ExportCommand) SetPatterns(patterns []string) {
	c.Config.Export.Patterns = patterns
}

// ParseFormats 解析格式字符串
func ParseFormats(formats string) []string {
	if formats == "" {
		return []string{"markdown"}
	}
	parts := strings.Split(formats, ",")
	var result []string
	for _, part := range parts {
		part = strings.TrimSpace(strings.ToLower(part))
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

// ParseTables 解析表名字符串
func ParseTables(tables string) []string {
	if tables == "" {
		return nil
	}
	parts := strings.Split(tables, ",")
	var result []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}
