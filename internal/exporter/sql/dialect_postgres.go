package sql

import (
	"fmt"
	"strings"

	"github.com/schema-export/schema-export/internal/model"
)

// PostgreSQLDialect PostgreSQL 方言
type PostgreSQLDialect struct{}

func (d *PostgreSQLDialect) GetName() string {
	return "postgres"
}

func (d *PostgreSQLDialect) QuoteIdentifier(name string) string {
	return "\"" + name + "\""
}

func (d *PostgreSQLDialect) GetDataType(col *model.Column) string {
	dataType := strings.ToUpper(col.DataType)

	switch dataType {
	case "CHARACTER VARYING", "VARCHAR", "VARCHAR2":
		if col.Length > 0 {
			return fmt.Sprintf("VARCHAR(%d)", col.Length)
		}
		return "VARCHAR"
	case "CHARACTER", "CHAR", "NCHAR":
		if col.Length > 0 {
			return fmt.Sprintf("CHAR(%d)", col.Length)
		}
		return "CHAR"
	case "NUMERIC", "DECIMAL", "NUMBER":
		if col.Precision > 0 && col.Scale > 0 {
			return fmt.Sprintf("NUMERIC(%d,%d)", col.Precision, col.Scale)
		} else if col.Precision > 0 {
			return fmt.Sprintf("NUMERIC(%d)", col.Precision)
		}
		return "NUMERIC"
	case "INTEGER", "INT4", "INT":
		return "INTEGER"
	case "BIGINT", "INT8":
		return "BIGINT"
	case "SMALLINT", "INT2":
		return "SMALLINT"
	case "REAL", "FLOAT4":
		return "REAL"
	case "DOUBLE PRECISION", "FLOAT8", "DOUBLE":
		return "DOUBLE PRECISION"
	case "BOOLEAN", "BOOL":
		return "BOOLEAN"
	case "TEXT", "CLOB", "NCLOB":
		return "TEXT"
	case "BYTEA", "BLOB", "RAW", "LONG RAW":
		return "BYTEA"
	case "TIMESTAMP", "DATETIME":
		if strings.Contains(dataType, "WITH TIME ZONE") || strings.Contains(dataType, "TIMEZONE") {
			return "TIMESTAMPTZ"
		}
		return "TIMESTAMP"
	case "DATE":
		return "DATE"
	case "TIME":
		if strings.Contains(dataType, "WITH TIME ZONE") || strings.Contains(dataType, "TIMEZONE") {
			return "TIMETZ"
		}
		return "TIME"
	case "UUID":
		return "UUID"
	case "JSONB", "JSON":
		return "JSONB"
	case "ARRAY":
		return "ARRAY"
	default:
		if col.Length > 0 {
			return fmt.Sprintf("%s(%d)", dataType, col.Length)
		}
		return dataType
	}
}

func (d *PostgreSQLDialect) GetDefaultValue(col *model.Column) string {
	if col.DefaultValue == "" {
		return ""
	}

	defaultVal := strings.TrimSpace(col.DefaultValue)

	upperVal := strings.ToUpper(defaultVal)
	switch {
	case strings.HasPrefix(upperVal, "NEXTVAL"):
		return defaultVal
	case upperVal == "NOW()", upperVal == "CURRENT_TIMESTAMP", upperVal == "CURRENT_DATE", upperVal == "CURRENT_TIME":
		return "DEFAULT " + defaultVal
	case strings.Contains(upperVal, "GEN_RANDOM_UUID"), strings.Contains(upperVal, "UUID"):
		return "DEFAULT gen_random_uuid()"
	case strings.Contains(upperVal, "TRUE"), strings.Contains(upperVal, "FALSE"):
		return "DEFAULT " + defaultVal
	}

	if !strings.HasPrefix(defaultVal, "'") && !strings.HasPrefix(defaultVal, "\"") {
		if _, err := fmt.Sscanf(defaultVal, "%f", new(float64)); err != nil {
			if _, err := fmt.Sscanf(defaultVal, "%d", new(int)); err != nil {
				defaultVal = "'" + strings.ReplaceAll(defaultVal, "'", "''") + "'"
			}
		}
	}

	return "DEFAULT " + defaultVal
}

func (d *PostgreSQLDialect) GetColumnDefinition(col *model.Column) string {
	var parts []string

	parts = append(parts, d.QuoteIdentifier(col.Name))
	parts = append(parts, d.GetDataType(col))

	if !col.IsNullable {
		parts = append(parts, "NOT NULL")
	}

	if col.IsAutoIncrement {
		parts = append(parts, "GENERATED ALWAYS AS IDENTITY")
	} else if col.DefaultValue != "" {
		parts = append(parts, d.GetDefaultValue(col))
	}

	return strings.Join(parts, " ")
}

func (d *PostgreSQLDialect) GetCheckConstraint(cc *model.CheckConstraint) string {
	definition := cc.Definition
	if strings.HasPrefix(definition, "(") && strings.HasSuffix(definition, ")") {
		definition = definition[1 : len(definition)-1]
	}
	return fmt.Sprintf("CONSTRAINT %s CHECK (%s)", d.QuoteIdentifier(cc.Name), definition)
}

func (d *PostgreSQLDialect) GetColumnCommentSQL(tableName string, col *model.Column) string {
	if col.Comment == "" {
		return ""
	}
	escapedComment := strings.ReplaceAll(col.Comment, "'", "''")
	return fmt.Sprintf("COMMENT ON COLUMN %s.%s IS '%s';",
		d.QuoteIdentifier(tableName),
		d.QuoteIdentifier(col.Name),
		escapedComment,
	)
}

func (d *PostgreSQLDialect) GetTableCommentSQL(tableName string, comment string) string {
	if comment == "" {
		return ""
	}
	escapedComment := strings.ReplaceAll(comment, "'", "''")
	return fmt.Sprintf("COMMENT ON TABLE %s IS '%s';", d.QuoteIdentifier(tableName), escapedComment)
}

func (d *PostgreSQLDialect) GetViewCommentSQL(viewName string, comment string) string {
	if comment == "" {
		return ""
	}
	escapedComment := strings.ReplaceAll(comment, "'", "''")
	return fmt.Sprintf("COMMENT ON VIEW %s IS '%s';", d.QuoteIdentifier(viewName), escapedComment)
}

func (d *PostgreSQLDialect) SupportsInlineComment() bool {
	return false
}

func (d *PostgreSQLDialect) SupportsInlineCheck() bool {
	return true
}
