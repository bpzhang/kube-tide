package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"kube-tide/internal/database"
	"kube-tide/internal/database/models"
)

// podRepository implements PodRepository interface
type podRepository struct {
	db     *database.Database
	logger *zap.Logger
}

// NewPodRepository creates a new pod repository
func NewPodRepository(db *database.Database, logger *zap.Logger) PodRepository {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &podRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new pod
func (r *podRepository) Create(ctx context.Context, pod *models.Pod) error {
	pod.ID = uuid.New().String()
	pod.CreatedAt = time.Now()
	pod.UpdatedAt = time.Now()

	query := `
		INSERT INTO pods (id, cluster_id, namespace, name, status, phase, node_name, pod_ip, host_ip,
			restart_count, ready_containers, total_containers, cpu_requests, memory_requests,
			cpu_limits, memory_limits, labels, annotations, owner_references, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)`

	_, err := r.db.ExecContext(ctx, query,
		pod.ID, pod.ClusterID, pod.Namespace, pod.Name, pod.Status, pod.Phase, pod.NodeName,
		pod.PodIP, pod.HostIP, pod.RestartCount, pod.ReadyContainers, pod.TotalContainers,
		pod.CPURequests, pod.MemoryRequests, pod.CPULimits, pod.MemoryLimits,
		pod.Labels, pod.Annotations, pod.OwnerReferences, pod.CreatedAt, pod.UpdatedAt)

	if err != nil {
		r.logger.Error("failed to create pod", zap.Error(err), zap.String("name", pod.Name))
		return fmt.Errorf("failed to create pod: %w", err)
	}

	r.logger.Info("pod created successfully", zap.String("id", pod.ID), zap.String("name", pod.Name))
	return nil
}

// CreateTx creates a new pod within a transaction
func (r *podRepository) CreateTx(ctx context.Context, tx *sql.Tx, pod *models.Pod) error {
	pod.ID = uuid.New().String()
	pod.CreatedAt = time.Now()
	pod.UpdatedAt = time.Now()

	query := `
		INSERT INTO pods (id, cluster_id, namespace, name, status, phase, node_name, pod_ip, host_ip,
			restart_count, ready_containers, total_containers, cpu_requests, memory_requests,
			cpu_limits, memory_limits, labels, annotations, owner_references, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)`

	_, err := tx.ExecContext(ctx, query,
		pod.ID, pod.ClusterID, pod.Namespace, pod.Name, pod.Status, pod.Phase, pod.NodeName,
		pod.PodIP, pod.HostIP, pod.RestartCount, pod.ReadyContainers, pod.TotalContainers,
		pod.CPURequests, pod.MemoryRequests, pod.CPULimits, pod.MemoryLimits,
		pod.Labels, pod.Annotations, pod.OwnerReferences, pod.CreatedAt, pod.UpdatedAt)

	if err != nil {
		r.logger.Error("failed to create pod in transaction", zap.Error(err), zap.String("name", pod.Name))
		return fmt.Errorf("failed to create pod: %w", err)
	}

	return nil
}

// GetByID retrieves a pod by ID
func (r *podRepository) GetByID(ctx context.Context, id string) (*models.Pod, error) {
	query := `
		SELECT id, cluster_id, namespace, name, status, phase, node_name, pod_ip, host_ip,
			restart_count, ready_containers, total_containers, cpu_requests, memory_requests,
			cpu_limits, memory_limits, labels, annotations, owner_references, created_at, updated_at
		FROM pods WHERE id = $1`

	var pod models.Pod
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&pod.ID, &pod.ClusterID, &pod.Namespace, &pod.Name, &pod.Status, &pod.Phase, &pod.NodeName,
		&pod.PodIP, &pod.HostIP, &pod.RestartCount, &pod.ReadyContainers, &pod.TotalContainers,
		&pod.CPURequests, &pod.MemoryRequests, &pod.CPULimits, &pod.MemoryLimits,
		&pod.Labels, &pod.Annotations, &pod.OwnerReferences, &pod.CreatedAt, &pod.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("pod not found: %s", id)
		}
		r.logger.Error("failed to get pod by ID", zap.Error(err), zap.String("id", id))
		return nil, fmt.Errorf("failed to get pod: %w", err)
	}

	return &pod, nil
}

// GetByClusterNamespaceAndName retrieves a pod by cluster ID, namespace, and name
func (r *podRepository) GetByClusterNamespaceAndName(ctx context.Context, clusterID, namespace, name string) (*models.Pod, error) {
	query := `
		SELECT id, cluster_id, namespace, name, status, phase, node_name, pod_ip, host_ip,
			restart_count, ready_containers, total_containers, cpu_requests, memory_requests,
			cpu_limits, memory_limits, labels, annotations, owner_references, created_at, updated_at
		FROM pods WHERE cluster_id = $1 AND namespace = $2 AND name = $3`

	var pod models.Pod
	err := r.db.QueryRowContext(ctx, query, clusterID, namespace, name).Scan(
		&pod.ID, &pod.ClusterID, &pod.Namespace, &pod.Name, &pod.Status, &pod.Phase, &pod.NodeName,
		&pod.PodIP, &pod.HostIP, &pod.RestartCount, &pod.ReadyContainers, &pod.TotalContainers,
		&pod.CPURequests, &pod.MemoryRequests, &pod.CPULimits, &pod.MemoryLimits,
		&pod.Labels, &pod.Annotations, &pod.OwnerReferences, &pod.CreatedAt, &pod.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("pod not found: cluster=%s, namespace=%s, name=%s", clusterID, namespace, name)
		}
		r.logger.Error("failed to get pod by cluster, namespace and name", zap.Error(err),
			zap.String("cluster_id", clusterID), zap.String("namespace", namespace), zap.String("name", name))
		return nil, fmt.Errorf("failed to get pod: %w", err)
	}

	return &pod, nil
}

// ListByCluster retrieves pods by cluster ID with pagination
func (r *podRepository) ListByCluster(ctx context.Context, clusterID string, params models.PaginationParams) (*models.PaginatedResult, error) {
	// Count total records
	countQuery := "SELECT COUNT(*) FROM pods WHERE cluster_id = $1"
	var totalCount int
	err := r.db.QueryRowContext(ctx, countQuery, clusterID).Scan(&totalCount)
	if err != nil {
		r.logger.Error("failed to count pods", zap.Error(err))
		return nil, fmt.Errorf("failed to count pods: %w", err)
	}

	// Get paginated results
	query := `
		SELECT id, cluster_id, namespace, name, status, phase, node_name, pod_ip, host_ip,
			restart_count, ready_containers, total_containers, cpu_requests, memory_requests,
			cpu_limits, memory_limits, labels, annotations, owner_references, created_at, updated_at
		FROM pods WHERE cluster_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, clusterID, params.PageSize, params.Offset())
	if err != nil {
		r.logger.Error("failed to query pods", zap.Error(err))
		return nil, fmt.Errorf("failed to query pods: %w", err)
	}
	defer rows.Close()

	var pods []*models.Pod
	for rows.Next() {
		var pod models.Pod
		err := rows.Scan(
			&pod.ID, &pod.ClusterID, &pod.Namespace, &pod.Name, &pod.Status, &pod.Phase, &pod.NodeName,
			&pod.PodIP, &pod.HostIP, &pod.RestartCount, &pod.ReadyContainers, &pod.TotalContainers,
			&pod.CPURequests, &pod.MemoryRequests, &pod.CPULimits, &pod.MemoryLimits,
			&pod.Labels, &pod.Annotations, &pod.OwnerReferences, &pod.CreatedAt, &pod.UpdatedAt)
		if err != nil {
			r.logger.Error("failed to scan pod", zap.Error(err))
			return nil, fmt.Errorf("failed to scan pod: %w", err)
		}
		pods = append(pods, &pod)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating pod rows", zap.Error(err))
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return models.NewPaginatedResult(pods, totalCount, params), nil
}

// ListByNamespace retrieves pods by cluster ID and namespace with pagination
func (r *podRepository) ListByNamespace(ctx context.Context, clusterID, namespace string, params models.PaginationParams) (*models.PaginatedResult, error) {
	// Count total records
	countQuery := "SELECT COUNT(*) FROM pods WHERE cluster_id = $1 AND namespace = $2"
	var totalCount int
	err := r.db.QueryRowContext(ctx, countQuery, clusterID, namespace).Scan(&totalCount)
	if err != nil {
		r.logger.Error("failed to count pods", zap.Error(err))
		return nil, fmt.Errorf("failed to count pods: %w", err)
	}

	// Get paginated results
	query := `
		SELECT id, cluster_id, namespace, name, status, phase, node_name, pod_ip, host_ip,
			restart_count, ready_containers, total_containers, cpu_requests, memory_requests,
			cpu_limits, memory_limits, labels, annotations, owner_references, created_at, updated_at
		FROM pods WHERE cluster_id = $1 AND namespace = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`

	rows, err := r.db.QueryContext(ctx, query, clusterID, namespace, params.PageSize, params.Offset())
	if err != nil {
		r.logger.Error("failed to query pods", zap.Error(err))
		return nil, fmt.Errorf("failed to query pods: %w", err)
	}
	defer rows.Close()

	var pods []*models.Pod
	for rows.Next() {
		var pod models.Pod
		err := rows.Scan(
			&pod.ID, &pod.ClusterID, &pod.Namespace, &pod.Name, &pod.Status, &pod.Phase, &pod.NodeName,
			&pod.PodIP, &pod.HostIP, &pod.RestartCount, &pod.ReadyContainers, &pod.TotalContainers,
			&pod.CPURequests, &pod.MemoryRequests, &pod.CPULimits, &pod.MemoryLimits,
			&pod.Labels, &pod.Annotations, &pod.OwnerReferences, &pod.CreatedAt, &pod.UpdatedAt)
		if err != nil {
			r.logger.Error("failed to scan pod", zap.Error(err))
			return nil, fmt.Errorf("failed to scan pod: %w", err)
		}
		pods = append(pods, &pod)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating pod rows", zap.Error(err))
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return models.NewPaginatedResult(pods, totalCount, params), nil
}

// Update updates a pod
func (r *podRepository) Update(ctx context.Context, id string, updates models.PodUpdateRequest) error {
	var setParts []string
	var args []interface{}
	argIndex := 1

	if updates.Status != nil {
		setParts = append(setParts, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *updates.Status)
		argIndex++
	}

	if updates.Phase != nil {
		setParts = append(setParts, fmt.Sprintf("phase = $%d", argIndex))
		args = append(args, *updates.Phase)
		argIndex++
	}

	if updates.NodeName != nil {
		setParts = append(setParts, fmt.Sprintf("node_name = $%d", argIndex))
		args = append(args, *updates.NodeName)
		argIndex++
	}

	if updates.PodIP != nil {
		setParts = append(setParts, fmt.Sprintf("pod_ip = $%d", argIndex))
		args = append(args, *updates.PodIP)
		argIndex++
	}

	if updates.HostIP != nil {
		setParts = append(setParts, fmt.Sprintf("host_ip = $%d", argIndex))
		args = append(args, *updates.HostIP)
		argIndex++
	}

	if updates.RestartCount != nil {
		setParts = append(setParts, fmt.Sprintf("restart_count = $%d", argIndex))
		args = append(args, *updates.RestartCount)
		argIndex++
	}

	if updates.ReadyContainers != nil {
		setParts = append(setParts, fmt.Sprintf("ready_containers = $%d", argIndex))
		args = append(args, *updates.ReadyContainers)
		argIndex++
	}

	if updates.TotalContainers != nil {
		setParts = append(setParts, fmt.Sprintf("total_containers = $%d", argIndex))
		args = append(args, *updates.TotalContainers)
		argIndex++
	}

	if updates.CPURequests != nil {
		setParts = append(setParts, fmt.Sprintf("cpu_requests = $%d", argIndex))
		args = append(args, *updates.CPURequests)
		argIndex++
	}

	if updates.MemoryRequests != nil {
		setParts = append(setParts, fmt.Sprintf("memory_requests = $%d", argIndex))
		args = append(args, *updates.MemoryRequests)
		argIndex++
	}

	if updates.CPULimits != nil {
		setParts = append(setParts, fmt.Sprintf("cpu_limits = $%d", argIndex))
		args = append(args, *updates.CPULimits)
		argIndex++
	}

	if updates.MemoryLimits != nil {
		setParts = append(setParts, fmt.Sprintf("memory_limits = $%d", argIndex))
		args = append(args, *updates.MemoryLimits)
		argIndex++
	}

	if updates.Labels != nil {
		setParts = append(setParts, fmt.Sprintf("labels = $%d", argIndex))
		args = append(args, *updates.Labels)
		argIndex++
	}

	if updates.Annotations != nil {
		setParts = append(setParts, fmt.Sprintf("annotations = $%d", argIndex))
		args = append(args, *updates.Annotations)
		argIndex++
	}

	if updates.OwnerReferences != nil {
		setParts = append(setParts, fmt.Sprintf("owner_references = $%d", argIndex))
		args = append(args, *updates.OwnerReferences)
		argIndex++
	}

	if len(setParts) == 0 {
		return fmt.Errorf("no fields to update")
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	args = append(args, id)

	query := fmt.Sprintf("UPDATE pods SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIndex)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("failed to update pod", zap.Error(err), zap.String("id", id))
		return fmt.Errorf("failed to update pod: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("pod not found: %s", id)
	}

	r.logger.Info("pod updated successfully", zap.String("id", id))
	return nil
}

// UpdateTx updates a pod within a transaction
func (r *podRepository) UpdateTx(ctx context.Context, tx *sql.Tx, id string, updates models.PodUpdateRequest) error {
	// Implementation similar to Update but using tx
	return fmt.Errorf("UpdateTx not implemented yet")
}

// Delete deletes a pod
func (r *podRepository) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM pods WHERE id = $1"

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("failed to delete pod", zap.Error(err), zap.String("id", id))
		return fmt.Errorf("failed to delete pod: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("pod not found: %s", id)
	}

	r.logger.Info("pod deleted successfully", zap.String("id", id))
	return nil
}

// DeleteTx deletes a pod within a transaction
func (r *podRepository) DeleteTx(ctx context.Context, tx *sql.Tx, id string) error {
	query := "DELETE FROM pods WHERE id = $1"

	result, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("failed to delete pod in transaction", zap.Error(err), zap.String("id", id))
		return fmt.Errorf("failed to delete pod: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("pod not found: %s", id)
	}

	return nil
}

// DeleteByCluster deletes all pods in a cluster
func (r *podRepository) DeleteByCluster(ctx context.Context, clusterID string) error {
	query := "DELETE FROM pods WHERE cluster_id = $1"

	result, err := r.db.ExecContext(ctx, query, clusterID)
	if err != nil {
		r.logger.Error("failed to delete pods by cluster", zap.Error(err), zap.String("cluster_id", clusterID))
		return fmt.Errorf("failed to delete pods by cluster: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	r.logger.Info("pods deleted by cluster", zap.String("cluster_id", clusterID), zap.Int64("count", rowsAffected))
	return nil
}

// DeleteByNamespace deletes all pods in a namespace
func (r *podRepository) DeleteByNamespace(ctx context.Context, clusterID, namespace string) error {
	query := "DELETE FROM pods WHERE cluster_id = $1 AND namespace = $2"

	result, err := r.db.ExecContext(ctx, query, clusterID, namespace)
	if err != nil {
		r.logger.Error("failed to delete pods by namespace", zap.Error(err),
			zap.String("cluster_id", clusterID), zap.String("namespace", namespace))
		return fmt.Errorf("failed to delete pods by namespace: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	r.logger.Info("pods deleted by namespace",
		zap.String("cluster_id", clusterID), zap.String("namespace", namespace), zap.Int64("count", rowsAffected))
	return nil
}

// Count counts pods in a cluster
func (r *podRepository) Count(ctx context.Context, clusterID string) (int, error) {
	query := "SELECT COUNT(*) FROM pods WHERE cluster_id = $1"

	var count int
	err := r.db.QueryRowContext(ctx, query, clusterID).Scan(&count)
	if err != nil {
		r.logger.Error("failed to count pods", zap.Error(err))
		return 0, fmt.Errorf("failed to count pods: %w", err)
	}

	return count, nil
}

// CountByNamespace counts pods in a namespace
func (r *podRepository) CountByNamespace(ctx context.Context, clusterID, namespace string) (int, error) {
	query := "SELECT COUNT(*) FROM pods WHERE cluster_id = $1 AND namespace = $2"

	var count int
	err := r.db.QueryRowContext(ctx, query, clusterID, namespace).Scan(&count)
	if err != nil {
		r.logger.Error("failed to count pods by namespace", zap.Error(err))
		return 0, fmt.Errorf("failed to count pods by namespace: %w", err)
	}

	return count, nil
}
