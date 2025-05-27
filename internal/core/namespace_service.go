package core

import (
	"context"

	"go.uber.org/zap"

	"kube-tide/internal/database/models"
	"kube-tide/internal/repository"
)

// namespaceService implements NamespaceService interface
type namespaceService struct {
	repo   repository.NamespaceRepository
	logger *zap.Logger
}

// Create creates a new namespace
func (s *namespaceService) Create(ctx context.Context, namespace *models.Namespace) error {
	if err := namespace.Validate(); err != nil {
		s.logger.Error("namespace validation failed", zap.Error(err))
		return err
	}

	return s.repo.Create(ctx, namespace)
}

// GetByID retrieves a namespace by ID
func (s *namespaceService) GetByID(ctx context.Context, id string) (*models.Namespace, error) {
	return s.repo.GetByID(ctx, id)
}

// GetByClusterAndName retrieves a namespace by cluster ID and name
func (s *namespaceService) GetByClusterAndName(ctx context.Context, clusterID, name string) (*models.Namespace, error) {
	return s.repo.GetByClusterAndName(ctx, clusterID, name)
}

// ListByCluster retrieves namespaces by cluster ID with pagination
func (s *namespaceService) ListByCluster(ctx context.Context, clusterID string, params models.PaginationParams) (*models.PaginatedResult, error) {
	return s.repo.ListByCluster(ctx, clusterID, params)
}

// Update updates a namespace
func (s *namespaceService) Update(ctx context.Context, id string, updates models.NamespaceUpdateRequest) error {
	return s.repo.Update(ctx, id, updates)
}

// Delete deletes a namespace
func (s *namespaceService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// DeleteByCluster deletes all namespaces in a cluster
func (s *namespaceService) DeleteByCluster(ctx context.Context, clusterID string) error {
	return s.repo.DeleteByCluster(ctx, clusterID)
}

// Count counts namespaces in a cluster
func (s *namespaceService) Count(ctx context.Context, clusterID string) (int, error) {
	return s.repo.Count(ctx, clusterID)
}
