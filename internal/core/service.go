package core

import (
	"context"

	"go.uber.org/zap"

	"kube-tide/internal/database"
	"kube-tide/internal/database/models"
	"kube-tide/internal/repository"
)

// Service represents the core service layer
type Service struct {
	repos  *repository.Repositories
	db     *database.Database
	logger *zap.Logger
}

// NewService creates a new core service
func NewService(repos *repository.Repositories, db *database.Database, logger *zap.Logger) *Service {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Service{
		repos:  repos,
		db:     db,
		logger: logger,
	}
}

// ClusterService provides cluster-related operations
func (s *Service) ClusterService() ClusterService {
	return &clusterService{
		repo:   s.repos.Cluster,
		logger: s.logger,
	}
}

// NodeService provides node-related operations
func (s *Service) NodeService() NodeService {
	return &nodeService{
		repo:   s.repos.Node,
		logger: s.logger,
	}
}

// PodService provides pod-related operations
func (s *Service) PodService() PodService {
	return &podService{
		repo:   s.repos.Pod,
		logger: s.logger,
	}
}

// NamespaceService provides namespace-related operations
func (s *Service) NamespaceService() NamespaceService {
	return &namespaceService{
		repo:   s.repos.Namespace,
		logger: s.logger,
	}
}

// DeploymentService provides deployment-related operations
func (s *Service) DeploymentService() DeploymentService {
	return &deploymentService{
		repo:   s.repos.Deployment,
		logger: s.logger,
	}
}

// ServiceService provides service-related operations
func (s *Service) ServiceService() ServiceService {
	return &serviceService{
		repo:   s.repos.Service,
		logger: s.logger,
	}
}

// Service interfaces
type ClusterService interface {
	Create(ctx context.Context, cluster *models.Cluster) error
	GetByID(ctx context.Context, id string) (*models.Cluster, error)
	GetByName(ctx context.Context, name string) (*models.Cluster, error)
	List(ctx context.Context, filters models.ClusterFilters, params models.PaginationParams) (*models.PaginatedResult, error)
	Update(ctx context.Context, id string, updates models.ClusterUpdateRequest) error
	Delete(ctx context.Context, id string) error
}

type NodeService interface {
	Create(ctx context.Context, node *models.Node) error
	GetByID(ctx context.Context, id string) (*models.Node, error)
	GetByClusterAndName(ctx context.Context, clusterID, name string) (*models.Node, error)
	ListByCluster(ctx context.Context, clusterID string, params models.PaginationParams) (*models.PaginatedResult, error)
	Update(ctx context.Context, id string, updates models.NodeUpdateRequest) error
	Delete(ctx context.Context, id string) error
	DeleteByCluster(ctx context.Context, clusterID string) error
	Count(ctx context.Context, clusterID string) (int, error)
}

type PodService interface {
	Create(ctx context.Context, pod *models.Pod) error
	GetByID(ctx context.Context, id string) (*models.Pod, error)
	GetByClusterNamespaceAndName(ctx context.Context, clusterID, namespace, name string) (*models.Pod, error)
	ListByCluster(ctx context.Context, clusterID string, params models.PaginationParams) (*models.PaginatedResult, error)
	ListByNamespace(ctx context.Context, clusterID, namespace string, params models.PaginationParams) (*models.PaginatedResult, error)
	Update(ctx context.Context, id string, updates models.PodUpdateRequest) error
	Delete(ctx context.Context, id string) error
	DeleteByCluster(ctx context.Context, clusterID string) error
	DeleteByNamespace(ctx context.Context, clusterID, namespace string) error
	Count(ctx context.Context, clusterID string) (int, error)
	CountByNamespace(ctx context.Context, clusterID, namespace string) (int, error)
}

type NamespaceService interface {
	Create(ctx context.Context, namespace *models.Namespace) error
	GetByID(ctx context.Context, id string) (*models.Namespace, error)
	GetByClusterAndName(ctx context.Context, clusterID, name string) (*models.Namespace, error)
	ListByCluster(ctx context.Context, clusterID string, params models.PaginationParams) (*models.PaginatedResult, error)
	Update(ctx context.Context, id string, updates models.NamespaceUpdateRequest) error
	Delete(ctx context.Context, id string) error
	DeleteByCluster(ctx context.Context, clusterID string) error
	Count(ctx context.Context, clusterID string) (int, error)
}

type DeploymentService interface {
	Create(ctx context.Context, deployment *models.Deployment) error
	GetByID(ctx context.Context, id string) (*models.Deployment, error)
	GetByClusterNamespaceAndName(ctx context.Context, clusterID, namespace, name string) (*models.Deployment, error)
	ListByCluster(ctx context.Context, clusterID string, params models.PaginationParams) (*models.PaginatedResult, error)
	ListByNamespace(ctx context.Context, clusterID, namespace string, params models.PaginationParams) (*models.PaginatedResult, error)
	Update(ctx context.Context, id string, updates models.DeploymentUpdateRequest) error
	Delete(ctx context.Context, id string) error
	DeleteByCluster(ctx context.Context, clusterID string) error
	DeleteByNamespace(ctx context.Context, clusterID, namespace string) error
	Count(ctx context.Context, clusterID string) (int, error)
	CountByNamespace(ctx context.Context, clusterID, namespace string) (int, error)
}

type ServiceService interface {
	Create(ctx context.Context, service *models.Service) error
	GetByID(ctx context.Context, id string) (*models.Service, error)
	GetByClusterNamespaceAndName(ctx context.Context, clusterID, namespace, name string) (*models.Service, error)
	ListByCluster(ctx context.Context, clusterID string, params models.PaginationParams) (*models.PaginatedResult, error)
	ListByNamespace(ctx context.Context, clusterID, namespace string, params models.PaginationParams) (*models.PaginatedResult, error)
	Update(ctx context.Context, id string, updates models.ServiceUpdateRequest) error
	Delete(ctx context.Context, id string) error
	DeleteByCluster(ctx context.Context, clusterID string) error
	DeleteByNamespace(ctx context.Context, clusterID, namespace string) error
	Count(ctx context.Context, clusterID string) (int, error)
	CountByNamespace(ctx context.Context, clusterID, namespace string) (int, error)
}
