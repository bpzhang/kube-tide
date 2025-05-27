# 数据库 API 整合使用说明

## 概述

Kube-Tide 现在支持双轨制架构：
- **Kubernetes API** - 用于实时操作和管理
- **数据库 API** - 用于数据持久化、统计分析和历史查询

## 架构设计

### 双轨制架构
```
前端 → API 层 → 服务层 → 数据层
              ↓
         Kubernetes API (实时)
              ↓
         数据库 API (持久化)
```

### 数据同步机制
- **异步同步**：Kubernetes 操作成功后，异步同步数据到数据库
- **智能更新**：检查数据库中是否已存在记录，存在则更新，不存在则创建
- **错误容错**：数据库操作失败不影响 Kubernetes 操作

## 启用数据库功能

### 环境变量配置
```bash
# 启用数据库功能
export ENABLE_DATABASE=true

# 数据库类型 (sqlite 或 postgres)
export DB_TYPE=sqlite

# SQLite 配置
export DB_SQLITE_PATH=./data/kube_tide.db

# PostgreSQL 配置 (如果使用 postgres)
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=your_password
export DB_NAME=kube_tide
export DB_SSL_MODE=disable

# 连接池配置
export DB_MAX_OPEN_CONNS=25
export DB_MAX_IDLE_CONNS=5
export DB_CONN_MAX_LIFETIME=5m
```

### 启动服务器
```bash
# 使用 SQLite (推荐用于开发和小规模部署)
ENABLE_DATABASE=true DB_TYPE=sqlite ./bin/server

# 使用 PostgreSQL (推荐用于生产环境)
ENABLE_DATABASE=true DB_TYPE=postgres DB_HOST=localhost DB_USER=postgres DB_PASSWORD=password ./bin/server
```

## API 端点

### Kubernetes API (现有)
保持不变，用于实时操作：
```
GET    /api/clusters/{cluster}/deployments
GET    /api/clusters/{cluster}/services
POST   /api/clusters/{cluster}/namespaces/{namespace}/deployments
DELETE /api/clusters/{cluster}/namespaces/{namespace}/deployments/{name}
```

### 数据库 API (新增)
用于数据查询和统计：
```
# 部署相关
GET    /api/v1/db/deployments/{id}                                              # 根据 ID 获取部署
GET    /api/v1/db/clusters/{cluster_id}/deployments                             # 获取集群中的部署列表
GET    /api/v1/db/clusters/{cluster_id}/deployments/count                       # 统计集群中的部署数量
GET    /api/v1/db/clusters/{cluster_id}/namespaces/{namespace}/deployments      # 获取命名空间中的部署列表
GET    /api/v1/db/clusters/{cluster_id}/namespaces/{namespace}/deployments/count # 统计命名空间中的部署数量

# 服务相关
GET    /api/v1/db/services/{id}                                                 # 根据 ID 获取服务
GET    /api/v1/db/clusters/{cluster_id}/services                                # 获取集群中的服务列表
GET    /api/v1/db/clusters/{cluster_id}/services/count                          # 统计集群中的服务数量
GET    /api/v1/db/clusters/{cluster_id}/namespaces/{namespace}/services         # 获取命名空间中的服务列表
GET    /api/v1/db/clusters/{cluster_id}/namespaces/{namespace}/services/count   # 统计命名空间中的服务数量
```

## 使用示例

### 1. 创建部署 (Kubernetes API)
```bash
curl -X POST http://localhost:8080/api/clusters/my-cluster/namespaces/default/deployments \
  -H "Content-Type: application/json" \
  -d '{
    "name": "nginx-deployment",
    "replicas": 3,
    "containers": [
      {
        "name": "nginx",
        "image": "nginx:1.20",
        "ports": [{"containerPort": 80}]
      }
    ]
  }'
```

### 2. 查询部署统计 (数据库 API)
```bash
# 获取集群中的部署数量
curl http://localhost:8080/api/v1/db/clusters/my-cluster/deployments/count

# 获取命名空间中的部署列表 (分页)
curl "http://localhost:8080/api/v1/db/clusters/my-cluster/namespaces/default/deployments?page=1&page_size=10"
```

### 3. 查询服务信息 (数据库 API)
```bash
# 获取集群中的服务列表
curl "http://localhost:8080/api/v1/db/clusters/my-cluster/services?page=1&page_size=20"

# 获取特定服务详情
curl http://localhost:8080/api/v1/db/services/service-uuid
```

## 分页查询

所有列表 API 都支持分页：
```bash
# 参数说明
page=1          # 页码，从 1 开始
page_size=10    # 每页大小，默认 10，最大 100

# 响应格式
{
  "code": 200,
  "message": "success",
  "data": {
    "deployments": [...],
    "total_count": 50,
    "page": 1,
    "page_size": 10,
    "total_pages": 5,
    "has_next": true,
    "has_previous": false
  }
}
```

## 数据模型

### 部署 (Deployment)
```json
{
  "id": "uuid",
  "cluster_id": "cluster-name",
  "namespace": "default",
  "name": "nginx-deployment",
  "replicas": 3,
  "ready_replicas": 3,
  "available_replicas": 3,
  "unavailable_replicas": 0,
  "updated_replicas": 3,
  "strategy_type": "RollingUpdate",
  "labels": "{\"app\":\"nginx\"}",
  "annotations": "{}",
  "selector": "{\"app\":\"nginx\"}",
  "template": "{\"containers\":[...]}",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### 服务 (Service)
```json
{
  "id": "uuid",
  "cluster_id": "cluster-name",
  "namespace": "default",
  "name": "nginx-service",
  "type": "ClusterIP",
  "cluster_ip": "10.96.0.1",
  "external_ips": "[]",
  "ports": "[{\"port\":80,\"targetPort\":80,\"protocol\":\"TCP\"}]",
  "selector": "{\"app\":\"nginx\"}",
  "session_affinity": "None",
  "labels": "{\"app\":\"nginx\"}",
  "annotations": "{}",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

## 前端集成

### 保持现有调用方式
前端代码无需修改，继续使用现有的 API 调用：
```typescript
// 现有的 API 调用保持不变
import { listAllDeployments, getDeploymentDetails } from '@/api/deployment';

// 获取部署列表 (会自动同步到数据库)
const deployments = await listAllDeployments('my-cluster');

// 获取部署详情 (会自动同步到数据库)
const deployment = await getDeploymentDetails('my-cluster', 'default', 'nginx-deployment');
```

### 新增数据库查询 API (可选)
如果需要使用数据库 API 进行统计查询，可以添加新的 API 调用：
```typescript
// 新增的数据库 API 调用
import api from '@/api/axios';

// 获取部署统计
export const getDeploymentCount = (clusterName: string) => {
  return api.get(`/api/v1/db/clusters/${clusterName}/deployments/count`);
};

// 获取历史数据 (分页)
export const getDeploymentHistory = (clusterName: string, page: number = 1, pageSize: number = 10) => {
  return api.get(`/api/v1/db/clusters/${clusterName}/deployments?page=${page}&page_size=${pageSize}`);
};
```

## 监控和日志

### 健康检查
```bash
# 检查服务健康状态
curl http://localhost:8080/health

# 响应示例
{
  "status": "healthy",
  "timestamp": "2024-01-01T00:00:00Z",
  "services": {
    "database": "healthy"
  }
}
```

### 日志记录
- 所有数据库操作都有详细的结构化日志
- 同步失败不会影响 Kubernetes 操作
- 可以通过日志监控数据同步状态

## 最佳实践

### 1. 生产环境建议
- 使用 PostgreSQL 作为数据库
- 配置适当的连接池大小
- 启用数据库备份
- 监控数据库性能

### 2. 开发环境建议
- 使用 SQLite 简化部署
- 启用详细日志记录
- 定期清理测试数据

### 3. 性能优化
- 数据库操作是异步的，不影响 Kubernetes API 性能
- 使用分页查询避免大量数据传输
- 合理设置连接池参数

## 故障排除

### 常见问题

1. **数据库连接失败**
   ```bash
   # 检查数据库配置
   echo $DB_TYPE $DB_HOST $DB_PORT
   
   # 检查数据库服务状态
   systemctl status postgresql  # PostgreSQL
   ls -la ./data/kube_tide.db   # SQLite
   ```

2. **数据同步失败**
   - 查看应用日志中的错误信息
   - 检查数据库权限和连接
   - 验证数据模型是否正确

3. **迁移失败**
   ```bash
   # 手动运行迁移
   ./bin/migrate up
   
   # 检查迁移状态
   ./bin/migrate version
   ```

## 扩展开发

### 添加新的资源类型
1. 创建数据模型 (`internal/database/models/`)
2. 实现仓储接口 (`internal/repository/`)
3. 创建服务层 (`internal/core/`)
4. 添加 API 处理器 (`internal/api/`)
5. 更新路由配置

### 自定义查询
可以在仓储层添加自定义查询方法，支持复杂的统计和分析需求。

## 总结

数据库 API 整合为 Kube-Tide 提供了强大的数据持久化和分析能力，同时保持了原有 Kubernetes API 的实时性和可靠性。通过双轨制架构，用户可以：

- 使用 Kubernetes API 进行实时操作
- 使用数据库 API 进行历史查询和统计分析
- 享受异步数据同步带来的性能优势
- 获得完整的数据持久化保障 