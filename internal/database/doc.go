// Package database 提供数据库 Inspector 的基础实现和公共功能。
//
// 该包包含了 Inspector 接口的基础实现，为各数据库类型的 Inspector 提供公共功能。
// 各数据库类型的 Inspector 实现位于独立的子包中：
//   - dm: 达梦数据库 Inspector
//   - oracle: Oracle 数据库 Inspector
//   - sqlserver: Microsoft SQL Server Inspector
//   - mysql: MySQL 数据库 Inspector
//   - postgres: PostgreSQL 数据库 Inspector
//
// 核心组件:
//   - BaseInspector: Inspector 接口的基础实现，提供连接管理、配置访问等公共功能
//
// 常量:
//   - DefaultPingTimeout: 默认 Ping 超时时间
//   - DefaultTestTimeout: 默认测试连接超时时间
//   - DefaultExportTimeout: 默认导出操作超时时间
//   - DefaultMaxOpenConns: 默认最大打开连接数
//   - DefaultMaxIdleConns: 默认最大空闲连接数
//   - DefaultConnMaxLifetime: 默认连接最大生命周期
//
// 使用示例:
//
//	// 创建 MySQL Inspector
//	inspector := mysql.NewInspector(config)
//
//	// 连接数据库
//	if err := inspector.Connect(ctx); err != nil {
//	    log.Fatal(err)
//	}
//	defer inspector.Close()
package database
