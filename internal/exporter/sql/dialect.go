package sql

import "github.com/schema-export/schema-export/internal/model"

// Dialect 数据库方言接口
type Dialect interface {
	// GetName 获取方言名称
	GetName() string

	// QuoteIdentifier 引用标识符（表名、字段名）
	QuoteIdentifier(name string) string

	// GetDataType 获取数据类型定义
	GetDataType(col *model.Column) string

	// GetDefaultValue 获取默认值表达式
	GetDefaultValue(col *model.Column) string

	// GetColumnDefinition 获取完整的字段定义
	GetColumnDefinition(col *model.Column) string

	// GetCheckConstraint 获取 CHECK 约束语句
	GetCheckConstraint(cc *model.CheckConstraint) string

	// GetColumnCommentSQL 获取字段注释 SQL 语句
	GetColumnCommentSQL(tableName string, col *model.Column) string

	// GetTableCommentSQL 获取表注释 SQL 语句
	GetTableCommentSQL(tableName string, comment string) string

	// GetViewCommentSQL 获取视图注释 SQL 语句
	GetViewCommentSQL(viewName string, comment string) string

	// SupportsInlineComment 是否支持内联注释（MySQL 支持在字段定义中写注释）
	SupportsInlineComment() bool

	// SupportsInlineCheck 是否支持内联 CHECK 约束
	SupportsInlineCheck() bool
}
