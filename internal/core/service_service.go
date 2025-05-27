package core

import (
	"context"

	"go.uber.org/zap"

	"kube-tide/internal/database/models"
	"kube-tide/internal/repository"
)

// serviceService implements ServiceService interface
type serviceService struct {
	repo   repository.ServiceRepository
	logger *zap.Logger
}

// Create creates a new service
func (s *serviceService) Create(ctx context.Context, service *models.Service) error {
	if err := service.Validate(); err != nil {
		s.logger.Error("service validation failed", zap.Error(err))
		return err
	}

	return s.repo.Create(ctx, service)
}

// GetByID retrieves a service by ID
func (s *serviceService) GetByID(ctx context.Context, id string) (*models.Service, error) {
	return s.repo.GetByID(ctx, id)
}

// GetByClusterNamespaceAndName retrieves a service by cluster ID, namespace, and name
func (s *serviceService) GetByClusterNamespaceAndName(ctx context.Context, clusterID, namespace, name string) (*models.Service, error) {
	return s.repo.GetByClusterNamespaceAndName(ctx, clusterID, namespace, name)
}

// ListByCluster retrieves services by cluster ID with pagination
func (s *serviceService) ListByCluster(ctx context.Context, clusterID string, params models.PaginationParams) (*models.PaginatedResult, error) {
	return s.repo.ListByCluster(ctx, clusterID, params)
}

// ListByNamespace retrieves services by cluster ID and namespace with pagination
func (s *serviceService) ListByNamespace(ctx context.Context, clusterID, namespace string, params models.PaginationParams) (*models.PaginatedResult, error) {
	return s.repo.ListByNamespace(ctx, clusterID, namespace, params)
}

// Update updates a service
func (s *serviceService) Update(ctx context.Context, id string, updates models.ServiceUpdateRequest) error {
	return s.repo.Update(ctx, id, updates)
}

// Delete deletes a service
func (s *serviceService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// DeleteByCluster deletes all services in a cluster
func (s *serviceService) DeleteByCluster(ctx context.Context, clusterID string) error {
	return s.repo.DeleteByCluster(ctx, clusterID)
}

// DeleteByNamespace deletes all services in a namespace
func (s *serviceService) DeleteByNamespace(ctx context.Context, clusterID, namespace string) error {
	return s.repo.DeleteByNamespace(ctx, clusterID, namespace)
}

// Count counts services in a cluster
func (s *serviceService) Count(ctx context.Context, clusterID string) (int, error) {
	return s.repo.Count(ctx, clusterID)
}

// CountByNamespace counts services in a namespace
func (s *serviceService) CountByNamespace(ctx context.Context, clusterID, namespace string) (int, error) {
	return s.repo.CountByNamespace(ctx, clusterID, namespace)
}
