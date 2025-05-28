# 数据库迁移文件整合说明

## 概述

本次整合将原来分散的用户角色体系迁移文件（006-009）合并为一个统一的迁移文件，简化了迁移管理并提高了部署效率。

## 整合前的文件结构

```textplain
internal/database/migrations/
├── 006_create_users_table.up.sql
├── 006_create_users_table.down.sql
├── 007_create_roles_permissions.up.sql
├── 007_create_roles_permissions.down.sql
├── 008_create_auth_tables.up.sql
├── 008_create_auth_tables.down.sql
├── 009_insert_initial_data.up.sql
├── 009_insert_initial_data.down.sql
├── migrations.go
└── migrate.go
```

## 整合后的文件结构

```
internal/database/migrations/
├── migrations.go (已更新，包含新的迁移版本7)
└── migrate.go (保持不变)
```

## 变更详情

### 1. 迁移版本调整

- **原版本**: 006, 007, 008, 009 (4个独立迁移)
- **新版本**: 7 (1个整合迁移)

### 2. 迁移内容整合

新的迁移版本7包含以下完整内容：

#### 2.1 用户表结构

- `users` 表：用户基本信息
- 用户表索引和触发器
- 更新时间自动触发器函数

#### 2.2 角色权限体系

- `roles` 表：角色定义
- `permissions` 表：权限定义
- `user_roles` 表：用户角色关联
- `role_permissions` 表：角色权限关联
- 相关索引和约束

#### 2.3 认证相关表

- `user_sessions` 表：用户会话管理
- `audit_logs` 表：审计日志
- 会话清理函数

#### 2.4 初始数据

- 系统权限数据（29个权限）
- 系统角色数据（6个角色）
- 角色权限关联数据
- 默认管理员用户（用户名：admin，密码：admin123）

### 3. 权限体系设计

#### 3.1 权限分类

- **集群权限**: cluster:create, cluster:read, cluster:update, cluster:delete
- **部署权限**: deployment:create, deployment:read, deployment:update, deployment:delete, deployment:scale, deployment:restart
- **服务权限**: service:create, service:read, service:update, service:delete
- **Pod权限**: pod:read, pod:delete, pod:logs, pod:exec
- **节点权限**: node:read, node:update, node:drain, node:cordon
- **命名空间权限**: namespace:read, namespace:create, namespace:delete
- **用户管理权限**: user:create, user:read, user:update, user:delete
- **角色管理权限**: role:create, role:read, role:update, role:delete
- **审计权限**: audit:read

#### 3.2 系统角色

- **super_admin**: 超级管理员，拥有所有权限
- **cluster_admin**: 集群管理员，管理特定集群的所有资源
- **namespace_admin**: 命名空间管理员，管理特定命名空间的资源
- **developer**: 开发者，可以部署和管理应用
- **viewer**: 查看者，只读权限（默认角色）
- **operator**: 运维人员，可以管理基础设施

#### 3.3 权限作用域

- **global**: 全局权限
- **cluster**: 集群级权限
- **namespace**: 命名空间级权限

## 迁移执行

### 使用内嵌迁移系统

项目使用内嵌的迁移系统，迁移定义在 `migrations.go` 文件中。执行迁移的方式：

```bash
# 运行所有迁移
go run cmd/migrate/main.go -action=migrate

# 查看当前版本
go run cmd/migrate/main.go -action=version

# 回滚到指定版本
go run cmd/migrate/main.go -action=rollback -version=6
```

### 数据库配置

支持 PostgreSQL 和 SQLite 数据库：

```bash
# PostgreSQL (默认)
go run cmd/migrate/main.go -type=postgres -host=localhost -port=5432 -user=postgres -password=yourpassword -database=kube_tide

# SQLite
go run cmd/migrate/main.go -type=sqlite -sqlite-file=./data/kube_tide.db
```

## 回滚策略

如果需要回滚到用户角色体系之前的状态：

```bash
go run cmd/migrate/main.go -action=rollback -version=6
```

这将完全删除用户角色体系的所有表和数据。

## 验证

迁移完成后，可以验证以下内容：

1. **表结构**: 确认所有表都已创建
2. **索引**: 确认所有索引都已创建
3. **初始数据**: 确认权限、角色和默认用户已插入
4. **约束**: 确认外键约束和检查约束正常工作

```sql
-- 检查表是否存在
SELECT table_name FROM information_schema.tables 
WHERE table_schema = 'public' AND table_name IN ('users', 'roles', 'permissions', 'user_roles', 'role_permissions', 'user_sessions', 'audit_logs');

-- 检查权限数量
SELECT COUNT(*) FROM permissions;

-- 检查角色数量
SELECT COUNT(*) FROM roles;

-- 检查默认管理员
SELECT username, email, display_name FROM users WHERE username = 'admin';
```

## 注意事项

1. **密码安全**: 默认管理员密码为 `admin123`，生产环境中应立即修改
2. **权限检查**: 确保应用代码中的权限检查逻辑与新的权限名称匹配
3. **数据备份**: 在生产环境执行迁移前，请务必备份数据库
4. **测试验证**: 在测试环境充分验证迁移和回滚功能

## 相关文件

- `internal/database/migrations/migrations.go`: 迁移定义
- `internal/database/migrations/migrate.go`: 迁移执行逻辑
- `cmd/migrate/main.go`: 迁移命令行工具
- `docs/user-role-system.md`: 用户角色体系详细文档

## 更新日期

2024年12月 - 初始整合完成
