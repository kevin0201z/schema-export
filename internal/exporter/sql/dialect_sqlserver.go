package sql

import (
	"fmt"
	"strings"

	"github.com/schema-export/schema-export/internal/model"
)

// SQLServerDialect SQL Server 方言
type SQLServerDialect struct{}

func (d *SQLServerDialect) GetName() string {
	return "sqlserver"
}

func (d *SQLServerDialect) QuoteIdentifier(name string) string {
	return "[" + name + "]"
}

func (d *SQLServerDialect) GetDataType(col *model.Column) string {
	dataType := strings.ToUpper(col.DataType)

	switch dataType {
	case "VARCHAR", "NVARCHAR", "CHAR", "NCHAR":
		if col.Length == -1 {
			return dataType + "(MAX)"
		}
		if col.Length > 0 {
			return fmt.Sprintf("%s(%d)", dataType, col.Length)
		}
		return dataType
	case "DECIMAL", "NUMERIC":
		if col.Precision > 0 && col.Scale > 0 {
			return fmt.Sprintf("%s(%d,%d)", dataType, col.Precision, col.Scale)
		} else if col.Precision > 0 {
			return fmt.Sprintf("%s(%d)", dataType, col.Precision)
		}
		return dataType
	case "DATETIME", "DATETIME2", "DATETIMEOFFSET":
		if col.Scale > 0 {
			return fmt.Sprintf("%s(%d)", dataType, col.Scale)
		}
		return dataType
	case "TIME":
		if col.Scale > 0 {
			return fmt.Sprintf("TIME(%d)", col.Scale)
		}
		return "TIME"
	case "VARBINARY", "BINARY":
		if col.Length == -1 {
			return "VARBINARY(MAX)"
		}
		if col.Length > 0 {
			return fmt.Sprintf("%s(%d)", dataType, col.Length)
		}
		return dataType
	case "FLOAT":
		if col.Precision > 0 {
			return fmt.Sprintf("FLOAT(%d)", col.Precision)
		}
		return "FLOAT"
	case "TEXT", "NTEXT", "IMAGE":
		return dataType
	default:
		return dataType
	}
}

func (d *SQLServerDialect) GetDefaultValue(col *model.Column) string {
	if col.DefaultValue == "" {
		return ""
	}

	defaultVal := strings.TrimSpace(col.DefaultValue)

	// 处理函数默认值
	upperVal := strings.ToUpper(defaultVal)
	switch {
	case strings.Contains(upperVal, "GETDATE"):
		return "DEFAULT GETDATE()"
	case strings.Contains(upperVal, "NEWID"):
		return "DEFAULT NEWID()"
	case strings.Contains(upperVal, "CURRENT_TIMESTAMP"):
		return "DEFAULT CURRENT_TIMESTAMP"
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

func (d *SQLServerDialect) GetColumnDefinition(col *model.Column) string {
	var parts []string

	parts = append(parts, d.QuoteIdentifier(col.Name))
	parts = append(parts, d.GetDataType(col))

	if col.IsAutoIncrement {
		parts = append(parts, "IDENTITY(1,1)")
	}

	if !col.IsNullable {
		parts = append(parts, "NOT NULL")
	}

	if col.DefaultValue != "" && !col.IsAutoIncrement {
		parts = append(parts, d.GetDefaultValue(col))
	}

	if col.IsPrimaryKey {
		parts = append(parts, "PRIMARY KEY")
	}

	return strings.Join(parts, " ")
}

func (d *SQLServerDialect) GetCheckConstraint(cc *model.CheckConstraint) string {
	return fmt.Sprintf("CONSTRAINT [%s] CHECK (%s)", cc.Name, cc.Definition)
}

func (d *SQLServerDialect) GetColumnCommentSQL(tableName string, col *model.Column) string {
	if col.Comment == "" {
		return ""
	}
	escapedComment := strings.ReplaceAll(col.Comment, "'", "''")
	return fmt.Sprintf("EXEC sp_addextendedproperty 'MS_Description', N'%s', 'SCHEMA', 'dbo', 'TABLE', '%s', 'COLUMN', '%s';",
		escapedComment, tableName, col.Name)
}

func (d *SQLServerDialect) GetTableCommentSQL(tableName string, comment string) string {
	if comment == "" {
		return ""
	}
	escapedComment := strings.ReplaceAll(comment, "'", "''")
	return fmt.Sprintf("EXEC sp_addextendedproperty 'MS_Description', N'%s', 'SCHEMA', 'dbo', 'TABLE', '%s';",
		escapedComment, tableName)
}

func (d *SQLServerDialect) GetViewCommentSQL(viewName string, comment string) string {
	if comment == "" {
		return ""
	}
	escapedComment := strings.ReplaceAll(comment, "'", "''")
	return fmt.Sprintf("EXEC sp_addextendedproperty 'MS_Description', N'%s', 'SCHEMA', 'dbo', 'VIEW', '%s';",
		escapedComment, viewName)
}

func (d *SQLServerDialect) SupportsInlineComment() bool {
	return false
}

func (d *SQLServerDialect) SupportsInlineCheck() bool {
	return true
}
