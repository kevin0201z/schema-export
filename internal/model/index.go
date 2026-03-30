package model

import "strings"

// IndexType 索引类型
type IndexType string

const (
	IndexTypeNormal  IndexType = "NORMAL"  // 普通索引
	IndexTypeUnique  IndexType = "UNIQUE"  // 唯一索引
	IndexTypePrimary IndexType = "PRIMARY" // 主键索引
)

// Index 索引元数据
type Index struct {
	Name       string    // 索引名
	Type       IndexType // 索引类型
	Columns    []string  // 索引字段列表
	IsPrimary  bool      // 是否主键索引
	IsUnique   bool      // 是否唯一索引
}

// GetColumnsString 获取字段列表字符串
func (i *Index) GetColumnsString() string {
	return strings.Join(i.Columns, ", ")
}
