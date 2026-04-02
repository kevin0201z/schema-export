package sql

import (
	"fmt"
	"strings"

	"github.com/schema-export/schema-export/internal/model"
)

// GenericDialect 通用方言（默认）
type GenericDialect struct{}

func (d *GenericDialect) GetName() string {
	return "generic"
}

func (d *GenericDialect) QuoteIdentifier(name string) string {
	return "\"" + name + "\""
}

func (d *GenericDialect) GetDataType(col *model.Column) string {
	dataType := strings.ToUpper(col.DataType)

	switch dataType {
	case "VARCHAR", "VARCHAR2", "NVARCHAR", "NVARCHAR2":
		if col.Length > 0 {
			return fmt.Sprintf("VARCHAR(%d)", col.Length)
		}
		return "VARCHAR"
	case "CHAR", "NCHAR":
		if col.Length > 0 {
			return fmt.Sprintf("CHAR(%d)", col.Length)
		}
		return "CHAR"
	case "DECIMAL", "NUMERIC", "NUMBER":
		if col.Precision > 0 && col.Scale > 0 {
			return fmt.Sprintf("DECIMAL(%d,%d)", col.Precision, col.Scale)
		} else if col.Precision > 0 {
			return fmt.Sprintf("DECIMAL(%d)", col.Precision)
		}
		return "DECIMAL"
	case "TIMESTAMP", "DATETIME":
		return "TIMESTAMP"
	default:
		return dataType
	}
}

func (d *GenericDialect) GetDefaultValue(col *model.Column) string {
	if col.DefaultValue == "" {
		return ""
	}
	return "DEFAULT " + strings.TrimSpace(col.DefaultValue)
}

func (d *GenericDialect) GetColumnDefinition(col *model.Column) string {
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

func (d *GenericDialect) GetCheckConstraint(cc *model.CheckConstraint) string {
	return fmt.Sprintf("CONSTRAINT %s CHECK (%s)", d.QuoteIdentifier(cc.Name), cc.Definition)
}

func (d *GenericDialect) GetColumnCommentSQL(tableName string, col *model.Column) string {
	if col.Comment == "" {
		return ""
	}
	return fmt.Sprintf("COMMENT ON COLUMN %s.%s IS '%s';",
		d.QuoteIdentifier(tableName),
		d.QuoteIdentifier(col.Name),
		strings.ReplaceAll(col.Comment, "'", "''"))
}

func (d *GenericDialect) GetTableCommentSQL(tableName string, comment string) string {
	if comment == "" {
		return ""
	}
	return fmt.Sprintf("COMMENT ON TABLE %s IS '%s';",
		d.QuoteIdentifier(tableName),
		strings.ReplaceAll(comment, "'", "''"))
}

func (d *GenericDialect) GetViewCommentSQL(viewName string, comment string) string {
	if comment == "" {
		return ""
	}
	return fmt.Sprintf("COMMENT ON VIEW %s IS '%s';",
		d.QuoteIdentifier(viewName),
		strings.ReplaceAll(comment, "'", "''"))
}

func (d *GenericDialect) SupportsInlineComment() bool {
	return false
}

func (d *GenericDialect) SupportsInlineCheck() bool {
	return true
}
