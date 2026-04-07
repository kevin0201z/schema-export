// Package config 提供数据库结构导出工具的配置管理功能。
//
// 该包定义了应用程序的配置结构，支持从多种来源加载配置：
//   - 命令行参数（由 CLI 包处理）
//   - 环境变量
//   - 配置文件（计划支持）
//
// 配置结构:
//   - Config: 顶层配置容器
//   - DatabaseConfig: 数据库连接配置
//   - ExportConfig: 导出选项配置
//
// 环境变量映射:
//   - DB_TYPE: 数据库类型
//   - DB_HOST: 数据库主机
//   - DB_PORT: 数据库端口
//   - DB_DATABASE: 数据库名称
//   - DB_USERNAME: 数据库用户名
//   - DB_PASSWORD: 数据库密码
//   - DB_DSN: 数据源名称（完整连接字符串）
//   - DB_SCHEMA: 数据库 Schema
//   - EXPORT_OUTPUT: 输出目录
//   - EXPORT_FORMATS: 导出格式（逗号分隔）
//   - EXPORT_SPLIT: 是否分文件导出
//   - EXPORT_INCLUDE_VIEWS: 是否包含视图
//   - EXPORT_INCLUDE_PROCEDURES: 是否包含存储过程
//   - EXPORT_INCLUDE_FUNCTIONS: 是否包含函数
//   - EXPORT_INCLUDE_TRIGGERS: 是否包含触发器
//   - EXPORT_INCLUDE_SEQUENCES: 是否包含序列
//
// 使用示例:
//
//	cfg := config.DefaultConfig()
//	cfg.LoadFromEnv()
//	if err := cfg.Validate(); err != nil {
//	    log.Fatal(err)
//	}
package config
