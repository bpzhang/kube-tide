package repository

import (
	"context"
	"database/sql"

	"kube-tide/internal/database/models"
)

// ClusterRepository defines the interface for cluster data operations
type ClusterRepository interface {
	Create(ctx context.Context, cluster *models.Cluster) error
	CreateTx(ctx context.Context, tx *sql.Tx, cluster *models.Cluster) error
	GetByID(ctx context.Context, id string) (*models.Cluster, error)
	GetByName(ctx context.Context, name string) (*models.Cluster, error)
	List(ctx context.Context, filters models.ClusterFilters, params models.PaginationParams) (*models.PaginatedResult, error)
	Update(ctx context.Context, id string, updates models.ClusterUpdateRequest) error
	UpdateTx(ctx context.Context, tx *sql.Tx, id string, updates models.ClusterUpdateRequest) error
	Delete(ctx context.Context, id string) error
	DeleteTx(ctx context.Context, tx *sql.Tx, id string) error
	Count(ctx context.Context, filters models.ClusterFilters) (int, error)
	Exists(ctx context.Context, id string) (bool, error)
	ExistsByName(ctx context.Context, name string) (bool, error)
}

// NodeRepository defines the interface for node data operations
type NodeRepository interface {
	Create(ctx context.Context, node *models.Node) error
	CreateTx(ctx context.Context, tx *sql.Tx, node *models.Node) error
	GetByID(ctx context.Context, id string) (*models.Node, error)
	GetByClusterAndName(ctx context.Context, clusterID, name string) (*models.Node, error)
	ListByCluster(ctx context.Context, clusterID string, params models.PaginationParams) (*models.PaginatedResult, error)
	Update(ctx context.Context, id string, updates models.NodeUpdateRequest) error
	UpdateTx(ctx context.Context, tx *sql.Tx, id string, updates models.NodeUpdateRequest) error
	Delete(ctx context.Context, id string) error
	DeleteTx(ctx context.Context, tx *sql.Tx, id string) error
	DeleteByCluster(ctx context.Context, clusterID string) error
	Count(ctx context.Context, clusterID string) (int, error)
}

// PodRepository defines the interface for pod data operations
type PodRepository interface {
	Create(ctx context.Context, pod *models.Pod) error
	CreateTx(ctx context.Context, tx *sql.Tx, pod *models.Pod) error
	GetByID(ctx context.Context, id string) (*models.Pod, error)
	GetByClusterNamespaceAndName(ctx context.Context, clusterID, namespace, name string) (*models.Pod, error)
	ListByCluster(ctx context.Context, clusterID string, params models.PaginationParams) (*models.PaginatedResult, error)
	ListByNamespace(ctx context.Context, clusterID, namespace string, params models.PaginationParams) (*models.PaginatedResult, error)
	Update(ctx context.Context, id string, updates models.PodUpdateRequest) error
	UpdateTx(ctx context.Context, tx *sql.Tx, id string, updates models.PodUpdateRequest) error
	Delete(ctx context.Context, id string) error
	DeleteTx(ctx context.Context, tx *sql.Tx, id string) error
	DeleteByCluster(ctx context.Context, clusterID string) error
	DeleteByNamespace(ctx context.Context, clusterID, namespace string) error
	Count(ctx context.Context, clusterID string) (int, error)
	CountByNamespace(ctx context.Context, clusterID, namespace string) (int, error)
}

// NamespaceRepository defines the interface for namespace data operations
type NamespaceRepository interface {
	Create(ctx context.Context, namespace *models.Namespace) error
	CreateTx(ctx context.Context, tx *sql.Tx, namespace *models.Namespace) error
	GetByID(ctx context.Context, id string) (*models.Namespace, error)
	GetByClusterAndName(ctx context.Context, clusterID, name string) (*models.Namespace, error)
	ListByCluster(ctx context.Context, clusterID string, params models.PaginationParams) (*models.PaginatedResult, error)
	Update(ctx context.Context, id string, updates models.NamespaceUpdateRequest) error
	UpdateTx(ctx context.Context, tx *sql.Tx, id string, updates models.NamespaceUpdateRequest) error
	Delete(ctx context.Context, id string) error
	DeleteTx(ctx context.Context, tx *sql.Tx, id string) error
	DeleteByCluster(ctx context.Context, clusterID string) error
	Count(ctx context.Context, clusterID string) (int, error)
}

// DeploymentRepository defines the interface for deployment data operations
type DeploymentRepository interface {
	Create(ctx context.Context, deployment *models.Deployment) error
	CreateTx(ctx context.Context, tx *sql.Tx, deployment *models.Deployment) error
	GetByID(ctx context.Context, id string) (*models.Deployment, error)
	GetByClusterNamespaceAndName(ctx context.Context, clusterID, namespace, name string) (*models.Deployment, error)
	ListByCluster(ctx context.Context, clusterID string, params models.PaginationParams) (*models.PaginatedResult, error)
	ListByNamespace(ctx context.Context, clusterID, namespace string, params models.PaginationParams) (*models.PaginatedResult, error)
	Update(ctx context.Context, id string, updates models.DeploymentUpdateRequest) error
	UpdateTx(ctx context.Context, tx *sql.Tx, id string, updates models.DeploymentUpdateRequest) error
	Delete(ctx context.Context, id string) error
	DeleteTx(ctx context.Context, tx *sql.Tx, id string) error
	DeleteByCluster(ctx context.Context, clusterID string) error
	DeleteByNamespace(ctx context.Context, clusterID, namespace string) error
	Count(ctx context.Context, clusterID string) (int, error)
	CountByNamespace(ctx context.Context, clusterID, namespace string) (int, error)
}

// ServiceRepository defines the interface for service data operations
type ServiceRepository interface {
	Create(ctx context.Context, service *models.Service) error
	CreateTx(ctx context.Context, tx *sql.Tx, service *models.Service) error
	GetByID(ctx context.Context, id string) (*models.Service, error)
	GetByClusterNamespaceAndName(ctx context.Context, clusterID, namespace, name string) (*models.Service, error)
	ListByCluster(ctx context.Context, clusterID string, params models.PaginationParams) (*models.PaginatedResult, error)
	ListByNamespace(ctx context.Context, clusterID, namespace string, params models.PaginationParams) (*models.PaginatedResult, error)
	Update(ctx context.Context, id string, updates models.ServiceUpdateRequest) error
	UpdateTx(ctx context.Context, tx *sql.Tx, id string, updates models.ServiceUpdateRequest) error
	Delete(ctx context.Context, id string) error
	DeleteTx(ctx context.Context, tx *sql.Tx, id string) error
	DeleteByCluster(ctx context.Context, clusterID string) error
	DeleteByNamespace(ctx context.Context, clusterID, namespace string) error
	Count(ctx context.Context, clusterID string) (int, error)
	CountByNamespace(ctx context.Context, clusterID, namespace string) (int, error)
}
