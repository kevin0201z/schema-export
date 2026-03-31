package config

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/schema-export/schema-export/internal/inspector"
	"github.com/schema-export/schema-export/internal/model"
)

// Config 应用配置
type Config struct {
	Database DatabaseConfig `yaml:"database" json:"database"`
	Export   ExportConfig   `yaml:"export" json:"export"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type     string `yaml:"type" json:"type"`
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	Database string `yaml:"database" json:"database"`
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
	DSN      string `yaml:"dsn" json:"dsn"`
	Schema   string `yaml:"schema" json:"schema"`
	SSLMode  string `yaml:"ssl_mode" json:"ssl_mode"`
}

// ExportConfig 导出配置
type ExportConfig struct {
	OutputDir  string   `yaml:"output_dir" json:"output_dir"`
	Formats    []string `yaml:"formats" json:"formats"`
	SplitFiles bool     `yaml:"split_files" json:"split_files"`
	Tables     []string `yaml:"tables" json:"tables"`
	Exclude    []string `yaml:"exclude" json:"exclude"`
	Patterns   []string `yaml:"patterns" json:"patterns"`
}

// ToConnectionConfig 转换为 Inspector 连接配置
func (d *DatabaseConfig) ToConnectionConfig() inspector.ConnectionConfig {
	return inspector.ConnectionConfig{
		Type:     d.Type,
		Host:     d.Host,
		Port:     d.Port,
		Database: d.Database,
		Username: d.Username,
		Password: d.Password,
		DSN:      d.DSN,
		Schema:   d.Schema,
		SSLMode:  d.SSLMode,
	}
}

// LoadFromEnv 从环境变量加载配置
func (c *Config) LoadFromEnv() {
	if v := os.Getenv("DB_TYPE"); v != "" {
		c.Database.Type = v
	}
	if v := os.Getenv("DB_HOST"); v != "" {
		c.Database.Host = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		var port int
		fmt.Sscanf(v, "%d", &port)
		c.Database.Port = port
	}
	if v := os.Getenv("DB_DATABASE"); v != "" {
		c.Database.Database = v
	}
	if v := os.Getenv("DB_USERNAME"); v != "" {
		c.Database.Username = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		c.Database.Password = v
	}
	if v := os.Getenv("DB_DSN"); v != "" {
		c.Database.DSN = v
	}
	if v := os.Getenv("DB_SCHEMA"); v != "" {
		c.Database.Schema = v
	}
	if v := os.Getenv("EXPORT_OUTPUT"); v != "" {
		c.Export.OutputDir = v
	}
	if v := os.Getenv("EXPORT_FORMATS"); v != "" {
		c.Export.Formats = strings.Split(v, ",")
	}
	if v := os.Getenv("EXPORT_SPLIT"); v != "" {
		c.Export.SplitFiles = (v == "true" || v == "1")
	}
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.Database.Type == "" {
		return fmt.Errorf("database type is required")
	}

	if c.Database.DSN == "" {
		if c.Database.Host == "" {
			return fmt.Errorf("database host or DSN is required")
		}
		if c.Database.Username == "" {
			return fmt.Errorf("database username is required")
		}
	} else {
		// 如果 DSN 中有 schema 参数，提取出来
		if c.Database.Schema == "" {
			c.Database.Schema = extractSchemaFromDSN(c.Database.DSN)
		}
	}

	if len(c.Export.Formats) == 0 {
		c.Export.Formats = []string{"markdown"}
	}

	if c.Export.OutputDir == "" {
		c.Export.OutputDir = "./output"
	}

	return nil
}

// extractSchemaFromDSN 从 DSN 中提取 schema 参数
func extractSchemaFromDSN(dsn string) string {
	// 支持格式: dm://user:password@host:port?schema=SCHEMA_NAME
	// 或: dm://user:password@host:port?schema=SCHEMA_NAME&other=params
	u, err := url.Parse(dsn)
	if err != nil {
		return ""
	}
	schema := u.Query().Get("schema")
	return schema
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Database: DatabaseConfig{
			Type: "dm",
			Port: 5236,
		},
		Export: ExportConfig{
			OutputDir:  "./output",
			Formats:    []string{"markdown"},
			SplitFiles: false,
		},
	}
}

// TableFilter 表过滤器
type TableFilter struct {
	Include  []string
	Exclude  []string
	Patterns []*regexp.Regexp
}

// NewTableFilter 创建表过滤器
func NewTableFilter(include, exclude, patterns []string) (*TableFilter, error) {
	filter := &TableFilter{
		Include: include,
		Exclude: exclude,
	}

	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern %s: %w", pattern, err)
		}
		filter.Patterns = append(filter.Patterns, re)
	}

	return filter, nil
}

// ShouldInclude 判断表是否应该被包含
func (f *TableFilter) ShouldInclude(tableName string) bool {
	// 如果有指定包含列表，检查是否在列表中
	if len(f.Include) > 0 {
		found := false
		for _, name := range f.Include {
			if strings.EqualFold(name, tableName) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// 检查排除列表
	for _, name := range f.Exclude {
		if strings.EqualFold(name, tableName) {
			return false
		}
	}

	// 检查正则模式
	if len(f.Patterns) > 0 {
		matched := false
		for _, pattern := range f.Patterns {
			if pattern.MatchString(tableName) {
				matched = true
				break
			}
		}
		if !matched && len(f.Include) == 0 {
			return false
		}
	}

	return true
}

// FilterTables 过滤表列表
func (f *TableFilter) FilterTables(tables []model.Table) []model.Table {
	var filtered []model.Table
	for _, table := range tables {
		if f.ShouldInclude(table.Name) {
			filtered = append(filtered, table)
		}
	}
	return filtered
}
