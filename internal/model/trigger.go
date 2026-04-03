package model

type Trigger struct {
	Name        string // 触发器名称
	TableName   string // 所属表名
	Event       string // 触发事件（INSERT, UPDATE, DELETE）
	Timing      string // 触发时机（BEFORE, AFTER, INSTEAD OF）
	Definition  string // 触发器定义
	Status      string // 状态（ENABLED, DISABLED）
}
