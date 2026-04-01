package cli

import (
	"strings"

	exportapp "github.com/schema-export/schema-export/internal/app/export"
	"github.com/schema-export/schema-export/internal/config"
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
	c.Config.LoadFromEnv()
	if err := c.Config.Validate(); err != nil {
		return err
	}

	return exportapp.NewService(c.Config).Run()
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
