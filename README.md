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
  --dsn "dm://SYSDBA:password@localhost:5236?schema=sc" \
  --formats markdown,sql \
  --output ./docs/schema.md
# 将生成：schema.md 和 schema.sql
```

### 使用 DSN

```bash
# 达梦 DSN 格式（推荐指定 schema）
./schema-export export \
  --type dm \
  --dsn "dm://SYSDBA:password@localhost:5236?schema=sc" \
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

# 导出指定 schema 下以 tb_ 开头的表
./schema-export export \
  --type dm \
  --dsn "dm://SYSDBA:password@localhost:5236?schema=sc" \
  --patterns "^tb_"
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

**总表数:** 2

## 概览

### 表清单
- **users**: 用户信息表
- **orders**: 订单表

### 目录
- [users](#表-users)
- [orders](#表-orders)

## 表: users

**描述:** 用户信息表

**表类型:** TABLE

### 基本信息

| 属性 | 值 |
|------|-----|
| 表名 | users |
| 类型 | TABLE |
| 描述 | 用户信息表 |
| 字段数 | 5 |
| 索引数 | 2 |
| 外键数 | 0 |

### 字段详情

| 字段名 | 数据类型 | 长度/精度 | 可空 | 默认值 | 约束 | 注释 |
|--------|----------|-----------|------|--------|------|------|
| id | BIGINT | - | 否 | - | 主键 非空 | 主键ID |
| username | VARCHAR | 50 | 否 | - | 非空 | 用户名 |
| email | VARCHAR | 100 | 是 | - |  | 邮箱地址 |
| created_at | TIMESTAMP | - | 是 | CURRENT_TIMESTAMP |  | 创建时间 |
| status | INT | - | 是 | 1 |  | 用户状态 |

### 约束

#### 主键
- **id**: BIGINT - 主键ID

#### 唯一约束
- **idx_username**: username

### 索引

| 索引名 | 类型 | 字段 | 是否唯一 | 是否主键 |
|--------|------|------|----------|----------|
| PRIMARY | PRIMARY | id | 否 | 是 |
| idx_username | UNIQUE | username | 是 | 否 |

### 外键

*未定义外键*

此表没有与其他表的外键关联关系。

---

## 表: orders

**描述:** 订单表

**表类型:** TABLE

### 基本信息

| 属性 | 值 |
|------|-----|
| 表名 | orders |
| 类型 | TABLE |
| 描述 | 订单表 |
| 字段数 | 6 |
| 索引数 | 2 |
| 外键数 | 1 |

### 字段详情

| 字段名 | 数据类型 | 长度/精度 | 可空 | 默认值 | 约束 | 注释 |
|--------|----------|-----------|------|--------|------|------|
| id | BIGINT | - | 否 | - | 主键 非空 | 订单ID |
| user_id | BIGINT | - | 否 | - | 非空 | 用户ID |
| order_no | VARCHAR | 32 | 否 | - | 非空 | 订单编号 |
| amount | DECIMAL | 10,2 | 是 | 0.00 |  | 订单金额 |

### 外键

| 外键名 | 字段 | 引用表 | 引用字段 | 删除规则 | 更新规则 |
|--------|------|--------|----------|----------|----------|
| fk_order_user | user_id | users | id | CASCADE | NO ACTION |

#### 关联关系
- **fk_order_user**: `orders.user_id` → `users.id` (删除时CASCADE)

---
```

### SQL DDL 格式

```sql
-- ========================================================
-- Database Schema DDL
-- Generated by schema-export tool
-- ========================================================

-- Total Tables: 2

-- --------------------------------------------------------
-- Table List
-- --------------------------------------------------------
-- 1. users - 用户信息表
-- 2. orders - 订单表

-- --------------------------------------------------------
-- Table: users
-- --------------------------------------------------------
-- Description: 用户信息表
-- Type: TABLE
-- Columns: 5
-- Indexes: 2
-- Foreign Keys: 0

CREATE TABLE users (
    -- id: 主键ID
    id BIGINT PRIMARY KEY NOT NULL,
    -- username: 用户名
    username VARCHAR(50) NOT NULL,
    -- email: 邮箱地址
    email VARCHAR(100),
    -- created_at: 创建时间
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    -- status: 用户状态
    status INT DEFAULT 1
);

-- --------------------------------------------------------
-- Indexes for users
-- --------------------------------------------------------
-- Primary Key: PRIMARY on (id)
CREATE UNIQUE INDEX idx_username ON users (username);

-- --------------------------------------------------------
-- Table: orders
-- --------------------------------------------------------
-- Description: 订单表
-- Type: TABLE
-- Columns: 6
-- Indexes: 2
-- Foreign Keys: 1

CREATE TABLE orders (
    -- id: 订单ID
    id BIGINT PRIMARY KEY NOT NULL,
    -- user_id: 用户ID
    user_id BIGINT NOT NULL,
    -- order_no: 订单编号
    order_no VARCHAR(32) NOT NULL,
    -- amount: 订单金额
    amount DECIMAL(10,2) DEFAULT 0.00,
    -- status: 订单状态
    status INT DEFAULT 0,
    -- created_at: 创建时间
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- --------------------------------------------------------
-- Foreign Keys for orders
-- --------------------------------------------------------
-- Relationship: orders.user_id -> users.id
ALTER TABLE orders ADD CONSTRAINT fk_order_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
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
| MySQL      | 🚧 计划中 | -                     |
| PostgreSQL | 🚧 计划中 | -                     |
| SQL Server | 🚧 计划中 | -                     |
| SQLite     | 🚧 计划中 | -                     |

## 架构

```
schema-export/
├── cmd/schema-export/          # CLI 入口
│   └── main.go                 # 程序入口，初始化 CLI
├── internal/
│   ├── config/                 # 配置管理
│   │   ├── config.go           # 配置结构定义、环境变量读取
│   │   └── filter.go           # 表过滤逻辑（包含/排除/正则）
│   ├── database/               # 数据库驱动
│   │   ├── dm/                 # 达梦驱动
│   │   │   └── inspector.go    # 达梦数据库 Inspector 实现
│   │   ├── oracle/             # Oracle 驱动
│   │   │   └── inspector.go    # Oracle 数据库 Inspector 实现
│   │   └── base.go             # 基础 Inspector，提供通用功能
│   ├── inspector/              # Inspector 接口
│   │   └── interface.go        # Inspector 接口定义
│   ├── model/                  # 数据模型
│   │   ├── table.go            # 表结构模型
│   │   ├── column.go           # 字段结构模型
│   │   ├── index.go            # 索引结构模型
│   │   └── foreign_key.go      # 外键结构模型
│   ├── exporter/               # 导出器
│   │   ├── interface.go        # Exporter 接口定义
│   │   ├── markdown/           # Markdown 导出器
│   │   │   └── exporter.go     # Markdown 格式导出实现
│   │   └── sql/                # SQL 导出器
│   │       └── exporter.go     # SQL DDL 格式导出实现
│   └── cli/                    # CLI 命令
│       ├── root.go             # 根命令定义
│       └── export.go           # export 子命令实现
├── internal/dm-go-driver/      # 达梦数据库 Go 驱动（本地）
│   └── dm/                     # 达梦驱动源码
├── docs/                       # 文档输出目录（自动生成）
├── go.mod                      # Go 模块定义
├── go.sum                      # Go 依赖校验
├── Makefile                    # 构建脚本
└── README.md                   # 项目说明文档
```

### 核心组件说明

| 组件 | 职责 | 关键文件 |
|------|------|----------|
| **CLI** | 命令行参数解析、命令路由 | `cmd/schema-export/main.go`, `internal/cli/` |
| **Config** | 配置管理、环境变量、表过滤 | `internal/config/` |
| **Inspector** | 数据库元数据查询接口 | `internal/inspector/interface.go` |
| **Database** | 各数据库 Inspector 实现 | `internal/database/dm/`, `internal/database/oracle/` |
| **Model** | 数据模型定义 | `internal/model/` |
| **Exporter** | 导出格式实现 | `internal/exporter/markdown/`, `internal/exporter/sql/` |

### 数据流向

```
CLI 参数 → Config → Inspector → Database Driver → Model → Exporter → 输出文件
                ↓
            Filter (表过滤)
```

1. **CLI** 解析用户输入的参数
2. **Config** 整合 CLI 参数和环境变量
3. **Inspector** 根据数据库类型创建对应的驱动实例
4. **Database Driver** 查询数据库元数据（表、字段、索引、外键）
5. **Filter** 根据配置过滤表
6. **Model** 存储查询结果
7. **Exporter** 将模型转换为指定格式（Markdown/SQL）
8. **输出文件** 写入磁盘

## 许可证

MIT 许可证
