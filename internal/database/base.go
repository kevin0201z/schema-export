package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/schema-export/schema-export/internal/inspector"
)

// 连接池和超时相关常量
const (
	// DefaultPingTimeout 默认 Ping 超时时间
	DefaultPingTimeout = 10 * time.Second
	// DefaultTestTimeout 默认测试连接超时时间
	DefaultTestTimeout = 5 * time.Second
	// DefaultExportTimeout 默认导出操作超时时间
	DefaultExportTimeout = 30 * time.Minute
	// DefaultMaxOpenConns 默认最大打开连接数
	DefaultMaxOpenConns = 10
	// DefaultMaxIdleConns 默认最大空闲连接数
	DefaultMaxIdleConns = 5
	// DefaultConnMaxLifetime 默认连接最大生命周期
	DefaultConnMaxLifetime = 30 * time.Minute
)

// BaseInspector Inspector 接口的基础实现。
//
// BaseInspector 提供了 Inspector 接口的通用功能实现，包括：
//   - 连接配置管理
//   - 数据库连接管理
//   - 连接测试功能
//
// 各数据库类型的 Inspector 可以嵌入 BaseInspector 来复用这些通用功能，
// 只需要实现 Connect 和 BuildDSN 等特定方法。
//
// 使用示例:
//
//	type MySQLInspector struct {
//	    *database.BaseInspector
//	}
//
//	func (i *MySQLInspector) Connect(ctx context.Context) error {
//	    dsn := i.BuildDSN()
//	    db, err := sql.Open("mysql", dsn)
//	    if err != nil {
//	        return err
//	    }
//	    i.SetDB(db)
//	    return nil
//	}
type BaseInspector struct {
	config inspector.ConnectionConfig // 数据库连接配置
	db     *sql.DB                    // 数据库连接实例
}

// NewBaseInspector 创建基础 Inspector 实例。
//
// 参数:
//   - config: 数据库连接配置
//
// 返回值:
//   - *BaseInspector: 新创建的基础 Inspector 实例
func NewBaseInspector(config inspector.ConnectionConfig) *BaseInspector {
	return &BaseInspector{
		config: config,
	}
}

// Connect 连接数据库。
//
// 该方法需要由子类重写，实现特定数据库类型的连接逻辑。
// 默认实现返回 "not implemented" 错误。
//
// 参数:
//   - ctx: 上下文，用于控制连接超时
//
// 返回值:
//   - error: 连接失败时返回错误
func (b *BaseInspector) Connect(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}

// Close 关闭数据库连接。
//
// 如果连接存在，则关闭它。如果连接不存在，则直接返回 nil。
//
// 返回值:
//   - error: 关闭失败时返回错误
func (b *BaseInspector) Close() error {
	if b.db != nil {
		return b.db.Close()
	}
	return nil
}

// TestConnection 测试数据库连接是否正常。
//
// 使用 Ping 方法验证连接是否仍然有效。
//
// 参数:
//   - ctx: 上下文，用于控制超时
//
// 返回值:
//   - error: 连接测试失败时返回错误
func (b *BaseInspector) TestConnection(ctx context.Context) error {
	if b.db == nil {
		return fmt.Errorf("database not connected")
	}

	ctx, cancel := context.WithTimeout(ctx, DefaultTestTimeout)
	defer cancel()

	return b.db.PingContext(ctx)
}

// GetDB 获取数据库连接实例。
//
// 返回值:
//   - *sql.DB: 数据库连接实例
func (b *BaseInspector) GetDB() *sql.DB {
	return b.db
}

// SetDB 设置数据库连接实例。
//
// 参数:
//   - db: 数据库连接实例
func (b *BaseInspector) SetDB(db *sql.DB) {
	b.db = db
}

// GetConfig 获取数据库连接配置。
//
// 返回值:
//   - inspector.ConnectionConfig: 数据库连接配置
func (b *BaseInspector) GetConfig() inspector.ConnectionConfig {
	return b.config
}

// BuildDSN 构建数据源名称（DSN）。
//
// 该方法需要由子类重写，实现特定数据库类型的 DSN 构建逻辑。
// 默认实现直接返回配置中的 DSN 字段。
//
// 返回值:
//   - string: 数据源名称字符串
func (b *BaseInspector) BuildDSN() string {
	if b.config.DSN != "" {
		return b.config.DSN
	}
	return ""
}
