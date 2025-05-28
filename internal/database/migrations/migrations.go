package migrations

// getAllMigrations returns all available migrations
func getAllMigrations() []Migration {
	return []Migration{
		{
			Version:     1,
			Description: "Create clusters table",
			UpSQL:       createClustersTableUp,
			DownSQL:     createClustersTableDown,
		},
		{
			Version:     2,
			Description: "Create nodes table",
			UpSQL:       createNodesTableUp,
			DownSQL:     createNodesTableDown,
		},
		{
			Version:     3,
			Description: "Create pods table",
			UpSQL:       createPodsTableUp,
			DownSQL:     createPodsTableDown,
		},
		{
			Version:     4,
			Description: "Create namespaces table",
			UpSQL:       createNamespacesTableUp,
			DownSQL:     createNamespacesTableDown,
		},
		{
			Version:     5,
			Description: "Create deployments table",
			UpSQL:       createDeploymentsTableUp,
			DownSQL:     createDeploymentsTableDown,
		},
		{
			Version:     6,
			Description: "Create services table",
			UpSQL:       createServicesTableUp,
			DownSQL:     createServicesTableDown,
		},
		{
			Version:     7,
			Description: "Create user role system",
			UpSQL:       createUserRoleSystemUp,
			DownSQL:     createUserRoleSystemDown,
		},
	}
}

// Migration 1: Create clusters table
const createClustersTableUp = `
CREATE TABLE clusters (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    config TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'inactive',
    description TEXT,
    kubeconfig TEXT,
    endpoint VARCHAR(255),
    version VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_clusters_name ON clusters(name);
CREATE INDEX idx_clusters_status ON clusters(status);
CREATE INDEX idx_clusters_created_at ON clusters(created_at);
`

const createClustersTableDown = `
DROP INDEX IF EXISTS idx_clusters_created_at;
DROP INDEX IF EXISTS idx_clusters_status;
DROP INDEX IF EXISTS idx_clusters_name;
DROP TABLE IF EXISTS clusters;
`

// Migration 2: Create nodes table
const createNodesTableUp = `
CREATE TABLE nodes (
    id VARCHAR(36) PRIMARY KEY,
    cluster_id VARCHAR(36) NOT NULL,
    name VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL,
    roles TEXT,
    age VARCHAR(50),
    version VARCHAR(50),
    internal_ip VARCHAR(45),
    external_ip VARCHAR(45),
    os_image VARCHAR(255),
    kernel_version VARCHAR(100),
    container_runtime VARCHAR(100),
    cpu_capacity VARCHAR(20),
    memory_capacity VARCHAR(20),
    cpu_allocatable VARCHAR(20),
    memory_allocatable VARCHAR(20),
    conditions TEXT,
    labels TEXT,
    annotations TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (cluster_id) REFERENCES clusters(id) ON DELETE CASCADE
);

CREATE INDEX idx_nodes_cluster_id ON nodes(cluster_id);
CREATE INDEX idx_nodes_name ON nodes(name);
CREATE INDEX idx_nodes_status ON nodes(status);
CREATE INDEX idx_nodes_created_at ON nodes(created_at);
`

const createNodesTableDown = `
DROP INDEX IF EXISTS idx_nodes_created_at;
DROP INDEX IF EXISTS idx_nodes_status;
DROP INDEX IF EXISTS idx_nodes_name;
DROP INDEX IF EXISTS idx_nodes_cluster_id;
DROP TABLE IF EXISTS nodes;
`

// Migration 3: Create pods table
const createPodsTableUp = `
CREATE TABLE pods (
    id VARCHAR(36) PRIMARY KEY,
    cluster_id VARCHAR(36) NOT NULL,
    namespace VARCHAR(100) NOT NULL,
    name VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL,
    phase VARCHAR(20),
    node_name VARCHAR(100),
    pod_ip VARCHAR(45),
    host_ip VARCHAR(45),
    restart_count INTEGER DEFAULT 0,
    ready_containers INTEGER DEFAULT 0,
    total_containers INTEGER DEFAULT 0,
    cpu_requests VARCHAR(20),
    memory_requests VARCHAR(20),
    cpu_limits VARCHAR(20),
    memory_limits VARCHAR(20),
    labels TEXT,
    annotations TEXT,
    owner_references TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (cluster_id) REFERENCES clusters(id) ON DELETE CASCADE
);

CREATE INDEX idx_pods_cluster_id ON pods(cluster_id);
CREATE INDEX idx_pods_namespace ON pods(namespace);
CREATE INDEX idx_pods_name ON pods(name);
CREATE INDEX idx_pods_status ON pods(status);
CREATE INDEX idx_pods_node_name ON pods(node_name);
CREATE INDEX idx_pods_created_at ON pods(created_at);
`

const createPodsTableDown = `
DROP INDEX IF EXISTS idx_pods_created_at;
DROP INDEX IF EXISTS idx_pods_node_name;
DROP INDEX IF EXISTS idx_pods_status;
DROP INDEX IF EXISTS idx_pods_name;
DROP INDEX IF EXISTS idx_pods_namespace;
DROP INDEX IF EXISTS idx_pods_cluster_id;
DROP TABLE IF EXISTS pods;
`

// Migration 4: Create namespaces table
const createNamespacesTableUp = `
CREATE TABLE namespaces (
    id VARCHAR(36) PRIMARY KEY,
    cluster_id VARCHAR(36) NOT NULL,
    name VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL,
    phase VARCHAR(20),
    labels TEXT,
    annotations TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (cluster_id) REFERENCES clusters(id) ON DELETE CASCADE,
    UNIQUE(cluster_id, name)
);

CREATE INDEX idx_namespaces_cluster_id ON namespaces(cluster_id);
CREATE INDEX idx_namespaces_name ON namespaces(name);
CREATE INDEX idx_namespaces_status ON namespaces(status);
CREATE INDEX idx_namespaces_created_at ON namespaces(created_at);
`

const createNamespacesTableDown = `
DROP INDEX IF EXISTS idx_namespaces_created_at;
DROP INDEX IF EXISTS idx_namespaces_status;
DROP INDEX IF EXISTS idx_namespaces_name;
DROP INDEX IF EXISTS idx_namespaces_cluster_id;
DROP TABLE IF EXISTS namespaces;
`

// Migration 5: Create deployments table
const createDeploymentsTableUp = `
CREATE TABLE deployments (
    id VARCHAR(36) PRIMARY KEY,
    cluster_id VARCHAR(36) NOT NULL,
    namespace VARCHAR(100) NOT NULL,
    name VARCHAR(100) NOT NULL,
    replicas INTEGER DEFAULT 0,
    ready_replicas INTEGER DEFAULT 0,
    available_replicas INTEGER DEFAULT 0,
    unavailable_replicas INTEGER DEFAULT 0,
    updated_replicas INTEGER DEFAULT 0,
    strategy_type VARCHAR(50),
    labels TEXT,
    annotations TEXT,
    selector TEXT,
    template TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (cluster_id) REFERENCES clusters(id) ON DELETE CASCADE
);

CREATE INDEX idx_deployments_cluster_id ON deployments(cluster_id);
CREATE INDEX idx_deployments_namespace ON deployments(namespace);
CREATE INDEX idx_deployments_name ON deployments(name);
CREATE INDEX idx_deployments_created_at ON deployments(created_at);
`

const createDeploymentsTableDown = `
DROP INDEX IF EXISTS idx_deployments_created_at;
DROP INDEX IF EXISTS idx_deployments_name;
DROP INDEX IF EXISTS idx_deployments_namespace;
DROP INDEX IF EXISTS idx_deployments_cluster_id;
DROP TABLE IF EXISTS deployments;
`

// Migration 6: Create services table
const createServicesTableUp = `
CREATE TABLE services (
    id VARCHAR(36) PRIMARY KEY,
    cluster_id VARCHAR(36) NOT NULL,
    namespace VARCHAR(100) NOT NULL,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(50) NOT NULL,
    cluster_ip VARCHAR(45),
    external_ips TEXT,
    ports TEXT,
    selector TEXT,
    session_affinity VARCHAR(50),
    labels TEXT,
    annotations TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (cluster_id) REFERENCES clusters(id) ON DELETE CASCADE
);

CREATE INDEX idx_services_cluster_id ON services(cluster_id);
CREATE INDEX idx_services_namespace ON services(namespace);
CREATE INDEX idx_services_name ON services(name);
CREATE INDEX idx_services_type ON services(type);
CREATE INDEX idx_services_created_at ON services(created_at);
`

const createServicesTableDown = `
DROP INDEX IF EXISTS idx_services_created_at;
DROP INDEX IF EXISTS idx_services_type;
DROP INDEX IF EXISTS idx_services_name;
DROP INDEX IF EXISTS idx_services_namespace;
DROP INDEX IF EXISTS idx_services_cluster_id;
DROP TABLE IF EXISTS services;
`

// Migration 7: Create user role system
const createUserRoleSystemUp = `
-- ========================================
-- 1. 创建用户表
-- ========================================

-- 创建用户表
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    display_name VARCHAR(100),
    avatar_url VARCHAR(255),
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'suspended')),
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES users(id)
);

-- 创建用户表索引
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);

-- 创建更新时间触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 创建用户表更新时间触发器
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ========================================
-- 2. 创建角色权限表
-- ========================================

-- 创建角色表
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) UNIQUE NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    type VARCHAR(20) DEFAULT 'custom' CHECK (type IN ('system', 'custom')),
    is_default BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES users(id)
);

-- 创建权限表
CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    resource_type VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    scope VARCHAR(20) DEFAULT 'global' CHECK (scope IN ('global', 'cluster', 'namespace'))
);

-- 创建用户角色关联表
CREATE TABLE IF NOT EXISTS user_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    scope_type VARCHAR(20) DEFAULT 'global' CHECK (scope_type IN ('global', 'cluster', 'namespace')),
    scope_value VARCHAR(100),
    granted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    granted_by UUID REFERENCES users(id),
    expires_at TIMESTAMP WITH TIME ZONE,
    
    -- 唯一约束
    UNIQUE(user_id, role_id, scope_type, scope_value)
);

-- 创建角色权限关联表
CREATE TABLE IF NOT EXISTS role_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- 唯一约束
    UNIQUE(role_id, permission_id)
);

-- 创建角色权限相关索引
CREATE INDEX IF NOT EXISTS idx_roles_name ON roles(name);
CREATE INDEX IF NOT EXISTS idx_roles_type ON roles(type);
CREATE INDEX IF NOT EXISTS idx_roles_is_default ON roles(is_default);

CREATE INDEX IF NOT EXISTS idx_permissions_resource_action ON permissions(resource_type, action);
CREATE INDEX IF NOT EXISTS idx_permissions_scope ON permissions(scope);

CREATE INDEX IF NOT EXISTS idx_user_roles_user ON user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_role ON user_roles(role_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_scope ON user_roles(scope_type, scope_value);
CREATE INDEX IF NOT EXISTS idx_user_roles_expires ON user_roles(expires_at);

CREATE INDEX IF NOT EXISTS idx_role_permissions_role ON role_permissions(role_id);
CREATE INDEX IF NOT EXISTS idx_role_permissions_permission ON role_permissions(permission_id);

-- 创建角色表更新时间触发器
CREATE TRIGGER update_roles_updated_at BEFORE UPDATE ON roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ========================================
-- 3. 创建认证相关表
-- ========================================

-- 创建用户会话表
CREATE TABLE IF NOT EXISTS user_sessions (
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

-- 创建审计日志表
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50),
    resource_id VARCHAR(100),
    cluster_name VARCHAR(100),
    namespace VARCHAR(100),
    details JSONB,
    ip_address INET,
    user_agent TEXT,
    status VARCHAR(20) CHECK (status IN ('success', 'failed', 'denied')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 创建认证相关表索引
CREATE INDEX IF NOT EXISTS idx_user_sessions_user ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_token ON user_sessions(token_hash);
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires ON user_sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_user_sessions_last_used ON user_sessions(last_used_at);

CREATE INDEX IF NOT EXISTS idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_cluster ON audit_logs(cluster_name);
CREATE INDEX IF NOT EXISTS idx_audit_logs_namespace ON audit_logs(namespace);
CREATE INDEX IF NOT EXISTS idx_audit_logs_status ON audit_logs(status);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created ON audit_logs(created_at);

-- 创建清理过期会话的函数
CREATE OR REPLACE FUNCTION cleanup_expired_sessions()
RETURNS void AS $$
BEGIN
    DELETE FROM user_sessions WHERE expires_at < NOW();
END;
$$ LANGUAGE plpgsql;

-- ========================================
-- 4. 插入初始数据
-- ========================================

-- 插入系统权限
INSERT INTO permissions (id, name, display_name, description, resource_type, action, scope) VALUES
-- 集群权限
('11111111-1111-1111-1111-111111111111', 'cluster:create', '创建集群', '创建新的 Kubernetes 集群', 'cluster', 'create', 'global'),
('11111111-1111-1111-1111-111111111112', 'cluster:read', '查看集群', '查看集群信息和状态', 'cluster', 'read', 'global'),
('11111111-1111-1111-1111-111111111113', 'cluster:update', '更新集群', '更新集群配置', 'cluster', 'update', 'global'),
('11111111-1111-1111-1111-111111111114', 'cluster:delete', '删除集群', '删除集群', 'cluster', 'delete', 'global'),

-- 部署权限
('22222222-2222-2222-2222-222222222221', 'deployment:create', '创建部署', '创建新的部署', 'deployment', 'create', 'namespace'),
('22222222-2222-2222-2222-222222222222', 'deployment:read', '查看部署', '查看部署信息', 'deployment', 'read', 'namespace'),
('22222222-2222-2222-2222-222222222223', 'deployment:update', '更新部署', '更新部署配置', 'deployment', 'update', 'namespace'),
('22222222-2222-2222-2222-222222222224', 'deployment:delete', '删除部署', '删除部署', 'deployment', 'delete', 'namespace'),
('22222222-2222-2222-2222-222222222225', 'deployment:scale', '扩缩容部署', '调整部署副本数', 'deployment', 'scale', 'namespace'),
('22222222-2222-2222-2222-222222222226', 'deployment:restart', '重启部署', '重启部署', 'deployment', 'restart', 'namespace'),

-- 服务权限
('33333333-3333-3333-3333-333333333331', 'service:create', '创建服务', '创建新的服务', 'service', 'create', 'namespace'),
('33333333-3333-3333-3333-333333333332', 'service:read', '查看服务', '查看服务信息', 'service', 'read', 'namespace'),
('33333333-3333-3333-3333-333333333333', 'service:update', '更新服务', '更新服务配置', 'service', 'update', 'namespace'),
('33333333-3333-3333-3333-333333333334', 'service:delete', '删除服务', '删除服务', 'service', 'delete', 'namespace'),

-- Pod 权限
('44444444-4444-4444-4444-444444444441', 'pod:read', '查看 Pod', '查看 Pod 信息', 'pod', 'read', 'namespace'),
('44444444-4444-4444-4444-444444444442', 'pod:delete', '删除 Pod', '删除 Pod', 'pod', 'delete', 'namespace'),
('44444444-4444-4444-4444-444444444443', 'pod:logs', '查看 Pod 日志', '查看 Pod 日志', 'pod', 'logs', 'namespace'),
('44444444-4444-4444-4444-444444444444', 'pod:exec', '执行 Pod 命令', '在 Pod 中执行命令', 'pod', 'exec', 'namespace'),

-- 节点权限
('55555555-5555-5555-5555-555555555551', 'node:read', '查看节点', '查看节点信息', 'node', 'read', 'cluster'),
('55555555-5555-5555-5555-555555555552', 'node:update', '更新节点', '更新节点配置', 'node', 'update', 'cluster'),
('55555555-5555-5555-5555-555555555553', 'node:drain', '驱逐节点', '驱逐节点上的 Pod', 'node', 'drain', 'cluster'),
('55555555-5555-5555-5555-555555555554', 'node:cordon', '封锁节点', '封锁/解封节点', 'node', 'cordon', 'cluster'),

-- 命名空间权限
('66666666-6666-6666-6666-666666666661', 'namespace:read', '查看命名空间', '查看命名空间信息', 'namespace', 'read', 'cluster'),
('66666666-6666-6666-6666-666666666662', 'namespace:create', '创建命名空间', '创建新的命名空间', 'namespace', 'create', 'cluster'),
('66666666-6666-6666-6666-666666666663', 'namespace:delete', '删除命名空间', '删除命名空间', 'namespace', 'delete', 'cluster'),

-- 用户管理权限
('77777777-7777-7777-7777-777777777771', 'user:create', '创建用户', '创建新用户', 'user', 'create', 'global'),
('77777777-7777-7777-7777-777777777772', 'user:read', '查看用户', '查看用户信息', 'user', 'read', 'global'),
('77777777-7777-7777-7777-777777777773', 'user:update', '更新用户', '更新用户信息', 'user', 'update', 'global'),
('77777777-7777-7777-7777-777777777774', 'user:delete', '删除用户', '删除用户', 'user', 'delete', 'global'),

-- 角色管理权限
('88888888-8888-8888-8888-888888888881', 'role:create', '创建角色', '创建新角色', 'role', 'create', 'global'),
('88888888-8888-8888-8888-888888888882', 'role:read', '查看角色', '查看角色信息', 'role', 'read', 'global'),
('88888888-8888-8888-8888-888888888883', 'role:update', '更新角色', '更新角色信息', 'role', 'update', 'global'),
('88888888-8888-8888-8888-888888888884', 'role:delete', '删除角色', '删除角色', 'role', 'delete', 'global'),

-- 审计日志权限
('99999999-9999-9999-9999-999999999991', 'audit:read', '查看审计日志', '查看系统审计日志', 'audit', 'read', 'global');

-- 插入系统角色
INSERT INTO roles (id, name, display_name, description, type, is_default) VALUES
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'super_admin', '超级管理员', '拥有所有权限的系统管理员', 'system', false),
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'cluster_admin', '集群管理员', '管理特定集群的所有资源', 'system', false),
('cccccccc-cccc-cccc-cccc-cccccccccccc', 'namespace_admin', '命名空间管理员', '管理特定命名空间的资源', 'system', false),
('dddddddd-dddd-dddd-dddd-dddddddddddd', 'developer', '开发者', '开发人员角色，可以部署和管理应用', 'system', false),
('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'viewer', '查看者', '只读权限，可以查看资源信息', 'system', true),
('ffffffff-ffff-ffff-ffff-ffffffffffff', 'operator', '运维人员', '运维人员角色，可以管理基础设施', 'system', false);

-- 为超级管理员角色分配所有权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', id FROM permissions;

-- 为集群管理员角色分配权限
INSERT INTO role_permissions (role_id, permission_id) VALUES
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '11111111-1111-1111-1111-111111111112'), -- cluster:read
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '11111111-1111-1111-1111-111111111113'), -- cluster:update
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '22222222-2222-2222-2222-222222222221'), -- deployment:create
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '22222222-2222-2222-2222-222222222222'), -- deployment:read
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '22222222-2222-2222-2222-222222222223'), -- deployment:update
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '22222222-2222-2222-2222-222222222224'), -- deployment:delete
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '22222222-2222-2222-2222-222222222225'), -- deployment:scale
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '22222222-2222-2222-2222-222222222226'), -- deployment:restart
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '33333333-3333-3333-3333-333333333331'), -- service:create
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '33333333-3333-3333-3333-333333333332'), -- service:read
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '33333333-3333-3333-3333-333333333333'), -- service:update
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '33333333-3333-3333-3333-333333333334'), -- service:delete
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '44444444-4444-4444-4444-444444444441'), -- pod:read
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '44444444-4444-4444-4444-444444444442'), -- pod:delete
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '44444444-4444-4444-4444-444444444443'), -- pod:logs
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '44444444-4444-4444-4444-444444444444'), -- pod:exec
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '55555555-5555-5555-5555-555555555551'), -- node:read
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '55555555-5555-5555-5555-555555555552'), -- node:update
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '55555555-5555-5555-5555-555555555553'), -- node:drain
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '55555555-5555-5555-5555-555555555554'), -- node:cordon
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '66666666-6666-6666-6666-666666666661'), -- namespace:read
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '66666666-6666-6666-6666-666666666662'), -- namespace:create
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '66666666-6666-6666-6666-666666666663'); -- namespace:delete

-- 为命名空间管理员角色分配权限
INSERT INTO role_permissions (role_id, permission_id) VALUES
('cccccccc-cccc-cccc-cccc-cccccccccccc', '22222222-2222-2222-2222-222222222221'), -- deployment:create
('cccccccc-cccc-cccc-cccc-cccccccccccc', '22222222-2222-2222-2222-222222222222'), -- deployment:read
('cccccccc-cccc-cccc-cccc-cccccccccccc', '22222222-2222-2222-2222-222222222223'), -- deployment:update
('cccccccc-cccc-cccc-cccc-cccccccccccc', '22222222-2222-2222-2222-222222222224'), -- deployment:delete
('cccccccc-cccc-cccc-cccc-cccccccccccc', '22222222-2222-2222-2222-222222222225'), -- deployment:scale
('cccccccc-cccc-cccc-cccc-cccccccccccc', '22222222-2222-2222-2222-222222222226'), -- deployment:restart
('cccccccc-cccc-cccc-cccc-cccccccccccc', '33333333-3333-3333-3333-333333333331'), -- service:create
('cccccccc-cccc-cccc-cccc-cccccccccccc', '33333333-3333-3333-3333-333333333332'), -- service:read
('cccccccc-cccc-cccc-cccc-cccccccccccc', '33333333-3333-3333-3333-333333333333'), -- service:update
('cccccccc-cccc-cccc-cccc-cccccccccccc', '33333333-3333-3333-3333-333333333334'), -- service:delete
('cccccccc-cccc-cccc-cccc-cccccccccccc', '44444444-4444-4444-4444-444444444441'), -- pod:read
('cccccccc-cccc-cccc-cccc-cccccccccccc', '44444444-4444-4444-4444-444444444442'), -- pod:delete
('cccccccc-cccc-cccc-cccc-cccccccccccc', '44444444-4444-4444-4444-444444444443'), -- pod:logs
('cccccccc-cccc-cccc-cccc-cccccccccccc', '44444444-4444-4444-4444-444444444444'); -- pod:exec

-- 为开发者角色分配权限
INSERT INTO role_permissions (role_id, permission_id) VALUES
('dddddddd-dddd-dddd-dddd-dddddddddddd', '22222222-2222-2222-2222-222222222221'), -- deployment:create
('dddddddd-dddd-dddd-dddd-dddddddddddd', '22222222-2222-2222-2222-222222222222'), -- deployment:read
('dddddddd-dddd-dddd-dddd-dddddddddddd', '22222222-2222-2222-2222-222222222223'), -- deployment:update
('dddddddd-dddd-dddd-dddd-dddddddddddd', '22222222-2222-2222-2222-222222222224'), -- deployment:delete
('dddddddd-dddd-dddd-dddd-dddddddddddd', '22222222-2222-2222-2222-222222222225'), -- deployment:scale
('dddddddd-dddd-dddd-dddd-dddddddddddd', '33333333-3333-3333-3333-333333333331'), -- service:create
('dddddddd-dddd-dddd-dddd-dddddddddddd', '33333333-3333-3333-3333-333333333332'), -- service:read
('dddddddd-dddd-dddd-dddd-dddddddddddd', '33333333-3333-3333-3333-333333333333'), -- service:update
('dddddddd-dddd-dddd-dddd-dddddddddddd', '33333333-3333-3333-3333-333333333334'), -- service:delete
('dddddddd-dddd-dddd-dddd-dddddddddddd', '44444444-4444-4444-4444-444444444441'), -- pod:read
('dddddddd-dddd-dddd-dddd-dddddddddddd', '44444444-4444-4444-4444-444444444442'), -- pod:delete
('dddddddd-dddd-dddd-dddd-dddddddddddd', '44444444-4444-4444-4444-444444444443'); -- pod:logs

-- 为查看者角色分配权限
INSERT INTO role_permissions (role_id, permission_id) VALUES
('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', '11111111-1111-1111-1111-111111111112'), -- cluster:read
('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', '22222222-2222-2222-2222-222222222222'), -- deployment:read
('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', '33333333-3333-3333-3333-333333333332'), -- service:read
('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', '44444444-4444-4444-4444-444444444441'), -- pod:read
('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', '44444444-4444-4444-4444-444444444443'), -- pod:logs
('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', '55555555-5555-5555-5555-555555555551'), -- node:read
('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', '66666666-6666-6666-6666-666666666661'); -- namespace:read

-- 为运维人员角色分配权限
INSERT INTO role_permissions (role_id, permission_id) VALUES
('ffffffff-ffff-ffff-ffff-ffffffffffff', '11111111-1111-1111-1111-111111111112'), -- cluster:read
('ffffffff-ffff-ffff-ffff-ffffffffffff', '11111111-1111-1111-1111-111111111113'), -- cluster:update
('ffffffff-ffff-ffff-ffff-ffffffffffff', '22222222-2222-2222-2222-222222222222'), -- deployment:read
('ffffffff-ffff-ffff-ffff-ffffffffffff', '22222222-2222-2222-2222-222222222223'), -- deployment:update
('ffffffff-ffff-ffff-ffff-ffffffffffff', '22222222-2222-2222-2222-222222222225'), -- deployment:scale
('ffffffff-ffff-ffff-ffff-ffffffffffff', '22222222-2222-2222-2222-222222222226'), -- deployment:restart
('ffffffff-ffff-ffff-ffff-ffffffffffff', '33333333-3333-3333-3333-333333333332'), -- service:read
('ffffffff-ffff-ffff-ffff-ffffffffffff', '33333333-3333-3333-3333-333333333333'), -- service:update
('ffffffff-ffff-ffff-ffff-ffffffffffff', '44444444-4444-4444-4444-444444444441'), -- pod:read
('ffffffff-ffff-ffff-ffff-ffffffffffff', '44444444-4444-4444-4444-444444444442'), -- pod:delete
('ffffffff-ffff-ffff-ffff-ffffffffffff', '44444444-4444-4444-4444-444444444443'), -- pod:logs
('ffffffff-ffff-ffff-ffff-ffffffffffff', '44444444-4444-4444-4444-444444444444'), -- pod:exec
('ffffffff-ffff-ffff-ffff-ffffffffffff', '55555555-5555-5555-5555-555555555551'), -- node:read
('ffffffff-ffff-ffff-ffff-ffffffffffff', '55555555-5555-5555-5555-555555555552'), -- node:update
('ffffffff-ffff-ffff-ffff-ffffffffffff', '55555555-5555-5555-5555-555555555553'), -- node:drain
('ffffffff-ffff-ffff-ffff-ffffffffffff', '55555555-5555-5555-5555-555555555554'), -- node:cordon
('ffffffff-ffff-ffff-ffff-ffffffffffff', '66666666-6666-6666-6666-666666666661'); -- namespace:read

-- 创建默认管理员用户 (密码: admin123)
INSERT INTO users (id, username, email, password_hash, display_name, status) VALUES
('00000000-0000-0000-0000-000000000000', 'admin', 'admin@kube-tide.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', '系统管理员', 'active');

-- 为默认管理员分配超级管理员角色
INSERT INTO user_roles (user_id, role_id, scope_type, granted_by) VALUES
('00000000-0000-0000-0000-000000000000', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'global', '00000000-0000-0000-0000-000000000000');
`

const createUserRoleSystemDown = `
-- 删除用户角色关联
DELETE FROM user_roles;

-- 删除默认管理员用户
DELETE FROM users WHERE username = 'admin';

-- 删除角色权限关联
DELETE FROM role_permissions;

-- 删除系统角色
DELETE FROM roles WHERE type = 'system';

-- 删除系统权限
DELETE FROM permissions;

-- 删除清理过期会话函数
DROP FUNCTION IF EXISTS cleanup_expired_sessions();

-- 删除审计日志表索引
DROP INDEX IF EXISTS idx_audit_logs_created;
DROP INDEX IF EXISTS idx_audit_logs_status;
DROP INDEX IF EXISTS idx_audit_logs_namespace;
DROP INDEX IF EXISTS idx_audit_logs_cluster;
DROP INDEX IF EXISTS idx_audit_logs_resource;
DROP INDEX IF EXISTS idx_audit_logs_action;
DROP INDEX IF EXISTS idx_audit_logs_user;

-- 删除用户会话表索引
DROP INDEX IF EXISTS idx_user_sessions_last_used;
DROP INDEX IF EXISTS idx_user_sessions_expires;
DROP INDEX IF EXISTS idx_user_sessions_token;
DROP INDEX IF EXISTS idx_user_sessions_user;

-- 删除认证相关表
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS user_sessions;

-- 删除角色表触发器
DROP TRIGGER IF EXISTS update_roles_updated_at ON roles;

-- 删除角色权限相关索引
DROP INDEX IF EXISTS idx_role_permissions_permission;
DROP INDEX IF EXISTS idx_role_permissions_role;
DROP INDEX IF EXISTS idx_user_roles_expires;
DROP INDEX IF EXISTS idx_user_roles_scope;
DROP INDEX IF EXISTS idx_user_roles_role;
DROP INDEX IF EXISTS idx_user_roles_user;
DROP INDEX IF EXISTS idx_permissions_scope;
DROP INDEX IF EXISTS idx_permissions_resource_action;
DROP INDEX IF EXISTS idx_roles_is_default;
DROP INDEX IF EXISTS idx_roles_type;
DROP INDEX IF EXISTS idx_roles_name;

-- 删除角色权限关联表
DROP TABLE IF EXISTS role_permissions;

-- 删除用户角色关联表
DROP TABLE IF EXISTS user_roles;

-- 删除权限表
DROP TABLE IF EXISTS permissions;

-- 删除角色表
DROP TABLE IF EXISTS roles;

-- 删除用户表触发器
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- 删除更新时间触发器函数
DROP FUNCTION IF EXISTS update_updated_at_column();

-- 删除用户表索引
DROP INDEX IF EXISTS idx_users_created_at;
DROP INDEX IF EXISTS idx_users_status;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_username;

-- 删除用户表
DROP TABLE IF EXISTS users;
`
