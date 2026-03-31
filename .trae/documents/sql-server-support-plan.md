# SQL Server 数据库支持实施计划

## 📌 概述

为 schema-export 工具添加 SQL Server 数据库支持，使用户能够连接 SQL Server 数据库并导出表结构。

***

## 🎯 实施步骤

### 步骤 1: 添加 SQL Server 驱动依赖

**文件**: `go.mod`

添加 Microsoft 官方 Go SQL Server 驱动：

```
github.com/microsoft/go-mssqldb v1.7.0
```

**操作**:

```bash
go get github.com/microsoft/go-mssqldb@v1.7.0
```

***

### 步骤 2: 创建 SQL Server Inspector 实现

**文件**: `internal/database/sqlserver/inspector.go`

实现内容：

1. **Inspector 结构体** - 继承 BaseInspector
2. **Connect()** - 连接 SQL Server 数据库
3. **BuildDSN()** - 构建 SQL Server DSN 连接字符串
4. **GetTables()** - 查询所有表列表
5. **GetTable()** - 获取单个表的完整元数据
6. **GetColumns()** - 查询表字段信息
7. **GetIndexes()** - 查询表索引信息
8. **GetForeignKeys()** - 查询表外键信息

**SQL Server 特有考虑**:

* 支持 Windows 身份验证和 SQL Server 身份验证

* DSN 格式: `sqlserver://user:password@host:port?database=dbname`

* 使用 `?` 占位符

* 系统表过滤（排除 sys、INFORMATION\_SCHEMA 等）

***

### 步骤 3: 实现元数据查询方法

#### 3.1 GetTables 查询

```sql
SELECT 
    t.name AS table_name,
    ep.value AS table_comment
FROM sys.tables t
LEFT JOIN sys.extended_properties ep 
    ON t.object_id = ep.major_id 
    AND ep.minor_id = 0 
    AND ep.name = 'MS_Description'
WHERE t.is_ms_shipped = 0
ORDER BY t.name
```

#### 3.2 GetColumns 查询

```sql
SELECT 
    c.name AS column_name,
    ty.name AS data_type,
    c.max_length,
    c.precision,
    c.scale,
    c.is_nullable,
    dc.definition AS default_value,
    ep.value AS column_comment,
    CASE WHEN pk.column_id IS NOT NULL THEN 1 ELSE 0 END AS is_primary_key,
    c.is_identity AS is_auto_increment
FROM sys.columns c
INNER JOIN sys.types ty ON c.user_type_id = ty.user_type_id
INNER JOIN sys.tables t ON c.object_id = t.object_id
LEFT JOIN sys.default_constraints dc ON c.default_object_id = dc.object_id
LEFT JOIN sys.extended_properties ep ON c.object_id = ep.major_id AND c.column_id = ep.minor_id AND ep.name = 'MS_Description'
LEFT JOIN (
    SELECT ic.object_id, ic.column_id
    FROM sys.indexes i
    INNER JOIN sys.index_columns ic ON i.object_id = ic.object_id AND i.index_id = ic.index_id
    WHERE i.is_primary_key = 1
) pk ON c.object_id = pk.object_id AND c.column_id = pk.column_id
WHERE t.name = ?
ORDER BY c.column_id
```

#### 3.3 GetIndexes 查询

```sql
SELECT 
    i.name AS index_name,
    i.type_desc AS index_type,
    i.is_unique,
    i.is_primary_key,
    c.name AS column_name
FROM sys.indexes i
INNER JOIN sys.index_columns ic ON i.object_id = ic.object_id AND i.index_id = ic.index_id
INNER JOIN sys.columns c ON ic.object_id = c.object_id AND ic.column_id = c.column_id
INNER JOIN sys.tables t ON i.object_id = t.object_id
WHERE t.name = ? AND i.type > 0
ORDER BY i.name, ic.key_ordinal
```

#### 3.4 GetForeignKeys 查询

```sql
SELECT 
    fk.name AS fk_name,
    pc.name AS column_name,
    rt.name AS ref_table,
    rc.name AS ref_column,
    fk.delete_referential_action_desc AS on_delete
FROM sys.foreign_keys fk
INNER JOIN sys.foreign_key_columns fkc ON fk.object_id = fkc.constraint_object_id
INNER JOIN sys.tables pt ON fk.parent_object_id = pt.object_id
INNER JOIN sys.columns pc ON fkc.parent_object_id = pc.object_id AND fkc.parent_column_id = pc.column_id
INNER JOIN sys.tables rt ON fk.referenced_object_id = rt.object_id
INNER JOIN sys.columns rc ON fkc.referenced_object_id = rc.object_id AND fkc.referenced_column_id = rc.column_id
WHERE pt.name = ?
```

***

### 步骤 4: 注册 SQL Server 驱动

**文件**: `internal/database/sqlserver/inspector.go`

在文件末尾添加工厂注册：

```go
type Factory struct{}

func (f *Factory) Create(config inspector.ConnectionConfig) (inspector.Inspector, error) {
    return NewInspector(config), nil
}

func (f *Factory) GetType() string {
    return "sqlserver"
}

func init() {
    inspector.Register("sqlserver", &Factory{})
}
```

***

### 步骤 5: 在 main.go 中导入 SQL Server 驱动

**文件**: `cmd/schema-export/main.go`

添加导入：

```go
import (
    // ... 其他导入
    _ "github.com/schema-export/schema-export/internal/database/sqlserver"
)
```

***

### 步骤 6: 更新文档

**文件**: `README.md`（如存在）

添加 SQL Server 到支持的数据库列表。

***

## 📁 文件结构

```
internal/database/sqlserver/
└── inspector.go          # SQL Server Inspector 实现
```

***

## ✅ 验收标准

* [ ] SQL Server 驱动依赖已添加

* [ ] SQL Server Inspector 实现完成

* [ ] 支持 SQL Server 身份验证连接

* [ ] 支持 Windows 身份验证连接（可选）

* [ ] 能够正确读取表列表

* [ ] 能够正确读取字段信息（含注释）

* [ ] 能够正确读取索引信息

* [ ] 能够正确读取外键信息

* [ ] 驱动已注册到全局工厂

* [ ] 可以通过 `--type sqlserver` 参数使用

* [ ] 代码通过编译

***

## 🛠️ 技术选型

| 组件            | 技术         | 说明                                     |
| ------------- | ---------- | -------------------------------------- |
| SQL Server 驱动 | go-mssqldb | Microsoft 官方 Go 驱动                     |
| 占位符           | `?`        | 与 DM 驱动一致                              |
| DSN 格式        | URL 格式     | \`sqlserver://user:password\@host:port |

