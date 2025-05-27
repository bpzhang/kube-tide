# Kube-Tide Database API Integration

## 概述

本文档介绍了 Kube-Tide 项目的完整数据库 API 集成实现，包括部署和服务的完整仓储实现以及 RESTful API 层。

## 架构概览

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   HTTP Client   │───▶│   API Handler   │───▶│  Core Service   │───▶│   Repository    │
└─────────────────┘    └─────────────────┘    └─────────────────┘    └─────────────────┘
                                                                                │
                                                                                ▼
                                                                      ┌─────────────────┐
                                                                      │    Database     │
                                                                      │ (PostgreSQL/    │
                                                                      │    SQLite)      │
                                                                      └─────────────────┘
```

## 功能特性

### ✅ 完整实现的功能

1. **数据库层**
   - PostgreSQL 和 SQLite 支持
   - 连接池管理
   - 事务支持
   - 健康检查

2. **仓储层**
   - 部署仓储完整实现
   - 服务仓储完整实现
   - CRUD 操作
   - 分页查询
   - 条件过滤
   - 事务方法

3. **服务层**
   - 业务逻辑封装
   - 数据验证
   - 错误处理
   - 结构化日志

4. **API 层**
   - RESTful API 设计
   - JSON 请求/响应
   - 分页支持
   - 错误处理
   - CORS 支持

## API 端点

### 部署 (Deployments)

| 方法 | 端点 | 描述 |
|------|------|------|
| POST | `/api/v1/db/deployments` | 创建部署 |
| GET | `/api/v1/db/deployments/:id` | 根据 ID 获取部署 |
| PUT | `/api/v1/db/deployments/:id` | 更新部署 |
| DELETE | `/api/v1/db/deployments/:id` | 删除部署 |
| GET | `/api/v1/db/clusters/:cluster_id/deployments` | 列出集群中的部署 |
| DELETE | `/api/v1/db/clusters/:cluster_id/deployments` | 删除集群中的所有部署 |
| GET | `/api/v1/db/clusters/:cluster_id/deployments/count` | 统计集群中的部署数量 |
| GET | `/api/v1/db/clusters/:cluster_id/namespaces/:namespace/deployments` | 列出命名空间中的部署 |
| GET | `/api/v1/db/clusters/:cluster_id/namespaces/:namespace/deployments/:name` | 根据名称获取部署 |
| DELETE | `/api/v1/db/clusters/:cluster_id/namespaces/:namespace/deployments` | 删除命名空间中的所有部署 |
| GET | `/api/v1/db/clusters/:cluster_id/namespaces/:namespace/deployments/count` | 统计命名空间中的部署数量 |

### 服务 (Services)

| 方法 | 端点 | 描述 |
|------|------|------|
| POST | `/api/v1/db/services` | 创建服务 |
| GET | `/api/v1/db/services/:id` | 根据 ID 获取服务 |
| PUT | `/api/v1/db/services/:id` | 更新服务 |
| DELETE | `/api/v1/db/services/:id` | 删除服务 |
| GET | `/api/v1/db/clusters/:cluster_id/services` | 列出集群中的服务 |
| DELETE | `/api/v1/db/clusters/:cluster_id/services` | 删除集群中的所有服务 |
| GET | `/api/v1/db/clusters/:cluster_id/services/count` | 统计集群中的服务数量 |
| GET | `/api/v1/db/clusters/:cluster_id/namespaces/:namespace/services` | 列出命名空间中的服务 |
| GET | `/api/v1/db/clusters/:cluster_id/namespaces/:namespace/services/:name` | 根据名称获取服务 |
| DELETE | `/api/v1/db/clusters/:cluster_id/namespaces/:namespace/services` | 删除命名空间中的所有服务 |
| GET | `/api/v1/db/clusters/:cluster_id/namespaces/:namespace/services/count` | 统计命名空间中的服务数量 |

### 系统端点

| 方法 | 端点 | 描述 |
|------|------|------|
| GET | `/health` | 健康检查 |
| GET | `/api/v1/db` | API 文档 |

## 快速开始

### 1. 启动 API 服务器

```bash
# 使用 SQLite (默认)
go run cmd/api-server/main.go

# 使用 PostgreSQL
DB_TYPE=postgres \
DB_HOST=localhost \
DB_PORT=5432 \
DB_USER=postgres \
DB_PASSWORD=password \
DB_NAME=kube_tide \
go run cmd/api-server/main.go
```

### 2. 环境变量配置

```bash
# 数据库配置
export DB_TYPE=sqlite                    # 数据库类型: postgres, sqlite
export DB_HOST=localhost                 # PostgreSQL 主机
export DB_PORT=5432                      # PostgreSQL 端口
export DB_USER=postgres                  # PostgreSQL 用户名
export DB_PASSWORD=password              # PostgreSQL 密码
export DB_NAME=kube_tide                 # 数据库名称
export DB_SSL_MODE=disable               # PostgreSQL SSL 模式
export DB_SQLITE_PATH=./data/kube_tide.db # SQLite 文件路径

# 连接池配置
export DB_MAX_OPEN_CONNS=25              # 最大打开连接数
export DB_MAX_IDLE_CONNS=5               # 最大空闲连接数
export DB_CONN_MAX_LIFETIME=5m           # 连接最大生命周期

# 服务器配置
export PORT=8080                         # 服务器端口
```

### 3. 运行测试

```bash
# 启动服务器
go run cmd/api-server/main.go &

# 等待服务器启动
sleep 2

# 运行 API 测试
chmod +x scripts/test-api.sh
./scripts/test-api.sh
```

## API 使用示例

### 创建部署

```bash
curl -X POST http://localhost:8080/api/v1/db/deployments \
  -H "Content-Type: application/json" \
  -d '{
    "cluster_id": "cluster-1",
    "namespace": "default",
    "name": "nginx-deployment",
    "replicas": 3,
    "ready_replicas": 3,
    "available_replicas": 3,
    "unavailable_replicas": 0,
    "updated_replicas": 3,
    "strategy_type": "RollingUpdate",
    "labels": {"app": "nginx", "version": "v1"},
    "annotations": {"deployment.kubernetes.io/revision": "1"},
    "selector": {"app": "nginx"},
    "template": {
      "spec": {
        "containers": [
          {
            "name": "nginx",
            "image": "nginx:1.21"
          }
        ]
      }
    }
  }'
```

### 获取部署列表

```bash
# 获取集群中的所有部署
curl "http://localhost:8080/api/v1/db/clusters/cluster-1/deployments?page=1&page_size=10"

# 获取命名空间中的部署
curl "http://localhost:8080/api/v1/db/clusters/cluster-1/namespaces/default/deployments"
```

### 创建服务

```bash
curl -X POST http://localhost:8080/api/v1/db/services \
  -H "Content-Type: application/json" \
  -d '{
    "cluster_id": "cluster-1",
    "namespace": "default",
    "name": "nginx-service",
    "type": "ClusterIP",
    "cluster_ip": "10.96.0.100",
    "ports": [
      {
        "name": "http",
        "port": 80,
        "target_port": 8080,
        "protocol": "TCP"
      }
    ],
    "selector": {"app": "nginx"},
    "labels": {"app": "nginx", "version": "v1"},
    "annotations": {"service.kubernetes.io/load-balancer-class": "internal"}
  }'
```

### 分页查询

```bash
# 第一页，每页 10 条记录
curl "http://localhost:8080/api/v1/db/clusters/cluster-1/deployments?page=1&page_size=10"

# 第二页，每页 20 条记录
curl "http://localhost:8080/api/v1/db/clusters/cluster-1/deployments?page=2&page_size=20"
```

### 统计数量

```bash
# 统计集群中的部署数量
curl "http://localhost:8080/api/v1/db/clusters/cluster-1/deployments/count"

# 统计命名空间中的服务数量
curl "http://localhost:8080/api/v1/db/clusters/cluster-1/namespaces/default/services/count"
```

## 响应格式

### 成功响应

```json
{
  "deployment": {
    "id": "uuid-string",
    "cluster_id": "cluster-1",
    "namespace": "default",
    "name": "nginx-deployment",
    "replicas": 3,
    "ready_replicas": 3,
    "available_replicas": 3,
    "unavailable_replicas": 0,
    "updated_replicas": 3,
    "strategy_type": "RollingUpdate",
    "labels": {"app": "nginx", "version": "v1"},
    "annotations": {"deployment.kubernetes.io/revision": "1"},
    "selector": {"app": "nginx"},
    "template": {...},
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

### 分页响应

```json
{
  "deployments": [...],
  "total_count": 100,
  "page": 1,
  "page_size": 10,
  "total_pages": 10,
  "has_next": true,
  "has_previous": false
}
```

### 错误响应

```json
{
  "error": "Deployment not found"
}
```

## 数据模型

### 部署模型

```go
type Deployment struct {
    ID                   string                 `json:"id"`
    ClusterID            string                 `json:"cluster_id"`
    Namespace            string                 `json:"namespace"`
    Name                 string                 `json:"name"`
    Replicas             int32                  `json:"replicas"`
    ReadyReplicas        int32                  `json:"ready_replicas"`
    AvailableReplicas    int32                  `json:"available_replicas"`
    UnavailableReplicas  int32                  `json:"unavailable_replicas"`
    UpdatedReplicas      int32                  `json:"updated_replicas"`
    StrategyType         string                 `json:"strategy_type"`
    Labels               map[string]string      `json:"labels"`
    Annotations          map[string]string      `json:"annotations"`
    Selector             map[string]string      `json:"selector"`
    Template             map[string]interface{} `json:"template"`
    CreatedAt            time.Time              `json:"created_at"`
    UpdatedAt            time.Time              `json:"updated_at"`
}
```

### 服务模型

```go
type Service struct {
    ID           string                   `json:"id"`
    ClusterID    string                   `json:"cluster_id"`
    Namespace    string                   `json:"namespace"`
    Name         string                   `json:"name"`
    Type         string                   `json:"type"`
    ClusterIP    string                   `json:"cluster_ip"`
    ExternalIPs  []string                 `json:"external_ips"`
    Ports        []ServicePort            `json:"ports"`
    Selector     map[string]string        `json:"selector"`
    Labels       map[string]string        `json:"labels"`
    Annotations  map[string]string        `json:"annotations"`
    CreatedAt    time.Time                `json:"created_at"`
    UpdatedAt    time.Time                `json:"updated_at"`
}
```

## 性能优化

### 数据库优化

1. **连接池配置**
   - 合理设置最大连接数
   - 配置连接生命周期
   - 监控连接使用情况

2. **查询优化**
   - 使用索引优化查询
   - 实现分页查询
   - 避免 N+1 查询问题

3. **事务管理**
   - 保持事务简短
   - 适当使用事务隔离级别
   - 实现超时控制

### API 优化

1. **响应优化**
   - 实现 GZIP 压缩
   - 使用 HTTP 缓存头
   - 分页大数据集

2. **并发控制**
   - 实现请求限流
   - 使用连接池
   - 优雅关闭

## 监控和日志

### 结构化日志

```go
logger.Info("deployment created",
    zap.String("deployment_id", deployment.ID),
    zap.String("cluster_id", deployment.ClusterID),
    zap.String("namespace", deployment.Namespace),
    zap.String("name", deployment.Name))
```

### 健康检查

```bash
curl http://localhost:8080/health
```

响应：
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T00:00:00Z",
  "services": {
    "database": "healthy"
  }
}
```

## 错误处理

### 常见错误码

- `400 Bad Request`: 请求参数错误
- `404 Not Found`: 资源不存在
- `500 Internal Server Error`: 服务器内部错误
- `503 Service Unavailable`: 数据库连接失败

### 错误响应格式

```json
{
  "error": "详细错误信息",
  "code": "错误代码",
  "details": "额外的错误详情"
}
```

## 安全考虑

1. **输入验证**
   - 验证所有输入参数
   - 防止 SQL 注入
   - 限制请求大小

2. **访问控制**
   - 实现认证机制
   - 添加授权检查
   - 记录访问日志

3. **数据保护**
   - 使用 HTTPS
   - 加密敏感数据
   - 定期备份

## 部署建议

### 生产环境

1. **数据库配置**
   - 使用 PostgreSQL
   - 配置主从复制
   - 定期备份

2. **服务器配置**
   - 使用反向代理
   - 配置负载均衡
   - 实现健康检查

3. **监控告警**
   - 监控 API 响应时间
   - 监控数据库连接
   - 设置告警阈值

## 总结

本实现提供了一个完整的、生产就绪的数据库 API 集成解决方案，包括：

✅ **完整的部署和服务仓储实现**
✅ **RESTful API 层集成**
✅ **分页和过滤功能**
✅ **错误处理和日志记录**
✅ **健康检查和监控**
✅ **多数据库支持**
✅ **事务支持**
✅ **性能优化**

该实现遵循了最佳实践，具有良好的可扩展性和可维护性，可以作为 Kube-Tide 项目的核心数据持久化解决方案。 