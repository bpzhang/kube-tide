package core

import (
	"context"

	"go.uber.org/zap"

	"kube-tide/internal/database/models"
	"kube-tide/internal/repository"
)

// nodeService implements NodeService interface
type nodeService struct {
	repo   repository.NodeRepository
	logger *zap.Logger
}

// Create creates a new node
func (s *nodeService) Create(ctx context.Context, node *models.Node) error {
	if err := node.Validate(); err != nil {
		s.logger.Error("node validation failed", zap.Error(err))
		return err
	}

	return s.repo.Create(ctx, node)
}

// GetByID retrieves a node by ID
func (s *nodeService) GetByID(ctx context.Context, id string) (*models.Node, error) {
	return s.repo.GetByID(ctx, id)
}

// GetByClusterAndName retrieves a node by cluster ID and name
func (s *nodeService) GetByClusterAndName(ctx context.Context, clusterID, name string) (*models.Node, error) {
	return s.repo.GetByClusterAndName(ctx, clusterID, name)
}

// ListByCluster retrieves nodes by cluster ID with pagination
func (s *nodeService) ListByCluster(ctx context.Context, clusterID string, params models.PaginationParams) (*models.PaginatedResult, error) {
	return s.repo.ListByCluster(ctx, clusterID, params)
}

// Update updates a node
func (s *nodeService) Update(ctx context.Context, id string, updates models.NodeUpdateRequest) error {
	return s.repo.Update(ctx, id, updates)
}

// Delete deletes a node
func (s *nodeService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// DeleteByCluster deletes all nodes in a cluster
func (s *nodeService) DeleteByCluster(ctx context.Context, clusterID string) error {
	return s.repo.DeleteByCluster(ctx, clusterID)
}

// Count counts nodes in a cluster
func (s *nodeService) Count(ctx context.Context, clusterID string) (int, error) {
	return s.repo.Count(ctx, clusterID)
}
