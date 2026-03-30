package inspector

import (
	"context"
	"github.com/schema-export/schema-export/internal/model"
)

// Inspector 数据库元数据读取接口
type Inspector interface {
	// Connect 连接数据库
	Connect(ctx context.Context) error
	
	// Close 关闭连接
	Close() error
	
	// TestConnection 测试连接
	TestConnection(ctx context.Context) error
	
	// GetTables 获取所有表列表
	GetTables(ctx context.Context) ([]model.Table, error)
	
	// GetTable 获取单个表的完整元数据
	GetTable(ctx context.Context, tableName string) (*model.Table, error)
	
	// GetColumns 获取表字段列表
	GetColumns(ctx context.Context, tableName string) ([]model.Column, error)
	
	// GetIndexes 获取表索引列表
	GetIndexes(ctx context.Context, tableName string) ([]model.Index, error)
	
	// GetForeignKeys 获取表外键列表
	GetForeignKeys(ctx context.Context, tableName string) ([]model.ForeignKey, error)
}

// InspectorFactory Inspector 工厂接口
type InspectorFactory interface {
	// Create 创建 Inspector 实例
	Create(config ConnectionConfig) (Inspector, error)
	
	// GetType 获取数据库类型
	GetType() string
}

// ConnectionConfig 数据库连接配置
type ConnectionConfig struct {
	Type     string // 数据库类型（dm, oracle, mysql, postgres等）
	Host     string // 主机地址
	Port     int    // 端口
	Database string // 数据库名
	Username string // 用户名
	Password string // 密码
	DSN      string // DSN连接字符串（优先级高于其他参数）
	Schema   string // Schema（Oracle等需要）
	SSLMode  string // SSL模式
}

// Registry Inspector 工厂注册表
type Registry struct {
	factories map[string]InspectorFactory
}

// NewRegistry 创建新的注册表
func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[string]InspectorFactory),
	}
}

// Register 注册工厂
func (r *Registry) Register(dbType string, factory InspectorFactory) {
	r.factories[dbType] = factory
}

// Get 获取工厂
func (r *Registry) Get(dbType string) (InspectorFactory, bool) {
	factory, ok := r.factories[dbType]
	return factory, ok
}

// GetSupportedTypes 获取支持的数据库类型列表
func (r *Registry) GetSupportedTypes() []string {
	types := make([]string, 0, len(r.factories))
	for t := range r.factories {
		types = append(types, t)
	}
	return types
}

// 全局注册表
var globalRegistry = NewRegistry()

// Register 注册工厂到全局注册表
func Register(dbType string, factory InspectorFactory) {
	globalRegistry.Register(dbType, factory)
}

// GetFactory 从全局注册表获取工厂
func GetFactory(dbType string) (InspectorFactory, bool) {
	return globalRegistry.Get(dbType)
}

// GetSupportedTypes 获取全局注册表支持的数据库类型
func GetSupportedTypes() []string {
	return globalRegistry.GetSupportedTypes()
}
