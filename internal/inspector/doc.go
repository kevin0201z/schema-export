// Package inspector 定义了数据库元数据读取的抽象接口。
//
// 该包提供了数据库 Inspector 的核心接口定义，用于从不同类型的数据库中读取元数据。
// 通过接口抽象，支持多种数据库类型的统一访问方式。
//
// 核心接口:
//   - Inspector: 数据库元数据读取接口，定义了获取表、字段、索引等元数据的方法
//   - InspectorFactory: Inspector 工厂接口，用于创建特定数据库类型的 Inspector 实例
//
// 支持的数据库类型:
//   - dm: 达梦数据库
//   - oracle: Oracle 数据库
//   - sqlserver: Microsoft SQL Server
//   - mysql: MySQL 数据库
//   - postgres: PostgreSQL 数据库
//
// 使用示例:
//
//	// 获取工厂
//	factory, ok := inspector.GetFactory("mysql")
//	if !ok {
//	    log.Fatal("unsupported database type")
//	}
//
//	// 创建 Inspector
//	config := inspector.ConnectionConfig{
//	    Host:     "localhost",
//	    Port:     3306,
//	    Database: "mydb",
//	    Username: "root",
//	    Password: "password",
//	}
//	inspector, err := factory.Create(config)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer inspector.Close()
//
//	// 获取表列表
//	tables, err := inspector.GetTables(ctx)
package inspector
