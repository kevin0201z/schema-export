package model

import "fmt"

// ForeignKey 外键元数据
type ForeignKey struct {
	Name           string // 外键名
	Column         string // 源字段
	RefTable       string // 目标表
	RefColumn      string // 目标字段
	OnDelete       string // 删除级联规则
	OnUpdate       string // 更新级联规则
}

// GetReferenceString 获取引用描述字符串
func (fk *ForeignKey) GetReferenceString() string {
	return fmt.Sprintf("%s(%s)", fk.RefTable, fk.RefColumn)
}

// GetOnDeleteRule 获取删除规则描述
func (fk *ForeignKey) GetOnDeleteRule() string {
	if fk.OnDelete == "" {
		return "NO ACTION"
	}
	return fk.OnDelete
}

// GetOnUpdateRule 获取更新规则描述
func (fk *ForeignKey) GetOnUpdateRule() string {
	if fk.OnUpdate == "" {
		return "NO ACTION"
	}
	return fk.OnUpdate
}
