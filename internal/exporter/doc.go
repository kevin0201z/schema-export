// Package exporter 定义了数据库结构导出的抽象接口。
//
// 该包提供了导出器的核心接口定义，用于将数据库元数据导出为各种格式。
// 通过接口抽象，支持多种导出格式的统一访问方式。
//
// 核心接口:
//   - Exporter: 导出器接口，定义了导出数据库元数据的方法
//   - ExporterFactory: 导出器工厂接口，用于创建特定格式的导出器实例
//
// 支持的导出格式:
//   - markdown: Markdown 格式文档
//   - sql: SQL DDL 脚本
//   - json: JSON 格式
//   - yaml: YAML 格式
//
// 使用示例:
//
//	// 获取工厂
//	factory, ok := exporter.GetFactory("markdown")
//	if !ok {
//	    log.Fatal("unsupported export format")
//	}
//
//	// 创建导出器
//	exp, err := factory.Create()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// 执行导出
//	options := exporter.ExportOptions{
//	    OutputDir:    "./output",
//	    DbType:       "mysql",
//	    IncludeViews: true,
//	}
//	err = exp.Export(tables, views, procedures, functions, triggers, sequences, options)
package exporter
