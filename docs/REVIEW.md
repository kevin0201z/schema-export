**审查报告 — schema-export**

**总体概览**
- **CI 状态**: 最新提交 a82d6fe 在 GitHub Actions 上已通过：`gofmt`、`go vet`、`golangci-lint`、单元测试矩阵和构建产物上传均成功。
- **范围**: 代码质量与 CI 配置审查，包含已发现问题、已修复项与改进建议。

**已完成的关键动作**
- 添加并收敛了 GitHub Actions CI workflow（格式化、静态检查、linter、测试、构建）。
- 解决了 golangci-lint 在 runner 上的预编译二进制兼容问题（改为在 runner 上 `go install` 编译运行）。
- 修复并提交了 golangci-lint 报告的问题：安全解析 DB_PORT 与替换 `errors.New(fmt.Sprintf(...))`。
- 为 `inspector` 实现添加了必要的前向方法以满足接口要求（单独提交）。

**修改列表（已提交）**
- 配置解析: [internal/config/config.go](internal/config/config.go) — 使用 `strconv.Atoi` 解析 `DB_PORT` 并检查错误。
- 错误工具包: [internal/errors/errors.go](internal/errors/errors.go) — 将 `errors.New(fmt.Sprintf(...))` 替换为 `fmt.Errorf(...)`。
- Inspector 适配: 多个 `inspector` 实现添加前向方法（见各实现文件）。

**CI 与日志摘要**
- 问题来源: 预编译的 `golangci-lint` 与 runner 当前 Go 工具链存在 `export data` 格式不匹配，导致分析器崩溃（已通过改为在 runner 上构建 linter 解决）。
- 修复后，linter 报告了可修复的代码问题（errcheck、gosimple），已按建议修复并再次通过 CI。

**发现/建议（按优先级）**
1. **固定 linter 版本**: 在 workflow 中指定 `golangci-lint` 的稳定版本（例如 `cmd/golangci-lint@v1.64.8`），避免每次编译获得不同插件集。将 `go install ...@latest` 改为明确版本可提高可重复性。
2. **缓存依赖**: 在 CI 中缓存模块下载（GOMODCACHE）和构建缓存，加速重复运行（已在部分 job 使用 cache，但可进一步强化）。
3. **预提交钩子**: 增加 `gofmt`/`go vet` 和 `golangci-lint` 的 pre-commit 检查（使用 `pre-commit` 或 `lefthook`），阻止低质量提交进入主分支。
4. **测试覆盖**: 为 `internal/database/*`、`internal/inspector` 增加单元/集成测试，模拟连接失败、不同 DB 配置（DSN/schema）和边界场景。
5. **文档补充**: 在仓库根目录添加 `CONTRIBUTING.md` 或 `docs/REVIEW.md`（当前文件）和 `DEVELOPER.md`，说明本地运行步骤、CI 要点（包括 linter 构建说明）以及常见问题解决方法。
6. **Actions Node.js 兼容性**: 日志提醒当前 actions 使用 Node.js 20 即将弃用，建议升级到支持 Node.js 24 的 action 版本或在 workflow 中设置 `FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true`（短期权宜）。
7. **静态分析策略**: 将 `.golangci.yml` 明确配置为启用必要 checks（如 `errcheck`、`gosimple`、`govet` 等），并区分阻塞级别与警告级别。

**如何在本地复现 CI 步骤（建议命令）**
```bash
# 安装/更新 golangci-lint（固定版本更可控）
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8

# 格式检查
gofmt -s -l .

# 静态检查
go vet ./...

# 运行 golangci-lint
$(go env GOPATH)/bin/golangci-lint run --timeout 5m

# 单元测试
go test ./...

# 构建可执行文件
make build
```

