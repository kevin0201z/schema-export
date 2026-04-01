package filter

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/schema-export/schema-export/internal/model"
)

// TableFilter 表过滤器
type TableFilter struct {
	Include  []string
	Exclude  []string
	Patterns []*regexp.Regexp
}

// NewTableFilter 创建表过滤器
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

// ShouldInclude 判断表是否应该被包含
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

// FilterTables 过滤表列表
func (f *TableFilter) FilterTables(tables []model.Table) []model.Table {
	var filtered []model.Table
	for _, table := range tables {
		if f.ShouldInclude(table.Name) {
			filtered = append(filtered, table)
		}
	}
	return filtered
}
