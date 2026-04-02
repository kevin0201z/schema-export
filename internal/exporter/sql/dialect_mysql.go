package sql

import (
	"fmt"
	"strings"

	"github.com/schema-export/schema-export/internal/model"
)

// MySQLDialect MySQL 方言
type MySQLDialect struct{}

func (d *MySQLDialect) GetName() string {
	return "mysql"
}

func (d *MySQLDialect) QuoteIdentifier(name string) string {
	return "`" + name + "`"
}

// typeMapping Oracle/DM 类型到 MySQL 类型的映射
var oracleToMySQLTypeMap = map[string]string{
	"VARCHAR2":      "VARCHAR",
	"NVARCHAR2":     "VARCHAR",
	"CHAR":          "CHAR",
	"NCHAR":         "CHAR",
	"NUMBER":        "DECIMAL",
	"DECIMAL":       "DECIMAL",
	"NUMERIC":       "DECIMAL",
	"INTEGER":       "INT",
	"SMALLINT":      "SMALLINT",
	"BIGINT":        "BIGINT",
	"FLOAT":         "DOUBLE",
	"BINARY_FLOAT":  "FLOAT",
	"BINARY_DOUBLE": "DOUBLE",
	"TIMESTAMP":     "DATETIME",
	"DATE":          "DATE",
	"CLOB":          "LONGTEXT",
	"NCLOB":         "LONGTEXT",
	"BLOB":          "LONGBLOB",
	"RAW":           "VARBINARY",
	"LONG RAW":      "LONGBLOB",
}

func (d *MySQLDialect) GetDataType(col *model.Column) string {
	oracleType := strings.ToUpper(col.DataType)
	mysqlType, ok := oracleToMySQLTypeMap[oracleType]
	if !ok {
		mysqlType = oracleType
	}

	switch mysqlType {
	case "VARCHAR", "CHAR":
		if col.Length > 0 {
			return fmt.Sprintf("%s(%d)", mysqlType, col.Length)
		}
		return mysqlType
	case "DECIMAL", "DOUBLE":
		if col.Precision > 0 && col.Scale > 0 {
			return fmt.Sprintf("DECIMAL(%d,%d)", col.Precision, col.Scale)
		} else if col.Precision > 0 {
			return fmt.Sprintf("DECIMAL(%d)", col.Precision)
		}
		return "DECIMAL"
	case "INT", "SMALLINT", "BIGINT":
		return mysqlType
	case "DATETIME", "TIMESTAMP":
		return "DATETIME"
	case "VARBINARY":
		if col.Length > 0 {
			return fmt.Sprintf("VARBINARY(%d)", col.Length)
		}
		return "VARBINARY"
	default:
		return mysqlType
	}
}

func (d *MySQLDialect) GetDefaultValue(col *model.Column) string {
	if col.DefaultValue == "" {
		return ""
	}

	defaultVal := strings.TrimSpace(col.DefaultValue)

	// 处理函数默认值
	upperVal := strings.ToUpper(defaultVal)
	switch {
	case strings.Contains(upperVal, "SYSDATE"), strings.Contains(upperVal, "CURRENT_TIMESTAMP"):
		return "DEFAULT CURRENT_TIMESTAMP"
	case strings.Contains(upperVal, "SYS_GUID"), strings.Contains(upperVal, "UUID"):
		return "DEFAULT (UUID())"
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

func (d *MySQLDialect) GetColumnDefinition(col *model.Column) string {
	var parts []string

	parts = append(parts, d.QuoteIdentifier(col.Name))
	parts = append(parts, d.GetDataType(col))

	if col.IsAutoIncrement {
		parts = append(parts, "AUTO_INCREMENT")
	}

	if !col.IsNullable {
		parts = append(parts, "NOT NULL")
	} else {
		parts = append(parts, "NULL")
	}

	if col.DefaultValue != "" && !col.IsAutoIncrement {
		parts = append(parts, d.GetDefaultValue(col))
	}

	if col.Comment != "" {
		escapedComment := strings.ReplaceAll(col.Comment, "'", "\\'")
		parts = append(parts, fmt.Sprintf("COMMENT '%s'", escapedComment))
	}

	if col.IsPrimaryKey {
		parts = append(parts, "PRIMARY KEY")
	}

	return strings.Join(parts, " ")
}

func (d *MySQLDialect) GetCheckConstraint(cc *model.CheckConstraint) string {
	return fmt.Sprintf("CONSTRAINT `%s` CHECK (%s)", cc.Name, d.convertCheckExpression(cc.Definition))
}

// convertCheckExpression 转换 CHECK 表达式中的 Oracle 语法到 MySQL
func (d *MySQLDialect) convertCheckExpression(expr string) string {
	// 转换常见的 Oracle 函数到 MySQL
	expr = strings.ReplaceAll(strings.ToUpper(expr), "SYSDATE", "NOW()")
	return expr
}

func (d *MySQLDialect) GetColumnCommentSQL(tableName string, col *model.Column) string {
	// MySQL 使用内联注释，不需要单独的语句
	return ""
}

func (d *MySQLDialect) GetTableCommentSQL(tableName string, comment string) string {
	if comment == "" {
		return ""
	}
	escapedComment := strings.ReplaceAll(comment, "'", "\\'")
	return fmt.Sprintf("ALTER TABLE %s COMMENT = '%s';", d.QuoteIdentifier(tableName), escapedComment)
}

func (d *MySQLDialect) SupportsInlineComment() bool {
	return true
}

func (d *MySQLDialect) SupportsInlineCheck() bool {
	return true
}
