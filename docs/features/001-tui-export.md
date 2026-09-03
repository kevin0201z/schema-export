# 功能：通过 TUI 界面导出数据库结构

## 1. 功能目标

为现有 `schema-export export` 命令新增终端图形界面（TUI）入口，让用户可以在不记忆完整命令行参数的情况下，通过交互式表单完成数据库连接配置、导出范围选择、输出格式选择和导出执行。

第一版目标是复用现有 CLI 配置模型和导出服务，不改变数据库 Inspector、Exporter 和输出文件规则。

## 2. 背景与问题

当前项目已经支持通过 CLI 参数和环境变量导出数据库结构，但在以下场景中使用成本较高：

- 数据库类型、DSN、Schema、输出格式、对象类型等参数较多，首次使用者容易漏填或填错。
- 用户需要反复调整导出范围时，每次都要重新拼接命令。
- 部分用户在服务器环境中不能使用 Web UI，但可以使用交互式终端。

TUI 入口用于补足交互式配置体验，同时保持工具的命令行可脚本化能力。

## 3. 功能影响分析

| 影响面 | 是否涉及 | 需要读取或更新 |
|---|---|---|
| 前端 | 是 | 新增终端 TUI 页面与交互状态，属于终端界面，不涉及浏览器前端 |
| 后端 | 是 | 复用 `internal/app/export` 导出编排，可能新增 TUI 到配置的适配层 |
| 数据库 | 否 | 不修改业务数据库结构，只读取目标库元数据 |
| API | 否 | 不新增 HTTP API |
| 权限 | 否 | 不修改应用权限模型；目标数据库权限仍由连接账号决定 |
| 环境变量 | 是 | 读取现有环境变量作为默认值，不新增环境变量 |
| 第三方服务 | 否 | 不接入外部服务；如实现时新增 TUI 库需单独确认 |

## 4. 风险等级

当前文档编写阶段为 `Continue With Note`。

后续实现阶段如需要新增 TUI 依赖，属于 `Ask First`，需要确认依赖选型、维护风险和兼容性；如涉及凭据保存、配置文件落盘或日志策略调整，也需要先确认安全方案。

## 5. 用户流程

1. 用户执行 `schema-export tui` 进入交互式界面。
2. TUI 读取现有环境变量和默认配置，预填数据库类型、输出目录、导出格式等字段。
3. 用户选择数据库类型。
4. 用户选择连接方式：
   - DSN 连接。
   - 分离参数连接：host、port、database、username、password、schema。
5. 用户填写连接信息。
6. 用户选择导出内容：
   - 表。
   - 视图。
   - 存储过程。
   - 函数。
   - 触发器。
   - 序列。
7. 用户配置表过滤规则：
   - 指定表名。
   - 排除表名。
   - 正则匹配。
8. 用户选择导出格式：
   - Markdown。
   - SQL DDL。
   - JSON。
   - YAML。
9. 用户配置输出路径和是否按表分文件导出。
10. 用户在确认页检查配置摘要。
11. 用户确认后执行导出。
12. TUI 展示执行进度、成功结果或失败信息。
13. 导出完成后，用户可以退出、返回修改配置或再次执行。

## 6. MVP 范围

第一版必须包含：

- 新增 `schema-export tui` 命令入口。
- 支持现有数据库类型选项：`dm`、`oracle`、`sqlserver`、`mysql`、`postgres`。
- 支持 DSN 和分离参数两种连接方式。
- 支持现有导出格式：`markdown`、`sql`、`json`、`yaml`。
- 支持现有导出选项：输出路径、分文件导出、指定表、排除表、正则匹配。
- 支持现有对象开关：视图、存储过程、函数、触发器、序列。
- 支持从现有环境变量加载默认值。
- 密码输入必须隐藏显示。
- 执行导出时复用现有 `config.Config`、校验逻辑和导出服务。
- 导出失败时展示可读错误，并允许用户返回修改配置。

## 7. 非目标

第一版明确不做：

- 不提供浏览器 Web UI。
- 不保存连接配置到本地文件。
- 不保存密码、DSN 或任何凭据。
- 不新增数据库元数据预览页。
- 不在 TUI 中编辑导出的文件内容。
- 不改变现有 `export` 命令参数语义。
- 不改变输出文件命名、覆盖和格式规则。
- 不新增 HTTP API。
- 不新增数据库表或迁移。

## 8. 页面与交互设计

TUI 建议采用分步向导式界面，避免一次性展示过多字段。

| 页面 | 目标 | 主要控件 | 关键行为 |
|---|---|---|---|
| 欢迎页 | 说明当前进入 TUI 导出流程 | 开始、退出 | 展示版本信息和当前默认输出目录 |
| 数据库类型页 | 选择目标数据库 | 单选列表 | 默认值来自 `DB_TYPE` 或 `dm` |
| 连接方式页 | 选择 DSN 或分离参数 | 单选列表 | 切换后进入对应表单 |
| DSN 表单页 | 填写 DSN、Schema | 文本输入、密码遮罩策略说明 | DSN 不落盘，不在摘要中完整展示密码部分 |
| 分离参数表单页 | 填写 host、port、database、username、password、schema | 文本输入、数字输入、密码输入 | password 必须隐藏输入 |
| 导出内容页 | 选择对象类型 | 多选列表 | 表默认启用；其他对象默认按现有配置关闭 |
| 表过滤页 | 配置 tables、exclude、patterns | 文本输入 | 多值使用逗号分隔，复用现有解析规则 |
| 输出设置页 | 配置 output、formats、split | 文本输入、多选、开关 | 默认 output 为 `./output`，默认格式为 markdown |
| 确认页 | 展示脱敏后的配置摘要 | 确认、返回修改、退出 | 不展示明文密码，不展示 DSN 中的密码 |
| 执行页 | 展示导出进度和结果 | 进度文本、完成操作 | 成功展示输出路径；失败展示错误和返回入口 |

## 9. 业务规则

- TUI 生成的配置必须映射为现有 `config.Config`。
- 配置校验必须复用现有 `Config.Validate()`，避免 TUI 和 CLI 校验逻辑分叉。
- TUI 执行导出必须复用现有导出服务，避免重复实现导出编排。
- 环境变量只作为默认值来源，用户在 TUI 中修改后仅影响本次执行。
- 如果用户选择 DSN，分离参数中的 host、port、database、username、password 不应参与本次连接。
- 如果用户选择分离参数，必须至少满足现有校验规则：有数据库类型、有 host、有 username。
- `schema` 继续遵循现有规则：DM 和 Oracle 规范化为大写；DSN 中带 schema 且未显式填写时可被提取。
- 多选导出格式最终必须转换为 `[]string`，并继续由现有格式规范化逻辑处理。
- 表过滤输入继续沿用逗号分隔规则，空输入等同于未配置。
- 用户取消导出时，不应创建输出文件。
- 导出文件覆盖行为沿用现有规则：目标文件已存在时直接覆盖。
  分文件导出须先通过 [macOS 兼容性修复](003-macos-compatibility.md) 的批次文件名冲突检查；失败时不覆盖已有文件，TUI 展示导出错误。

## 10. 安全规则

- 密码输入必须使用隐藏输入控件。
- 确认页、错误信息和日志不得输出明文密码。
- DSN 中的密码在确认页必须脱敏，例如 `dm://user:******@host:5236?...`。
- TUI 不得默认保存连接配置、DSN 或密码。
- 如后续需要“保存配置”能力，必须作为独立需求重新做安全评审。
- 示例文档中不得写入真实数据库密码或生产 DSN。
- 导出目标路径由用户显式确认；TUI 不应静默写入用户未确认的目录。

## 11. 字段映射

| 业务含义 | TUI 字段 | CLI 参数 | 环境变量 | 后端字段 | 说明 |
|---|---|---|---|---|---|
| 数据库类型 | database.type | `--type` | `DB_TYPE` | `Config.Database.Type` | 支持 dm、oracle、sqlserver、mysql、postgres |
| 数据库主机 | database.host | `--host` | `DB_HOST` | `Config.Database.Host` | 分离参数连接时使用 |
| 数据库端口 | database.port | `--port` | `DB_PORT` | `Config.Database.Port` | 数字输入 |
| 数据库名称 | database.database | `--database` | `DB_DATABASE` | `Config.Database.Database` | 按数据库类型解释 |
| 用户名 | database.username | `--username` | `DB_USERNAME` | `Config.Database.Username` | 分离参数连接时使用 |
| 密码 | database.password | `--password` | `DB_PASSWORD` | `Config.Database.Password` | TUI 必须隐藏输入和脱敏展示 |
| DSN | database.dsn | `--dsn` | `DB_DSN` | `Config.Database.DSN` | DSN 连接时使用，摘要中脱敏 |
| Schema | database.schema | `--schema` | `DB_SCHEMA` | `Config.Database.Schema` | DM/Oracle 会规范化为大写 |
| 输出路径 | export.output | `--output` | `EXPORT_OUTPUT` | `Config.Export.OutputDir` | 目录或文件路径 |
| 导出格式 | export.formats | `--formats` | `EXPORT_FORMATS` | `Config.Export.Formats` | 多选，默认 markdown |
| 分文件导出 | export.split | `--split` | `EXPORT_SPLIT` | `Config.Export.SplitFiles` | 布尔开关 |
| 指定表 | export.tables | `--tables` | 无 | `Config.Export.Tables` | 逗号分隔 |
| 排除表 | export.exclude | `--exclude` | 无 | `Config.Export.Exclude` | 逗号分隔 |
| 表名正则 | export.patterns | `--patterns` | 无 | `Config.Export.Patterns` | 逗号分隔 |
| 包含视图 | export.includeViews | `--include-views` | `EXPORT_INCLUDE_VIEWS` | `Config.Export.IncludeViews` | 布尔开关 |
| 包含存储过程 | export.includeProcedures | `--include-procedures` | `EXPORT_INCLUDE_PROCEDURES` | `Config.Export.IncludeProcedures` | 布尔开关 |
| 包含函数 | export.includeFunctions | `--include-functions` | `EXPORT_INCLUDE_FUNCTIONS` | `Config.Export.IncludeFunctions` | 布尔开关 |
| 包含触发器 | export.includeTriggers | `--include-triggers` | `EXPORT_INCLUDE_TRIGGERS` | `Config.Export.IncludeTriggers` | 布尔开关 |
| 包含序列 | export.includeSequences | `--include-sequences` | `EXPORT_INCLUDE_SEQUENCES` | `Config.Export.IncludeSequences` | 布尔开关 |

## 12. 异常情况

- 用户按取消键：返回上一页或退出，退出前不执行导出。
- 必填字段为空：在当前页提示并阻止进入下一步。
- 端口不是数字：提示格式错误并保留当前输入。
- 导出格式未选择：提示至少选择一种格式，或恢复默认 markdown。
- DSN 解析失败：执行校验或连接时展示错误，并允许返回修改。
- 数据库连接失败：展示连接失败原因，不泄露密码。
- 用户无目标库元数据读取权限：展示权限相关错误，并建议检查数据库账号授权。
- 输出路径不可写：展示路径错误，并允许返回修改输出路径。
- 正则表达式非法：展示校验错误，并允许修改 patterns。
- 导出过程中发生错误：保留错误信息，允许返回修改配置后重试。

## 13. 验收标准

- [ ] 执行 `schema-export tui` 可以进入 TUI。
- [ ] TUI 可以读取现有环境变量作为默认值。
- [ ] 用户可以通过 DSN 完成一次 Markdown 导出。
- [ ] 用户可以通过分离参数完成一次 Markdown 导出。
- [ ] 用户可以选择多个导出格式并生成对应文件。
- [ ] 用户可以配置 tables、exclude、patterns 并影响导出范围。
- [ ] 用户可以开启视图、存储过程、函数、触发器、序列导出选项。
- [ ] 密码输入不会明文显示。
- [ ] 确认页不会展示明文密码或 DSN 密码。
- [ ] 用户取消导出时不会创建输出文件。
- [ ] 导出失败时 TUI 展示可读错误并允许返回修改。
- [ ] 现有 `schema-export export` 命令行为不受影响。
- [ ] 自动化测试覆盖 TUI 配置到 `config.Config` 的映射。

## 14. 测试要求

- 自动化测试：
  - TUI 表单状态到 `config.Config` 的映射测试。
  - DSN 脱敏展示测试。
  - 密码字段不进入摘要文本测试。
  - 环境变量默认值加载测试。
  - 取消流程不调用导出服务测试。
  - 导出服务错误返回后的状态转换测试。
- 手动验证：
  - 使用 DSN 完成一次导出。
  - 使用 host、port、username、password 完成一次导出。
  - 使用多格式导出。
  - 使用表过滤导出。
  - 在窄终端窗口下检查页面不重叠、文本不截断关键字段。
- 边界场景：
  - 空 DSN。
  - 非数字端口。
  - 非法正则。
  - 输出目录不存在。
  - 输出目录不可写。
- 失败场景：
  - 数据库连接失败。
  - 用户名或密码错误。
  - 目标 schema 不存在或无权限。

## 15. 开发步骤

- [ ] 确认 TUI 技术选型和是否允许新增依赖。
- [ ] 新增 `tui` 命令入口。
- [ ] 设计 TUI 状态模型和表单字段。
- [ ] 实现环境变量和默认配置预填。
- [ ] 实现分步表单、确认页和执行页。
- [ ] 实现密码与 DSN 脱敏展示。
- [ ] 将 TUI 状态转换为 `config.Config`。
- [ ] 复用现有导出服务执行导出。
- [ ] 增加配置映射、脱敏、取消流程和错误流程测试。
- [ ] 更新 README 的使用说明和 CLI 参考。
- [ ] 运行 `go test ./...` 验证现有功能不回归。

## 16. 待确认问题

- TUI 依赖选型是否优先使用 Bubble Tea 系列库，还是选择更轻量的终端表单库。
- 第一版是否需要在执行前提供数据库连接测试按钮。
- 第一版是否需要支持从历史输入中恢复配置；如需要，必须另行确认凭据保存和脱敏策略。
- 第一版是否需要支持只预览将要执行的等价 CLI 命令；如支持，命令中的密码和 DSN 密码必须脱敏。
