package dm

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "dm" // 达梦 Go 驱动（本地驱动）

	"github.com/schema-export/schema-export/internal/database"
	"github.com/schema-export/schema-export/internal/inspector"
)

// Inspector 达梦数据库 Inspector 实现
// 使用 dm-go-driver 纯 Go 驱动，无需安装 ODBC
type Inspector struct {
	*database.OracleCompatibleInspector
}

// NewInspector 创建达梦 Inspector
func NewInspector(config inspector.ConnectionConfig) *Inspector {
	return &Inspector{
		OracleCompatibleInspector: database.NewOracleCompatibleInspector(config, database.PlaceholderQuestion),
	}
}

// Connect 连接达梦数据库
func (i *Inspector) Connect(ctx context.Context) error {
	dsn := i.BuildDSN()

	// 使用 dm-go-driver 连接
	db, err := sql.Open("dm", dsn)
	if err != nil {
		return fmt.Errorf("failed to open dm connection: %w", err)
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
		return fmt.Errorf("failed to ping dm database: %w", err)
	}

	i.SetDB(db)
	return nil
}

// BuildDSN 构建达梦 DSN
func (i *Inspector) BuildDSN() string {
	config := i.GetConfig()
	if config.DSN != "" {
		// 如果 DSN 中已经有 dm:// 前缀，直接返回
		if strings.HasPrefix(config.DSN, "dm://") {
			return config.DSN
		}
		// 添加 dm:// 前缀
		return "dm://" + config.DSN
	}

	// 达梦 DSN 格式: dm://user:password@host:port
	return fmt.Sprintf("dm://%s:%s@%s:%d",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
	)
}

// Factory 达梦 Inspector 工厂
type Factory struct{}

// Create 创建 Inspector 实例
func (f *Factory) Create(config inspector.ConnectionConfig) (inspector.Inspector, error) {
	ins := NewInspector(config)
	return ins, nil
}

// GetType 获取数据库类型
func (f *Factory) GetType() string {
	return "dm"
}

func init() {
	inspector.Register("dm", &Factory{})
}

// Forwarding methods to satisfy analyzers and method promotion
// These simply delegate to the embedded OracleCompatibleInspector/BaseInspector.
func (i *Inspector) SetDB(db *sql.DB) {
	i.OracleCompatibleInspector.SetDB(db)
}

func (i *Inspector) GetDB() *sql.DB {
	return i.OracleCompatibleInspector.GetDB()
}

func (i *Inspector) GetConfig() inspector.ConnectionConfig {
	return i.OracleCompatibleInspector.GetConfig()
}
