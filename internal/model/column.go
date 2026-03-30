package model

import "fmt"

// Column 字段元数据
type Column struct {
	Name         string // 字段名
	DataType     string // 数据类型
	Length       int    // 长度（适用于字符串类型）
	Precision    int    // 精度（适用于数值类型）
	Scale        int    // 小数位数（适用于数值类型）
	IsPrimaryKey bool   // 是否主键
	IsAutoIncrement bool // 是否自增
	IsNullable   bool   // 是否可空
	DefaultValue string // 默认值
	Comment      string // 字段注释
}

// GetFullDataType 获取完整数据类型（包含长度/精度）
func (c *Column) GetFullDataType() string {
	dt := c.DataType
	if c.Length > 0 {
		dt = fmt.Sprintf("%s(%d)", dt, c.Length)
	} else if c.Precision > 0 {
		if c.Scale > 0 {
			dt = fmt.Sprintf("%s(%d,%d)", dt, c.Precision, c.Scale)
		} else {
			dt = fmt.Sprintf("%s(%d)", dt, c.Precision)
		}
	}
	return dt
}

// IsNumeric 是否为数值类型
func (c *Column) IsNumeric() bool {
	numericTypes := []string{"INT", "INTEGER", "BIGINT", "SMALLINT", "TINYINT", 
		"DECIMAL", "NUMERIC", "FLOAT", "DOUBLE", "REAL", "NUMBER"}
	for _, t := range numericTypes {
		if c.DataType == t {
			return true
		}
	}
	return false
}

// IsString 是否为字符串类型
func (c *Column) IsString() bool {
	stringTypes := []string{"VARCHAR", "VARCHAR2", "CHAR", "NCHAR", "NVARCHAR", 
		"NVARCHAR2", "TEXT", "CLOB", "NCLOB"}
	for _, t := range stringTypes {
		if c.DataType == t {
			return true
		}
	}
	return false
}
