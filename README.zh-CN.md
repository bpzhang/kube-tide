# Kube Tide

基于 Go 和 React 的现代化 Kubernetes 多集群管理平台，提供直观的 Web 界面来简化 Kubernetes 资源管理与运维操作。

[中文文档](README.zh-CN.md) | [English](README.md)

## 主要特性

### 集群管理

- 多集群支持和管理
- 集群连接测试
- 集群资源概览

### 节点管理

- 节点状态监控和详情查看
- 节点资源使用情况可视化
- 节点排水 (Drain)、禁止/允许调度 (Cordon/Uncordon)
- 节点污点、标签管理
- 节点池 (Node Pools) 创建和管理
- 集群自动扩缩容 (Cluster Autoscaler) 配置

### 工作负载管理

#### Pod 管理

- Pod 查看、详情和删除
- 实时 Pod 日志与终端连接
- Pod 资源监控（CPU、内存、磁盘）与指标历史
- Pod 事件、生命周期与重启策略管理

#### Deployment 管理

- Deployment 创建、扩缩容、重启
- 发布历史与回滚
- 独立详情页 + Tab（概览、容器组、状态、Pod、访问方式、事件）
- 访问方式 Tab 展示关联 Service、Endpoints 与 Ingress 路由
- 更新策略、健康检查、资源限制、节点亲和性配置

#### StatefulSet 管理

- 基本管理、扩缩容与详情查看

#### Service 与 Ingress

- Service 创建与管理、端点监控
- 按命名空间查询 Ingress（Deployment 详情「路由」Tab 展示）

### 监控和可观测性

- 实时资源监控与 Recharts 可视化
- 集群/节点资源概览
- Pod 指标历史（内存缓存）

### 国际化

- 中英文多语言支持
- 动态语言切换

## 技术栈

### 后端

- **Go 1.26+** — 主语言
- **Gin** — Web 框架
- **client-go v0.36** — Kubernetes 客户端
- **coder/websocket** — Pod 终端
- **zap + lumberjack** — 日志与轮转
- **viper** — 配置管理

### 前端

- **React 19** — UI 框架
- **TypeScript** — 类型安全
- **Vite 8** — 构建工具
- **Ant Design 6** — UI 组件
- **React Router 7** — 路由
- **Axios** — HTTP 客户端
- **Recharts** — 图表
- **xterm.js** — 终端

## 系统架构

单体 Go HTTP 服务，生产构建将前端 embed 进二进制；后端通过 client-go 连接多个 Kubernetes 集群，前端通过 REST 与 WebSocket 交互。

详见 [架构文档](docs/architecture.md)。

**生产环境请注意（ECS / 云主机部署，非 K8s 内部署）：**

- 平台 **无内置认证**，需 Nginx + TLS + 网络隔离
- 集群注册信息 **仅存内存**，重启后需重新添加
- 当前建议 **单实例** 跑在一台 ECS 上

部署步骤见 [运维手册 §2 ECS 部署](docs/operations.md)。

## 目录结构

```plaintext
kube-tide/
├── cmd/server/           # 应用入口
├── configs/              # 配置文件
├── docs/                 # 文档
├── internal/
│   ├── api/              # HTTP 处理器与路由
│   ├── core/k8s/         # K8s 业务逻辑
│   └── utils/            # 日志、国际化
├── pkg/embed/            # 生产模式 embed 的前端
├── web/                  # React 前端
├── scripts/              # 工具脚本
├── Makefile
└── README.md
```

完整目录：[代码架构](docs/code_arch.md)

## 安装和使用

### 环境要求

- Go 1.26+
- Node.js 18+（推荐 LTS）
- pnpm
- 可访问的 Kubernetes 集群

### 快速开始（生产）

```bash
git clone https://github.com/bpzhang/kube-tide.git
cd kube-tide

make build-prod
make run-prod
```

访问：`http://localhost:8080`

### 开发环境

**方式一 — Makefile（前端热更新 + 后端）**

```bash
make run
# 前端 http://127.0.0.1:5173
# 后端 http://127.0.0.1:8080
```

**方式二 — 分终端**

```bash
# 终端 1
cd web && pnpm install && pnpm dev

# 终端 2
go mod download
go run ./cmd/server/main.go
```

### Make 命令

| 命令 | 说明 |
|------|------|
| `make build` | 构建开发版二进制（含前端 embed） |
| `make build-prod` | 构建生产版二进制 |
| `make run` | 开发模式运行 |
| `make run-prod` | 构建并运行生产版 |
| `make clean` | 清理构建产物 |
| `make docker-build` | 构建 Docker 镜像（ECS 上可选） |
| `make docker-run` | 本地运行容器 |
| `make help` | 查看帮助 |

## 配置

- **文件：** `configs/config.yaml` — 服务端口、日志
- **环境变量：** `K8S_PLATFORM_ENV=production` — 强制生产模式
- **集群：** 运行时通过 UI 或 API 添加（不在配置文件中）

生产配置详见 [运维手册](docs/operations.md)。

## 文档

- [系统架构](docs/architecture.md)
- [代码目录结构](docs/code_arch.md)
- [运维部署手册](docs/operations.md)
- [功能路线图](docs/TODO.md)

## 路线图

- ConfigMap / Secret 管理
- 存储（PV、PVC、StorageClass）
- Prometheus 集成
- 平台 RBAC / 认证
- 平台 RBAC / 认证

部署方式：**ECS / 云主机 + systemd**，见 [运维手册](docs/operations.md)。可选 Docker：`deployments/docker/`。

完整列表见 [TODO](docs/TODO.md)。

## 贡献

欢迎提交 Pull Request 或 Issue。

- 遵循 Go 官方代码规范
- 保持前端 TypeScript 类型准确
- 行为变更时同步更新文档

## 升级依赖

```bash
# 后端
./scripts/upgrade-deps.sh

# 或手动
go get -u ./...
go mod tidy

# 前端
cd web && pnpm update
```

## 许可证

[MIT License](LICENSE)

## 致谢

- [Kubernetes](https://kubernetes.io/)
- [client-go](https://github.com/kubernetes/client-go)
- [Ant Design](https://ant.design/)
- [Gin](https://gin-gonic.com/)
