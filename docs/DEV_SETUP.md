# 开发环境快速设置

下面说明如何在本地安装并启用预提交钩子与 `golangci-lint`，尽量与 CI 保持一致。

1. 安装 `golangci-lint`（固定版本）

使用支持当前 Go 工具链的 v2 版本，`.golangci.yml` 按官方迁移规则保留原 v1 的检查范围和默认排除项，避免工具升级混入新的样式规则。

```bash
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.13.2
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

## macOS 构建与验证

需要 Go 1.25 或更高版本和 Xcode Command Line Tools。使用 `make build` 或 `make install` 自动应用 macOS 12 部署目标；`MACOSX_DEPLOYMENT_TARGET` 可覆盖该目标。直接执行 `go build` 不应用 Makefile 的设置。

```bash
make build-darwin
sh scripts/check-macos-build.sh build/schema-export-darwin-amd64 amd64 1
sh scripts/check-macos-build.sh build/schema-export-darwin-arm64 arm64 1
```

`make build-darwin` 和 `make build-all` 需要 macOS。Linux 等平台可执行 `make build-darwin-portable`，产物名称带 `-portable`，不支持达梦第三方加密插件。正式发布使用 Mac 上启用 CGO 的产物，最低目标为 macOS 12。

CI 在 Linux、Intel Mac、Apple Silicon Mac 上测试 Go 1.25 / 1.26，并检查两种 Mac 架构的 CGO 配置、最低系统标记和本机启动。发布工具链从 `go.mod` 读取。

达梦服务配置需显式传入 `DM_SVC_PATH` 或 DSN 的 `svcConfPath`；驱动在 macOS 上不会自动读取 `/etc/dm_svc.conf`。
