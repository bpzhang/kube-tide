package core

import (
	"context"

	"go.uber.org/zap"

	"kube-tide/internal/database/models"
	"kube-tide/internal/repository"
)

// podService implements PodService interface
type podService struct {
	repo   repository.PodRepository
	logger *zap.Logger
}

// Create creates a new pod
func (s *podService) Create(ctx context.Context, pod *models.Pod) error {
	if err := pod.Validate(); err != nil {
		s.logger.Error("pod validation failed", zap.Error(err))
		return err
	}

	return s.repo.Create(ctx, pod)
}

// GetByID retrieves a pod by ID
func (s *podService) GetByID(ctx context.Context, id string) (*models.Pod, error) {
	return s.repo.GetByID(ctx, id)
}

// GetByClusterNamespaceAndName retrieves a pod by cluster ID, namespace, and name
func (s *podService) GetByClusterNamespaceAndName(ctx context.Context, clusterID, namespace, name string) (*models.Pod, error) {
	return s.repo.GetByClusterNamespaceAndName(ctx, clusterID, namespace, name)
}

// ListByCluster retrieves pods by cluster ID with pagination
func (s *podService) ListByCluster(ctx context.Context, clusterID string, params models.PaginationParams) (*models.PaginatedResult, error) {
	return s.repo.ListByCluster(ctx, clusterID, params)
}

// ListByNamespace retrieves pods by cluster ID and namespace with pagination
func (s *podService) ListByNamespace(ctx context.Context, clusterID, namespace string, params models.PaginationParams) (*models.PaginatedResult, error) {
	return s.repo.ListByNamespace(ctx, clusterID, namespace, params)
}

// Update updates a pod
func (s *podService) Update(ctx context.Context, id string, updates models.PodUpdateRequest) error {
	return s.repo.Update(ctx, id, updates)
}

// Delete deletes a pod
func (s *podService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// DeleteByCluster deletes all pods in a cluster
func (s *podService) DeleteByCluster(ctx context.Context, clusterID string) error {
	return s.repo.DeleteByCluster(ctx, clusterID)
}

// DeleteByNamespace deletes all pods in a namespace
func (s *podService) DeleteByNamespace(ctx context.Context, clusterID, namespace string) error {
	return s.repo.DeleteByNamespace(ctx, clusterID, namespace)
}

// Count counts pods in a cluster
func (s *podService) Count(ctx context.Context, clusterID string) (int, error) {
	return s.repo.Count(ctx, clusterID)
}

// CountByNamespace counts pods in a namespace
func (s *podService) CountByNamespace(ctx context.Context, clusterID, namespace string) (int, error) {
	return s.repo.CountByNamespace(ctx, clusterID, namespace)
} 