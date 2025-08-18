# Kube Tide

![Kube Tide Logo](docs/images/logo.png)

一个基于 Go 和 React 的现代化 Kubernetes 多集群管理平台，提供直观的 Web 界面来简化 Kubernetes 资源管理和运维操作。

[中文文档](README.zh-CN.md) | [English](README.md)

## 主要特性

### 集群管理

- 多集群支持和管理
- 集群连接测试
- 集群资源概览
- 集群健康状态监控

### 节点管理

- 节点状态监控和详情查看
- 节点资源使用情况可视化
- 节点排水(Drain)操作
- 禁止/允许调度(Cordon/Uncordon)
- 节点污点(Taints)管理
- 节点标签(Labels)管理
- 节点池(Node Pools)创建和管理

### 工作负载管理

#### Pod 管理

- Pod 查看、详情和删除
- 实时 Pod 日志查看
- Pod 终端连接
- Pod 资源监控（CPU、内存、磁盘使用情况）
- Pod 指标历史数据可视化
- Pod 事件查看

#### Deployment 管理

- Deployment 创建和管理
- Deployment 扩缩容和重启
- Deployment 详情查看
- Deployment 更新策略配置
- Deployment 健康检查配置
- Deployment 资源限制配置
- Deployment 节点亲和性配置

#### StatefulSet 管理

- StatefulSet 基本管理功能
- StatefulSet 扩缩容
- StatefulSet 详情查看

#### Service 管理

- Service 创建和管理
- Service 详情查看
- Service 端点监控

### 监控和可观测性

- 实时资源监控
- 指标数据可视化
- 集群和节点资源概览
- Pod 性能指标历史记录

### 国际化支持

- 中英文多语言支持
- 动态语言切换

## 技术栈

### 后端

- **Go** - 主要编程语言
- **Gin** - Web 框架
- **client-go** - Kubernetes 客户端库
- **WebSocket** - 实时通信
- **Logrus** - 日志记录

### 前端

- **React 18** - 前端框架
- **TypeScript** - 类型安全
- **Ant Design** - UI 组件库
- **Vite** - 构建工具
- **React Router** - 路由管理
- **Axios** - HTTP 客户端
- **ECharts** - 数据可视化

## 系统架构

平台采用前后端分离架构：

- **前端**：React SPA 应用，通过 RESTful API 和 WebSocket 与后端通信
- **后端**：Go 微服务，通过 client-go 与多个 Kubernetes 集群交互
- **实时通信**：WebSocket 支持实时日志查看和终端连接

### 架构特点

- **多集群支持**
- **高性能缓存**
- **安全认证**
- **实时监控**
- **国际化**

## 目录结构

```plaintext
kube-tide/
├── cmd/                    # 应用程序入口点
│   ├── kube-tide/          # CLI 入口
│   └── server/             # 服务器入口
├── configs/                # 配置文件
├── docs/                   # 文档
│   ├── architecture.md     # 架构文档
│   ├── code_arch.md        # 代码架构
│   └── images/             # 文档图片
├── internal/               # 内部包
│   ├── api/                # API 处理器和路由
│   │   └── middleware/     # HTTP 中间件
│   ├── core/               # 核心业务逻辑
│   │   └── k8s/            # Kubernetes 资源管理
│   └── utils/              # 工具函数
│       ├── i18n/           # 国际化
│       └── logger/         # 日志工具
├── pkg/                    # 可导出的包
│   └── embed/              # 嵌入式资源
├── web/                    # 前端代码
│   ├── public/             # 静态资源
│   └── src/                # 源代码
│       ├── api/            # API 客户端
│       ├── components/     # React 组件
│       ├── i18n/           # 国际化
│       ├── layouts/        # 页面布局
│       ├── pages/          # 页面组件
│       └── utils/          # 工具函数
└── Makefile                # 构建脚本
```

## 安装和使用

### 环境要求

- Go 1.19 或更高版本
- Node.js 16 或更高版本
- pnpm 包管理器
- 可访问的 Kubernetes 集群

### 快速开始

1. **克隆仓库**

   ```bash
   git clone https://github.com/bpzhang/kube-tide.git
   cd kube-tide
   ```

2. **构建和运行**

   ```bash
   # 构建生产版本（前端和后端）
   make build-prod
   
   # 运行应用程序
   make run-prod
   ```

3. **访问 Web 界面**

   ```text
   http://localhost:8080
   ```

### 开发环境搭建

1. **后端开发**

   ```bash
   # 安装 Go 依赖
   go mod download
   
   # 以开发模式运行后端
   make dev
   ```

2. **前端开发**

   ```bash
   cd web
   pnpm install
   pnpm dev
   ```

### 可用的 Make 命令

- `make build` - 构建项目（前端和后端）
- `make build-prod` - 构建生产版本
- `make build-web` - 仅构建前端
- `make build-backend` - 仅构建后端
- `make run` - 运行应用程序
- `make run-prod` - 运行生产版本
- `make dev` - 以开发模式运行
- `make test` - 运行测试
- `make verify` - 运行验证（Maven 风格）
- `make clean` - 清理构建产物

## 配置

应用程序可以通过以下方式配置：

- **环境变量**
- **配置文件**（`configs/config.yaml`）
- **命令行参数**

主要配置选项：

- 服务器端口和主机
- Kubernetes 集群配置
- 日志级别
- 前端构建设置

## 文档

- [架构文档](docs/architecture.md)
- [代码架构](docs/code_arch.md)
- [TODO 列表](docs/TODO.md)

## 路线图

### 即将推出的功能

- ConfigMap 和 Secret 管理
- 存储管理（PV、PVC、StorageClass）
- 监控系统集成（Prometheus）
- RBAC 权限管理
- CI/CD 集成
- Helm Chart 支持

查看完整的 [TODO 列表](docs/TODO.md) 了解详细规划。

## 贡献

我们欢迎贡献！请随时提交 Pull Request 或 Issue 来改进项目。

### 开发指南

- 遵循 Go 官方代码标准
- 为新功能包含适当的测试
- 必要时更新文档
- 确保前端代码的 TypeScript 类型安全

## 许可证

本项目采用 [MIT 许可证](LICENSE)。

## 致谢

- [Kubernetes](https://kubernetes.io/) - 出色的容器编排平台
- [client-go](https://github.com/kubernetes/client-go) - 官方 Kubernetes Go 客户端库
- [Ant Design](https://ant.design/) - 优秀的 React UI 库
- [Gin](https://gin-gonic.com/) - 高性能 Go Web 框架
