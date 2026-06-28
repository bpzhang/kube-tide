# 代码目录结构

本文档描述 **当前仓库的实际结构**（截至最近一次文档同步）。若目录有变动，请随代码一并更新。

## 顶层概览

```text
kube-tide/
├── cmd/
│   └── server/
│       └── main.go                 # 应用入口：配置、服务初始化、HTTP 启动、优雅关闭
├── configs/
│   ├── config.go                   # 配置加载（viper）、开发/生产模式判断
│   └── config.yaml                 # 默认配置（服务端口、日志）
├── docs/                           # 项目文档
│   ├── architecture.md             # 系统架构
│   ├── code_arch.md                # 本文档
│   ├── operations.md               # 运维部署手册
│   └── TODO.md                     # 功能路线图
├── internal/
│   ├── api/                        # HTTP API 层
│   ├── core/k8s/                   # Kubernetes 业务逻辑
│   └── utils/                      # 日志、国际化等工具
├── pkg/
│   └── embed/                      # 生产模式 embed 的前端静态资源
│       ├── static.go
│       └── web/dist/               # 构建产物（make build 时生成）
├── web/                            # React 前端
│   ├── src/
│   ├── package.json
│   └── vite.config.ts
├── scripts/
│   └── upgrade-deps.sh             # 依赖升级脚本
├── deployments/                    # 部署配置
│   ├── ecs/                        # ECS / VM 推荐：systemd、Nginx、生产配置
│   └── docker/                     # 可选：Docker 镜像（非 K8s 编排）
├── dist/                           # Go 二进制输出（构建时生成）
├── logs/                           # 运行时日志（默认路径）
├── go.mod
├── go.sum
├── Makefile
├── README.md
└── README.zh-CN.md
```

## 后端

### `cmd/server/main.go`

- 加载配置、初始化 logger
- 创建 `ClientManager` 与各 K8s Service
- 启动 Pod 指标定时采集与缓存清理
- 注册 API Handler、启动 HTTP Server
- 处理 SIGTERM/SIGINT 优雅关闭

### `configs/`

| 文件 | 职责 |
|------|------|
| `config.go` | `LoadConfig()`、`IsDevMode()` / `IsProductionMode()` |
| `config.yaml` | server / logging 默认值 |

### `internal/api/`

扁平 Handler 结构（非 `handlers/` 子目录）：

| 文件 | 职责 |
|------|------|
| `router.go` | 路由注册、CORS、静态资源、SPA fallback |
| `response.go` | 统一成功/错误响应 |
| `health_handler.go` | `/api/health` |
| `cluster_handler.go` | 集群增删查、连接测试、指标与事件 |
| `namespace_handler.go` | 命名空间列表 |
| `node_handler.go` | 节点 CRUD、Drain/Cordon、污点/标签 |
| `nodepool_handler.go` | 节点池管理 |
| `autoscaler_handler.go` | 集群自动扩缩容配置 |
| `pod_handler.go` | Pod CRUD、日志、生命周期、重启策略 |
| `pod_metrics_handler.go` | Pod 指标查询 |
| `pod_terminal_handler.go` | Pod Exec WebSocket |
| `deployment_handler.go` | Deployment CRUD、扩缩容、历史与回滚 |
| `statefulset_handler.go` | StatefulSet 管理 |
| `service_handler.go` | Service 管理 |
| `ingress_handler.go` | Ingress 列表（按命名空间） |
| `middleware/language.go` | 请求语言检测 |

### `internal/core/k8s/`

| 文件 | 职责 |
|------|------|
| `client.go` | 多集群 client-go 连接管理（内存） |
| `namespace.go` | 命名空间 |
| `node.go` / `nodepool.go` | 节点与节点池 |
| `pod.go` / `pod_lifecycle.go` | Pod 与生命周期 |
| `pod_metrics*.go` | 指标采集、内存缓存 |
| `pod_resource_usage.go` / `pod_disk_usage.go` | 资源用量 |
| `deployment.go` | Deployment |
| `statefulset.go` / `statefulset_converters.go` | StatefulSet |
| `service.go` | Service |
| `ingress.go` | Ingress |
| `autoscaler.go` | Cluster Autoscaler 配置 |
| `metrics.go` / `container.go` / `storage_format.go` | 辅助逻辑 |

### `internal/utils/`

```text
utils/
├── logger/
│   ├── core.go       # zap 初始化
│   ├── config.go     # 文件与轮转配置
│   └── logger.go     # Logger 接口封装
└── i18n/
    ├── i18n.go
    └── locales/
        ├── en/translation.json
        └── zh/translation.json
```

### `pkg/embed/`

- `static.go`：`//go:embed web/dist`，生产模式 HTTP 静态文件服务

## 前端 (`web/src/`)

```text
web/src/
├── main.tsx / App.tsx          # 入口与路由
├── api/                        # 后端 API 客户端
│   ├── axios.ts                # Axios 实例
│   ├── cluster.ts / node.ts / pod.ts / deployment.ts
│   ├── service.ts / ingress.ts / statefulset.ts
│   ├── namespace.ts / nodepool.ts / autoscaler.ts
│   └── pod_metrics.ts
├── components/
│   ├── common/                 # LanguageSwitcher 等
│   └── k8s/                    # K8s 业务组件
│       ├── cluster/
│       ├── node/
│       ├── pod/
│       ├── deployment/
│       ├── statefulset/
│       ├── service/
│       └── common/             # 命名空间选择、标签/污点、事件等
├── layouts/
│   ├── MainLayout.tsx
│   └── menuConfig.tsx
├── pages/
│   ├── Dashboard.tsx
│   ├── Clusters.tsx / ClusterDetail.tsx
│   ├── Nodes.tsx / NodeDetail.tsx
│   └── workloads/
│       ├── Pods.tsx / PodDetailPage.tsx
│       ├── PodLogsPage.tsx / PodTerminalPage.tsx
│       ├── Deployments.tsx / DeploymentDetailPage.tsx
│       ├── StatefulSets.tsx / StatefulSetDetailPage.tsx
│       └── Services.tsx
├── i18n/                       # 前端国际化
└── utils/format.ts
```

### 前端技术说明

- **状态管理**：组件本地 state + React Router，未使用 Redux
- **图表**：Recharts（Dashboard、PodMonitoring、ClusterDetail 等）
- **终端**：@xterm/xterm + attach/fit 插件

## 构建产物（gitignore / 生成目录）

| 路径 | 说明 |
|------|------|
| `dist/kube-tide` | 开发构建二进制 |
| `dist/kube-tide-prod` | 生产构建二进制 |
| `web/dist/` | Vite 前端构建输出 |
| `pkg/embed/web/dist/` | 复制后 embed 的前端资源 |
| `logs/` | 运行时日志 |

## 尚未实现的规划目录

以下路径曾出现在早期设计稿中，**当前仓库不存在**，请勿按此部署或开发：

- `internal/models/`、`internal/repository/`（PostgreSQL / Redis 持久化）
- `internal/core/auth/`（JWT / OIDC）
- `cmd/kube-tide/`（独立 CLI）
- `web/src/store/`（Redux）

以下目录 **已提供**（面向 ECS / VM，非「平台部署进 K8s」）：

- `deployments/ecs/` — systemd 单元、生产配置、Nginx 示例
- `deployments/docker/Dockerfile` — 可选容器镜像
