# Kube Tide 系统架构

Kube Tide 是一个基于 Go + React 的 Kubernetes 多集群 Web 管理平台。前后端分离，后端通过 client-go 与多个集群交互，前端以 SPA 形式提供资源管理与运维操作界面。

## 文档索引

- [代码目录结构](./code_arch.md)
- [运维部署手册](./operations.md)
- [功能路线图](./TODO.md)

## 架构概览

```text
┌─────────────┐     REST / WebSocket      ┌──────────────────┐
│  React SPA  │ ◄───────────────────────► │  Gin HTTP Server │
│  (web/)     │                           │  (cmd/server)    │
└─────────────┘                           └────────┬─────────┘
                                                   │ client-go
                                          ┌────────┴─────────┐
                                          │  Cluster A / B…  │
                                          └──────────────────┘
```

### 设计特点

| 特点 | 说明 |
|------|------|
| 单体应用 | 前后端打包为单一二进制（生产模式 embed 前端静态资源） |
| 多集群 | 运行时动态注册多个 kubeconfig，内存中维护 client 连接 |
| 实时能力 | Pod 日志流、Exec 终端通过 WebSocket 实现 |
| 指标缓存 | Pod 指标定时采集并缓存在内存，供历史图表使用 |
| 国际化 | 前后端均支持中英文切换 |

### 当前限制（运维相关）

- **无平台级认证**：API 未启用 JWT/OIDC/RBAC，生产环境需通过网络层隔离或反向代理鉴权
- **集群配置无持久化**：通过 UI 添加的集群保存在进程内存，重启后需重新注册
- **单实例设计**：不支持多副本共享状态，水平扩展需额外改造
- **健康检查较浅**：`/api/health` 仅返回进程存活，不探测 K8s 连通性（详见 [operations.md](./operations.md)）

## 技术栈

### 后端

| 组件 | 选型 | 用途 |
|------|------|------|
| 语言 | Go 1.26+ | 主服务 |
| Web 框架 | Gin | HTTP 路由与中间件 |
| K8s 客户端 | client-go v0.36 | 集群 API 交互 |
| WebSocket | coder/websocket | Pod 终端 |
| 配置 | viper | 读取 `configs/config.yaml` |
| 日志 | zap + lumberjack | 结构化日志与文件轮转 |

### 前端

| 组件 | 选型 | 用途 |
|------|------|------|
| 框架 | React 19 | UI |
| 语言 | TypeScript | 类型安全 |
| 构建 | Vite 8 | 开发与打包 |
| UI | Ant Design 6 | 组件库 |
| 路由 | React Router 7 | SPA 路由 |
| HTTP | Axios | API 请求 |
| 图表 | Recharts | 指标可视化 |
| 终端 | xterm.js | Pod Exec |
| 国际化 | i18next | 中英文 |

> 说明：项目**未使用** PostgreSQL、Redis、GORM、Redux、ECharts 等；若未来引入持久化或状态管理，需同步更新本文档。

## 运行模式

通过构建标志与环境变量区分开发/生产模式（见 `configs/config.go`）：

| 模式 | 触发条件 | 前端资源 | 典型用途 |
|------|----------|----------|----------|
| 开发 | 默认（未设 `K8S_PLATFORM_ENV=production` 且非 prod 构建） | 文件系统 / Vite 热更新 | 本地开发 |
| 生产 | `make build-prod` 或 `K8S_PLATFORM_ENV=production` | embed 进二进制 | 部署运行 |

生产构建会将 `web/dist` 复制到 `pkg/embed/web/dist` 并编译进二进制。

## 核心模块

### API 层 (`internal/api`)

- `router.go`：路由注册，区分 dev/prod 静态资源服务
- `*_handler.go`：按资源划分的 HTTP 处理器（Cluster、Node、Pod、Deployment 等）
- `middleware/language.go`：语言检测
- `response.go`：统一响应格式

### 业务层 (`internal/core/k8s`)

- `client.go`：`ClientManager`，管理多集群 client-go 连接
- 各资源 `*.go`：Deployment、Pod、Service、Ingress、StatefulSet、Node 等
- `pod_metrics*.go`：指标采集与内存缓存
- `autoscaler.go`、`nodepool.go`：节点池与自动扩缩容

### 工具层 (`internal/utils`)

- `logger/`：zap 封装，支持文件输出与轮转
- `i18n/`：后端国际化

### 嵌入资源 (`pkg/embed`)

- `static.go`：生产模式提供 embed 的前端静态文件

## 开发

### 环境要求

- Go 1.26+
- Node.js 18+（推荐 LTS）
- pnpm
- 可访问的 Kubernetes 集群（kubeconfig）

### 克隆与依赖

```bash
git clone https://github.com/bpzhang/kube-tide.git
cd kube-tide

# 后端依赖
go mod download

# 前端依赖
cd web && pnpm install && cd ..
```

### 本地开发（推荐）

**方式一：前后端分离热更新**

```bash
# 终端 1：前端
cd web && pnpm dev
# 访问 http://127.0.0.1:5173

# 终端 2：后端
go run ./cmd/server/main.go
# API http://127.0.0.1:8080
```

**方式二：Makefile 一键启动**

```bash
make run
# 前端 http://127.0.0.1:5173，后端 http://127.0.0.1:8080
```

### 构建

```bash
make build       # 测试环境二进制 → dist/kube-tide
make build-prod  # 生产二进制（embed 前端）→ dist/kube-tide-prod
make run-prod    # 构建并运行生产版本
make clean       # 清理 dist 与前端构建产物
make help        # 查看可用命令
```

生产访问地址：`http://localhost:8080`（端口见 `configs/config.yaml`）。

## 配置

主配置文件：`configs/config.yaml`

```yaml
server:
  port: 8080
  host: 127.0.0.1

logging:
  level: info
  file:
    enabled: true
    path: "./logs/kube-tide.log"
  rotate:
    enabled: true
    max_size: 100
    max_age: 30
    max_backups: 10
```

环境变量：

| 变量 | 说明 |
|------|------|
| `K8S_PLATFORM_ENV=production` | 强制生产模式（与 prod 构建等效） |

集群 kubeconfig 通过 Web UI 或 API 动态添加，**不写入配置文件**。详见 [operations.md](./operations.md)。

## API 概览

- 基础路径：`/api/v1`（健康检查为 `/api/health`）
- 资源路径模式：`/api/clusters/:cluster/namespaces/:namespace/...`
- WebSocket：Pod Exec 端点见 `router.go` 中 `/exec` 路由

完整接口列表以 `internal/api/router.go` 为准；OpenAPI 文档尚未生成（见 TODO）。

## 优雅关闭

`cmd/server/main.go` 监听 `SIGINT` / `SIGTERM`：

1. 停止 Pod 指标采集 goroutine
2. HTTP Server 5 秒超时 shutdown
3. 记录退出日志

容器/K8s 编排部署时建议预留 graceful shutdown 时间；**推荐部署方式为 ECS + systemd**（见 [operations.md](./operations.md)）。

## 相关文档

- [代码目录结构](./code_arch.md)
- [运维部署手册](./operations.md)
- [功能路线图](./TODO.md)
