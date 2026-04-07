# 测试增补计划

本文档用于整理 schema-export 项目的测试补齐路线图，目标是优先提升核心导出链路的自动化保护，而不是单纯追求表面覆盖率。

## 当前情况

基于现状分析，测试覆盖存在明显偏科：

- `internal/model`、`internal/filter`、`internal/inspector` 覆盖较好。
- `internal/config` 覆盖尚可。
- `internal/cli` 有部分测试，但偏重纯解析函数。
- `cmd/schema-export`、`internal/app/export`、各类 `exporter`、`postgres/oracle/dm inspector` 基本缺少有效测试。

当前主要问题不是“完全没有测试”，而是“最关键的用户路径没有被测试保护”：

1. 命令入口未验证。
2. 导出主流程未验证。
3. 真实文件输出结果未验证。
4. 多数据库 Inspector 支持严重不均衡。

## 目标

优先建立以下自动化保障：

- 用户执行 CLI 命令时，参数能正确进入导出流程。
- 导出服务在成功、部分失败、全部失败、降级告警等场景下行为稳定。
- 导出器生成的文件名、目录结构和关键内容可被回归测试保护。
- MySQL、PostgreSQL、SQL Server、Oracle、DM 的关键元数据解析逻辑有最基本的单元测试覆盖。

建议阶段性目标：

- `cmd/schema-export` 和 `internal/app/export` 不再为 `0%`
- 所有 exporter 包达到 `70%+`
- `postgres/mysql/sqlserver` inspector 各达到 `60%+`
- 主链路相关包达到 `70%+`

## 执行顺序

建议按以下顺序推进：

1. 先补主流程编排测试
2. 再补 CLI 命令测试
3. 再补导出器文件内容测试
4. 再补各数据库 Inspector
5. 最后补基础设施和工具层测试

这样做的收益最高，因为可以先保护最接近用户价值的行为。

## 阶段一：导出主流程编排

目标文件：

- `internal/app/export/service_test.go`

重点覆盖：

- `Run()` 在数据库类型未注册时返回错误
- `Run()` 在 Inspector 工厂 `Create()` 失败时返回错误
- `Run()` 在 `Connect()` 失败时返回错误
- `Run()` 在 `TestConnection()` 失败时返回错误
- `Run()` 在 `GetTables()` 失败时返回错误
- `Run()` 在过滤后无成功表时返回 `no tables were successfully processed`
- `loadTables()` 遇到单表失败时继续处理其他表
- `loadTables()` 在 context 取消时返回取消错误
- 开启 `IncludeViews`
- 开启 `IncludeProcedures`
- 开启 `IncludeFunctions`
- 开启 `IncludeTriggers`
- 开启 `IncludeSequences`
- 附属对象查询失败时只打印告警，不中断整个导出
- `ExportAllFormats()` 在全部成功时返回 `nil`
- `ExportAllFormats()` 在部分失败时返回 `partial export failure`
- `ExportAllFormats()` 在全部失败时返回 `all exports failed`
- `exportFormat()` 对 unsupported format 返回错误
- `exportFormat()` 在 exporter factory 创建失败时返回错误
- `exportFormat()` 能正确传递 `ExportOptions`

建议实现方式：

- 在测试中定义 `mockInspector`、`mockInspectorFactory`
- 在测试中注册临时 exporter factory
- 重点断言返回错误、输入参数和降级行为

## 阶段二：CLI 命令测试

目标文件：

- `cmd/schema-export/main_test.go`
- `internal/cli/export_test.go`

重点覆盖：

- `newRootCmd()` 是否挂载 `export` 和 `version` 子命令
- `version` 命令输出是否包含 `version`、`commit`、`date`
- `export` 子命令 flags 是否正确绑定到 `cmd.Config`
- `--type`
- `--dsn`
- `--host`
- `--port`
- `--database`
- `--username`
- `--password`
- `--schema`
- `--output`
- `--formats`
- `--split`
- `--tables`
- `--exclude`
- `--patterns`
- `-V/-P/-F/-T/-S`
- `ExportCommand.Run()` 在配置无效时返回错误
- `ExportCommand.Run()` 在配置有效时进入 service

说明：

- 这里重点不是测试 Cobra 本身，而是测试本项目的命令组装和参数接线。

## 阶段三：导出器文件内容测试

目标文件：

- `internal/exporter/markdown/exporter_test.go`
- `internal/exporter/sql/exporter_test.go`
- `internal/exporter/json/exporter_test.go`
- `internal/exporter/yaml/exporter_test.go`

重点覆盖：

- 单文件导出默认文件名正确
- 输出路径传目录时行为正确
- 输出路径传文件时行为正确
- `SplitFiles=true` 时生成预期目录结构
- 表导出内容正确
- 视图导出内容正确
- 存储过程导出内容正确
- 函数导出内容正确
- 触发器导出内容正确
- 序列导出内容正确
- 空集合时统计信息与文件头仍合理
- SQL 导出能根据 `DbType` 选择正确方言
- SQL 导出包含主键、索引、外键、检查约束
- SQL 导出包含表注释和列注释
- Markdown 导出包含概览、目录、详情区块
- JSON/YAML 输出结构完整且可反序列化

建议实现方式：

- 使用 `t.TempDir()` 生成输出目录
- 执行导出后读取文件内容进行断言
- 不只检查文件存在，还要检查关键文本片段

## 阶段四：PostgreSQL Inspector

目标文件：

- `internal/database/postgres/inspector_test.go`

重点覆盖：

- `BuildDSN()` 原始 DSN 透传
- `BuildDSN()` 自动补前缀
- `BuildDSN()` 组件拼装
- `BuildDSN()` `SSLMode` 分支
- `GetTables()`
- `GetTable()`
- `GetColumns()`
- `GetIndexes()`
- `GetForeignKeys()`
- `GetCheckConstraints()`
- `GetViews()`
- `GetProcedures()`
- `GetFunctions()`
- `GetTriggers()`
- `GetSequences()`
- 查询失败分支
- `rows.Err()` 分支

补充说明：

- PostgreSQL 当前几乎是完整空白区，应优先补齐。

## 阶段五：SQL Server Inspector 深化

目标文件：

- `internal/database/sqlserver/inspector_test.go`

重点覆盖：

- `BuildDSN()`
- `GetColumns()` 中 Unicode 长度处理
- `GetColumns()` 中默认值清理
- `GetColumns()` 中主键、自增、注释解析
- `GetForeignKeys()`
- `GetCheckConstraints()`
- `GetViews()`
- `GetProcedures()`
- `GetFunctions()`
- `GetTriggers()`
- `GetSequences()`
- `getTableComment()`
- `isUnicodeType()`
- `cleanDefaultValue()`

## 阶段六：MySQL Inspector 缺口补齐

目标文件：

- `internal/database/mysql/inspector_test.go`

重点覆盖：

- `GetTable()` 聚合成功路径
- `GetTable()` 在列/索引/外键/检查约束子调用失败时的返回
- `GetCheckConstraints()`
- `GetViews()`
- `GetProcedures()`
- `GetFunctions()`
- `GetTriggers()`
- `GetSequences()`
- `getTableComment()`
- `BuildDSN()` 前缀处理
- `BuildDSN()` SSL 参数分支
- 查询失败与空结果分支

## 阶段七：Oracle / DM 公共逻辑

目标文件：

- `internal/database/oracle_compatible_test.go`
- `internal/database/oracle/inspector_test.go`
- `internal/database/dm/inspector_test.go`

重点覆盖：

- `placeholderStr()`
- `parseTriggerTiming()`
- `querySource()` 多行拼接
- `queryTableComment()`
- `queryCheckConstraints()`
- `queryViews()`
- `queryProcedures()`
- `queryFunctions()`
- `queryTriggers()`
- `querySequences()`
- schema 非空分支
- schema 为空分支
- Oracle `BuildDSN()`
- DM `BuildDSN()`
- 工厂 `Create()` 与 `GetType()`

## 阶段八：基础设施与回归补充

目标文件：

- `internal/database/base_test.go`
- `internal/exporter/interface_test.go`
- `internal/errors/errors_test.go`

重点覆盖：

- `BaseInspector` 的 `BuildDSN()`
- `BaseInspector` 的 `Connect()`
- `BaseInspector` 的 `Close()`
- `BaseInspector` 的 `TestConnection()`
- exporter registry 的 `Register()`、`Get()`、`GetSupportedTypes()`
- errors 包的 `Wrap()`、`Wrapf()`、`Is()`、`As()`

## 建议排期

可按四周推进：

- 第 1 周：阶段一 + 阶段二
- 第 2 周：阶段三
- 第 3 周：阶段四 + 阶段五
- 第 4 周：阶段六 + 阶段七 + 阶段八

如果时间更紧，建议先完成前三个阶段，它们对整体稳定性提升最大。

## 完成标准

每个阶段完成后，至少做以下检查：

```bash
go test ./...
go test ./... -cover
```

建议在阶段结束时额外关注：

- 是否覆盖了成功路径与失败路径
- 是否覆盖了默认值与边界值
- 是否覆盖了文件内容而不只是文件存在
- 是否覆盖了降级行为而不只是直接报错

## 备注

本计划优先强调“关键行为保护”而不是单纯追求某个覆盖率数字。覆盖率可以作为辅助指标，但不应替代对核心业务路径的测试设计。
