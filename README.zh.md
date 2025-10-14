# Go 项目模板（Gin）

一个开箱可用的 Go Web 项目模板，基于 Gin，提供：

- Gin 路由与中间件（日志、恢复、限流、鉴权）
- 完整的可观测方案集成（Opentelemetry： 日志、链路、指标）
- 异步任务管理（异步任务或定时任务）
- CI 与代码检查示例
- Dockerfile 与镜像构建目标
- 用于从模板创建仓库后重命名模块的脚本/工作流

本说明包含快速上手指南、常用开发命令和配置说明。

## 快速上手

1. 在 GitHub 上使用 "Use this template" 从本模板创建新仓库。
2. 按需将 Go module 重命名为你的仓库路径（见下文）。
3. 在本地运行应用。

### 重命名 Go module

方式 A — 本地脚本（推荐）：

```bash
./scripts/rename-module.sh github.com/<your>/<repo>
```

方式 B — GitHub Actions：

在你克隆的仓库 Actions 面板中运行 "Rename Go module (one-time)" 工作流；留空则使用默认 `github.com/<owner>/<repo>`。

## 开发

安装依赖并运行：

```bash
make tidy
make run
```

打开健康检查和示例接口：

- http://localhost:8080/healthz
- http://localhost:8080/version

## 测试与 Lint

```bash
make test
make lint
```

## Docker

构建生产镜像：

```bash
make docker-build
```

## 配置

设置环境变量以控制运行时行为：

- PORT：监听端口（默认 8080）
- APP_ENV：debug | release | test（来自 gin）
- APP_NAME：应用名称

更多配置请查看 `pkg/config` 或 `config.yaml`。

## 版本信息

构建时会通过 `-ldflags` 将版本信息注入 `model/entity/version.go`，具体示例见 Makefile 与 CI 工作流。

## 项目结构（节选）

- `cmd/` — 入口与 CLI
- `api/` — HTTP 服务与控制器
- `pkg/` — 可复用包（config、jwt、logger、otel）
- `model/` — 数据模型与 DAO
- `scripts/` — 辅助脚本（模块重命名）

## 说明

- 本模板偏向轻量实用，可根据需要裁剪或替换组件。
- 如在运行或定制过程中遇到问题，请在你的仓库中打开 issue。
