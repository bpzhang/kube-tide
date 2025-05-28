# Kube-Tide 用户角色体系开发文档

## 📋 目录

- [项目概述](#项目概述)
- [功能特性](#功能特性)
- [系统架构](#系统架构)
- [技术栈](#技术栈)
- [数据模型](#数据模型)
- [API 设计](#api-设计)
- [权限控制](#权限控制)
- [安装部署](#安装部署)
- [开发指南](#开发指南)
- [测试指南](#测试指南)
- [故障排除](#故障排除)

## 项目概述

Kube-Tide 是一个基于 Go 和 React 的 Kubernetes 多集群管理平台，现已完成**完整的用户角色体系**开发。该系统提供了企业级的用户管理、角色权限控制和审计日志功能，支持细粒度的权限管理和多级作用域控制。

### 🎯 核心目标

- **安全性**: 提供企业级的认证授权机制
- **灵活性**: 支持多级权限作用域（全局、集群、命名空间）
- **可扩展性**: 模块化设计，易于扩展新功能
- **易用性**: 直观的 API 设计和完善的文档

## 功能特性

### ✅ 已完成功能

#### 🔐 认证系统

- **JWT 令牌认证**: 基于 JWT 的无状态认证
- **会话管理**: 用户会话创建、验证和销毁
- **密码安全**: bcrypt 密码哈希和验证
- **令牌刷新**: 支持令牌自动刷新机制

#### 👥 用户管理

- **用户 CRUD**: 完整的用户创建、读取、更新、删除操作
- **用户状态管理**: 支持激活、停用、暂停等状态
- **密码管理**: 用户密码修改和管理员重置
- **用户查询**: 支持多条件查询和分页

#### 🎭 角色管理

- **角色 CRUD**: 完整的角色管理功能
- **系统角色保护**: 防止误删系统关键角色
- **默认角色**: 支持新用户自动分配默认角色
- **角色权限关联**: 灵活的角色权限绑定

#### 🔑 权限管理

- **细粒度权限**: 支持资源级别的权限控制
- **权限作用域**: 全局、集群、命名空间三级作用域
- **通配符权限**: 支持 `*` 通配符权限
- **权限检查**: 高效的权限验证机制

#### 📊 审计日志

- **操作记录**: 记录所有关键操作
- **用户追踪**: 追踪用户行为和操作历史
- **安全审计**: 支持安全审计和合规检查
- **日志查询**: 支持审计日志查询和分析

#### 🛡️ 中间件系统

- **认证中间件**: 自动验证用户身份
- **权限中间件**: 基于路由的权限检查
- **角色中间件**: 基于角色的访问控制
- **可选认证**: 支持可选的用户认证

### 🚧 待完成功能

#### API 处理器层

- [ ] 用户管理 API 处理器 (UserHandler)
- [ ] 角色管理 API 处理器 (RoleHandler)
- [ ] 权限管理 API 处理器 (PermissionHandler)
- [ ] 审计日志 API 处理器 (AuditHandler)

#### 前端集成

- [ ] 用户管理界面
- [ ] 角色权限配置界面
- [ ] 审计日志查看界面
- [ ] 权限检查组件

## 系统架构

### 🏗️ 整体架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   前端 React    │───▶│   API 网关      │───▶│  Kubernetes     │
│   用户界面      │    │   Gin Router    │    │   集群          │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │   认证中间件    │
                       │   权限检查      │
                       └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │   业务服务层    │
                       │   Service Layer │
                       └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │   数据访问层    │
                       │   Repository    │
                       └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │   数据库        │
                       │   PostgreSQL    │
                       └─────────────────┘
```

### 📦 分层架构

#### 1. API 层 (internal/api/)

- **路由管理**: 定义 RESTful API 路由
- **请求处理**: 处理 HTTP 请求和响应
- **参数验证**: 验证请求参数和数据
- **错误处理**: 统一的错误响应格式

#### 2. 中间件层 (internal/api/middleware/)

- **认证中间件**: JWT 令牌验证
- **权限中间件**: 基于权限的访问控制
- **角色中间件**: 基于角色的访问控制
- **日志中间件**: 请求日志记录

#### 3. 服务层 (internal/core/)

- **业务逻辑**: 核心业务逻辑实现
- **数据验证**: 业务数据验证
- **事务管理**: 跨仓储的事务处理
- **审计日志**: 操作审计记录

#### 4. 仓储层 (internal/repository/)

- **数据访问**: 数据库操作抽象
- **查询优化**: 高效的数据查询
- **事务支持**: 数据库事务管理
- **连接池**: 数据库连接池管理

#### 5. 数据层 (internal/database/)

- **数据模型**: 数据结构定义
- **数据库迁移**: 数据库版本管理
- **连接管理**: 数据库连接配置

## 技术栈

### 🔧 后端技术栈

| 组件 | 技术选型 | 版本 | 说明 |
|------|----------|------|------|
| **Web 框架** | Gin | v1.9+ | 高性能 HTTP Web 框架 |
| **数据库** | PostgreSQL | 13+ | 企业级关系型数据库 |
| **ORM** | 原生 SQL | - | 使用 database/sql 包 |
| **认证** | JWT | v5+ | JSON Web Token 认证 |
| **密码加密** | bcrypt | - | 安全的密码哈希算法 |
| **日志** | Zap | v1.24+ | 高性能结构化日志 |
| **UUID** | Google UUID | v1.3+ | UUID 生成库 |
| **验证** | Validator | v10+ | 数据验证库 |

### 🎨 前端技术栈

| 组件 | 技术选型 | 版本 | 说明 |
|------|----------|------|------|
| **框架** | React | 18+ | 现代化前端框架 |
| **语言** | TypeScript | 4.9+ | 类型安全的 JavaScript |
| **UI 库** | Ant Design | 5+ | 企业级 UI 组件库 |
| **状态管理** | Redux Toolkit | 1.9+ | 现代化状态管理 |
| **HTTP 客户端** | Axios | 1.4+ | Promise 基础的 HTTP 库 |
| **构建工具** | Vite | 4+ | 快速的前端构建工具 |

## 数据模型

### 📊 核心实体关系图

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│    User     │    │  UserRole   │    │    Role     │
│             │───▶│             │◀───│             │
│ - id        │    │ - user_id   │    │ - id        │
│ - username  │    │ - role_id   │    │ - name      │
│ - email     │    │ - scope     │    │ - type      │
│ - status    │    │ - expires   │    │ - default   │
└─────────────┘    └─────────────┘    └─────────────┘
                                             │
                                             ▼
                                      ┌─────────────┐
                                      │RolePermission│
                                      │             │
                                      │ - role_id   │
                                      │ - perm_id   │
                                      └─────────────┘
                                             │
                                             ▼
                                      ┌─────────────┐
                                      │ Permission  │
                                      │             │
                                      │ - id        │
                                      │ - name      │
                                      │ - resource  │
                                      │ - action    │
                                      │ - scope     │
                                      └─────────────┘
```

### 🗃️ 数据表结构

#### 用户表 (users)

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    display_name VARCHAR(100),
    avatar_url TEXT,
    status VARCHAR(20) DEFAULT 'active',
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES users(id)
);
```

#### 角色表 (roles)

```sql
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) UNIQUE NOT NULL,
    display_name VARCHAR(100),
    description TEXT,
    type VARCHAR(20) DEFAULT 'custom',
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES users(id)
);
```

#### 权限表 (permissions)

```sql
CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    display_name VARCHAR(100),
    description TEXT,
    resource_type VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    scope VARCHAR(20) DEFAULT 'global'
);
```

#### 用户角色关联表 (user_roles)

```sql
CREATE TABLE user_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    scope_type VARCHAR(20) DEFAULT 'global',
    scope_value VARCHAR(255),
    granted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    granted_by UUID REFERENCES users(id),
    expires_at TIMESTAMP WITH TIME ZONE
);
```

#### 角色权限关联表 (role_permissions)

```sql
CREATE TABLE role_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

#### 用户会话表 (user_sessions)

```sql
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    refresh_token_hash VARCHAR(255),
    ip_address INET,
    user_agent TEXT,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_used_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

#### 审计日志表 (audit_logs)

```sql
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    action VARCHAR(50) NOT NULL,
    resource VARCHAR(50) NOT NULL,
    resource_type VARCHAR(50),
    resource_id VARCHAR(255),
    cluster_name VARCHAR(100),
    namespace VARCHAR(100),
    result VARCHAR(20) NOT NULL,
    details TEXT,
    ip_address INET,
    user_agent TEXT,
    status VARCHAR(20) DEFAULT 'success',
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

## API 设计

### 🔗 RESTful API 端点

#### 认证 API

```
POST   /api/auth/login           # 用户登录
POST   /api/auth/logout          # 用户登出
POST   /api/auth/register        # 用户注册
POST   /api/auth/refresh         # 刷新令牌
POST   /api/auth/change-password # 修改密码
GET    /api/auth/me              # 获取当前用户信息
```

#### 用户管理 API

```
GET    /api/users                # 获取用户列表
POST   /api/users                # 创建用户
GET    /api/users/:id             # 获取用户详情
PUT    /api/users/:id             # 更新用户信息
DELETE /api/users/:id             # 删除用户
GET    /api/users/:id/roles       # 获取用户角色
POST   /api/users/:id/roles       # 分配角色
DELETE /api/users/:id/roles/:role # 移除角色
```

#### 角色管理 API

```
GET    /api/roles                     # 获取角色列表
POST   /api/roles                     # 创建角色
GET    /api/roles/:id                 # 获取角色详情
PUT    /api/roles/:id                 # 更新角色信息
DELETE /api/roles/:id                 # 删除角色
GET    /api/roles/:id/permissions     # 获取角色权限
POST   /api/roles/:id/permissions     # 分配权限
DELETE /api/roles/:id/permissions     # 移除权限
```

#### 权限管理 API

```
GET    /api/permissions               # 获取权限列表
GET    /api/permissions/:id           # 获取权限详情
POST   /api/permissions/check         # 检查权限
GET    /api/permissions/resources     # 按资源分组权限
```

#### 审计日志 API

```
GET    /api/audit-logs               # 获取审计日志列表
GET    /api/audit-logs/:id           # 获取审计日志详情
```

### 📝 API 请求/响应示例

#### 用户登录

```bash
# 请求
POST /api/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "password123"
}

# 响应
{
  "code": 0,
  "message": "登录成功",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2024-01-01T12:00:00Z",
    "user": {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "username": "admin",
      "email": "admin@example.com",
      "display_name": "系统管理员",
      "status": "active"
    }
  }
}
```

#### 创建用户

```bash
# 请求
POST /api/users
Authorization: Bearer <token>
Content-Type: application/json

{
  "username": "newuser",
  "email": "newuser@example.com",
  "password": "password123",
  "display_name": "新用户",
  "assign_default_role": true
}

# 响应
{
  "code": 0,
  "message": "用户创建成功",
  "data": {
    "id": "456e7890-e89b-12d3-a456-426614174001",
    "username": "newuser",
    "email": "newuser@example.com",
    "display_name": "新用户",
    "status": "active",
    "created_at": "2024-01-01T10:00:00Z"
  }
}
```

#### 权限检查

```bash
# 请求
POST /api/permissions/check
Authorization: Bearer <token>
Content-Type: application/json

{
  "resource": "deployment",
  "action": "create",
  "cluster_name": "prod-cluster",
  "namespace": "default"
}

# 响应
{
  "code": 0,
  "message": "权限检查完成",
  "data": {
    "allowed": true,
    "reason": ""
  }
}
```

## 权限控制

### 🔐 权限模型

#### 权限作用域

1. **全局 (global)**: 对所有集群和命名空间有效
2. **集群 (cluster)**: 对特定集群有效
3. **命名空间 (namespace)**: 对特定命名空间有效

#### 权限格式

权限名称格式：`resource:action`

- **resource**: 资源类型（如 deployment, service, pod）
- **action**: 操作类型（如 create, read, update, delete）

#### 通配符支持

- `*:*`: 所有资源的所有操作
- `deployment:*`: 部署资源的所有操作
- `*:read`: 所有资源的读取操作

### 🛡️ 权限检查流程

```
1. 提取用户令牌 → 2. 验证令牌有效性 → 3. 获取用户角色
                                              ↓
6. 返回检查结果 ← 5. 匹配权限规则 ← 4. 获取角色权限
```

#### 权限检查算法

```go
func CheckPermission(userID, resource, action, scope string) bool {
    // 1. 获取用户在指定作用域的所有权限
    permissions := GetUserPermissions(userID, scope)
    
    // 2. 检查精确匹配
    if HasPermission(permissions, resource, action) {
        return true
    }
    
    // 3. 检查通配符权限
    if HasPermission(permissions, resource, "*") ||
       HasPermission(permissions, "*", action) ||
       HasPermission(permissions, "*", "*") {
        return true
    }
    
    return false
}
```

### 🎭 预定义角色

#### 系统管理员 (system-admin)

- 权限: `*:*` (全局)
- 描述: 拥有系统所有权限

#### 集群管理员 (cluster-admin)

- 权限: `*:*` (集群级别)
- 描述: 拥有特定集群的所有权限

#### 开发者 (developer)

- 权限:
  - `deployment:*` (命名空间级别)
  - `service:*` (命名空间级别)
  - `pod:read,logs,exec` (命名空间级别)
- 描述: 开发人员常用权限

#### 只读用户 (viewer)

- 权限: `*:read` (指定作用域)
- 描述: 只能查看资源，不能修改

## 安装部署

### 🚀 快速开始

#### 1. 环境要求

- Go 1.19+
- PostgreSQL 13+
- Node.js 16+
- Git

#### 2. 克隆项目

```bash
git clone https://github.com/your-org/kube-tide.git
cd kube-tide
```

#### 3. 配置数据库

```bash
# 创建数据库
createdb kube_tide

# 设置环境变量
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=your_password
export DB_NAME=kube_tide
export JWT_SECRET=your_jwt_secret
```

#### 4. 运行数据库迁移

```bash
go run cmd/migrate/main.go up
```

#### 5. 启动后端服务

```bash
go run cmd/server/main.go
```

#### 6. 启动前端服务

```bash
cd web
npm install
npm run dev
```

#### 7. 访问应用

- 前端: <http://localhost:3000>
- 后端: <http://localhost:8080>

### 🐳 Docker 部署

#### 1. 使用 Docker Compose

```yaml
version: '3.8'
services:
  postgres:
    image: postgres:13
    environment:
      POSTGRES_DB: kube_tide
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  kube-tide:
    build: .
    environment:
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: postgres
      DB_PASSWORD: password
      DB_NAME: kube_tide
      JWT_SECRET: your_jwt_secret
    ports:
      - "8080:8080"
    depends_on:
      - postgres

volumes:
  postgres_data:
```

#### 2. 启动服务

```bash
docker-compose up -d
```

### ☸️ Kubernetes 部署

#### 1. 创建配置文件

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: kube-tide-config
data:
  DB_HOST: "postgres-service"
  DB_PORT: "5432"
  DB_NAME: "kube_tide"
---
apiVersion: v1
kind: Secret
metadata:
  name: kube-tide-secret
type: Opaque
stringData:
  DB_USER: "postgres"
  DB_PASSWORD: "password"
  JWT_SECRET: "your_jwt_secret"
```

#### 2. 部署应用

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kube-tide
spec:
  replicas: 3
  selector:
    matchLabels:
      app: kube-tide
  template:
    metadata:
      labels:
        app: kube-tide
    spec:
      containers:
      - name: kube-tide
        image: kube-tide:latest
        ports:
        - containerPort: 8080
        envFrom:
        - configMapRef:
            name: kube-tide-config
        - secretRef:
            name: kube-tide-secret
```

## 开发指南

### 🛠️ 开发环境设置

#### 1. 安装依赖

```bash
# Go 依赖
go mod tidy

# 前端依赖
cd web && npm install
```

#### 2. 代码规范

- 遵循 [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- 使用 `gofmt` 格式化代码
- 使用 `golint` 检查代码质量
- 编写单元测试，覆盖率 > 80%

#### 3. 提交规范

```bash
# 提交格式
<type>(<scope>): <subject>

# 示例
feat(auth): add JWT token refresh functionality
fix(user): resolve user creation validation issue
docs(api): update API documentation
```

### 📁 项目结构

```
kube-tide/
├── cmd/                          # 应用程序入口
│   ├── server/                   # 主服务器
│   └── migrate/                  # 数据库迁移工具
├── internal/                     # 内部包
│   ├── api/                      # API 层
│   │   ├── middleware/           # 中间件
│   │   │   ├── auth.go          # 认证中间件
│   │   │   └── language.go      # 语言中间件
│   │   ├── auth_handler.go      # 认证处理器
│   │   ├── user_handler.go      # 用户处理器
│   │   ├── role_handler.go      # 角色处理器
│   │   └── response.go          # 响应工具
│   ├── core/                     # 业务逻辑层
│   │   ├── auth_service.go      # 认证服务
│   │   ├── user_service.go      # 用户服务
│   │   ├── role_service.go      # 角色服务
│   │   └── permission_service.go # 权限服务
│   ├── repository/               # 数据访问层
│   │   ├── user_repository.go   # 用户仓储
│   │   ├── role_repository.go   # 角色仓储
│   │   └── auth_repository.go   # 认证仓储
│   ├── database/                 # 数据库相关
│   │   ├── models/              # 数据模型
│   │   ├── migrations/          # 数据库迁移
│   │   └── connection.go        # 数据库连接
│   └── utils/                    # 工具函数
│       ├── errors.go            # 错误定义
│       ├── pagination.go        # 分页工具
│       └── validator.go         # 验证工具
├── web/                          # 前端代码
│   ├── src/
│   │   ├── components/          # React 组件
│   │   ├── pages/               # 页面组件
│   │   ├── api/                 # API 客户端
│   │   └── utils/               # 工具函数
│   └── package.json
├── docs/                         # 文档
├── scripts/                      # 脚本文件
├── Dockerfile                    # Docker 构建文件
├── docker-compose.yml           # Docker Compose 配置
└── Makefile                     # 构建脚本
```

### 🔧 添加新功能

#### 1. 添加新的权限

```go
// 1. 在数据库中添加权限记录
INSERT INTO permissions (name, display_name, resource_type, action, scope)
VALUES ('configmap:create', '创建配置映射', 'configmap', 'create', 'namespace');

// 2. 在代码中使用权限检查
func (h *ConfigMapHandler) CreateConfigMap(c *gin.Context) {
    // 权限检查会自动通过中间件进行
}

// 3. 在路由中添加权限中间件
router.POST("/configmaps", 
    authMiddleware.RequireAuth(),
    authMiddleware.RequirePermission("create", "configmap"),
    handler.CreateConfigMap)
```

#### 2. 添加新的角色

```go
// 在迁移文件中添加新角色
INSERT INTO roles (name, display_name, description, type, is_default)
VALUES ('namespace-admin', '命名空间管理员', '管理特定命名空间的所有资源', 'system', false);

// 为角色分配权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'namespace-admin'
  AND p.resource_type IN ('deployment', 'service', 'pod', 'configmap')
  AND p.scope = 'namespace';
```

### 🧪 测试指南

#### 1. 单元测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./internal/core

# 运行测试并显示覆盖率
go test -cover ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

#### 2. 集成测试

```bash
# 启动测试数据库
docker run -d --name test-postgres \
  -e POSTGRES_DB=kube_tide_test \
  -e POSTGRES_USER=test \
  -e POSTGRES_PASSWORD=test \
  -p 5433:5432 postgres:13

# 运行集成测试
DB_HOST=localhost \
DB_PORT=5433 \
DB_USER=test \
DB_PASSWORD=test \
DB_NAME=kube_tide_test \
go test -tags=integration ./...
```

#### 3. API 测试

```bash
# 使用提供的测试脚本
chmod +x scripts/test-api.sh
./scripts/test-api.sh

# 或使用 curl 手动测试
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

### 📊 性能优化

#### 1. 数据库优化

```sql
-- 添加索引
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_expires ON user_roles(expires_at);
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_timestamp ON audit_logs(timestamp);

-- 分区表（大量审计日志）
CREATE TABLE audit_logs_2024 PARTITION OF audit_logs
FOR VALUES FROM ('2024-01-01') TO ('2025-01-01');
```

#### 2. 缓存策略

```go
// 使用 Redis 缓存用户权限
func (s *PermissionService) GetUserPermissions(ctx context.Context, userID string) ([]*models.Permission, error) {
    // 1. 尝试从缓存获取
    cacheKey := fmt.Sprintf("user_permissions:%s", userID)
    if cached := s.cache.Get(cacheKey); cached != nil {
        return cached.([]*models.Permission), nil
    }
    
    // 2. 从数据库获取
    permissions, err := s.permissionRepo.GetUserPermissions(ctx, userID)
    if err != nil {
        return nil, err
    }
    
    // 3. 存入缓存
    s.cache.Set(cacheKey, permissions, 5*time.Minute)
    return permissions, nil
}
```

## 故障排除

### 🐛 常见问题

#### 1. 数据库连接失败

```bash
# 检查数据库连接
psql -h localhost -p 5432 -U postgres -d kube_tide

# 检查环境变量
echo $DB_HOST $DB_PORT $DB_USER $DB_NAME

# 检查数据库日志
docker logs postgres-container
```

#### 2. JWT 令牌验证失败

```bash
# 检查 JWT 密钥配置
echo $JWT_SECRET

# 验证令牌格式
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/auth/me

# 检查令牌过期时间
# 令牌默认有效期为 24 小时
```

#### 3. 权限检查失败

```bash
# 检查用户角色
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/users/me/roles

# 检查角色权限
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/roles/<role_id>/permissions

# 检查权限定义
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/permissions
```

#### 4. 数据库迁移失败

```bash
# 检查迁移状态
go run cmd/migrate/main.go status

# 手动运行迁移
go run cmd/migrate/main.go up

# 回滚迁移
go run cmd/migrate/main.go down 1
```

### 📋 调试技巧

#### 1. 启用调试日志

```bash
# 设置日志级别
export LOG_LEVEL=debug

# 启用 SQL 查询日志
export DB_LOG_QUERIES=true
```

#### 2. 使用调试工具

```bash
# 使用 delve 调试器
go install github.com/go-delve/delve/cmd/dlv@latest
dlv debug cmd/server/main.go

# 在代码中添加断点
runtime.Breakpoint()
```

#### 3. 监控和指标

```go
// 添加 Prometheus 指标
var (
    authAttempts = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "auth_attempts_total",
            Help: "Total number of authentication attempts",
        },
        []string{"status"},
    )
)

// 在认证处理器中使用
func (h *AuthHandler) Login(c *gin.Context) {
    // ... 认证逻辑
    if err != nil {
        authAttempts.WithLabelValues("failed").Inc()
        return
    }
    authAttempts.WithLabelValues("success").Inc()
}
```

## 📚 参考资料

### 🔗 相关文档

- [Kubernetes RBAC 文档](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)
- [JWT 规范](https://tools.ietf.org/html/rfc7519)
- [Go 编码规范](https://golang.org/doc/effective_go.html)
- [PostgreSQL 文档](https://www.postgresql.org/docs/)

### 📖 推荐阅读

- [微服务安全模式](https://microservices.io/patterns/security/)
- [RESTful API 设计指南](https://restfulapi.net/)
- [数据库设计最佳实践](https://www.postgresql.org/docs/current/ddl-best-practices.html)

### 🛠️ 工具推荐

- [Postman](https://www.postman.com/) - API 测试工具
- [pgAdmin](https://www.pgadmin.org/) - PostgreSQL 管理工具
- [JWT.io](https://jwt.io/) - JWT 令牌调试工具
- [Grafana](https://grafana.com/) - 监控和可视化工具

---

## 📞 联系我们

如有问题或建议，请通过以下方式联系：

- 📧 Email: <support@kube-tide.com>
- 🐛 Issues: [GitHub Issues](https://github.com/your-org/kube-tide/issues)
- 📖 Wiki: [项目 Wiki](https://github.com/your-org/kube-tide/wiki)
- 💬 讨论: [GitHub Discussions](https://github.com/your-org/kube-tide/discussions)

---

*最后更新时间: 2024-01-01*
*文档版本: v1.0.0*
