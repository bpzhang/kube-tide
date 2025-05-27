package core

import (
	"context"

	"go.uber.org/zap"

	"kube-tide/internal/database/models"
	"kube-tide/internal/repository"
)

// clusterService implements ClusterService interface
type clusterService struct {
	repo   repository.ClusterRepository
	logger *zap.Logger
}

// Create creates a new cluster
func (s *clusterService) Create(ctx context.Context, cluster *models.Cluster) error {
	if err := cluster.Validate(); err != nil {
		s.logger.Error("cluster validation failed", zap.Error(err))
		return err
	}

	return s.repo.Create(ctx, cluster)
}

// GetByID retrieves a cluster by ID
func (s *clusterService) GetByID(ctx context.Context, id string) (*models.Cluster, error) {
	return s.repo.GetByID(ctx, id)
}

// GetByName retrieves a cluster by name
func (s *clusterService) GetByName(ctx context.Context, name string) (*models.Cluster, error) {
	return s.repo.GetByName(ctx, name)
}

// List retrieves clusters with filters and pagination
func (s *clusterService) List(ctx context.Context, filters models.ClusterFilters, params models.PaginationParams) (*models.PaginatedResult, error) {
	return s.repo.List(ctx, filters, params)
}

// Update updates a cluster
func (s *clusterService) Update(ctx context.Context, id string, updates models.ClusterUpdateRequest) error {
	return s.repo.Update(ctx, id, updates)
}

// Delete deletes a cluster
func (s *clusterService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
