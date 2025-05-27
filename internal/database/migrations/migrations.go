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
