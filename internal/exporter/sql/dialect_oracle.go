package sql

import (
	"fmt"
	"strings"

	"github.com/schema-export/schema-export/internal/model"
)

// OracleDialect Oracle/达梦 方言
type OracleDialect struct{}

func (d *OracleDialect) GetName() string {
	return "oracle"
}

func (d *OracleDialect) QuoteIdentifier(name string) string {
	return "\"" + name + "\""
}

func (d *OracleDialect) GetDataType(col *model.Column) string {
	dataType := strings.ToUpper(col.DataType)

	switch dataType {
	case "VARCHAR2", "NVARCHAR2":
		if col.Length > 0 {
			return fmt.Sprintf("%s(%d)", dataType, col.Length)
		}
		return dataType
	case "CHAR", "NCHAR":
		if col.Length > 0 {
			return fmt.Sprintf("%s(%d)", dataType, col.Length)
		}
		return dataType
	case "NUMBER":
		if col.Precision > 0 && col.Scale > 0 {
			return fmt.Sprintf("NUMBER(%d,%d)", col.Precision, col.Scale)
		} else if col.Precision > 0 {
			return fmt.Sprintf("NUMBER(%d)", col.Precision)
		}
		return "NUMBER"
	case "DECIMAL", "NUMERIC":
		if col.Precision > 0 && col.Scale > 0 {
			return fmt.Sprintf("%s(%d,%d)", dataType, col.Precision, col.Scale)
		} else if col.Precision > 0 {
			return fmt.Sprintf("%s(%d)", dataType, col.Precision)
		}
		return dataType
	case "TIMESTAMP":
		if col.Scale > 0 {
			return fmt.Sprintf("TIMESTAMP(%d)", col.Scale)
		}
		return "TIMESTAMP"
	case "CLOB", "BLOB", "NCLOB":
		return dataType
	default:
		return dataType
	}
}

func (d *OracleDialect) GetDefaultValue(col *model.Column) string {
	if col.DefaultValue == "" {
		return ""
	}

	defaultVal := strings.TrimSpace(col.DefaultValue)

	// 处理函数默认值
	upperVal := strings.ToUpper(defaultVal)
	switch {
	case strings.Contains(upperVal, "SYSDATE"):
		return "DEFAULT SYSDATE"
	case strings.Contains(upperVal, "CURRENT_TIMESTAMP"):
		return "DEFAULT CURRENT_TIMESTAMP"
	case strings.Contains(upperVal, "SYS_GUID"):
		return "DEFAULT SYS_GUID()"
	}

	// 字符串值需要加引号
	if !strings.HasPrefix(defaultVal, "'") && !strings.HasPrefix(defaultVal, "\"") {
		if _, err := fmt.Sscanf(defaultVal, "%f", new(float64)); err != nil {
			if _, err := fmt.Sscanf(defaultVal, "%d", new(int)); err != nil {
				defaultVal = "'" + defaultVal + "'"
			}
		}
	}

	return "DEFAULT " + defaultVal
}

func (d *OracleDialect) GetColumnDefinition(col *model.Column) string {
	var parts []string

	parts = append(parts, d.QuoteIdentifier(col.Name))
	parts = append(parts, d.GetDataType(col))

	if !col.IsNullable {
		parts = append(parts, "NOT NULL")
	}

	if col.DefaultValue != "" {
		parts = append(parts, d.GetDefaultValue(col))
	}

	if col.IsPrimaryKey {
		parts = append(parts, "PRIMARY KEY")
	}

	return strings.Join(parts, " ")
}

func (d *OracleDialect) GetCheckConstraint(cc *model.CheckConstraint) string {
	return fmt.Sprintf("CONSTRAINT %s CHECK (%s)", d.QuoteIdentifier(cc.Name), cc.Definition)
}

func (d *OracleDialect) GetColumnCommentSQL(tableName string, col *model.Column) string {
	if col.Comment == "" {
		return ""
	}
	return fmt.Sprintf("COMMENT ON COLUMN %s.%s IS '%s';",
		d.QuoteIdentifier(tableName),
		d.QuoteIdentifier(col.Name),
		strings.ReplaceAll(col.Comment, "'", "''"))
}

func (d *OracleDialect) GetTableCommentSQL(tableName string, comment string) string {
	if comment == "" {
		return ""
	}
	return fmt.Sprintf("COMMENT ON TABLE %s IS '%s';",
		d.QuoteIdentifier(tableName),
		strings.ReplaceAll(comment, "'", "''"))
}

func (d *OracleDialect) GetViewCommentSQL(viewName string, comment string) string {
	if comment == "" {
		return ""
	}
	return fmt.Sprintf("COMMENT ON TABLE %s IS '%s';",
		d.QuoteIdentifier(viewName),
		strings.ReplaceAll(comment, "'", "''"))
}

func (d *OracleDialect) SupportsInlineComment() bool {
	return false
}

func (d *OracleDialect) SupportsInlineCheck() bool {
	return true
}
