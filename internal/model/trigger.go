package model

// Trigger 表示数据库触发器的元数据。
//
// 触发器是一种特殊的存储过程，它在特定的表上执行特定的数据操作时自动触发。
// Trigger 结构体包含了触发器的基本信息。
//
// 字段说明:
//   - Name: 触发器名称
//   - TableName: 触发器所属的表名
//   - Event: 触发事件类型（INSERT, UPDATE, DELETE 或其组合）
//   - Timing: 触发时机（BEFORE, AFTER, INSTEAD OF）
//   - Definition: 触发器的完整定义（CREATE TRIGGER 语句）
//   - Status: 触发器状态（ENABLED, DISABLED）
type Trigger struct {
	Name       string // 触发器名称
	TableName  string // 所属表名
	Event      string // 触发事件（INSERT, UPDATE, DELETE）
	Timing     string // 触发时机（BEFORE, AFTER, INSTEAD OF）
	Definition string // 触发器定义
	Status     string // 状态（ENABLED, DISABLED）
}
