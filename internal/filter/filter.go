package filter

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/schema-export/schema-export/internal/model"
)

// TableFilter 表过滤器。
//
// TableFilter 实现了表名的多规则过滤功能。支持三种过滤方式：
//   - Include（白名单）：只包含列表中的表
//   - Exclude（黑名单）：排除列表中的表
//   - Patterns（正则）：表名必须匹配至少一个正则表达式
//
// 过滤逻辑：
//  1. 如果设置了 Include，表必须在 Include 列表中
//  2. 如果表在 Exclude 列表中，则排除
//  3. 如果设置了 Patterns 且未设置 Include，表名必须匹配至少一个正则
//
// 字段说明:
//   - Include: 白名单表名列表（大小写不敏感）
//   - Exclude: 黑名单表名列表（大小写不敏感）
//   - Patterns: 正则表达式模式列表
type TableFilter struct {
	Include  []string         // 白名单表名列表
	Exclude  []string         // 黑名单表名列表
	Patterns []*regexp.Regexp // 正则表达式模式列表
}

// NewTableFilter 创建表过滤器实例。
//
// 参数:
//   - include: 白名单表名列表（空表示不限制）
//   - exclude: 黑名单表名列表（空表示不排除）
//   - patterns: 正则表达式模式字符串列表（空表示不使用正则过滤）
//
// 返回值:
//   - *TableFilter: 新创建的过滤器实例
//   - error: 正则表达式编译失败时返回错误
//
// 示例:
//
//	filter, err := NewTableFilter(
//	    []string{"users", "orders"},
//	    []string{"temp_logs"},
//	    []string{"^app_.*"},
//	)
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

// ShouldInclude 判断表是否应该被包含在导出结果中。
//
// 该方法按照以下顺序应用过滤规则：
//  1. 白名单检查：如果设置了 Include，表名必须在列表中（大小写不敏感）
//  2. 黑名单检查：如果表名在 Exclude 列表中，则排除（大小写不敏感）
//  3. 正则匹配：如果设置了 Patterns 且未设置 Include，表名必须匹配至少一个正则
//
// 参数:
//   - tableName: 要检查的表名
//
// 返回值:
//   - bool: true 表示应该包含，false 表示应该排除
func (f *TableFilter) ShouldInclude(tableName string) bool {
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

	for _, name := range f.Exclude {
		if strings.EqualFold(name, tableName) {
			return false
		}
	}

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

// FilterTables 过滤表列表。
//
// 遍历表列表，使用 ShouldInclude 方法判断每个表是否应该包含，
// 返回过滤后的表列表。
//
// 参数:
//   - tables: 原始表列表
//
// 返回值:
//   - []model.Table: 过滤后的表列表
func (f *TableFilter) FilterTables(tables []model.Table) []model.Table {
	var filtered []model.Table
	for _, table := range tables {
		if f.ShouldInclude(table.Name) {
			filtered = append(filtered, table)
		}
	}
	return filtered
}
