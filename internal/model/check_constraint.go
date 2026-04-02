package model

// CheckConstraint CHECK 约束
type CheckConstraint struct {
	Name       string   // 约束名
	Definition string   // 约束定义表达式
	Columns    []string // 涉及的字段列表
}
