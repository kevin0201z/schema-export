package tui

import (
	"strconv"
	"strings"

	"github.com/schema-export/schema-export/internal/config"
)

// ToConfig 将 TUI 表单状态转换为配置对象。
//
// 映射规则：
//  1. DSN 模式下仅映射 DSN + Schema，清空分离参数字段
//  2. Params 模式下映射分离参数，清空 DSN 字段
//  3. 逗号分隔字符串解析为 []string
//  4. Port 从字符串转换为 int
//  5. 不调用 LoadFromEnv（环境变量已在 defaultTuiState 中读取为初始值）
//  6. 调用 Validate 进行规范化和默认值设置
func (s *TuiState) ToConfig() *config.Config {
	cfg := config.DefaultConfig()
	s.normalizeContentOptionsForDBType()

	// 数据库连接
	cfg.Database.Type = s.DBType

	switch s.ConnectionMethod {
	case ConnectionMethodDSN:
		cfg.Database.DSN = s.DSN
		cfg.Database.Schema = s.Schema
		// 清空分离参数字段
		cfg.Database.Host = ""
		cfg.Database.Port = 0
		cfg.Database.Username = ""
		cfg.Database.Password = ""
		cfg.Database.Database = ""
	case ConnectionMethodParams:
		cfg.Database.Database = s.Database
		cfg.Database.Schema = s.Schema
		if strings.EqualFold(strings.TrimSpace(s.DBType), "sqlite") {
			cfg.Database.Host = ""
			cfg.Database.Port = 0
			cfg.Database.Username = ""
			cfg.Database.Password = ""
		} else {
			cfg.Database.Host = s.Host
			cfg.Database.Username = s.Username
			cfg.Database.Password = s.Password
			cfg.Database.Port = defaultPortForDBType(s.DBType)
			if p := strings.TrimSpace(s.Port); p != "" {
				if port, err := strconv.Atoi(p); err == nil {
					cfg.Database.Port = port
				}
			}
		}
		// 清空 DSN 字段
		cfg.Database.DSN = ""
	}

	// 导出内容
	cfg.Export.IncludeViews = s.IncludeViews
	cfg.Export.IncludeProcedures = s.IncludeProcedures
	cfg.Export.IncludeFunctions = s.IncludeFunctions
	cfg.Export.IncludeTriggers = s.IncludeTriggers
	cfg.Export.IncludeSequences = s.IncludeSequences

	// 表过滤
	cfg.Export.Tables = parseCommaSeparated(s.Tables)
	cfg.Export.Exclude = parseCommaSeparated(s.Exclude)
	cfg.Export.Patterns = parseCommaSeparated(s.Patterns)

	// 输出设置
	if s.OutputDir != "" {
		cfg.Export.OutputDir = s.OutputDir
	}
	if len(s.Formats) > 0 {
		cfg.Export.Formats = s.Formats
	}
	cfg.Export.SplitFiles = s.SplitFiles

	// 规范化（Schema 大写、格式小写、默认值等）
	_ = cfg.Validate()

	return cfg
}

func defaultPortForDBType(dbType string) int {
	switch strings.ToLower(strings.TrimSpace(dbType)) {
	case "oracle":
		return 1521
	case "sqlserver":
		return 1433
	case "mysql":
		return 3306
	case "postgres":
		return 5432
	case "sqlite":
		return 0
	case "dm":
		fallthrough
	default:
		return 5236
	}
}

// parseCommaSeparated 解析逗号分隔的字符串为切片。
func parseCommaSeparated(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// applyContentCheckboxes 将复选框选择应用到 TuiState。
func (s *TuiState) applyContentCheckboxes(cb checkboxModel) {
	s.normalizeContentOptionsForDBType()
	for _, item := range cb.items {
		switch item.key {
		case "views":
			s.IncludeViews = item.selected
		case "procedures":
			s.IncludeProcedures = item.selected
		case "functions":
			s.IncludeFunctions = item.selected
		case "triggers":
			s.IncludeTriggers = item.selected
		case "sequences":
			s.IncludeSequences = item.selected
		}
	}
	s.normalizeContentOptionsForDBType()
}

// applyFormatCheckboxes 将格式选择应用到 TuiState。
func (s *TuiState) applyFormatCheckboxes(cb checkboxModel) {
	s.Formats = cb.selectedKeys()
}
