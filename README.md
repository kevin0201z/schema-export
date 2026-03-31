# 数据库结构导出工具

一个支持达梦数据库（DM）、Oracle、MySQL、PostgreSQL 等多种数据库的跨数据库结构导出工具。

可生成 Markdown 和 SQL DDL 格式的数据库结构文档。

## 功能特性

- **多数据库支持**：达梦（DM）、Oracle、MySQL、PostgreSQL（可扩展）
- **多种导出格式**：Markdown、SQL DDL
- **灵活的导出模式**：单文件或按表分文件导出
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
# 导出达梦数据库
./schema-export export \
  --type dm \
  --host localhost \
  --port 5236 \
  --username SYSDBA \
  --password password \
  --database DAMENG \
  --output ./docs

# 导出 Oracle 数据库
./schema-export export \
  --type oracle \
  --host localhost \
  --port 1521 \
  --username scott \
  --password tiger \
  --database ORCL \
  --output ./docs

# 导出指定表
./schema-export export \
  --type dm \
  --host localhost \
  --username SYSDBA \
  --password password \
  --tables users,orders,products \
  --output ./docs

# 导出 SQL DDL 格式
./schema-export export \
  --type dm \
  --host localhost \
  --username SYSDBA \
  --password password \
  --formats markdown,sql \
  --output ./docs

# 按表分文件导出
./schema-export export \
  --type dm \
  --host localhost \
  --username SYSDBA \
  --password password \
  --split \
  --output ./docs

# 指定输出文件名（多格式导出时会自动调整扩展名）
./schema-export export \
  --type dm \
  --dsn "dm://SYSDBA:password@localhost:5236?schema=MC_WYWXZJGLPT" \
  --formats markdown,sql \
  --output ./docs/schema.md
# 将生成：schema.md 和 schema.sql
```

### 使用 DSN

```bash
# 达梦 DSN 格式（推荐指定 schema）
./schema-export export \
  --type dm \
  --dsn "dm://SYSDBA:password@localhost:5236?schema=MC_WYWXZJGLPT" \
  --output ./docs

# Oracle DSN 格式（纯 Go 驱动，无需安装 Oracle Client）
./schema-export export \
  --type oracle \
  --dsn "oracle://scott:tiger@localhost:1521/ORCL" \
  --schema SCHEMA_NAME \
  --output ./docs
```

> **注意**：
> - DSN 中的 `schema` 参数会被自动提取
> - Oracle 使用 `go-ora` 纯 Go 驱动，无需安装 Oracle Instant Client

### 使用环境变量

```bash
export DB_TYPE=dm
export DB_HOST=localhost
export DB_PORT=5236
export DB_USERNAME=SYSDBA
export DB_PASSWORD=password
export DB_DATABASE=DAMENG
export EXPORT_OUTPUT=./docs
export EXPORT_FORMATS=markdown,sql
export EXPORT_SPLIT=true

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

# 导出指定 schema 下以 GGSY_ 开头的表
./schema-export export \
  --type dm \
  --dsn "dm://SYSDBA:password@localhost:5236?schema=MC_WYWXZJGLPT" \
  --patterns "^GGSY_"
```

## CLI 参考

### 全局参数

| 参数              | 说明   |
| --------------- | ---- |
| `-h, --help`    | 显示帮助 |
| `-v, --version` | 显示版本 |

### export 命令参数

| 参数           | 默认值      | 说明                              |
| ------------ | -------- | ------------------------------- |
| `--type`     | dm       | 数据库类型（dm、oracle、mysql、postgres） |
| `--host`     | <br />   | 数据库主机                           |
| `--port`     | 0        | 数据库端口                           |
| `--database` | <br />   | 数据库名                            |
| `--username` | <br />   | 数据库用户名                          |
| `--password` | <br />   | 数据库密码                           |
| `--dsn`      | <br />   | DSN 连接字符串                       |
| `--schema`   | <br />   | 数据库 Schema                      |
| `--output`   | ./output | 输出目录或文件路径                       |
| `--formats`  | markdown | 导出格式（逗号分隔：markdown,sql）         |
| `--split`    | false    | 按表分文件导出                         |
| `--tables`   | <br />   | 要导出的表（逗号分隔）                     |
| `--exclude`  | <br />   | 要排除的表（逗号分隔）                     |
| `--patterns` | <br />   | 表名正则匹配模式                        |

## 环境变量

| 变量               | 说明                |
| ---------------- | ----------------- |
| `DB_TYPE`        | 数据库类型             |
| `DB_HOST`        | 数据库主机             |
| `DB_PORT`        | 数据库端口             |
| `DB_DATABASE`    | 数据库名              |
| `DB_USERNAME`    | 数据库用户名            |
| `DB_PASSWORD`    | 数据库密码             |
| `DB_DSN`         | DSN 连接字符串         |
| `DB_SCHEMA`      | 数据库 Schema        |
| `EXPORT_OUTPUT`  | 输出目录              |
| `EXPORT_FORMATS` | 导出格式（逗号分隔）        |
| `EXPORT_SPLIT`   | 分文件导出（true/false） |

## 输出说明

### 输出路径

`--output` 参数支持两种形式：

1. **输出目录**：指定目录路径，文件名使用默认值
   ```bash
   --output ./docs
   # 生成：./docs/schema.md 或 ./docs/schema.sql
   ```
2. **输出文件**：指定完整文件路径，扩展名会根据格式自动调整
   ```bash
   --output ./docs/tables.md
   # Markdown 格式生成：./docs/tables.md
   # SQL 格式生成：./docs/tables.sql
   ```

### 文件覆盖

如果输出文件已存在，会直接覆盖，不会报错或创建备份。

## 输出格式

### Markdown 格式

```markdown
# 数据库结构文档

总表数：3

## 目录
- [users](#table-users)
- [orders](#table-orders)
...

## 表：users

**说明：** 用户信息表

### 字段

| 字段 | 类型 | 可空 | 默认值 | 注释 |
|------|------|------|--------|------|
| id | BIGINT PK AI | 否 | | 主键 |
| username | VARCHAR(50) | 否 | | 用户名 |
| email | VARCHAR(100) | 是 | | 邮箱地址 |
...

### 索引

| 索引 | 类型 | 字段 |
|------|------|------|
| idx_username | UNIQUE | username |
...

### 外键

| 外键 | 字段 | 引用 | 删除规则 |
|------|------|------|----------|
| fk_user_role | role_id | roles(id) | CASCADE |
```

### SQL DDL 格式

```sql
-- 数据库结构 DDL
-- 总表数：3

-- 表：users
-- 用户信息表

CREATE TABLE users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) NOT NULL,
    email VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- users 表的索引
CREATE UNIQUE INDEX idx_username ON users (username);

-- users 表的外键
ALTER TABLE users ADD CONSTRAINT fk_user_role FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE;
```

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
  --schema MC_WYWXZJGLPT

# Oracle 数据库
./schema-export export --type oracle \
  --dsn "oracle://user:password@localhost:1521/ORCL" \
  --schema OTHER_SCHEMA
```

**替代方式（直接在 DSN 中指定）：**

```bash
# 达梦数据库
./schema-export export --type dm \
  --dsn "dm://SYSDBA:password@localhost:5236?schema=MC_WYWXZJGLPT"

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
| MySQL      | 🚧 计划中 | -                     |
| PostgreSQL | 🚧 计划中 | -                     |
| SQL Server | 🚧 计划中 | -                     |
| SQLite     | 🚧 计划中 | -                     |

## 架构

```
schema-export/
├── cmd/schema-export/          # CLI 入口
│   └── main.go
├── internal/
│   ├── config/                 # 配置管理
│   ├── database/               # 数据库驱动
│   │   ├── dm/                 # 达梦驱动
│   │   ├── oracle/             # Oracle 驱动
│   │   └── base.go             # 基础 Inspector
│   ├── inspector/              # Inspector 接口
│   ├── model/                  # 数据模型
│   ├── exporter/               # 导出器
│   │   ├── markdown/           # Markdown 导出器
│   │   └── sql/                # SQL 导出器
│   └── cli/                    # CLI 命令
└── README.md
```

## 许可证

MIT 许可证
