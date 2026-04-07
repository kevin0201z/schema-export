package model

// Sequence 表示数据库序列的元数据。
//
// 序列是一种数据库对象，用于生成唯一的数值序列，通常用于自增主键。
// Sequence 结构体包含了序列的基本配置信息。
//
// 字段说明:
//   - Name: 序列名称
//   - MinValue: 序列的最小值
//   - MaxValue: 序列的最大值
//   - IncrementBy: 序列的增量值（正数递增，负数递减）
//   - Cycle: 是否在达到边界值后循环
//   - CacheSize: 缓存大小（用于提高性能）
//   - LastValue: 序列的当前值
//
// 注意: MySQL 不支持序列对象，该类型主要用于 Oracle、PostgreSQL 等数据库。
type Sequence struct {
	Name        string // 序列名称
	MinValue    int64  // 最小值
	MaxValue    int64  // 最大值
	IncrementBy int64  // 增量
	Cycle       bool   // 是否循环
	CacheSize   int64  // 缓存大小
	LastValue   int64  // 当前值
}
