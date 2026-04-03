package model

type Sequence struct {
	Name        string // 序列名称
	MinValue    int64  // 最小值
	MaxValue    int64  // 最大值
	IncrementBy int64  // 增量
	Cycle       bool   // 是否循环
	CacheSize   int64  // 缓存大小
	LastValue   int64  // 当前值
}
