package inspector

import (
	"context"
	"github.com/schema-export/schema-export/internal/model"
)

// Inspector 数据库元数据读取接口。
//
// Inspector 定义了从数据库读取元数据的统一接口。每种支持的数据库类型
// 都需要实现该接口，以提供表、字段、索引、外键等元数据的查询功能。
//
// 使用流程:
//  1. 调用 Connect 建立数据库连接
//  2. 调用各种 Get* 方法获取元数据
//  3. 调用 Close 关闭连接
//
// 示例:
//
//	err := inspector.Connect(ctx)
//	if err != nil {
//	    return err
//	}
//	defer inspector.Close()
//
//	tables, err := inspector.GetTables(ctx)
type Inspector interface {
	// Connect 建立数据库连接。
	//
	// 参数:
	//   - ctx: 上下文，用于控制连接超时
	//
	// 返回值:
	//   - error: 连接失败时返回错误
	Connect(ctx context.Context) error

	// Close 关闭数据库连接。
	//
	// 返回值:
	//   - error: 关闭失败时返回错误
	Close() error

	// TestConnection 测试数据库连接是否正常。
	//
	// 参数:
	//   - ctx: 上下文，用于控制超时
	//
	// 返回值:
	//   - error: 连接测试失败时返回错误
	TestConnection(ctx context.Context) error

	// GetTables 获取数据库中的所有表列表。
	//
	// 返回的 Table 对象仅包含基本信息（名称、注释），不包含字段、索引等详细信息。
	// 如需完整信息，请使用 GetTable 方法。
	//
	// 参数:
	//   - ctx: 上下文，用于控制超时
	//
	// 返回值:
	//   - []model.Table: 表列表
	//   - error: 查询失败时返回错误
	GetTables(ctx context.Context) ([]model.Table, error)

	// GetTable 获取单个表的完整元数据。
	//
	// 返回的 Table 对象包含完整的元数据，包括字段、索引、外键、CHECK 约束等。
	//
	// 参数:
	//   - ctx: 上下文，用于控制超时
	//   - tableName: 表名
	//
	// 返回值:
	//   - *model.Table: 表的完整元数据
	//   - error: 查询失败或表不存在时返回错误
	GetTable(ctx context.Context, tableName string) (*model.Table, error)

	// GetColumns 获取表的字段列表。
	//
	// 参数:
	//   - ctx: 上下文，用于控制超时
	//   - tableName: 表名
	//
	// 返回值:
	//   - []model.Column: 字段列表
	//   - error: 查询失败时返回错误
	GetColumns(ctx context.Context, tableName string) ([]model.Column, error)

	// GetIndexes 获取表的索引列表。
	//
	// 参数:
	//   - ctx: 上下文，用于控制超时
	//   - tableName: 表名
	//
	// 返回值:
	//   - []model.Index: 索引列表
	//   - error: 查询失败时返回错误
	GetIndexes(ctx context.Context, tableName string) ([]model.Index, error)

	// GetForeignKeys 获取表的外键列表。
	//
	// 参数:
	//   - ctx: 上下文，用于控制超时
	//   - tableName: 表名
	//
	// 返回值:
	//   - []model.ForeignKey: 外键列表
	//   - error: 查询失败时返回错误
	GetForeignKeys(ctx context.Context, tableName string) ([]model.ForeignKey, error)

	// GetCheckConstraints 获取表的 CHECK 约束列表。
	//
	// 参数:
	//   - ctx: 上下文，用于控制超时
	//   - tableName: 表名
	//
	// 返回值:
	//   - []model.CheckConstraint: CHECK 约束列表
	//   - error: 查询失败时返回错误
	GetCheckConstraints(ctx context.Context, tableName string) ([]model.CheckConstraint, error)

	// GetViews 获取数据库中的视图列表。
	//
	// 参数:
	//   - ctx: 上下文，用于控制超时
	//
	// 返回值:
	//   - []model.View: 视图列表
	//   - error: 查询失败时返回错误
	GetViews(ctx context.Context) ([]model.View, error)

	// GetProcedures 获取数据库中的存储过程列表。
	//
	// 参数:
	//   - ctx: 上下文，用于控制超时
	//
	// 返回值:
	//   - []model.Procedure: 存储过程列表
	//   - error: 查询失败时返回错误
	GetProcedures(ctx context.Context) ([]model.Procedure, error)

	// GetFunctions 获取数据库中的函数列表。
	//
	// 参数:
	//   - ctx: 上下文，用于控制超时
	//
	// 返回值:
	//   - []model.Function: 函数列表
	//   - error: 查询失败时返回错误
	GetFunctions(ctx context.Context) ([]model.Function, error)

	// GetTriggers 获取表的触发器列表。
	//
	// 参数:
	//   - ctx: 上下文，用于控制超时
	//   - tableName: 表名
	//
	// 返回值:
	//   - []model.Trigger: 触发器列表
	//   - error: 查询失败时返回错误
	GetTriggers(ctx context.Context, tableName string) ([]model.Trigger, error)

	// GetSequences 获取数据库中的序列列表。
	//
	// 注意: MySQL 不支持序列对象。
	//
	// 参数:
	//   - ctx: 上下文，用于控制超时
	//
	// 返回值:
	//   - []model.Sequence: 序列列表
	//   - error: 查询失败时返回错误
	GetSequences(ctx context.Context) ([]model.Sequence, error)
}

// InspectorFactory Inspector 工厂接口。
//
// 工厂接口用于创建特定数据库类型的 Inspector 实例。
// 每种数据库类型都应该实现该接口，并通过 Register 函数注册到全局注册表。
//
// 示例:
//
//	type MySQLFactory struct{}
//
//	func (f *MySQLFactory) Create(config ConnectionConfig) (Inspector, error) {
//	    return NewMySQLInspector(config), nil
//	}
//
//	func (f *MySQLFactory) GetType() string {
//	    return "mysql"
//	}
//
//	// 注册工厂
//	inspector.Register("mysql", &MySQLFactory{})
type InspectorFactory interface {
	// Create 创建 Inspector 实例。
	//
	// 参数:
	//   - config: 数据库连接配置
	//
	// 返回值:
	//   - Inspector: 新创建的 Inspector 实例
	//   - error: 创建失败时返回错误
	Create(config ConnectionConfig) (Inspector, error)

	// GetType 获取工厂支持的数据库类型标识。
	//
	// 返回值:
	//   - string: 数据库类型（如 "mysql", "postgres" 等）
	GetType() string
}

// ConnectionConfig 数据库连接配置。
//
// ConnectionConfig 包含建立数据库连接所需的所有参数。
// 支持两种连接方式：
//  1. 使用 DSN（推荐）：提供完整的连接字符串
//  2. 使用分离参数：分别提供 Host、Port、Username、Password、Database
//
// 字段说明:
//   - Type: 数据库类型（dm, oracle, sqlserver, mysql, postgres）
//   - Host: 数据库主机地址
//   - Port: 数据库端口号
//   - Database: 数据库名称
//   - Username: 数据库用户名
//   - Password: 数据库密码
//   - DSN: 完整的数据源名称（优先级高于其他参数）
//   - Schema: 数据库 Schema（用于 Oracle/达梦）
//   - SSLMode: SSL 连接模式（用于 PostgreSQL）
type ConnectionConfig struct {
	Type     string // 数据库类型（当前支持 dm、oracle、sqlserver）
	Host     string // 主机地址
	Port     int    // 端口
	Database string // 数据库名
	Username string // 用户名
	Password string // 密码
	DSN      string // DSN连接字符串（优先级高于其他参数）
	Schema   string // Schema（Oracle等需要）
	SSLMode  string // SSL模式
}

// Registry Inspector 工厂注册表。
//
// Registry 维护了数据库类型到 Inspector 工厂的映射关系。
// 支持动态注册新的数据库类型。
//
// 使用示例:
//
//	registry := inspector.NewRegistry()
//	registry.Register("mysql", &MySQLFactory{})
//	registry.Register("postgres", &PostgreSQLFactory{})
//
//	factory, ok := registry.Get("mysql")
type Registry struct {
	factories map[string]InspectorFactory
}

// NewRegistry 创建新的工厂注册表。
//
// 返回值:
//   - *Registry: 新创建的空注册表
func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[string]InspectorFactory),
	}
}

// Register 注册数据库工厂到注册表。
//
// 参数:
//   - dbType: 数据库类型标识（如 "mysql", "postgres"）
//   - factory: 工厂实例
func (r *Registry) Register(dbType string, factory InspectorFactory) {
	r.factories[dbType] = factory
}

// Get 从注册表获取指定类型的工厂。
//
// 参数:
//   - dbType: 数据库类型标识
//
// 返回值:
//   - InspectorFactory: 工厂实例
//   - bool: 是否存在该类型的工厂
func (r *Registry) Get(dbType string) (InspectorFactory, bool) {
	factory, ok := r.factories[dbType]
	return factory, ok
}

// GetSupportedTypes 获取注册表支持的所有数据库类型。
//
// 返回值:
//   - []string: 数据库类型列表
func (r *Registry) GetSupportedTypes() []string {
	types := make([]string, 0, len(r.factories))
	for t := range r.factories {
		types = append(types, t)
	}
	return types
}

// globalRegistry 全局 Inspector 工厂注册表。
//
// 各数据库驱动通过 init 函数自动注册到全局注册表。
var globalRegistry = NewRegistry()

// Register 注册工厂到全局注册表。
//
// 通常在数据库驱动包的 init 函数中调用。
//
// 参数:
//   - dbType: 数据库类型标识
//   - factory: 工厂实例
//
// 示例:
//
//	func init() {
//	    inspector.Register("mysql", &MySQLFactory{})
//	}
func Register(dbType string, factory InspectorFactory) {
	globalRegistry.Register(dbType, factory)
}

// GetFactory 从全局注册表获取工厂。
//
// 参数:
//   - dbType: 数据库类型标识
//
// 返回值:
//   - InspectorFactory: 工厂实例
//   - bool: 是否存在该类型的工厂
func GetFactory(dbType string) (InspectorFactory, bool) {
	return globalRegistry.Get(dbType)
}

// GetSupportedTypes 获取全局注册表支持的所有数据库类型。
//
// 返回值:
//   - []string: 数据库类型列表
func GetSupportedTypes() []string {
	return globalRegistry.GetSupportedTypes()
}
