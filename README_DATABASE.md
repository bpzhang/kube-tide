# Kube-Tide 数据库支持

## 概述

Kube-Tide 现在支持完整的数据库持久化功能，提供灵活的多数据库后端支持，用于存储和管理 Kubernetes 集群的相关数据。

## 🚀 功能特性

### 多数据库支持
- **PostgreSQL** - 推荐用于生产环境
- **SQLite** - 适用于开发和测试环境

### 完整的数据模型
- **集群 (Clusters)** - 存储 Kubernetes 集群信息
- **节点 (Nodes)** - 存储集群节点信息
- **Pod** - 存储 Pod 运行状态和资源信息
- **命名空间 (Namespaces)** - 存储命名空间信息
- **部署 (Deployments)** - 存储部署信息（待完善）
- **服务 (Services)** - 存储服务信息（待完善）

### 架构特性
- **仓储模式** - 分离业务逻辑和数据访问
- **事务支持** - 保证数据一致性
- **分页查询** - 高效处理大数据集
- **连接池** - 优化数据库性能
- **迁移系统** - 版本化数据库架构管理

## 📁 项目结构

```
internal/
├── database/                 # 数据库核心功能
│   ├── config.go            # 数据库配置
│   ├── database.go          # 数据库连接管理
│   ├── service.go           # 数据库服务
│   ├── models/              # 数据模型
│   │   ├── cluster.go       # 集群模型
│   │   ├── node.go          # 节点模型
│   │   ├── pod.go           # Pod 模型
│   │   ├── namespace.go     # 命名空间模型
│   │   ├── deployment.go    # 部署模型
│   │   ├── service.go       # 服务模型
│   │   └── common.go        # 通用模型
│   └── migrations/          # 数据库迁移
│       ├── migrate.go       # 迁移服务
│       └── migrations.go    # 迁移定义
├── repository/              # 数据访问层
│   ├── interfaces.go        # 仓储接口
│   ├── repository.go        # 仓储工厂
│   ├── cluster_repository.go # 集群仓储
│   ├── node_repository.go   # 节点仓储
│   ├── pod_repository.go    # Pod 仓储
│   ├── namespace_repository.go # 命名空间仓储
│   ├── deployment_repository.go # 部署仓储（基础版）
│   └── service_repository.go # 服务仓储（基础版）
cmd/
├── migrate/                 # 迁移工具
│   └── main.go
└── example/                 # 示例程序
    └── main.go
configs/
└── database.yaml           # 数据库配置示例
docs/
└── database.md             # 详细使用文档
```

## 🛠️ 快速开始

### 1. 配置数据库

创建配置文件 `configs/database.yaml`：

```yaml
# PostgreSQL 配置
database:
  type: "postgres"
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "password"
  database: "kube_tide"
  ssl_mode: "disable"
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: "5m"

# 或者 SQLite 配置
# database:
#   type: "sqlite"
#   database: "kube_tide"
#   sqlite_file_path: "./data/kube_tide.db"
#   max_open_conns: 10
#   max_idle_conns: 2
#   conn_max_lifetime: "5m"
```

### 2. 运行数据库迁移

```bash
# PostgreSQL
go run cmd/migrate/main.go -action=migrate -type=postgres -host=localhost -port=5432 -user=postgres -password=yourpassword -database=kube_tide

# SQLite
go run cmd/migrate/main.go -action=migrate -type=sqlite -sqlite-file=./data/kube_tide.db
```

### 3. 使用示例

```go
package main

import (
    "context"
    "log"
    "time"

    "go.uber.org/zap"
    "kube-tide/internal/database"
    "kube-tide/internal/database/models"
    "kube-tide/internal/repository"
)

func main() {
    // 创建日志器
    logger, _ := zap.NewDevelopment()
    
    // 创建数据库配置
    config := &database.DatabaseConfig{
        Type:            database.PostgreSQL,
        Host:            "localhost",
        Port:            5432,
        User:            "postgres",
        Password:        "password",
        Database:        "kube_tide",
        SSLMode:         "disable",
        MaxOpenConns:    25,
        MaxIdleConns:    5,
        ConnMaxLifetime: 5 * time.Minute,
    }
    
    // 创建数据库服务
    dbService, err := database.NewService(config, logger)
    if err != nil {
        log.Fatal(err)
    }
    defer dbService.Close()
    
    // 初始化数据库（运行迁移）
    ctx := context.Background()
    if err := dbService.Initialize(ctx); err != nil {
        log.Fatal(err)
    }
    
    // 创建仓储
    repos := repository.NewRepositories(dbService.GetDatabase(), logger)
    
    // 创建集群
    cluster := &models.Cluster{
        Name:        "my-cluster",
        Config:      "cluster-config",
        Status:      models.ClusterStatusActive,
        Description: "My Kubernetes cluster",
        Endpoint:    "https://api.my-cluster.com",
        Version:     "v1.28.0",
    }
    
    if err := repos.Cluster.Create(ctx, cluster); err != nil {
        log.Fatal(err)
    }
    
    // 查询集群
    clusters, err := repos.Cluster.List(ctx, models.ClusterFilters{}, models.DefaultPaginationParams())
    if err != nil {
        log.Fatal(err)
    }
    
    logger.Info("Found clusters", zap.Int("count", clusters.TotalCount))
}
```

## 🧪 测试

运行示例程序来测试数据库功能：

```bash
go run cmd/example/main.go
```

这将演示：
- 数据库连接和迁移
- 集群的 CRUD 操作
- 节点的创建和查询
- Pod 和命名空间的管理

## 🔧 迁移工具

### 查看当前版本
```bash
go run cmd/migrate/main.go -action=version -type=postgres -host=localhost -port=5432 -user=postgres -password=yourpassword -database=kube_tide
```

### 运行迁移
```bash
go run cmd/migrate/main.go -action=migrate -type=postgres -host=localhost -port=5432 -user=postgres -password=yourpassword -database=kube_tide
```

### 回滚到指定版本
```bash
go run cmd/migrate/main.go -action=rollback -version=1 -type=postgres -host=localhost -port=5432 -user=postgres -password=yourpassword -database=kube_tide
```

## 📊 数据模型

### 集群 (Clusters)
- ID, Name, Config, Status, Description
- Kubeconfig, Endpoint, Version
- 创建和更新时间

### 节点 (Nodes)
- 基本信息：ID, ClusterID, Name, Status, Roles
- 版本信息：Version, OSImage, KernelVersion
- 网络信息：InternalIP, ExternalIP
- 资源信息：CPU/Memory Capacity/Allocatable
- 元数据：Labels, Annotations, Conditions

### Pod
- 基本信息：ID, ClusterID, Namespace, Name
- 状态信息：Status, Phase, NodeName
- 网络信息：PodIP, HostIP
- 容器信息：RestartCount, ReadyContainers, TotalContainers
- 资源信息：CPU/Memory Requests/Limits
- 元数据：Labels, Annotations, OwnerReferences

## 🚧 待完善功能

- **部署仓储** - 完整的 Deployment 操作实现
- **服务仓储** - 完整的 Service 操作实现
- **事务方法** - 完善所有仓储的事务方法
- **索引优化** - 根据查询模式优化数据库索引
- **缓存层** - 添加 Redis 缓存支持
- **监控指标** - 数据库操作的 Prometheus 指标

## 📚 更多文档

- [详细使用文档](docs/database.md)
- [API 文档](docs/api.md)（待创建）
- [部署指南](docs/deployment.md)（待创建）

## 🤝 贡献

欢迎提交 Issue 和 Pull Request 来完善数据库功能！

## 📄 许可证

本项目采用 MIT 许可证。 