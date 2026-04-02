package model

type View struct {
	Name       string   // 视图名称
	Comment    string   // 视图注释
	Definition string   // 视图定义 SQL
	Columns    []Column // 视图字段列表
}
