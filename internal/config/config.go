package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/schema-export/schema-export/internal/inspector"
)

// Config 应用配置，包含数据库连接配置和导出选项配置。
//
// Config 是顶层配置容器，整合了所有配置项。配置可以通过以下方式加载：
//   - 命令行参数（优先级最高）
//   - 环境变量（通过 LoadFromEnv 方法）
//   - 默认值（通过 DefaultConfig 函数）
type Config struct {
	Database DatabaseConfig `yaml:"database" json:"database"` // 数据库连接配置
	Export   ExportConfig   `yaml:"export" json:"export"`     // 导出选项配置
}

// DatabaseConfig 数据库连接配置。
//
// 支持两种连接方式：
//  1. 使用 DSN（推荐）：提供完整的连接字符串
//  2. 使用分离参数：分别提供 Host、Port、Username、Password、Database
//
// 对于 Oracle 和达梦数据库，Schema 参数用于指定要导出的 Schema。
//
// 字段说明:
//   - Type: 数据库类型（dm, oracle, sqlserver, mysql, postgres）
//   - Host: 数据库主机地址
//   - Port: 数据库端口号
//   - Database: 数据库名称
//   - Username: 数据库用户名
//   - Password: 数据库密码
//   - DSN: 完整的数据源名称（Data Source Name）
//   - Schema: 数据库 Schema（用于 Oracle/达梦）
//   - SSLMode: SSL 连接模式（用于 PostgreSQL）
type DatabaseConfig struct {
	Type     string `yaml:"type" json:"type"`           // 数据库类型
	Host     string `yaml:"host" json:"host"`           // 数据库主机
	Port     int    `yaml:"port" json:"port"`           // 数据库端口
	Database string `yaml:"database" json:"database"`   // 数据库名
	Username string `yaml:"username" json:"username"`   // 用户名
	Password string `yaml:"password" json:"password"`   // 密码
	DSN      string `yaml:"dsn" json:"dsn"`             // DSN 连接字符串
	Schema   string `yaml:"schema" json:"schema"`       // 数据库 Schema
	SSLMode  string `yaml:"ssl_mode" json:"ssl_mode"`   // SSL 模式
}

// ExportConfig 导出选项配置。
//
// 控制导出的行为，包括输出目录、格式、表过滤规则等。
//
// 字段说明:
//   - OutputDir: 输出目录路径
//   - Formats: 导出格式列表（markdown, sql, json, yaml）
//   - SplitFiles: 是否按表分文件导出
//   - Tables: 要导出的表名列表（白名单）
//   - Exclude: 要排除的表名列表（黑名单）
//   - Patterns: 表名正则匹配模式列表
//   - IncludeViews: 是否导出视图
//   - IncludeProcedures: 是否导出存储过程
//   - IncludeFunctions: 是否导出函数
//   - IncludeTriggers: 是否导出触发器
//   - IncludeSequences: 是否导出序列
type ExportConfig struct {
	OutputDir         string   `yaml:"output_dir" json:"output_dir"`                   // 输出目录
	Formats           []string `yaml:"formats" json:"formats"`                         // 导出格式
	SplitFiles        bool     `yaml:"split_files" json:"split_files"`                 // 分文件导出
	Tables            []string `yaml:"tables" json:"tables"`                           // 要导出的表
	Exclude           []string `yaml:"exclude" json:"exclude"`                         // 要排除的表
	Patterns          []string `yaml:"patterns" json:"patterns"`                       // 表名匹配模式
	IncludeViews      bool     `yaml:"include_views" json:"include_views"`             // 包含视图
	IncludeProcedures bool     `yaml:"include_procedures" json:"include_procedures"`   // 包含存储过程
	IncludeFunctions  bool     `yaml:"include_functions" json:"include_functions"`     // 包含函数
	IncludeTriggers   bool     `yaml:"include_triggers" json:"include_triggers"`       // 包含触发器
	IncludeSequences  bool     `yaml:"include_sequences" json:"include_sequences"`     // 包含序列
}

// ToConnectionConfig 将 DatabaseConfig 转换为 Inspector 连接配置。
//
// 该方法用于将配置层的数据库配置转换为 Inspector 接口所需的连接配置格式。
//
// 返回值:
//   - inspector.ConnectionConfig: 可直接用于创建 Inspector 实例的连接配置
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

// LoadFromEnv 从环境变量加载配置。
//
// 该方法会检查所有支持的环境变量，并将非空值覆盖到当前配置中。
// 环境变量的优先级低于命令行参数，但高于默认值。
//
// 支持的环境变量:
//   - DB_TYPE: 数据库类型
//   - DB_HOST: 数据库主机
//   - DB_PORT: 数据库端口
//   - DB_DATABASE: 数据库名称
//   - DB_USERNAME: 数据库用户名
//   - DB_PASSWORD: 数据库密码
//   - DB_DSN: 数据源名称
//   - DB_SCHEMA: 数据库 Schema
//   - EXPORT_OUTPUT: 输出目录
//   - EXPORT_FORMATS: 导出格式（逗号分隔）
//   - EXPORT_SPLIT: 是否分文件导出（true/false）
//   - EXPORT_INCLUDE_VIEWS: 是否包含视图（true/false）
//   - EXPORT_INCLUDE_PROCEDURES: 是否包含存储过程（true/false）
//   - EXPORT_INCLUDE_FUNCTIONS: 是否包含函数（true/false）
//   - EXPORT_INCLUDE_TRIGGERS: 是否包含触发器（true/false）
//   - EXPORT_INCLUDE_SEQUENCES: 是否包含序列（true/false）
func (c *Config) LoadFromEnv() {
	if v := os.Getenv("DB_TYPE"); v != "" {
		c.Database.Type = v
	}
	if v := os.Getenv("DB_HOST"); v != "" {
		c.Database.Host = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			c.Database.Port = port
		}
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
		c.Export.Formats = normalizeFormats(strings.Split(v, ","))
	}
	if v := os.Getenv("EXPORT_SPLIT"); v != "" {
		c.Export.SplitFiles = (v == "true" || v == "1")
	}
	if v := os.Getenv("EXPORT_INCLUDE_VIEWS"); v != "" {
		c.Export.IncludeViews = (v == "true" || v == "1")
	}
	if v := os.Getenv("EXPORT_INCLUDE_PROCEDURES"); v != "" {
		c.Export.IncludeProcedures = (v == "true" || v == "1")
	}
	if v := os.Getenv("EXPORT_INCLUDE_FUNCTIONS"); v != "" {
		c.Export.IncludeFunctions = (v == "true" || v == "1")
	}
	if v := os.Getenv("EXPORT_INCLUDE_TRIGGERS"); v != "" {
		c.Export.IncludeTriggers = (v == "true" || v == "1")
	}
	if v := os.Getenv("EXPORT_INCLUDE_SEQUENCES"); v != "" {
		c.Export.IncludeSequences = (v == "true" || v == "1")
	}
}

// Validate 验证配置的有效性。
//
// 该方法执行以下验证和规范化操作：
//  1. 检查必需字段（数据库类型、主机或 DSN、用户名）
//  2. 从 DSN 中提取 Schema 参数（如果未显式指定）
//  3. 规范化 Schema 名称（Oracle/达梦转大写）
//  4. 规范化导出格式（转小写、去空白）
//  5. 设置默认值（输出目录、导出格式）
//
// 返回值:
//   - error: 如果配置无效，返回描述错误的 error；否则返回 nil
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

	c.Database.Schema = normalizeSchema(c.Database.Type, c.Database.Schema)
	c.Export.Formats = normalizeFormats(c.Export.Formats)

	if len(c.Export.Formats) == 0 {
		c.Export.Formats = []string{"markdown"}
	}

	if c.Export.OutputDir == "" {
		c.Export.OutputDir = "./output"
	}

	return nil
}

// extractSchemaFromDSN 从 DSN 中提取 schema 参数。
//
// 支持的 DSN 格式:
//   - dm://user:password@host:port?schema=SCHEMA_NAME
//   - dm://user:password@host:port?schema=SCHEMA_NAME&other=params
//
// 参数:
//   - dsn: 数据源名称字符串
//
// 返回值:
//   - string: 提取的 schema 名称，如果不存在或解析失败则返回空字符串
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

// normalizeFormats 规范化导出格式列表。
//
// 执行以下规范化操作：
//   - 去除每个格式字符串的前后空白
//   - 转换为小写
//   - 过滤空字符串
//
// 参数:
//   - formats: 原始格式列表
//
// 返回值:
//   - []string: 规范化后的格式列表
func normalizeFormats(formats []string) []string {
	var normalized []string
	for _, format := range formats {
		format = strings.TrimSpace(strings.ToLower(format))
		if format != "" {
			normalized = append(normalized, format)
		}
	}
	return normalized
}

// normalizeSchema 规范化 Schema 名称。
//
// 根据数据库类型对 Schema 名称进行规范化处理：
//   - Oracle/达梦: 转换为大写（除非已用双引号括起）
//   - 其他数据库: 保持原样
//
// 参数:
//   - dbType: 数据库类型
//   - schema: 原始 Schema 名称
//
// 返回值:
//   - string: 规范化后的 Schema 名称
func normalizeSchema(dbType, schema string) string {
	schema = strings.TrimSpace(schema)
	if schema == "" {
		return ""
	}

	switch strings.ToLower(strings.TrimSpace(dbType)) {
	case "oracle", "dm":
		if strings.HasPrefix(schema, "\"") && strings.HasSuffix(schema, "\"") {
			return schema
		}
		return strings.ToUpper(schema)
	default:
		return schema
	}
}

// DefaultConfig 返回默认配置。
//
// 默认配置包括：
//   - 数据库类型: dm（达梦）
//   - 数据库端口: 5236
//   - 输出目录: ./output
//   - 导出格式: markdown
//   - 分文件导出: false
//
// 返回值:
//   - *Config: 包含默认值的配置对象指针
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
