package core

import (
	"context"

	"go.uber.org/zap"

	"kube-tide/internal/database/models"
	"kube-tide/internal/repository"
)

// deploymentService implements DeploymentService interface
type deploymentService struct {
	repo   repository.DeploymentRepository
	logger *zap.Logger
}

// Create creates a new deployment
func (s *deploymentService) Create(ctx context.Context, deployment *models.Deployment) error {
	if err := deployment.Validate(); err != nil {
		s.logger.Error("deployment validation failed", zap.Error(err))
		return err
	}

	return s.repo.Create(ctx, deployment)
}

// GetByID retrieves a deployment by ID
func (s *deploymentService) GetByID(ctx context.Context, id string) (*models.Deployment, error) {
	return s.repo.GetByID(ctx, id)
}

// GetByClusterNamespaceAndName retrieves a deployment by cluster ID, namespace, and name
func (s *deploymentService) GetByClusterNamespaceAndName(ctx context.Context, clusterID, namespace, name string) (*models.Deployment, error) {
	return s.repo.GetByClusterNamespaceAndName(ctx, clusterID, namespace, name)
}

// ListByCluster retrieves deployments by cluster ID with pagination
func (s *deploymentService) ListByCluster(ctx context.Context, clusterID string, params models.PaginationParams) (*models.PaginatedResult, error) {
	return s.repo.ListByCluster(ctx, clusterID, params)
}

// ListByNamespace retrieves deployments by cluster ID and namespace with pagination
func (s *deploymentService) ListByNamespace(ctx context.Context, clusterID, namespace string, params models.PaginationParams) (*models.PaginatedResult, error) {
	return s.repo.ListByNamespace(ctx, clusterID, namespace, params)
}

// Update updates a deployment
func (s *deploymentService) Update(ctx context.Context, id string, updates models.DeploymentUpdateRequest) error {
	return s.repo.Update(ctx, id, updates)
}

// Delete deletes a deployment
func (s *deploymentService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// DeleteByCluster deletes all deployments in a cluster
func (s *deploymentService) DeleteByCluster(ctx context.Context, clusterID string) error {
	return s.repo.DeleteByCluster(ctx, clusterID)
}

// DeleteByNamespace deletes all deployments in a namespace
func (s *deploymentService) DeleteByNamespace(ctx context.Context, clusterID, namespace string) error {
	return s.repo.DeleteByNamespace(ctx, clusterID, namespace)
}

// Count counts deployments in a cluster
func (s *deploymentService) Count(ctx context.Context, clusterID string) (int, error) {
	return s.repo.Count(ctx, clusterID)
}

// CountByNamespace counts deployments in a namespace
func (s *deploymentService) CountByNamespace(ctx context.Context, clusterID, namespace string) (int, error) {
	return s.repo.CountByNamespace(ctx, clusterID, namespace)
}
