package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/schema-export/schema-export/internal/inspector"
)

const (
	DefaultPingTimeout      = 10 * time.Second
	DefaultTestTimeout      = 5 * time.Second
	DefaultExportTimeout    = 30 * time.Minute
	DefaultMaxOpenConns     = 10
	DefaultMaxIdleConns     = 5
	DefaultConnMaxLifetime  = 30 * time.Minute
)

// BaseInspector 基础 Inspector 实现
type BaseInspector struct {
	config inspector.ConnectionConfig
	db     *sql.DB
}

// NewBaseInspector 创建基础 Inspector
func NewBaseInspector(config inspector.ConnectionConfig) *BaseInspector {
	return &BaseInspector{
		config: config,
	}
}

// Connect 连接数据库（子类需要重写）
func (b *BaseInspector) Connect(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}

// Close 关闭连接
func (b *BaseInspector) Close() error {
	if b.db != nil {
		return b.db.Close()
	}
	return nil
}

// TestConnection 测试连接
func (b *BaseInspector) TestConnection(ctx context.Context) error {
	if b.db == nil {
		return fmt.Errorf("database not connected")
	}
	
	ctx, cancel := context.WithTimeout(ctx, DefaultTestTimeout)
	defer cancel()
	
	return b.db.PingContext(ctx)
}

// GetDB 获取数据库连接
func (b *BaseInspector) GetDB() *sql.DB {
	return b.db
}

// SetDB 设置数据库连接
func (b *BaseInspector) SetDB(db *sql.DB) {
	b.db = db
}

// GetConfig 获取配置
func (b *BaseInspector) GetConfig() inspector.ConnectionConfig {
	return b.config
}

// BuildDSN 构建 DSN（子类需要重写）
func (b *BaseInspector) BuildDSN() string {
	if b.config.DSN != "" {
		return b.config.DSN
	}
	return ""
}
