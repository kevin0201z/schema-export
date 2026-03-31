package oracle

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/sijms/go-ora/v2" // 纯 Go Oracle 驱动
	"github.com/schema-export/schema-export/internal/database"
	"github.com/schema-export/schema-export/internal/inspector"
)

// Inspector Oracle 数据库 Inspector 实现
type Inspector struct {
	*database.OracleCompatibleInspector
}

// NewInspector 创建 Oracle Inspector
func NewInspector(config inspector.ConnectionConfig) *Inspector {
	return &Inspector{
		OracleCompatibleInspector: database.NewOracleCompatibleInspector(config, database.PlaceholderColon),
	}
}

// Connect 连接 Oracle 数据库
func (i *Inspector) Connect(ctx context.Context) error {
	dsn := i.BuildDSN()

	db, err := sql.Open("oracle", dsn)
	if err != nil {
		return fmt.Errorf("failed to open oracle connection: %w", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(database.DefaultMaxOpenConns)
	db.SetMaxIdleConns(database.DefaultMaxIdleConns)
	db.SetConnMaxLifetime(database.DefaultConnMaxLifetime)

	// 使用带超时的 context
	pingCtx, cancel := context.WithTimeout(ctx, database.DefaultPingTimeout)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping oracle database: %w", err)
	}

	i.SetDB(db)
	return nil
}

// BuildDSN 构建 Oracle DSN
func (i *Inspector) BuildDSN() string {
	config := i.GetConfig()
	if config.DSN != "" {
		dsn := config.DSN
		// 如果 DSN 中已经有 oracle:// 前缀，直接返回
		if strings.HasPrefix(dsn, "oracle://") {
			return dsn
		}
		// 如果 DSN 中没有 oracle:// 前缀，添加它
		// 检查是否是传统格式 user/password@host:port/service
		if idx := strings.Index(dsn, "@"); idx > 0 {
			dsn = "oracle://" + dsn
		}
		return dsn
	}

	// Oracle DSN 格式: oracle://user:password@host:port/service_name
	serviceName := config.Database
	if serviceName == "" {
		serviceName = config.Schema
	}

	// 使用 go-ora 的 URL 格式
	return fmt.Sprintf("oracle://%s:%s@%s:%d/%s",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		serviceName,
	)
}

// Factory Oracle Inspector 工厂
type Factory struct{}

// Create 创建 Inspector 实例
func (f *Factory) Create(config inspector.ConnectionConfig) (inspector.Inspector, error) {
	ins := NewInspector(config)
	return ins, nil
}

// GetType 获取数据库类型
func (f *Factory) GetType() string {
	return "oracle"
}

func init() {
	inspector.Register("oracle", &Factory{})
}
