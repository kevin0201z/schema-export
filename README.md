# 数据库结构导出工具

一个支持达梦（DM）、Oracle、SQL Server、MySQL、PostgreSQL、SQLite 等数据库结构导出工具。

可生成 Markdown、SQL DDL、JSON、YAML 格式的数据库结构文档。

## 功能特性

- **多数据库支持**：达梦（DM）、Oracle、SQL Server、MySQL、PostgreSQL、SQLite
- **多种导出格式**：Markdown、SQL DDL、JSON、YAML
- **灵活的导出模式**：单文件或按表分文件导出
- **数据库对象支持**：表、视图、存储过程、函数、触发器、序列
- **表过滤功能**：包含/排除表、正则表达式匹配
- **CLI 界面**：易于使用的命令行界面
- **环境变量**：支持通过环境变量配置

## 安装

### 从源码编译

```bash
go build -o schema-export ./cmd/schema-export
```

### 跨平台编译

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o schema-export-linux ./cmd/schema-export

# Windows
GOOS=windows GOARCH=amd64 go build -o schema-export.exe ./cmd/schema-export

# macOS
GOOS=darwin GOARCH=amd64 go build -o schema-export-darwin ./cmd/schema-export
```

## 使用方法

### 基本用法

```bash
# 导出达梦数据库（基本用法）
./schema-export export \
  --type dm \
  --host localhost \
  --port 5236 \
  --username SYSDBA \
  --password password \
  --database DAMENG \
  --output ./docs

# 使用 DSN 导出（推荐）
./schema-export export \
  --type dm \
  --dsn "dm://SYSDBA:password@localhost:5236?schema=SC" \
  --output ./docs

# 导出 SQL DDL 格式
./schema-export export \
  --type dm \
  --dsn "dm://SYSDBA:password@localhost:5236" \
  --formats markdown,sql \
  --output ./docs

# 导出 JSON/YAML 格式（便于程序化处理）
./schema-export export \
  --type dm \
  --dsn "dm://SYSDBA:password@localhost:5236" \
  --formats json,yaml \
  --output ./docs

# 导出时包含视图
./schema-export export \
  --type dm \
  --dsn "dm://SYSDBA:password@localhost:5236" \
  -V \
  --output ./docs

# 导出指定表
./schema-export export \
  --type dm \
  --dsn "dm://SYSDBA:password@localhost:5236" \
  --tables users,orders,products \
  --output ./docs

# 按表分文件导出
./schema-export export \
  --type dm \
  --dsn "dm://SYSDBA:password@localhost:5236" \
  --split \
  --output ./docs

# 导出 SQLite 文件
./schema-export export \
  --type sqlite \
  --dsn ./app.db \
  --formats markdown,sql,json,yaml \
  --output ./docs

# 使用 --database 传递 SQLite 文件路径
./schema-export export \
  --type sqlite \
  --database ./app.db \
  --include-views \
  --include-triggers \
  --output ./docs
```

### DSN 格式参考

每种数据库的 DSN 连接字符串格式如下：

**达梦（DM）**

```bash
# 完整格式
--dsn "dm://user:password@host:port"

# 简化格式（自动补充 dm:// 前缀）
--dsn "user:password@host:port"

# 示例：带 schema 参数
--dsn "dm://SYSDBA:password@localhost:5236?schema=SCHEMA_NAME"
```

**Oracle**

```bash
# 完整格式
--dsn "oracle://user:password@host:port/service_name"

# 传统格式（自动补充 oracle:// 前缀）
--dsn "user/password@host:port/service_name"

# 示例
--dsn "oracle://scott:tiger@localhost:1521/ORCL"

# 带 schema
--dsn "oracle://scott:tiger@localhost:1521/ORCL?schema=OTHER_USER"
```

**SQL Server**

```bash
# 完整格式
--dsn "sqlserver://user:password@host:port?database=dbname"

# 简化格式（自动补充 sqlserver:// 前缀）
--dsn "user:password@host:port?database=dbname"

# 示例
--dsn "sqlserver://sa:password@localhost:1433?database=mydb"
```

**MySQL**

```bash
# 完整格式（推荐）
--dsn "mysql://user:password@host:port/dbname"

# DSN 原始格式（go-sql-driver 标准格式）
--dsn "user:password@tcp(host:port)/dbname"

# 示例
--dsn "mysql://root:password@localhost:3306/mydb"

# 带 SSL
--dsn "mysql://root:password@localhost:3306/mydb?tls=true"
```

**PostgreSQL**

```bash
# postgres:// 格式
--dsn "postgres://user:password@host:port/dbname"

# postgresql:// 格式（同样支持）
--dsn "postgresql://user:password@host:port/dbname"

# 示例：本地连接（自动禁 SSL）
--dsn "postgresql://postgres:password@localhost:5432/mydb"

# 示例：指定 SSL 模式
--dsn "postgres://user:password@host:5432/dbname?sslmode=require"

# 注意：不指定 sslmode 时，工具会自动追加 sslmode=disable
```

**SQLite**

```bash
# 本地文件路径
--dsn "./app.db"

# URI 形式
--dsn "file:app.db?mode=ro"

# 内存数据库
--dsn ":memory:"

# 也可使用分离参数模式，把文件路径填到 --database
--database "./app.db"
```

### 使用环境变量

```bash
export DB_TYPE=sqlite
export DB_DSN="./app.db"
export EXPORT_OUTPUT=./docs
export EXPORT_FORMATS=markdown,sql
export EXPORT_INCLUDE_VIEWS=true

./schema-export export
```

### 表过滤

```bash
# 导出指定表
./schema-export export --tables users,orders,products

# 排除表
./schema-export export --exclude temp_,log_

# 使用正则表达式
./schema-export export --patterns "^sys_.*","^log_.*"

# 组合过滤
./schema-export export --tables users,orders --exclude orders_archive
```

## TUI 交互模式

通过交互式向导完成数据库导出，无需记忆命令行参数。

```bash
./schema-export tui
```

支持功能：
- 数据库类型选择（dm、oracle、sqlserver、mysql、postgres、sqlite）
- DSN 和分离参数两种连接方式
- 导出对象选择（表、视图、存储过程、函数、触发器、序列）
- 表名过滤（指定表、排除表、正则匹配）
- 多格式导出（markdown、sql、json、yaml）
- 密码输入隐藏和确认页脱敏展示

## CLI 参考

### 全局参数

| 参数              | 说明   |
| --------------- | ---- |
| `-h, --help`    | 显示帮助 |
| `-v, --version` | 显示版本 |

### export 命令参数

| 参数           | 默认值      | 说明                                        |
| ------------ | -------- | ----------------------------------------- |
| `--type`     | dm       | 数据库类型（dm、oracle、sqlserver、mysql、postgres、sqlite） |
| `--host`     | <br />   | 数据库主机                           |
| `--port`     | 0        | 数据库端口                           |
| `--database` | <br />   | 数据库名；SQLite 下表示数据库文件路径          |
| `--username` | <br />   | 数据库用户名                          |
| `--password` | <br />   | 数据库密码                           |
| `--dsn`      | <br />   | DSN 连接字符串；SQLite 下可为文件路径、URI 或 `:memory:` |
| `--schema`   | <br />   | 数据库 Schema                      |
| `--output`   | ./output | 输出目录或文件路径                       |
| `--formats`  | markdown | 导出格式（逗号分隔：markdown,sql,json,yaml）   |
| `--split`    | false    | 按表分文件导出                         |
| `--tables`   | <br />   | 要导出的表（逗号分隔）                     |
| `--exclude`  | <br />   | 要排除的表（逗号分隔）                     |
| `--patterns` | <br />   | 表名正则匹配模式                        |
| `-V, --include-views` | false | 包含视图导出                      |
| `-P, --include-procedures` | false | 包含存储过程导出                |
| `-F, --include-functions` | false | 包含函数导出                    |
| `-T, --include-triggers` | false | 包含触发器导出                  |
| `-S, --include-sequences` | false | 包含序列导出                  |

## 环境变量

| 变量                    | 说明                  |
| --------------------- | ------------------- |
| `DB_TYPE`             | 数据库类型               |
| `DB_HOST`             | 数据库主机               |
| `DB_PORT`             | 数据库端口               |
| `DB_DATABASE`         | 数据库名                |
| `DB_USERNAME`         | 数据库用户名              |
| `DB_PASSWORD`         | 数据库密码               |
| `DB_DSN`              | DSN 连接字符串           |
| `DB_SCHEMA`           | 数据库 Schema          |
| `EXPORT_OUTPUT`       | 输出目录                |
| `EXPORT_FORMATS`      | 导出格式（逗号分隔）          |
| `EXPORT_SPLIT`        | 分文件导出（true/false）   |
| `EXPORT_INCLUDE_VIEWS` | 包含视图导出（true/false） |
| `EXPORT_INCLUDE_PROCEDURES` | 包含存储过程导出（true/false） |
| `EXPORT_INCLUDE_FUNCTIONS` | 包含函数导出（true/false） |
| `EXPORT_INCLUDE_TRIGGERS` | 包含触发器导出（true/false） |
| `EXPORT_INCLUDE_SEQUENCES` | 包含序列导出（true/false） |

## 输出说明

### 输出路径

`--output` 参数支持两种形式：

1. **输出目录**：指定目录路径，文件名使用默认值
   ```bash
   --output ./output
   # 生成：./output/schema.md 或 ./output/schema.sql
   ```
2. **输出文件**：指定完整文件路径，扩展名会根据格式自动调整
   ```bash
   --output ./output/tables.md
   # Markdown 格式生成：./output/tables.md
   # SQL 格式生成：./output/tables.sql
   ```

### 文件覆盖

如果输出文件已存在，会直接覆盖，不会报错或创建备份。


## 重要说明

### Schema 参数（达梦/Oracle）

对于达梦（DM）和 Oracle 数据库，**强烈建议**指定 `--schema` 参数来导出指定 schema 下的表：

- **不指定 schema**：导出当前连接用户拥有的表
- **指定 schema**：导出该 schema 下的所有表（需要相应权限）

**统一使用方式（推荐）：**

```bash
# 达梦数据库
./schema-export export --type dm \
  --dsn "dm://SYSDBA:password@localhost:5236" \
  --schema sc

# Oracle 数据库
./schema-export export --type oracle \
  --dsn "oracle://user:password@localhost:1521/ORCL" \
  --schema OTHER_SCHEMA
```

**替代方式（直接在 DSN 中指定）：**

```bash
# 达梦数据库
./schema-export export --type dm \
  --dsn "dm://SYSDBA:password@localhost:5236?schema=sc"

# Oracle 数据库
./schema-export export --type oracle \
  --dsn "oracle://user:password@localhost:1521/ORCL?schema=OTHER_SCHEMA"
```

**注意：**
- 达梦中 schema 参数用于指定模式名，查询 `ALL_` 前缀的数据字典视图
- Oracle 中 schema 等同于用户名，指定后可查询其他用户的表（需权限）

## 数据库支持

### 支持的数据库

| 数据库        | 状态     | 驱动                    |
| ---------- | ------ | --------------------- |
| 达梦（DM）     | ✅ 已支持  | dm-go-driver（纯 Go 驱动） |
| Oracle     | ✅ 已支持  | go-ora（纯 Go 驱动）       |
| SQL Server | ✅ 已支持  | go-mssqldb（纯 Go 驱动）   |
| MySQL      | ✅ 已支持  | go-sql-driver/mysql（纯 Go 驱动） |
| PostgreSQL | ✅ 已支持  | lib/pq（纯 Go 驱动）       |

## 架构

项目采用"CLI 负责入口、应用服务负责编排、Inspector/Exporter 负责扩展点"的分层结构。

这样的拆分有两个目的：

- 把不同数据库的查询差异隔离在 `internal/database/` 下，新增数据库时不需要改动主流程。
- 把不同输出格式的渲染逻辑隔离在 `internal/exporter/` 下，新增格式时不需要改动数据库读取逻辑。

### 核心组件说明

| 组件 | 职责 | 关键文件 |
|------|------|----------|
| **CLI** | 命令行参数解析、命令路由 | `cmd/schema-export/main.go`, `internal/cli/` |
| **App** | 导出流程编排 | `internal/app/export/` |
| **Config** | 配置管理、环境变量、配置校验 | `internal/config/` |
| **Filter** | 表过滤规则 | `internal/filter/` |
| **Inspector** | 数据库元数据查询接口 | `internal/inspector/interface.go` |
| **Database** | 各数据库 Inspector 实现 | `internal/database/dm/`, `internal/database/oracle/`, `internal/database/sqlserver/`, `internal/database/mysql/`, `internal/database/postgres/` |
| **Model** | 数据模型定义 | `internal/model/` |
| **Exporter** | 导出格式实现 | `internal/exporter/markdown/`, `internal/exporter/sql/` |
| **Third-Party** | 仓库内置第三方驱动源码 | `third_party/dm-go-driver/` |

### 目录重点

- `cmd/schema-export/`：程序入口与命令注册。
- `internal/app/export/`：完整导出流程的编排层。
- `internal/database/`：不同数据库的 Inspector 实现。
- `internal/exporter/`：不同输出格式的导出实现。
- `third_party/`：随仓库分发的第三方源码。

### 数据流向

```
CLI 参数 → Config → App Service → Inspector → Database Driver → Model → Exporter → 输出文件
                      ↓
                   Filter
```

1. **CLI** 解析用户输入的参数
2. **Config** 整合 CLI 参数和环境变量
3. **App Service** 编排导出流程
4. **Inspector** 根据数据库类型创建对应的驱动实例
5. **Database Driver** 查询数据库元数据（表、字段、索引、外键）
6. **Filter** 根据配置过滤表
7. **Model** 存储查询结果
8. **Exporter** 将模型转换为指定格式（Markdown/SQL）
9. **输出文件** 写入磁盘

## 许可证

MIT 许可证
