贡献指南

欢迎贡献到 schema-export。下面是快速指南：

1. 分支与提交
- 基于 `main` 创建 feature 分支：`git checkout -b feat/描述`。
- 使用清晰的 commit message（例如 `fix: 修复 X`、`feat: 添加 Y`）。

2. 本地检查
- 在提交前运行：

```bash
gofmt -s -l .
go vet ./...
golangci-lint run --timeout 5m
go test ./...
```

- 推荐启用 `pre-commit`：

```bash
pip install --user pre-commit
pre-commit install
pre-commit run --all-files
```

3. Pull Request 流程
- 创建 PR，选择 `main` 为目标分支。
- CI 将自动运行：格式检查、静态分析、单元测试与构建。请在 CI 通过前不要合并。

4. 本仓库约定
- 我们在 CI 中固定 `golangci-lint` 版本以确保结果可重复。
- 代码风格遵循 `gofmt` 与 `go vet` 的检查。

5. 常见问题
- 如果你在运行 `golangci-lint` 时遇到 analyzer export-data 错误，请确保你安装的 `golangci-lint` 版本与 CI 中使用的版本一致（见 `docs/DEV_SETUP.md`）。
- 如果你准备补测试或提升覆盖率，建议先查看 `docs/TEST_PLAN.md` 中的测试增补路线图。

感谢贡献！
