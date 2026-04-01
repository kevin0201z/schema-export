# 开发环境快速设置

下面说明如何在本地安装并启用预提交钩子与 golangci-lint（与 CI 保持一致）。

1. 安装 `golangci-lint`（固定版本）：

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8
```

确保 `$GOPATH/bin` 在 `PATH` 中：

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

2. 安装 `pre-commit`（用于运行 `.pre-commit-config.yaml`）：

```bash
# 使用 pip 安装
pip install pre-commit
# 或使用包管理器，例如 macOS: brew install pre-commit
```

3. 在仓库中启用预提交钩子：

```bash
pre-commit install
pre-commit run --all-files
```

4. 在本地运行 CI 检查：

```bash
gofmt -s -l .
go vet ./...
golangci-lint run --timeout 5m
go test ./...
```

如果你希望把 `golangci-lint` 安装到某个特定版本（与 CI 一致），请修改 `go install` 的版本号。
开发者设置指南

1) 安装 `golangci-lint`（固定版本）

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8
```

2) 安装 `pre-commit`（可选，但推荐）

Ubuntu/Debian:
```bash
sudo apt update && sudo apt install -y python3-pip
pip3 install --user pre-commit
```

macOS (Homebrew):
```bash
brew install pre-commit
```

3) 启用 pre-commit 钩子（仓库根目录运行）：

```bash
pre-commit install
pre-commit run --all-files
```

4) 常用本地检查命令

```bash
gofmt -s -l .
go vet ./...
$(go env GOPATH)/bin/golangci-lint run --timeout 5m
```

说明：CI 已在 workflow 中固定 `golangci-lint` 版本并对 linter 相关的依赖尝试使用缓存来加速构建。
