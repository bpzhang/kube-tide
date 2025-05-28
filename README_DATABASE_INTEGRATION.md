# Kube-Tide 数据库 API 整合

## 🎉 新功能发布

Kube-Tide 现在支持**数据库 API 整合**功能！这是一个重大更新，为平台带来了强大的数据持久化和分析能力。

## ✨ 主要特性

### 🔄 双轨制架构

- **Kubernetes API** - 实时操作和管理
- **数据库 API** - 数据持久化、统计分析和历史查询

### 🚀 核心优势

- ✅ **无缝集成** - 前端代码无需修改
- ✅ **异步同步** - 不影响 Kubernetes API 性能
- ✅ **智能更新** - 自动检测和更新数据库记录
- ✅ **错误容错** - 数据库故障不影响 Kubernetes 操作
- ✅ **多数据库支持** - SQLite (开发) 和 PostgreSQL (生产)

## 🚀 快速开始

### 1. 启用数据库功能

```bash
# 使用 SQLite (推荐用于开发)
export ENABLE_DATABASE=true
export DB_TYPE=sqlite
export DB_SQLITE_PATH=./data/kube_tide.db

# 启动服务器
./bin/server
```

### 2. 使用现有 API (自动同步)

```bash
# 创建部署 (会自动同步到数据库)
curl -X POST http://localhost:8080/api/clusters/my-cluster/namespaces/default/deployments \
  -H "Content-Type: application/json" \
  -d '{"name": "nginx", "replicas": 3, "containers": [{"name": "nginx", "image": "nginx:1.20"}]}'

# 查询部署 (会自动同步到数据库)
curl http://localhost:8080/api/clusters/my-cluster/deployments
```

### 3. 使用数据库 API (新增)

```bash
# 获取部署统计
curl http://localhost:8080/api/v1/db/clusters/my-cluster/deployments/count

# 获取历史数据 (分页)
curl "http://localhost:8080/api/v1/db/clusters/my-cluster/deployments?page=1&page_size=10"
```

## 📊 API 对比

| 功能 | Kubernetes API | 数据库 API | 说明 |
|------|----------------|------------|------|
| 实时操作 | ✅ | ❌ | 创建、更新、删除资源 |
| 数据查询 | ✅ | ✅ | 获取资源信息 |
| 历史记录 | ❌ | ✅ | 查看历史数据 |
| 统计分析 | ❌ | ✅ | 数量统计、趋势分析 |
| 分页查询 | ❌ | ✅ | 大数据量分页处理 |
| 性能 | 实时 | 缓存 | 不同的性能特征 |

## 🏗️ 架构图

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   前端 UI   │───▶│  API 网关   │───▶│ Kubernetes  │
└─────────────┘    └─────────────┘    │   集群      │
                          │           └─────────────┘
                          ▼
                   ┌─────────────┐           ▲
                   │  数据库 API │           │
                   └─────────────┘           │
                          │              异步同步
                          ▼                 │
                   ┌─────────────┐           │
                   │   数据库    │───────────┘
                   │ (SQLite/PG) │
                   └─────────────┘
```

## 🔧 配置选项

### 环境变量

```bash
# 基本配置
ENABLE_DATABASE=true          # 启用数据库功能
DB_TYPE=sqlite               # 数据库类型: sqlite 或 postgres

# SQLite 配置
DB_SQLITE_PATH=./data/kube_tide.db

# PostgreSQL 配置
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=kube_tide
DB_SSL_MODE=disable

# 连接池配置
DB_MAX_OPEN_CONNS=25         # 最大打开连接数
DB_MAX_IDLE_CONNS=5          # 最大空闲连接数
DB_CONN_MAX_LIFETIME=5m      # 连接最大生存时间
```

## 📈 使用场景

### 1. 开发环境

```bash
# 使用 SQLite，简单快速
ENABLE_DATABASE=true DB_TYPE=sqlite ./bin/server
```

### 2. 生产环境

```bash
# 使用 PostgreSQL，高性能高可用
ENABLE_DATABASE=true \
DB_TYPE=postgres \
DB_HOST=postgres.example.com \
DB_USER=kube_tide \
DB_PASSWORD=secure_password \
./bin/server
```

### 3. 数据分析

```bash
# 获取集群资源统计
curl http://localhost:8080/api/v1/db/clusters/prod/deployments/count
curl http://localhost:8080/api/v1/db/clusters/prod/services/count

# 获取历史数据进行分析
curl "http://localhost:8080/api/v1/db/clusters/prod/deployments?page=1&page_size=100"
```

## 🧪 测试

运行集成测试：

```bash
# 运行数据库集成测试
./scripts/test-database-integration.sh
```

## 📚 文档

- [详细使用说明](docs/database-api-integration.md)
- [API 参考文档](docs/api-reference.md)
- [架构设计文档](docs/architecture.md)

## 🔄 迁移指南

### 从旧版本升级

1. **无需修改前端代码** - 所有现有 API 保持兼容
2. **可选启用数据库** - 通过环境变量控制
3. **渐进式采用** - 可以逐步使用新的数据库 API

### 数据迁移

```bash
# 首次启动会自动运行数据库迁移
ENABLE_DATABASE=true ./bin/server

# 手动运行迁移（如果需要）
./bin/migrate up
```

## 🛠️ 开发指南

### 添加新资源类型

1. 创建数据模型 (`internal/database/models/`)
2. 实现仓储接口 (`internal/repository/`)
3. 创建服务层 (`internal/core/`)
4. 添加 API 处理器 (`internal/api/`)
5. 更新路由配置

### 自定义查询

```go
// 在仓储层添加自定义查询
func (r *DeploymentRepository) GetByDateRange(ctx context.Context, start, end time.Time) ([]*models.Deployment, error) {
    // 实现自定义查询逻辑
}
```

## 🐛 故障排除

### 常见问题

1. **数据库连接失败**

   ```bash
   # 检查配置
   echo $ENABLE_DATABASE $DB_TYPE $DB_HOST
   
   # 检查数据库服务
   systemctl status postgresql
   ```

2. **数据同步失败**
   - 查看应用日志
   - 检查数据库权限
   - 验证网络连接

3. **性能问题**
   - 调整连接池大小
   - 使用分页查询
   - 监控数据库性能

## 🤝 贡献

欢迎贡献代码和反馈！

1. Fork 项目
2. 创建功能分支
3. 提交更改
4. 创建 Pull Request

## 📄 许可证

本项目采用 MIT 许可证。详见 [LICENSE](LICENSE) 文件。

---

## 🎯 下一步计划

- [ ] 支持更多资源类型 (Pod, Node, ConfigMap 等)
- [ ] 添加数据导出功能
- [ ] 实现实时数据同步
- [ ] 添加数据可视化面板
- [ ] 支持多租户数据隔离

---

**🎉 享受新的数据库 API 整合功能！如有问题，请查看文档或提交 Issue。**
