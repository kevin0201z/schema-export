package model

// Procedure 表示数据库存储过程的元数据。
//
// 存储过程是一组预编译的 SQL 语句，存储在数据库中供重复使用。
// Procedure 结构体包含了存储过程的基本信息。
//
// 字段说明:
//   - Name: 存储过程名称
//   - Comment: 存储过程注释（如果数据库支持）
//   - Definition: 存储过程的完整定义（CREATE PROCEDURE 语句）
type Procedure struct {
	Name       string // 存储过程名称
	Comment    string // 存储过程注释
	Definition string // 存储过程定义
}

// Function 表示数据库函数的元数据。
//
// 函数与存储过程类似，但必须返回一个值。Function 结构体包含了函数的基本信息。
//
// 字段说明:
//   - Name: 函数名称
//   - Comment: 函数注释（如果数据库支持）
//   - Definition: 函数的完整定义（CREATE FUNCTION 语句）
//   - ReturnType: 函数返回值的数据类型
type Function struct {
	Name       string // 函数名称
	Comment    string // 函数注释
	Definition string // 函数定义
	ReturnType string // 返回值类型
}
