# 项目代码优化精简建议计划

## 📌 概述

针对 schema-export 项目当前代码结构，提出优化精简建议，提高代码复用性、可维护性和简洁性。

***

## 🎯 优化建议

### 建议 1: DM 和 Oracle 驱动代码去重（高优先级）

**问题分析**:

* `internal/database/dm/inspector.go` 和 `internal/database/oracle/inspector.go` 代码高度重复

* 两者都继承 `OracleCompatibleInspector`，但 `GetTable/GetTables/GetColumns/GetIndexes/GetForeignKeys` 方法实现几乎完全相同

* 只有 `Connect()` 和 `BuildDSN()` 方法有实质差异

**优化方案**:
将公共方法提升到 `OracleCompatibleInspector` 基类中：

```go
// OracleCompatibleInspector 添加公共方法
func (o *OracleCompatibleInspector) GetTables(ctx context.Context, schema string) ([]model.Table, error)
func (o *OracleCompatibleInspector) GetTable(ctx context.Context, tableName, schema string) (*model.Table, error)
func (o *OracleCompatibleInspector) GetColumns(ctx context.Context, tableName, schema string) ([]model.Column, error)
func (o *OracleCompatibleInspector) GetIndexes(ctx context.Context, tableName, schema string) ([]model.Index, error)
func (o *OracleCompatibleInspector) GetForeignKeys(ctx context.Context, tableName, schema string) ([]model.ForeignKey, error)
```

DM 和 Oracle 的 Inspector 只需实现：

* `Connect()` - 数据库连接

* `BuildDSN()` - DSN 构建

* 可选：覆盖基类方法（如果有特殊逻辑）

**预期效果**:

* DM Inspector 从 170 行减少到约 60 行

* Oracle Inspector 从 179 行减少到约 70 行

* 消除重复代码约 200+ 行

***

### 建议 2: 统一 SQL Server 驱动架构（中优先级）

**问题分析**:

* SQL Server 驱动 (`internal/database/sqlserver/inspector.go`) 独立实现，没有复用现有基类

* 但代码结构与 DM/Oracle 的公共方法非常相似

**优化方案**:
有两种选择：

**方案 A**: 保持独立（当前方式）

* 优点：SQL Server 与 Oracle 语法差异大，独立实现更灵活

* 缺点：代码重复

**方案 B**: 创建更通用的基类

* 提取所有驱动的公共逻辑到 `BaseInspector`

* 定义查询接口，各驱动只需实现 SQL 语句

**推荐**: 保持方案 A，SQL Server 与 Oracle 数据字典差异较大，独立实现更易于维护

***

### 建议 3: 简化 Model 方法（低优先级）

**问题分析**:

* `internal/model/table.go` 中的 `GetPrimaryKey()` 方法遍历查找主键，效率较低

* 每次调用都遍历所有字段

**优化方案**:
在解析时就将主键信息缓存：

```go
type Table struct {
    Name        string
    Comment     string
    Type        TableType
    Columns     []Column
    Indexes     []Index
    ForeignKeys []ForeignKey
    primaryKey  *Column  // 缓存主键字段
}

func (t *Table) GetPrimaryKey() *Column {
    if t.primaryKey == nil {
        for i := range t.Columns {
            if t.Columns[i].IsPrimaryKey {
                t.primaryKey = &t.Columns[i]
                break
            }
        }
    }
    return t.primaryKey
}
```

**预期效果**:

* 多次调用时性能提升

* 代码复杂度略微增加

***

### 建议 4: 合并输入参数结构体（中优先级）

**问题分析**:

* `oracle_compatible.go` 中定义了多个输入结构体：

  * `QueryTablesInput`

  * `QueryColumnsInput`

  * `QueryIndexesInput`

  * `QueryForeignKeysInput`

  * `QueryTableCommentInput`

* 这些结构体内容高度相似，都是 `TableName` + `Schema`

**优化方案**:
合并为一个通用结构体：

```go
// QueryInput 通用查询输入参数
type QueryInput struct {
    TableName string  // 可选，表名
    Schema    string  // 可选，Schema
}

// 使用示例
func (o *OracleCompatibleInspector) QueryColumns(ctx context.Context, input QueryInput) ([]model.Column, error)
func (o *OracleCompatibleInspector) QueryIndexes(ctx context.Context, input QueryInput) ([]model.Index, error)
```

**预期效果**:

* 减少类型定义数量

* 简化接口签名

* 提高代码一致性

***

### 建议 5: 提取 SQL 查询常量（低优先级）

**问题分析**:

* SQL 查询语句直接硬编码在代码中

* 查询语句较长，影响代码可读性

**优化方案**:
将 SQL 语句提取为常量：

```go
const (
    queryTablesSQL = `
        SELECT TABLE_NAME, COMMENTS 
        FROM %s_TAB_COMMENTS 
        WHERE TABLE_TYPE = 'TABLE'%s
        ORDER BY TABLE_NAME
    `
    // ...
)
```

**预期效果**:

* 提高代码可读性

* 便于 SQL 语句维护和调优

***

### 建议 6: 优化 Exporter 模板（中优先级）

**问题分析**:

* `markdown/exporter.go` 和 `sql/exporter.go` 中的模板较长

* 模板与代码混合，维护困难

**优化方案**:
将模板提取到单独的文件：

```
internal/exporter/markdown/
├── exporter.go
└── template.go  // 或 template.md

internal/exporter/sql/
├── exporter.go
└── template.go  // 或 template.sql
```

**预期效果**:

* 分离模板和业务逻辑

* 便于模板修改和国际化

***

### 建议 7: 统一错误处理（低优先级）

**问题分析**:

* 错误信息格式不统一

* 部分地方使用 `fmt.Errorf`，部分直接返回错误

**优化方案**:
定义统一的错误类型：

```go
package errors

var (
    ErrConnectionFailed = errors.New("database connection failed")
    ErrQueryFailed      = errors.New("query execution failed")
    ErrTableNotFound    = errors.New("table not found")
)

func Wrap(err error, msg string) error {
    return fmt.Errorf("%s: %w", msg, err)
}
```

**预期效果**:

* 统一的错误处理风格

* 便于错误分类和处理

***

## 📊 优化优先级总结

| 优先级  | 建议                   | 工作量 | 收益               |
| ---- | -------------------- | --- | ---------------- |
| 🔴 高 | 建议 1: DM/Oracle 代码去重 | 中   | 高（减少 200+ 行重复代码） |
| 🟡 中 | 建议 4: 合并输入参数结构体      | 小   | 中                |
| 🟡 中 | 建议 6: 优化 Exporter 模板 | 小   | 中                |
| 🟢 低 | 建议 3: 简化 Model 方法    | 小   | 低                |
| 🟢 低 | 建议 5: 提取 SQL 常量      | 小   | 低                |
| 🟢 低 | 建议 7: 统一错误处理         | 中   | 低                |
| ⚪ 可选 | 建议 2: SQL Server 架构  | -   | -（保持现状即可）        |

***

## ✅ 推荐实施顺序

1. **建议 1** - DM/Oracle 代码去重（收益最大）
2. **建议 4** - 合并输入参数结构体（简单易行）
3. **建议 6** - 优化 Exporter 模板（提高可维护性）
4. 其他建议根据实际情况选择性实施

***

## 📁 涉及文件

* `internal/database/oracle_compatible.go` - 需要大幅修改

* `internal/database/dm/inspector.go` - 可以大幅精简

* `internal/database/oracle/inspector.go` - 可以大幅精简

* `internal/model/table.go` - 可选优化

* `internal/exporter/markdown/exporter.go` - 可选优化

* `internal/exporter/sql/exporter.go` - 可选优化

