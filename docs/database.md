# Kube-Tide 数据库支持

Kube-Tide 支持多种数据库后端，提供灵活的持久化存储解决方案。

## 支持的数据库

- **PostgreSQL** - 推荐用于生产环境
- **SQLite** - 适用于开发和测试环境

## 配置

### PostgreSQL 配置

```yaml
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
```

### SQLite 配置

```yaml
database:
  type: "sqlite"
  database: "kube_tide"
  sqlite_file_path: "./data/kube_tide.db"
  max_open_conns: 10
  max_idle_conns: 2
  conn_max_lifetime: "5m"
```

## 数据库迁移

### 运行迁移

```bash
# 使用 PostgreSQL
go run cmd/migrate/main.go -action=migrate -type=postgres -host=localhost -port=5432 -user=postgres -password=yourpassword -database=kube_tide

# 使用 SQLite
go run cmd/migrate/main.go -action=migrate -type=sqlite -sqlite-file=./data/kube_tide.db
```

### 查看当前版本

```bash
go run cmd/migrate/main.go -action=version -type=postgres -host=localhost -port=5432 -user=postgres -password=yourpassword -database=kube_tide
```

### 回滚到指定版本

```bash
go run cmd/migrate/main.go -action=rollback -version=1 -type=postgres -host=localhost -port=5432 -user=postgres -password=yourpassword -database=kube_tide
```

## 数据模型

### 集群 (Clusters)

存储 Kubernetes 集群的基本信息：

- ID: 唯一标识符
- Name: 集群名称
- Config: 集群配置
- Status: 集群状态 (active/inactive)
- Description: 集群描述
- Kubeconfig: Kubernetes 配置文件
- Endpoint: 集群 API 端点
- Version: Kubernetes 版本

### 节点 (Nodes)

存储集群中的节点信息：

- ID: 唯一标识符
- ClusterID: 所属集群ID
- Name: 节点名称
- Status: 节点状态
- Roles: 节点角色
- Version: Kubernetes 版本
- 资源信息 (CPU, Memory)
- 网络信息 (Internal IP, External IP)

### Pod

存储 Pod 信息：

- ID: 唯一标识符
- ClusterID: 所属集群ID
- Namespace: 命名空间
- Name: Pod 名称
- Status: Pod 状态
- Phase: Pod 阶段
- 资源使用情况
- 容器信息

### 命名空间 (Namespaces)

存储命名空间信息：

- ID: 唯一标识符
- ClusterID: 所属集群ID
- Name: 命名空间名称
- Status: 状态
- Phase: 阶段

### 部署 (Deployments)

存储部署信息：

- ID: 唯一标识符
- ClusterID: 所属集群ID
- Namespace: 命名空间
- Name: 部署名称
- 副本信息
- 策略类型

### 服务 (Services)

存储服务信息：

- ID: 唯一标识符
- ClusterID: 所属集群ID
- Namespace: 命名空间
- Name: 服务名称
- Type: 服务类型
- 端口配置

## 使用示例

### 创建数据库服务

```go
package main

import (
    "context"
    "log"
    "time"
    
    "go.uber.org/zap"
    "kube-tide/internal/database"
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
    clusterRepo := repository.NewClusterRepository(dbService.GetDatabase(), logger)
    
    // 使用仓储进行数据操作
    // ...
}
```

### 集群操作示例

```go
// 创建集群
cluster := &models.Cluster{
    Name:        "test-cluster",
    Config:      "cluster-config",
    Status:      models.ClusterStatusActive,
    Description: "Test cluster",
    Endpoint:    "https://api.test-cluster.com",
    Version:     "v1.28.0",
}

err := clusterRepo.Create(ctx, cluster)
if err != nil {
    log.Fatal(err)
}

// 查询集群
clusters, err := clusterRepo.List(ctx, models.ClusterFilters{}, models.DefaultPaginationParams())
if err != nil {
    log.Fatal(err)
}

// 更新集群
updates := models.ClusterUpdateRequest{
    Status: &models.ClusterStatusInactive,
}
err = clusterRepo.Update(ctx, cluster.ID, updates)
if err != nil {
    log.Fatal(err)
}
```

## 性能优化

### 连接池配置

根据应用负载调整连接池参数：

```yaml
database:
  max_open_conns: 25    # 最大打开连接数
  max_idle_conns: 5     # 最大空闲连接数
  conn_max_lifetime: "5m"  # 连接最大生存时间
```

### 索引优化

数据库表已经创建了必要的索引：

- 集群名称索引
- 集群状态索引
- 创建时间索引
- 外键索引

### 查询优化

- 使用分页查询避免大结果集
- 使用过滤器减少数据传输
- 合理使用事务

## 故障排除

### 常见问题

1. **连接失败**
   - 检查数据库服务是否运行
   - 验证连接参数
   - 检查网络连接

2. **迁移失败**
   - 检查数据库权限
   - 查看迁移日志
   - 验证 SQL 语法

3. **性能问题**
   - 调整连接池参数
   - 检查查询执行计划
   - 优化索引

### 日志配置

启用详细日志以便调试：

```go
logger, _ := zap.NewDevelopment()
```

## 备份和恢复

### PostgreSQL

```bash
# 备份
pg_dump -h localhost -U postgres kube_tide > backup.sql

# 恢复
psql -h localhost -U postgres kube_tide < backup.sql
```

### SQLite

```bash
# 备份
cp ./data/kube_tide.db ./data/kube_tide_backup.db

# 恢复
cp ./data/kube_tide_backup.db ./data/kube_tide.db
``` 