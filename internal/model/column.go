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

// 预定义的类型集合，使用 map 提高查找效率
var (
	numericTypes = map[string]bool{
		"INT":      true,
		"INTEGER":  true,
		"BIGINT":   true,
		"SMALLINT": true,
		"TINYINT":  true,
		"DECIMAL":  true,
		"NUMERIC":  true,
		"FLOAT":    true,
		"DOUBLE":   true,
		"REAL":     true,
		"NUMBER":   true,
	}

	stringTypes = map[string]bool{
		"VARCHAR":    true,
		"VARCHAR2":   true,
		"CHAR":       true,
		"NCHAR":      true,
		"NVARCHAR":   true,
		"NVARCHAR2":  true,
		"TEXT":       true,
		"CLOB":       true,
		"NCLOB":      true,
	}
)

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
	return numericTypes[c.DataType]
}

// IsString 是否为字符串类型
func (c *Column) IsString() bool {
	return stringTypes[c.DataType]
}
