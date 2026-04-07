// Package model 定义了数据库结构导出工具的核心数据模型。
//
// 该包包含了表示数据库元数据的各种类型定义，包括：
//   - Table: 表结构元数据
//   - Column: 字段元数据
//   - Index: 索引元数据
//   - ForeignKey: 外键元数据
//   - CheckConstraint: CHECK 约束元数据
//   - View: 视图元数据
//   - Procedure: 存储过程元数据
//   - Function: 函数元数据
//   - Trigger: 触发器元数据
//   - Sequence: 序列元数据
//
// 这些模型类型被 Inspector 接口实现用于返回数据库元数据，
// 并被 Exporter 接口实现用于生成各种格式的输出文档。
//
// 使用示例:
//
//	table := model.Table{
//	    Name:    "users",
//	    Comment: "用户表",
//	    Type:    model.TableTypeTable,
//	    Columns: []model.Column{
//	        {
//	            Name:         "id",
//	            DataType:     "BIGINT",
//	            IsPrimaryKey: true,
//	            IsNullable:   false,
//	        },
//	    },
//	}
package model
