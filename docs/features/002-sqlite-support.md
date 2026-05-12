# 功能：支持 SQLite 数据库结构导出

## 1. 功能目标

为现有 `schema-export export` 和 `schema-export tui` 增加 SQLite 数据库支持，使用户可以从本地 SQLite 数据库文件导出表、字段、索引、外键、CHECK 约束、视图和触发器等结构信息，并继续生成 Markdown、SQL、JSON、YAML 格式的输出文件。

第一版目标是接入 SQLite 的读取能力，不改变现有 Inspector、Exporter、过滤器和输出文件的整体架构。

## 2. 背景与问题

当前项目已经支持 `dm`、`oracle`、`sqlserver`、`mysql`、`postgres`，但还不支持常见的嵌入式数据库 SQLite。SQLite 与现有服务端数据库的差异较大：

- SQLite 使用本地文件或内存数据库，不需要 host、port、username、password。
- SQLite 没有存储过程、独立函数、序列对象。
- SQLite 元数据主要来自 `sqlite_master`、`pragma_table_info`、`pragma_index_list`、`pragma_foreign_key_list` 等 PRAGMA 查询。
- SQLite 类型系统更宽松，字段类型、默认值、约束表达式需要保持原始定义，避免过度转换。

因此本功能应作为独立数据库 Inspector 接入，而不是复用 MySQL 或 PostgreSQL 的实现。

## 3. 功能影响分析

| 影响面 | 是否涉及 | 需要读取或更新 |
|---|---|---|
| 前端 | 是 | TUI 数据库类型选项、SQLite 连接表单提示和字段说明 |
| 后端 | 是 | 新增 SQLite Inspector、驱动注册、配置校验和 SQL 方言 |
| 数据库 | 否 | 不修改项目自身数据库；只读取用户提供的 SQLite 文件元数据 |
| API | 否 | 不新增 HTTP API |
| 权限 | 否 | 不修改应用权限模型；只依赖本地文件读取权限 |
| 环境变量 | 是 | 继续使用 `DB_TYPE`、`DB_DSN`、`DB_DATABASE`；可约定 SQLite 文件路径映射 |
| 第三方服务 | 否 | 不接入外部服务；但会新增 SQLite Go 驱动依赖 |

## 4. 风险等级

本计划文档阶段为 `Continue With Note`。

后续实现阶段包含新增数据库驱动依赖，属于 `Ask First`。推荐默认选择 `modernc.org/sqlite`，原因是纯 Go、无 CGO，和当前工具的跨平台构建目标更匹配。备选方案是 `github.com/mattn/go-sqlite3`，成熟度高但需要 CGO、gcc 和更复杂的跨平台构建配置。

## 5. MVP 范围

第一版必须包含：

- 新增数据库类型：`sqlite`。
- 支持通过 `--dsn` 指定 SQLite 文件路径或 URI。
- 支持通过 `--database` 指定 SQLite 文件路径，作为分离参数模式下的文件路径入口。
- 支持内存数据库 DSN：`:memory:`。
- 支持导出普通表。
- 支持导出字段、字段类型、是否可空、默认值、主键、自增标记。
- 支持导出索引，包括唯一索引和索引字段顺序。
- 支持导出外键。
- 支持导出 CHECK 约束，至少从建表 SQL 中提取表级 CHECK 表达式。
- 支持导出视图及视图定义。
- 支持导出触发器及触发器定义。
- 支持现有表过滤规则：`--tables`、`--exclude`、`--patterns`。
- 支持现有输出格式：`markdown`、`sql`、`json`、`yaml`。
- CLI、TUI、README、测试计划同步体现 SQLite 支持。

## 6. 非目标

第一版明确不做：

- 不支持 SQLite 加密数据库或 SQLCipher。
- 不支持附加数据库 `ATTACH DATABASE` 的跨库导出。
- 不支持虚拟表的专用元数据解析；可在表列表中跳过或标记为普通对象，具体以实现验证为准。
- 不支持 FTS、RTree 等扩展模块的专属 DDL 重建能力。
- 不支持存储过程、独立函数、序列导出；SQLite 无对应对象。
- 不新增配置文件持久化能力。
- 不改变其他数据库的连接参数语义。

## 7. 用户流程

CLI 推荐用法：

```bash
schema-export export --type sqlite --dsn ./app.db --formats markdown,sql
```

或：

```bash
schema-export export --type sqlite --database ./app.db --include-views --include-triggers
```

TUI 流程：

1. 用户执行 `schema-export tui`。
2. 用户选择数据库类型 `sqlite`。
3. 用户选择 DSN 或分离参数连接。
4. DSN 模式下填写 SQLite 文件路径、URI 或 `:memory:`。
5. 分离参数模式下只需要填写数据库文件路径，不需要 host、port、username、password。
6. 用户选择导出内容、过滤规则和输出格式。
7. TUI 展示确认页并执行导出。

## 8. 连接与配置规则

| 输入方式 | 字段 | SQLite 解释 |
|---|---|---|
| CLI DSN | `--dsn` | SQLite 文件路径、URI 或 `:memory:` |
| CLI 分离参数 | `--database` | SQLite 文件路径 |
| 环境变量 DSN | `DB_DSN` | SQLite 文件路径、URI 或 `:memory:` |
| 环境变量分离参数 | `DB_DATABASE` | SQLite 文件路径 |
| TUI DSN | `database.dsn` | SQLite 文件路径、URI 或 `:memory:` |
| TUI 分离参数 | `database.database` | SQLite 文件路径 |

SQLite 类型下的配置校验规则：

- `Config.Database.Type == "sqlite"` 时，不要求 `Host` 和 `Username`。
- `DSN` 非空时优先使用 `DSN`。
- `DSN` 为空时使用 `Database` 作为 SQLite 文件路径。
- `DSN` 和 `Database` 都为空时返回配置错误。
- `Port`、`Username`、`Password`、`Schema`、`SSLMode` 对 SQLite 不生效。

## 9. 元数据读取方案

SQLite Inspector 建议新增目录：

```text
internal/database/sqlite/
├── inspector.go
└── inspector_test.go
```

核心查询建议：

| 元数据 | 查询来源 | 说明 |
|---|---|---|
| 表列表 | `sqlite_master` / `sqlite_schema` | `type = 'table'`，排除 `sqlite_%` 系统表 |
| 视图列表 | `sqlite_master` / `sqlite_schema` | `type = 'view'`，读取 `sql` 作为定义 |
| 触发器列表 | `sqlite_master` / `sqlite_schema` | `type = 'trigger'`，按表名过滤 |
| 字段 | `PRAGMA table_info(table)` | 字段名、类型、notnull、默认值、主键序号 |
| 自增 | 建表 SQL | 识别 `INTEGER PRIMARY KEY AUTOINCREMENT` |
| 索引 | `PRAGMA index_list(table)` + `PRAGMA index_info(index)` | 排除 SQLite 自动索引，保留唯一性 |
| 外键 | `PRAGMA foreign_key_list(table)` | 合并同一外键多字段 |
| CHECK | 建表 SQL | MVP 可用正则/轻量解析提取 CHECK 片段，需测试复杂表达式 |
| 序列 | 不支持 | 返回空列表 |
| 存储过程 | 不支持 | 返回空列表 |
| 函数 | 不支持 | 返回空列表 |

注意：SQLite 标识符可能包含空格、引号和特殊字符，所有动态 PRAGMA 查询必须正确引用表名或通过安全 helper 生成标识符，避免直接拼接未转义的用户输入。

## 10. SQL 方言方案

新增 `SQLiteDialect`，注册到 `GetDialect("sqlite")`。

第一版方言行为：

- 标识符使用双引号引用，例如 `"users"`。
- 类型尽量保留 SQLite 原始类型，不强制映射为其他数据库类型。
- `INTEGER PRIMARY KEY` 字段保留主键语义。
- `AUTOINCREMENT` 仅在源表定义存在时输出。
- 字段注释、表注释、视图注释返回空字符串，SQLite 无原生 comment 语法。
- CHECK 约束支持内联输出。
- 视图和触发器优先使用 Inspector 读取到的原始 SQL 定义。

## 11. 文件改动计划

预计需要修改或新增：

| 文件 | 计划 |
|---|---|
| `go.mod`, `go.sum` | 新增 SQLite 驱动依赖，推荐 `modernc.org/sqlite` |
| `cmd/schema-export/main.go` | 导入 SQLite inspector 注册包，CLI 帮助文案加入 `sqlite` |
| `internal/config/config.go` | SQLite 类型下放宽 host/username 校验，更新注释 |
| `internal/database/sqlite/inspector.go` | 新增 SQLite Inspector 实现 |
| `internal/database/sqlite/inspector_test.go` | 覆盖 DSN、表、字段、索引、外键、视图、触发器 |
| `internal/database_test` | 增加跨包 SQLite 注册或 DSN 测试 |
| `internal/exporter/sql/dialect_factory.go` | 注册 SQLite SQL 方言 |
| `internal/exporter/sql/dialect_sqlite.go` | 新增 SQLite 方言 |
| `internal/exporter/sql/dialect_test.go` | 增加 SQLite 方言测试 |
| `internal/tui/types.go` | 数据库类型选项加入 `sqlite` |
| `internal/tui/mapper.go` | SQLite 分离参数模式下跳过 host/username/port |
| `internal/tui/mapper_test.go` | 覆盖 SQLite 配置映射 |
| `README.md` | 增加 SQLite 用法示例和支持列表 |
| `docs/TEST_PLAN.md` | 增加 SQLite 自动化和手动验证项 |

## 12. 开发步骤

1. 确认 SQLite 驱动选型。
2. 新增 `sqlite` 类型常量和文档说明。
3. 调整 `Config.Validate()`，允许 SQLite 使用 `DSN` 或 `Database` 文件路径连接。
4. 新增 SQLite Inspector 骨架、工厂注册和 DSN 构建逻辑。
5. 实现 `Connect()`，使用 `database/sql` 打开 SQLite 数据库。
6. 实现 `GetTables()` 和 `GetTable()`。
7. 实现字段、索引、外键、CHECK 约束读取。
8. 实现视图和触发器读取。
9. 对存储过程、函数、序列返回空列表，并在文档中说明。
10. 新增 SQLite SQL 方言。
11. 更新 CLI 和 TUI 的数据库类型选项、提示文案、默认端口逻辑。
12. 更新 README 和测试计划。
13. 增加单元测试和集成式临时 SQLite 文件测试。
14. 运行 `go test ./...`。
15. 使用临时 SQLite 文件做手动导出验证。

## 13. 测试要求

自动化测试：

- `BuildDSN` 覆盖文件路径、URI、`:memory:`、`Database` fallback。
- `Config.Validate()` 覆盖 SQLite 不需要 host/username。
- `GetTables()` 排除 `sqlite_%` 系统表。
- `GetColumns()` 覆盖 nullable、default、primary key、autoincrement。
- `GetIndexes()` 覆盖普通索引、唯一索引、复合索引、自动索引过滤。
- `GetForeignKeys()` 覆盖单字段和复合外键。
- `GetCheckConstraints()` 覆盖字段级和表级 CHECK。
- `GetViews()` 覆盖视图定义导出。
- `GetTriggers()` 覆盖按表读取触发器。
- SQL 方言覆盖双引号引用、SQLite 类型保留、注释为空。
- TUI 映射覆盖 `sqlite` DSN 和分离参数模式。

手动验证：

```bash
tmpdir=$(mktemp -d)
db="$tmpdir/sample.db"
sqlite3 "$db" <<'SQL'
PRAGMA foreign_keys = ON;
CREATE TABLE users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL CHECK (length(name) > 0),
  email TEXT UNIQUE,
  created_at TEXT DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE orders (
  id INTEGER PRIMARY KEY,
  user_id INTEGER NOT NULL,
  amount REAL CHECK (amount >= 0),
  FOREIGN KEY (user_id) REFERENCES users(id)
);
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE VIEW active_users AS SELECT id, name FROM users;
CREATE TRIGGER trg_orders_amount_check
BEFORE INSERT ON orders
WHEN NEW.amount < 0
BEGIN
  SELECT RAISE(ABORT, 'amount must be non-negative');
END;
SQL

go run ./cmd/schema-export export --type sqlite --dsn "$db" --include-views --include-triggers --formats markdown,sql,json,yaml --output "$tmpdir/out"
```

验收时检查：

- `schema.md` 包含 `users`、`orders`、字段、索引、外键、CHECK、视图、触发器。
- `schema.sql` 使用 SQLite 方言，标识符以双引号引用。
- `schema.json` 和 `schema.yaml` 包含结构化的视图和触发器数据。
- 不生成存储过程、函数、序列内容。

## 14. 验收标准

- [ ] `schema-export export --type sqlite --dsn ./sample.db` 可以成功导出。
- [ ] `schema-export export --type sqlite --database ./sample.db` 可以成功导出。
- [ ] `schema-export tui` 可选择 `sqlite`。
- [ ] SQLite 不要求 host、port、username、password。
- [ ] 表、字段、索引、外键、CHECK、视图、触发器导出正确。
- [ ] 存储过程、函数、序列在 SQLite 下返回空集合且不报错。
- [ ] 表过滤规则对 SQLite 生效。
- [ ] Markdown、SQL、JSON、YAML 输出均可生成。
- [ ] 现有数据库类型的测试不回归。
- [ ] `go test ./...` 通过。

## 15. 待确认问题

- 已确认默认使用 `modernc.org/sqlite` 作为 SQLite 驱动。
- 是否需要支持 SQLCipher 或加密 SQLite 文件；若需要，应作为独立安全敏感需求评审。
- 是否需要支持 `ATTACH DATABASE` 后的多数据库导出；第一版建议不做。
- 是否需要专门展示 SQLite 虚拟表、FTS 表和 RTree 表；第一版建议只处理普通表和视图。
- 是否允许在没有本机 `sqlite3` CLI 的环境中只依赖 Go 自动化测试完成验证。

## 16. 参考

- Go package `modernc.org/sqlite`: https://pkg.go.dev/modernc.org/sqlite
- Go package `github.com/mattn/go-sqlite3`: https://pkg.go.dev/github.com/mattn/go-sqlite3
- SQLite 文档入口: https://sqlite.org/docs.html
