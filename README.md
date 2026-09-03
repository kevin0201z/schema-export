# 数据库结构导出工具

用于导出数据库结构信息并生成文档或结构化文件，当前支持达梦（DM）、Oracle、SQL Server、MySQL、PostgreSQL、SQLite。

输出格式支持：

- Markdown
- SQL DDL
- JSON
- YAML

## 功能概览

- 支持多数据库统一导出
- 支持表、视图、存储过程、函数、触发器、序列等对象
- 支持单文件导出和按对象分文件导出
- 支持表过滤：`--tables`、`--exclude`、`--patterns`
- 支持 CLI 和 TUI 两种使用方式
- 支持环境变量配置

SQLite 特别说明：

- 支持普通表、字段、索引、外键、CHECK、视图、触发器导出
- 不要求 `host`、`port`、`username`、`password`
- 不支持存储过程、独立函数、序列对象

## 快速开始

### 从源码编译

需要 Go 1.25 或更高版本。macOS 支持目标为 12 或更高版本，支持 Intel 和 Apple Silicon。

```bash
make build
./build/schema-export --help
```

下文的 `./schema-export` 表示可执行文件路径；使用上述构建时请替换为 `./build/schema-export`。

### 查看帮助

```bash
./schema-export --help
./schema-export export --help
```

### 最常见用法

```bash
# 达梦：使用 DSN 导出
./schema-export export \
  --type dm \
  --dsn "dm://SYSDBA:password@localhost:5236?schema=SCHEMA_NAME" \
  --output ./docs

# SQLite：导出本地数据库文件
./schema-export export \
  --type sqlite \
  --dsn ./app.db \
  --formats markdown,sql,json,yaml \
  --output ./docs

# SQLite：使用 --database 作为文件路径入口
./schema-export export \
  --type sqlite \
  --database ./app.db \
  --include-views \
  --include-triggers \
  --output ./docs
```

## 安装与构建

### 本地构建

```bash
make build
```

macOS 构建需要 Xcode Command Line Tools（`xcode-select --install`）。Makefile 为 C 编译和链接统一设置 macOS 12 部署目标，避免产物只能在构建机的新版系统运行。直接执行 `go build` 不会自动应用这些设置。

需要兼容 macOS 12 时使用已验证的 Go 1.25 / 1.26；升级 Go 大版本后应重新确认其最低系统要求。

### 跨平台构建

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o schema-export-linux ./cmd/schema-export

# Windows
GOOS=windows GOARCH=amd64 go build -o schema-export.exe ./cmd/schema-export

# macOS：在 Mac 上构建，保留达梦第三方加密插件能力
make build-darwin
# build/schema-export-darwin-amd64：Intel
# build/schema-export-darwin-arm64：Apple Silicon

# 从其他系统构建 macOS 便携版本，不支持达梦第三方加密插件
make build-darwin-portable
# build/schema-export-darwin-{amd64,arm64}-portable
```

正式 macOS 发布在 Mac 上启用 CGO 构建。达梦第三方加密插件需要与程序匹配的系统、架构及 Go 工具链；普通数据库连接和 SQLite 不需要此插件。

## 使用方式

### CLI 导出

#### 1. 使用 DSN

```bash
./schema-export export \
  --type dm \
  --dsn "dm://SYSDBA:password@localhost:5236?schema=SCHEMA_NAME" \
  --output ./docs
```

#### 2. 使用分离参数

```bash
./schema-export export \
  --type mysql \
  --host localhost \
  --port 3306 \
  --username root \
  --password password \
  --database mydb \
  --output ./docs
```

#### 3. 指定输出格式

```bash
# 生成 Markdown 和 SQL
./schema-export export \
  --type postgres \
  --dsn "postgres://user:password@localhost:5432/mydb" \
  --formats markdown,sql \
  --output ./docs

# 生成 JSON 和 YAML
./schema-export export \
  --type postgres \
  --dsn "postgres://user:password@localhost:5432/mydb" \
  --formats json,yaml \
  --output ./docs
```

#### 4. 包含其他数据库对象

```bash
./schema-export export \
  --type dm \
  --dsn "dm://SYSDBA:password@localhost:5236" \
  --include-views \
  --include-procedures \
  --include-functions \
  --include-triggers \
  --include-sequences \
  --output ./docs
```

#### 5. 表过滤

```bash
# 只导出指定表
./schema-export export --tables users,orders,products

# 排除部分表
./schema-export export --exclude temp_,log_

# 使用正则匹配
./schema-export export --patterns "^sys_.*","^log_.*"

# 组合使用
./schema-export export --tables users,orders --exclude orders_archive
```

#### 6. 按表分文件导出

```bash
./schema-export export \
  --type dm \
  --dsn "dm://SYSDBA:password@localhost:5236" \
  --split \
  --output ./docs
```

### TUI 交互模式

```bash
./schema-export tui
```

TUI 支持：

- 选择数据库类型
- DSN / 分离参数两种连接方式
- 选择导出对象和输出格式
- 配置表过滤规则
- 密码输入隐藏和确认页脱敏显示

SQLite 下的 TUI 行为：

- DSN 模式下填写文件路径、URI 或 `:memory:`
- 分离参数模式下只填写数据库文件路径
- 不显示存储过程、函数、序列选项

## SQLite 使用说明

### 连接方式

```bash
# 本地文件
--dsn "./app.db"

# URI
--dsn "file:app.db?mode=ro"

# 内存数据库
--dsn ":memory:"

# 也可以使用 --database
--database "./app.db"
```

### 支持范围

- 支持：表、字段、索引、外键、CHECK、视图、触发器
- 不支持：存储过程、独立函数、序列、SQLCipher、`ATTACH DATABASE` 跨库导出

## DSN 参考

### 达梦（DM）

```bash
--dsn "dm://user:password@host:port"
--dsn "user:password@host:port"
--dsn "dm://SYSDBA:password@localhost:5236?schema=SCHEMA_NAME"
```

### Oracle

```bash
--dsn "oracle://user:password@host:port/service_name"
--dsn "user/password@host:port/service_name"
--dsn "oracle://scott:tiger@localhost:1521/ORCL"
--dsn "oracle://scott:tiger@localhost:1521/ORCL?schema=OTHER_USER"
```

### SQL Server

```bash
--dsn "sqlserver://user:password@host:port?database=dbname"
--dsn "user:password@host:port?database=dbname"
--dsn "sqlserver://sa:password@localhost:1433?database=mydb"
```

### MySQL

```bash
--dsn "mysql://user:password@host:port/dbname"
--dsn "user:password@tcp(host:port)/dbname"
--dsn "mysql://root:password@localhost:3306/mydb"
--dsn "mysql://root:password@localhost:3306/mydb?tls=true"
```

### PostgreSQL

```bash
--dsn "postgres://user:password@host:port/dbname"
--dsn "postgresql://user:password@host:port/dbname"
--dsn "postgresql://postgres:password@localhost:5432/mydb"
--dsn "postgres://user:password@host:5432/dbname?sslmode=require"
```

说明：

- PostgreSQL 未指定 `sslmode` 时，工具会自动追加 `sslmode=disable`

## 环境变量

### 示例

```bash
export DB_TYPE=sqlite
export DB_DSN="./app.db"
export EXPORT_OUTPUT=./docs
export EXPORT_FORMATS=markdown,sql
export EXPORT_INCLUDE_VIEWS=true

./schema-export export
```

### 变量列表

| 变量 | 说明 |
|---|---|
| `DB_TYPE` | 数据库类型 |
| `DB_HOST` | 数据库主机 |
| `DB_PORT` | 数据库端口 |
| `DB_DATABASE` | 数据库名；SQLite 下可为数据库文件路径 |
| `DB_USERNAME` | 数据库用户名 |
| `DB_PASSWORD` | 数据库密码 |
| `DB_DSN` | DSN 连接字符串 |
| `DB_SCHEMA` | 数据库 Schema |
| `EXPORT_OUTPUT` | 输出目录 |
| `EXPORT_FORMATS` | 导出格式，逗号分隔 |
| `EXPORT_SPLIT` | 是否按表分文件导出 |
| `EXPORT_INCLUDE_VIEWS` | 是否包含视图 |
| `EXPORT_INCLUDE_PROCEDURES` | 是否包含存储过程 |
| `EXPORT_INCLUDE_FUNCTIONS` | 是否包含函数 |
| `EXPORT_INCLUDE_TRIGGERS` | 是否包含触发器 |
| `EXPORT_INCLUDE_SEQUENCES` | 是否包含序列 |

## CLI 参数参考

### 全局参数

| 参数 | 说明 |
|---|---|
| `-h, --help` | 显示帮助 |
| `-v, --version` | 显示版本 |

### `export` 命令参数

| 参数 | 默认值 | 说明 |
|---|---|---|
| `--type` | `dm` | 数据库类型：`dm`、`oracle`、`sqlserver`、`mysql`、`postgres`、`sqlite` |
| `--host` |  | 数据库主机 |
| `--port` | `0` | 数据库端口 |
| `--database` |  | 数据库名；SQLite 下表示数据库文件路径 |
| `--username` |  | 数据库用户名 |
| `--password` |  | 数据库密码 |
| `--dsn` |  | DSN 连接字符串；SQLite 下可为文件路径、URI 或 `:memory:` |
| `--schema` |  | 数据库 Schema |
| `--output` | `./output` | 输出目录 |
| `--formats` | `markdown` | 导出格式：`markdown`、`sql`、`json`、`yaml` |
| `--split` | `false` | 按表分文件导出 |
| `--tables` |  | 仅导出指定表，支持逗号分隔 |
| `--exclude` |  | 排除指定表，支持逗号分隔 |
| `--patterns` |  | 表名正则匹配模式 |
| `-V, --include-views` | `false` | 包含视图 |
| `-P, --include-procedures` | `false` | 包含存储过程 |
| `-F, --include-functions` | `false` | 包含函数 |
| `-T, --include-triggers` | `false` | 包含触发器 |
| `-S, --include-sequences` | `false` | 包含序列 |

## 输出说明

### 默认文件名

不同格式默认输出文件名如下：

- Markdown：`schema.md`
- SQL：`schema.sql`
- JSON：`schema.json`
- YAML：`schema.yaml`

### `--output` 行为

`--output` 当前按“输出目录”使用：

```bash
--output ./output
```

会在该目录下生成对应格式的默认文件名。

### 文件覆盖

如果输出文件已存在，会直接覆盖，不会自动备份。

分文件导出会先检查本次对象的文件名：忽略大小写并进行 Unicode 规范化后相同的名称，以及表名与视图等对象后缀构成的重名，均会报错并停止写入。名称包含路径分隔符时也会报错。可以去掉 `--split` 使用合并导出；同名的已有文件不会在冲突检查失败时被覆盖。

## 重要说明

### 达梦 / Oracle 的 Schema

对于达梦（DM）和 Oracle，建议显式指定 `--schema`：

- 不指定：导出当前连接用户可见的默认对象
- 指定：导出目标 schema 下的对象（需具备相应权限）

推荐方式：

```bash
# DM
./schema-export export \
  --type dm \
  --dsn "dm://SYSDBA:password@localhost:5236" \
  --schema SC

# Oracle
./schema-export export \
  --type oracle \
  --dsn "oracle://user:password@localhost:1521/ORCL" \
  --schema OTHER_SCHEMA
```

也可在 DSN 中指定：

```bash
./schema-export export --type dm --dsn "dm://SYSDBA:password@localhost:5236?schema=SC"
./schema-export export --type oracle --dsn "oracle://user:password@localhost:1521/ORCL?schema=OTHER_SCHEMA"
```

### macOS 上的达梦服务配置

驱动在 macOS 上不自动读取 `/etc/dm_svc.conf`。需要服务名解析等配置时，请显式设置 `DM_SVC_PATH=/path/to/dm_svc.conf`，或在 DSN 中添加 `svcConfPath` 参数。普通主机和端口直连无需此配置。

## 数据库支持

| 数据库 | 状态 | 驱动 |
|---|---|---|
| 达梦（DM） | ✅ 已支持 | dm-go-driver |
| Oracle | ✅ 已支持 | go-ora |
| SQL Server | ✅ 已支持 | go-mssqldb |
| MySQL | ✅ 已支持 | go-sql-driver/mysql |
| PostgreSQL | ✅ 已支持 | lib/pq |
| SQLite | ✅ 已支持 | modernc.org/sqlite |

## 项目结构

项目采用“CLI 入口 + 应用服务编排 + Inspector/Exporter 扩展点”的分层结构。

### 核心组件

| 组件 | 职责 | 关键位置 |
|---|---|---|
| CLI | 命令解析、命令路由 | `cmd/schema-export/`, `internal/cli/` |
| App | 导出流程编排 | `internal/app/export/` |
| Config | 配置加载与校验 | `internal/config/` |
| Filter | 表过滤规则 | `internal/filter/` |
| Inspector | 统一数据库元数据接口 | `internal/inspector/` |
| Database | 各数据库 Inspector 实现 | `internal/database/` |
| Model | 统一结构模型 | `internal/model/` |
| Exporter | 各输出格式实现 | `internal/exporter/` |
| TUI | 交互式导出界面 | `internal/tui/` |

### 数据流

```text
CLI/TUI -> Config -> App Service -> Inspector -> Database Driver -> Model -> Exporter -> 输出文件
                               -> Filter
```

## 许可证

MIT
