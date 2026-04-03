package model

type Procedure struct {
	Name       string // 存储过程名称
	Comment    string // 存储过程注释
	Definition string // 存储过程定义
}

type Function struct {
	Name       string // 函数名称
	Comment    string // 函数注释
	Definition string // 函数定义
	ReturnType string // 返回值类型
}
