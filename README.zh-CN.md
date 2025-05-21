# Kube Tide

一个基于Go语言和React开发的Kubernetes多集群管理平台，提供直观的Web界面，简化Kubernetes资源的管理和操作。

## 功能特点

### 集群管理

- ✅ 多集群支持和管理
- ✅ 集群连接测试
- ✅ 集群资源概览

### 节点管理

- ✅ 节点状态监控和详情查看
- ✅ 节点资源使用情况可视化
- ✅ 节点排水(Drain)操作
- ✅ 禁止/允许调度(Cordon/Uncordon)
- ✅ 节点污点(Taints)管理
- ✅ 节点标签(Labels)管理
- ✅ 节点池(Node Pools)创建和管理

### 工作负载管理

- ✅ Pod查看、详情和删除
- ✅ Pod日志实时查看
- ✅ Pod终端(Terminal)连接
- ✅ Pod资源监控（CPU、内存、磁盘使用情况）
- ✅ Pod指标历史数据可视化
- ✅ Deployment创建和管理
- ✅ Deployment扩缩容和重启
- ✅ StatefulSet管理
- ✅ Service管理

## 技术栈

### 后端

- Go语言
- Gin Web框架
- client-go Kubernetes客户端库

### 前端

- React
- TypeScript
- Ant Design组件库
- Vite构建工具

## 系统架构

平台采用前后端分离架构:

- 后端提供RESTful API
- 前端通过API与后端通信
- 后端通过client-go与Kubernetes集群交互

## 详细目录结构

```
kube-tide/
├── cmd/                    # 应用程序入口点
│   └── server/             # 服务器入口
├── configs/                # 配置文件
├── dist/                   # 构建输出目录
├── docs/                   # 文档
├── internal/               # 内部包
│   ├── api/                # API处理器和路由
│   ├── core/               # 核心业务逻辑
│   │   └── k8s/            # Kubernetes资源管理
│   └── utils/              # 工具函数
├── pkg/                    # 可导出的包
│   └── embed/              # 嵌入式资源
├── web/                    # 前端代码
│   ├── public/             # 静态资源
│   └── src/                # 源代码
│       ├── api/            # API客户端
│       ├── components/     # React组件
│       ├── layouts/        # 页面布局
│       └── pages/          # 页面组件
└── Makefile                # 构建脚本
```

## 安装和使用

### 先决条件

- Go 1.16或更高版本
- Node.js 14或更高版本
- Yarn包管理器
- 可访问的Kubernetes集群

### 构建和运行

1. 克隆仓库

    ```bash
    git clone https://github.com/your-username/kube-tide.git
    cd kube-tide
    ```

2. 安装依赖并构建

    ```bash
    # 构建生产版本(前后端)
    make build-prod
    
    # 或者分别构建
    make build-web      # 仅构建前端
    make build-backend  # 仅构建后端
    ```

3. 运行应用

    ```bash
    # 运行生产版本
    make run-prod
    
    # 或运行开发版本
    make dev
    ```

4. 访问Web界面

    ```textplain
    http://localhost:8080
    ```

## 开发指南

### 开发环境设置

1. 后端开发

    ```bash
    # 仅构建后端并启动
    make build-backend
    make dev
    ```

2. 前端开发

    ```bash
    cd web
    yarn install
    yarn dev
    ```

### 可用的Make命令

- `make build` - 构建项目(前后端)
- `make build-prod` - 构建生产版本
- `make build-web` - 仅构建前端
- `make build-backend` - 仅构建后端
- `make run` - 运行应用
- `make dev` - 开发模式运行
- `make test` - 运行测试
- `make clean` - 清理构建产物

## 待办事项

- 实现StatefulSet、DaemonSet管理
- 添加ConfigMap和Secret管理
- 实现存储管理(PV、PVC、StorageClass)
- 集成监控系统(Prometheus)
- 实现RBAC权限管理
- 添加CI/CD集成

## 贡献指南

欢迎提交Pull Request或Issues来改进项目。请确保代码遵循Go语言官方代码规范，并包含适当的测试。

## 许可证

[MIT License](LICENSE)
