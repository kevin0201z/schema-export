# macOS 兼容性修复

任务来源：2026-09-03 兼容性检查及用户的修复要求。

## 影响与边界

本次修改四种格式的分文件导出校验、macOS 构建设置、CI 和发布构建。
CLI/TUI 共用导出器，因此同时受保护；数据库结构、连接权限、API 和输出数据字段保持不变。
复用已有 `golang.org/x/text` 依赖进行 Unicode 处理，不增加运行时依赖。
导出安全策略调整已由本次修复请求授权；只编辑本地文件和验证临时数据，不执行发布。

## 文件名冲突规则

- 在创建导出目录或写入文件之前，对本次启用的全部对象进行检查。
- 文件名沿用原规则；通过 Unicode 规范化和大小写折叠检测冲突，各平台采用相同规则，便于导出文件跨平台复制。
- 检查表以及启用的视图、存储过程、函数、触发器、序列，包括对象后缀造成的重名。
- 名称包含目录分隔符、空名称或 NUL 时拒绝分文件导出，避免名称作为路径解释。
- 冲突时返回明确错误，建议使用合并导出；本批次不写入文件，现有文件不被截断。
- 合并导出不受文件名冲突限制。不存在冲突时，重复导出覆盖已有同名文件的行为不变。

## macOS 构建策略

- 支持目标为 macOS 12 或更高版本，Intel amd64 和 Apple Silicon arm64。
- 源码要求 Go 1.25 或更高；CI 验证 Go 1.25、1.26，发布从 `go.mod` 读取版本。
- 将不支持当前 Go 工具链的 golangci-lint v1 更新为 v2.13.2，迁移原默认检查范围和排除项，同步本地 pre-commit 与开发说明。
- `make build` / `make install` 在 macOS 上固定最低部署版本，并为 C 编译与链接同时设置对应选项，避免复用较高系统目标的缓存。
- macOS 的正式双架构产物在 macOS 上启用 CGO 构建，保留达梦第三方加密插件加载能力。
- `make build-darwin` 需要 macOS 及 Xcode Command Line Tools；无法在 Linux 上悄悄退化为不支持插件的产物。
- 提供显式 `make build-darwin-portable` 跨平台构建入口，关闭 CGO；该版本不支持达梦第三方加密插件，文件名包含 `-portable`，不替代正式产物。
- 达梦服务配置在 macOS 上通过 `DM_SVC_PATH` 或 DSN 的 `svcConfPath` 指定，避免修改内置第三方驱动的默认查找行为。

## 验收标准

- 四种格式拒绝大小写、Unicode 规范等价、不同对象类型的文件名冲突，并保留已有文件。
- 不导出的对象不参与冲突检测；普通中文和空格名称正常；合并导出与正常重复导出不回归。
- 原复现 SQLite 数据库分文件导出失败且不生成文件，合并导出保留两张表。
- 两种 macOS 架构的正式和便携产物编译成功，架构和最低系统标记符合预期；正式产物 CGO 为 1，便携产物为 0。
- CI 在 Linux、Intel Mac、Apple Silicon Mac 执行测试，构建失败使工作流失败；发布前验证 macOS 原生产物。
- Go 全量测试、静态分析、工作流语法和文档链接检查通过。

## 状态

已完成实现和本机验证（2026-09-03）。未建立统一 `docs/tasks.md`，本次状态及验证结果记录在此。

验证环境：macOS 26.6.2、Apple Silicon、Go 1.26.3。

| 验证 | 结果 |
|---|---|
| `go test ./... -count=1` | 21 个测试包全部通过，新增跨格式的文件名冲突、写入前检查、已有文件保护和正常导出回归 |
| `go vet ./...` | 通过 |
| `gofmt -s -l .` / `git diff --check` | 通过 |
| golangci-lint v2.13.2 | 0 issues，使用迁移后的原默认规则范围 |
| actionlint v1.7.12 | 两个工作流语法检查通过 |
| 文档检查 | UTF-8、代码围栏、相对链接均通过 |
| `make build build-darwin build-darwin-portable` | 本机及双架构正式/便携版本全部构建成功 |
| `scripts/check-macos-build.sh` | 四个产物的架构、CGO、macOS 12.0 标记正确；旧 macOS 26 产物被检查器拒绝 |
| 真实 SQLite CLI 验证 | 本机构建、正式 arm64、便携 arm64 均通过普通四格式导出；Unicode 冲突时退出失败且未创建输出目录；合并导出保留两张表 |
| 达梦插件加载探针 | 使用正式构建设置，驱动成功打开测试 Go 插件并进入加密符号验证，未落入无 CGO 的不支持分支 |

本机 Go 缓存放在临时目录。构建时出现模块元数据缓存受沙箱限制的非致命提示，产物均生成并通过上述检查。

未执行：远程 GitHub Actions、实际发布、Intel 实机运行、macOS 12 实机运行、真实达梦第三方加密服务器连接。测试插件仅用于确认加载入口，不实现加密算法，不能代替完整连接验收。代码和工作流已准备好后续在相应环境验证。

参考：

- [Go 最低系统要求](https://go.dev/wiki/MinimumRequirements)
- [GitHub 托管运行器与架构](https://docs.github.com/en/actions/reference/runners/github-hosted-runners)
- [golangci-lint 配置迁移](https://golangci-lint.run/docs/product/migration-guide/)
