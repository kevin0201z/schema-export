# 开发环境快速设置

下面说明如何在本地安装并启用预提交钩子与 `golangci-lint`，尽量与 CI 保持一致。

1. 安装 `golangci-lint`（固定版本）

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8
```

确保 `$GOPATH/bin` 在 `PATH` 中：

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

2. 安装 `pre-commit`（可选，但推荐）

Ubuntu/Debian:
```bash
sudo apt update && sudo apt install -y python3-pip
pip3 install --user pre-commit
```

macOS (Homebrew):
```bash
brew install pre-commit
```

3. 启用 `pre-commit` 钩子（仓库根目录运行）

```bash
pre-commit install
pre-commit run --all-files
```

4. 常用本地检查命令

```bash
gofmt -s -l .
go vet ./...
$(go env GOPATH)/bin/golangci-lint run --timeout 5m
go test ./...
```

说明：CI 已固定 `golangci-lint` 版本，并会运行格式检查、静态分析、单元测试和构建。
