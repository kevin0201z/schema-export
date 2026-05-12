package sql

import (
	"fmt"
	"strings"

	"github.com/schema-export/schema-export/internal/model"
)

// SQLiteDialect 输出接近原始 SQLite DDL 的 SQL 方言。
type SQLiteDialect struct{}

func (d *SQLiteDialect) GetName() string {
	return "sqlite"
}

func (d *SQLiteDialect) QuoteIdentifier(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

func (d *SQLiteDialect) GetDataType(col *model.Column) string {
	return col.GetFullDataType()
}

func (d *SQLiteDialect) GetDefaultValue(col *model.Column) string {
	if strings.TrimSpace(col.DefaultValue) == "" {
		return ""
	}
	return "DEFAULT " + col.DefaultValue
}

func (d *SQLiteDialect) GetColumnDefinition(col *model.Column) string {
	parts := []string{fmt.Sprintf("%s %s", d.QuoteIdentifier(col.Name), d.GetDataType(col))}

	if col.IsPrimaryKey {
		parts = append(parts, "PRIMARY KEY")
	}
	if col.IsAutoIncrement {
		parts = append(parts, "AUTOINCREMENT")
	}
	if !col.IsNullable {
		parts = append(parts, "NOT NULL")
	}
	if def := d.GetDefaultValue(col); def != "" {
		parts = append(parts, def)
	}
	if col.CheckConstraint != "" {
		parts = append(parts, fmt.Sprintf("CHECK (%s)", trimWrappedParentheses(col.CheckConstraint)))
	}

	return strings.Join(parts, " ")
}

func (d *SQLiteDialect) GetCheckConstraint(cc *model.CheckConstraint) string {
	return fmt.Sprintf("CONSTRAINT %s CHECK (%s)", d.QuoteIdentifier(cc.Name), trimWrappedParentheses(cc.Definition))
}

func (d *SQLiteDialect) GetColumnCommentSQL(string, *model.Column) string {
	return ""
}

func (d *SQLiteDialect) GetTableCommentSQL(string, string) string {
	return ""
}

func (d *SQLiteDialect) GetViewCommentSQL(string, string) string {
	return ""
}

func (d *SQLiteDialect) SupportsInlineComment() bool {
	return false
}

func (d *SQLiteDialect) SupportsInlineCheck() bool {
	return true
}

func (d *SQLiteDialect) EmitPrimaryKeyInline() bool {
	return false
}

func trimWrappedParentheses(s string) string {
	trimmed := strings.TrimSpace(s)
	if strings.HasPrefix(trimmed, "(") && strings.HasSuffix(trimmed, ")") {
		return strings.TrimSpace(trimmed[1 : len(trimmed)-1])
	}
	return trimmed
}
