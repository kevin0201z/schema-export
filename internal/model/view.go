package model

// View 表示数据库视图的元数据。
//
// 视图是一个虚拟表，其内容由查询定义。View 结构体包含了视图的基本信息，
// 包括名称、注释、定义 SQL 和字段列表。
//
// 字段说明:
//   - Name: 视图名称
//   - Comment: 视图注释（如果数据库支持）
//   - Definition: 视图的 SELECT 语句定义
//   - Columns: 视图返回的字段列表
type View struct {
	Name       string   // 视图名称
	Comment    string   // 视图注释
	Definition string   // 视图定义 SQL
	Columns    []Column // 视图字段列表
}
