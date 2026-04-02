package model

import "strings"

// CheckConstraint CHECK 约束
type CheckConstraint struct {
	Name       string   // 约束名
	Definition string   // 约束定义表达式
	Columns    []string // 涉及的字段列表
}

// GetColumnsString 获取字段列表的字符串表示
func (c *CheckConstraint) GetColumnsString() string {
	return strings.Join(c.Columns, ", ")
}
