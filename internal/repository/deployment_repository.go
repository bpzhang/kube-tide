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

// deploymentRepository implements DeploymentRepository interface
type deploymentRepository struct {
	db     *database.Database
	logger *zap.Logger
}

// NewDeploymentRepository creates a new deployment repository
func NewDeploymentRepository(db *database.Database, logger *zap.Logger) DeploymentRepository {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &deploymentRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new deployment
func (r *deploymentRepository) Create(ctx context.Context, deployment *models.Deployment) error {
	deployment.ID = uuid.New().String()
	deployment.CreatedAt = time.Now()
	deployment.UpdatedAt = time.Now()

	query := `
		INSERT INTO deployments (id, cluster_id, namespace, name, replicas, ready_replicas, 
			available_replicas, unavailable_replicas, updated_replicas, strategy_type,
			labels, annotations, selector, template, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`

	_, err := r.db.ExecContext(ctx, query,
		deployment.ID, deployment.ClusterID, deployment.Namespace, deployment.Name,
		deployment.Replicas, deployment.ReadyReplicas, deployment.AvailableReplicas,
		deployment.UnavailableReplicas, deployment.UpdatedReplicas, deployment.StrategyType,
		deployment.Labels, deployment.Annotations, deployment.Selector, deployment.Template,
		deployment.CreatedAt, deployment.UpdatedAt)

	if err != nil {
		r.logger.Error("failed to create deployment", zap.Error(err), zap.String("name", deployment.Name))
		return fmt.Errorf("failed to create deployment: %w", err)
	}

	r.logger.Info("deployment created successfully", zap.String("id", deployment.ID), zap.String("name", deployment.Name))
	return nil
}

// CreateTx creates a new deployment within a transaction
func (r *deploymentRepository) CreateTx(ctx context.Context, tx *sql.Tx, deployment *models.Deployment) error {
	deployment.ID = uuid.New().String()
	deployment.CreatedAt = time.Now()
	deployment.UpdatedAt = time.Now()

	query := `
		INSERT INTO deployments (id, cluster_id, namespace, name, replicas, ready_replicas, 
			available_replicas, unavailable_replicas, updated_replicas, strategy_type,
			labels, annotations, selector, template, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`

	_, err := tx.ExecContext(ctx, query,
		deployment.ID, deployment.ClusterID, deployment.Namespace, deployment.Name,
		deployment.Replicas, deployment.ReadyReplicas, deployment.AvailableReplicas,
		deployment.UnavailableReplicas, deployment.UpdatedReplicas, deployment.StrategyType,
		deployment.Labels, deployment.Annotations, deployment.Selector, deployment.Template,
		deployment.CreatedAt, deployment.UpdatedAt)

	if err != nil {
		r.logger.Error("failed to create deployment in transaction", zap.Error(err), zap.String("name", deployment.Name))
		return fmt.Errorf("failed to create deployment: %w", err)
	}

	return nil
}

// GetByID retrieves a deployment by ID
func (r *deploymentRepository) GetByID(ctx context.Context, id string) (*models.Deployment, error) {
	query := `
		SELECT id, cluster_id, namespace, name, replicas, ready_replicas, 
			available_replicas, unavailable_replicas, updated_replicas, strategy_type,
			labels, annotations, selector, template, created_at, updated_at
		FROM deployments WHERE id = $1`

	var deployment models.Deployment
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&deployment.ID, &deployment.ClusterID, &deployment.Namespace, &deployment.Name,
		&deployment.Replicas, &deployment.ReadyReplicas, &deployment.AvailableReplicas,
		&deployment.UnavailableReplicas, &deployment.UpdatedReplicas, &deployment.StrategyType,
		&deployment.Labels, &deployment.Annotations, &deployment.Selector, &deployment.Template,
		&deployment.CreatedAt, &deployment.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("deployment not found: %s", id)
		}
		r.logger.Error("failed to get deployment by ID", zap.Error(err), zap.String("id", id))
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}

	return &deployment, nil
}

// GetByClusterNamespaceAndName retrieves a deployment by cluster ID, namespace, and name
func (r *deploymentRepository) GetByClusterNamespaceAndName(ctx context.Context, clusterID, namespace, name string) (*models.Deployment, error) {
	query := `
		SELECT id, cluster_id, namespace, name, replicas, ready_replicas, 
			available_replicas, unavailable_replicas, updated_replicas, strategy_type,
			labels, annotations, selector, template, created_at, updated_at
		FROM deployments WHERE cluster_id = $1 AND namespace = $2 AND name = $3`

	var deployment models.Deployment
	err := r.db.QueryRowContext(ctx, query, clusterID, namespace, name).Scan(
		&deployment.ID, &deployment.ClusterID, &deployment.Namespace, &deployment.Name,
		&deployment.Replicas, &deployment.ReadyReplicas, &deployment.AvailableReplicas,
		&deployment.UnavailableReplicas, &deployment.UpdatedReplicas, &deployment.StrategyType,
		&deployment.Labels, &deployment.Annotations, &deployment.Selector, &deployment.Template,
		&deployment.CreatedAt, &deployment.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("deployment not found: cluster=%s, namespace=%s, name=%s", clusterID, namespace, name)
		}
		r.logger.Error("failed to get deployment by cluster, namespace and name", zap.Error(err),
			zap.String("cluster_id", clusterID), zap.String("namespace", namespace), zap.String("name", name))
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}

	return &deployment, nil
}

// ListByCluster retrieves deployments by cluster ID with pagination
func (r *deploymentRepository) ListByCluster(ctx context.Context, clusterID string, params models.PaginationParams) (*models.PaginatedResult, error) {
	// Count total records
	countQuery := "SELECT COUNT(*) FROM deployments WHERE cluster_id = $1"
	var totalCount int
	err := r.db.QueryRowContext(ctx, countQuery, clusterID).Scan(&totalCount)
	if err != nil {
		r.logger.Error("failed to count deployments", zap.Error(err))
		return nil, fmt.Errorf("failed to count deployments: %w", err)
	}

	// Get paginated results
	query := `
		SELECT id, cluster_id, namespace, name, replicas, ready_replicas, 
			available_replicas, unavailable_replicas, updated_replicas, strategy_type,
			labels, annotations, selector, template, created_at, updated_at
		FROM deployments WHERE cluster_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, clusterID, params.PageSize, params.Offset())
	if err != nil {
		r.logger.Error("failed to query deployments", zap.Error(err))
		return nil, fmt.Errorf("failed to query deployments: %w", err)
	}
	defer rows.Close()

	var deployments []*models.Deployment
	for rows.Next() {
		var deployment models.Deployment
		err := rows.Scan(
			&deployment.ID, &deployment.ClusterID, &deployment.Namespace, &deployment.Name,
			&deployment.Replicas, &deployment.ReadyReplicas, &deployment.AvailableReplicas,
			&deployment.UnavailableReplicas, &deployment.UpdatedReplicas, &deployment.StrategyType,
			&deployment.Labels, &deployment.Annotations, &deployment.Selector, &deployment.Template,
			&deployment.CreatedAt, &deployment.UpdatedAt)
		if err != nil {
			r.logger.Error("failed to scan deployment", zap.Error(err))
			return nil, fmt.Errorf("failed to scan deployment: %w", err)
		}
		deployments = append(deployments, &deployment)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating deployment rows", zap.Error(err))
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return models.NewPaginatedResult(deployments, totalCount, params), nil
}

// ListByNamespace retrieves deployments by cluster ID and namespace with pagination
func (r *deploymentRepository) ListByNamespace(ctx context.Context, clusterID, namespace string, params models.PaginationParams) (*models.PaginatedResult, error) {
	// Count total records
	countQuery := "SELECT COUNT(*) FROM deployments WHERE cluster_id = $1 AND namespace = $2"
	var totalCount int
	err := r.db.QueryRowContext(ctx, countQuery, clusterID, namespace).Scan(&totalCount)
	if err != nil {
		r.logger.Error("failed to count deployments", zap.Error(err))
		return nil, fmt.Errorf("failed to count deployments: %w", err)
	}

	// Get paginated results
	query := `
		SELECT id, cluster_id, namespace, name, replicas, ready_replicas, 
			available_replicas, unavailable_replicas, updated_replicas, strategy_type,
			labels, annotations, selector, template, created_at, updated_at
		FROM deployments WHERE cluster_id = $1 AND namespace = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`

	rows, err := r.db.QueryContext(ctx, query, clusterID, namespace, params.PageSize, params.Offset())
	if err != nil {
		r.logger.Error("failed to query deployments", zap.Error(err))
		return nil, fmt.Errorf("failed to query deployments: %w", err)
	}
	defer rows.Close()

	var deployments []*models.Deployment
	for rows.Next() {
		var deployment models.Deployment
		err := rows.Scan(
			&deployment.ID, &deployment.ClusterID, &deployment.Namespace, &deployment.Name,
			&deployment.Replicas, &deployment.ReadyReplicas, &deployment.AvailableReplicas,
			&deployment.UnavailableReplicas, &deployment.UpdatedReplicas, &deployment.StrategyType,
			&deployment.Labels, &deployment.Annotations, &deployment.Selector, &deployment.Template,
			&deployment.CreatedAt, &deployment.UpdatedAt)
		if err != nil {
			r.logger.Error("failed to scan deployment", zap.Error(err))
			return nil, fmt.Errorf("failed to scan deployment: %w", err)
		}
		deployments = append(deployments, &deployment)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating deployment rows", zap.Error(err))
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return models.NewPaginatedResult(deployments, totalCount, params), nil
}

// Update updates a deployment
func (r *deploymentRepository) Update(ctx context.Context, id string, updates models.DeploymentUpdateRequest) error {
	var setParts []string
	var args []interface{}
	argIndex := 1

	if updates.Replicas != nil {
		setParts = append(setParts, fmt.Sprintf("replicas = $%d", argIndex))
		args = append(args, *updates.Replicas)
		argIndex++
	}

	if updates.ReadyReplicas != nil {
		setParts = append(setParts, fmt.Sprintf("ready_replicas = $%d", argIndex))
		args = append(args, *updates.ReadyReplicas)
		argIndex++
	}

	if updates.AvailableReplicas != nil {
		setParts = append(setParts, fmt.Sprintf("available_replicas = $%d", argIndex))
		args = append(args, *updates.AvailableReplicas)
		argIndex++
	}

	if updates.UnavailableReplicas != nil {
		setParts = append(setParts, fmt.Sprintf("unavailable_replicas = $%d", argIndex))
		args = append(args, *updates.UnavailableReplicas)
		argIndex++
	}

	if updates.UpdatedReplicas != nil {
		setParts = append(setParts, fmt.Sprintf("updated_replicas = $%d", argIndex))
		args = append(args, *updates.UpdatedReplicas)
		argIndex++
	}

	if updates.StrategyType != nil {
		setParts = append(setParts, fmt.Sprintf("strategy_type = $%d", argIndex))
		args = append(args, *updates.StrategyType)
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

	if updates.Selector != nil {
		setParts = append(setParts, fmt.Sprintf("selector = $%d", argIndex))
		args = append(args, *updates.Selector)
		argIndex++
	}

	if updates.Template != nil {
		setParts = append(setParts, fmt.Sprintf("template = $%d", argIndex))
		args = append(args, *updates.Template)
		argIndex++
	}

	if len(setParts) == 0 {
		return fmt.Errorf("no fields to update")
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	args = append(args, id)

	query := fmt.Sprintf("UPDATE deployments SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIndex)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("failed to update deployment", zap.Error(err), zap.String("id", id))
		return fmt.Errorf("failed to update deployment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("deployment not found: %s", id)
	}

	r.logger.Info("deployment updated successfully", zap.String("id", id))
	return nil
}

// UpdateTx updates a deployment within a transaction
func (r *deploymentRepository) UpdateTx(ctx context.Context, tx *sql.Tx, id string, updates models.DeploymentUpdateRequest) error {
	var setParts []string
	var args []interface{}
	argIndex := 1

	if updates.Replicas != nil {
		setParts = append(setParts, fmt.Sprintf("replicas = $%d", argIndex))
		args = append(args, *updates.Replicas)
		argIndex++
	}

	if updates.ReadyReplicas != nil {
		setParts = append(setParts, fmt.Sprintf("ready_replicas = $%d", argIndex))
		args = append(args, *updates.ReadyReplicas)
		argIndex++
	}

	if updates.AvailableReplicas != nil {
		setParts = append(setParts, fmt.Sprintf("available_replicas = $%d", argIndex))
		args = append(args, *updates.AvailableReplicas)
		argIndex++
	}

	if updates.UnavailableReplicas != nil {
		setParts = append(setParts, fmt.Sprintf("unavailable_replicas = $%d", argIndex))
		args = append(args, *updates.UnavailableReplicas)
		argIndex++
	}

	if updates.UpdatedReplicas != nil {
		setParts = append(setParts, fmt.Sprintf("updated_replicas = $%d", argIndex))
		args = append(args, *updates.UpdatedReplicas)
		argIndex++
	}

	if updates.StrategyType != nil {
		setParts = append(setParts, fmt.Sprintf("strategy_type = $%d", argIndex))
		args = append(args, *updates.StrategyType)
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

	if updates.Selector != nil {
		setParts = append(setParts, fmt.Sprintf("selector = $%d", argIndex))
		args = append(args, *updates.Selector)
		argIndex++
	}

	if updates.Template != nil {
		setParts = append(setParts, fmt.Sprintf("template = $%d", argIndex))
		args = append(args, *updates.Template)
		argIndex++
	}

	if len(setParts) == 0 {
		return fmt.Errorf("no fields to update")
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	args = append(args, id)

	query := fmt.Sprintf("UPDATE deployments SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIndex)

	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("failed to update deployment in transaction", zap.Error(err), zap.String("id", id))
		return fmt.Errorf("failed to update deployment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("deployment not found: %s", id)
	}

	return nil
}

// Delete deletes a deployment
func (r *deploymentRepository) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM deployments WHERE id = $1"

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("failed to delete deployment", zap.Error(err), zap.String("id", id))
		return fmt.Errorf("failed to delete deployment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("deployment not found: %s", id)
	}

	r.logger.Info("deployment deleted successfully", zap.String("id", id))
	return nil
}

// DeleteTx deletes a deployment within a transaction
func (r *deploymentRepository) DeleteTx(ctx context.Context, tx *sql.Tx, id string) error {
	query := "DELETE FROM deployments WHERE id = $1"

	result, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("failed to delete deployment in transaction", zap.Error(err), zap.String("id", id))
		return fmt.Errorf("failed to delete deployment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("deployment not found: %s", id)
	}

	return nil
}

// DeleteByCluster deletes all deployments in a cluster
func (r *deploymentRepository) DeleteByCluster(ctx context.Context, clusterID string) error {
	query := "DELETE FROM deployments WHERE cluster_id = $1"

	result, err := r.db.ExecContext(ctx, query, clusterID)
	if err != nil {
		r.logger.Error("failed to delete deployments by cluster", zap.Error(err), zap.String("cluster_id", clusterID))
		return fmt.Errorf("failed to delete deployments by cluster: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	r.logger.Info("deployments deleted by cluster", zap.String("cluster_id", clusterID), zap.Int64("count", rowsAffected))
	return nil
}

// DeleteByNamespace deletes all deployments in a namespace
func (r *deploymentRepository) DeleteByNamespace(ctx context.Context, clusterID, namespace string) error {
	query := "DELETE FROM deployments WHERE cluster_id = $1 AND namespace = $2"

	result, err := r.db.ExecContext(ctx, query, clusterID, namespace)
	if err != nil {
		r.logger.Error("failed to delete deployments by namespace", zap.Error(err),
			zap.String("cluster_id", clusterID), zap.String("namespace", namespace))
		return fmt.Errorf("failed to delete deployments by namespace: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	r.logger.Info("deployments deleted by namespace",
		zap.String("cluster_id", clusterID), zap.String("namespace", namespace), zap.Int64("count", rowsAffected))
	return nil
}

// Count counts deployments in a cluster
func (r *deploymentRepository) Count(ctx context.Context, clusterID string) (int, error) {
	query := "SELECT COUNT(*) FROM deployments WHERE cluster_id = $1"

	var count int
	err := r.db.QueryRowContext(ctx, query, clusterID).Scan(&count)
	if err != nil {
		r.logger.Error("failed to count deployments", zap.Error(err))
		return 0, fmt.Errorf("failed to count deployments: %w", err)
	}

	return count, nil
}

// CountByNamespace counts deployments in a namespace
func (r *deploymentRepository) CountByNamespace(ctx context.Context, clusterID, namespace string) (int, error) {
	query := "SELECT COUNT(*) FROM deployments WHERE cluster_id = $1 AND namespace = $2"

	var count int
	err := r.db.QueryRowContext(ctx, query, clusterID, namespace).Scan(&count)
	if err != nil {
		r.logger.Error("failed to count deployments by namespace", zap.Error(err))
		return 0, fmt.Errorf("failed to count deployments by namespace: %w", err)
	}

	return count, nil
}
