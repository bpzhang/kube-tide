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

// namespaceRepository implements NamespaceRepository interface
type namespaceRepository struct {
	db     *database.Database
	logger *zap.Logger
}

// NewNamespaceRepository creates a new namespace repository
func NewNamespaceRepository(db *database.Database, logger *zap.Logger) NamespaceRepository {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &namespaceRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new namespace
func (r *namespaceRepository) Create(ctx context.Context, namespace *models.Namespace) error {
	namespace.ID = uuid.New().String()
	namespace.CreatedAt = time.Now()
	namespace.UpdatedAt = time.Now()

	query := `
		INSERT INTO namespaces (id, cluster_id, name, status, phase, labels, annotations, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.db.ExecContext(ctx, query,
		namespace.ID, namespace.ClusterID, namespace.Name, namespace.Status, namespace.Phase,
		namespace.Labels, namespace.Annotations, namespace.CreatedAt, namespace.UpdatedAt)

	if err != nil {
		r.logger.Error("failed to create namespace", zap.Error(err), zap.String("name", namespace.Name))
		return fmt.Errorf("failed to create namespace: %w", err)
	}

	r.logger.Info("namespace created successfully", zap.String("id", namespace.ID), zap.String("name", namespace.Name))
	return nil
}

// CreateTx creates a new namespace within a transaction
func (r *namespaceRepository) CreateTx(ctx context.Context, tx *sql.Tx, namespace *models.Namespace) error {
	namespace.ID = uuid.New().String()
	namespace.CreatedAt = time.Now()
	namespace.UpdatedAt = time.Now()

	query := `
		INSERT INTO namespaces (id, cluster_id, name, status, phase, labels, annotations, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := tx.ExecContext(ctx, query,
		namespace.ID, namespace.ClusterID, namespace.Name, namespace.Status, namespace.Phase,
		namespace.Labels, namespace.Annotations, namespace.CreatedAt, namespace.UpdatedAt)

	if err != nil {
		r.logger.Error("failed to create namespace in transaction", zap.Error(err), zap.String("name", namespace.Name))
		return fmt.Errorf("failed to create namespace: %w", err)
	}

	return nil
}

// GetByID retrieves a namespace by ID
func (r *namespaceRepository) GetByID(ctx context.Context, id string) (*models.Namespace, error) {
	query := `
		SELECT id, cluster_id, name, status, phase, labels, annotations, created_at, updated_at
		FROM namespaces WHERE id = $1`

	var namespace models.Namespace
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&namespace.ID, &namespace.ClusterID, &namespace.Name, &namespace.Status, &namespace.Phase,
		&namespace.Labels, &namespace.Annotations, &namespace.CreatedAt, &namespace.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("namespace not found: %s", id)
		}
		r.logger.Error("failed to get namespace by ID", zap.Error(err), zap.String("id", id))
		return nil, fmt.Errorf("failed to get namespace: %w", err)
	}

	return &namespace, nil
}

// GetByClusterAndName retrieves a namespace by cluster ID and name
func (r *namespaceRepository) GetByClusterAndName(ctx context.Context, clusterID, name string) (*models.Namespace, error) {
	query := `
		SELECT id, cluster_id, name, status, phase, labels, annotations, created_at, updated_at
		FROM namespaces WHERE cluster_id = $1 AND name = $2`

	var namespace models.Namespace
	err := r.db.QueryRowContext(ctx, query, clusterID, name).Scan(
		&namespace.ID, &namespace.ClusterID, &namespace.Name, &namespace.Status, &namespace.Phase,
		&namespace.Labels, &namespace.Annotations, &namespace.CreatedAt, &namespace.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("namespace not found: cluster=%s, name=%s", clusterID, name)
		}
		r.logger.Error("failed to get namespace by cluster and name", zap.Error(err),
			zap.String("cluster_id", clusterID), zap.String("name", name))
		return nil, fmt.Errorf("failed to get namespace: %w", err)
	}

	return &namespace, nil
}

// ListByCluster retrieves namespaces by cluster ID with pagination
func (r *namespaceRepository) ListByCluster(ctx context.Context, clusterID string, params models.PaginationParams) (*models.PaginatedResult, error) {
	// Count total records
	countQuery := "SELECT COUNT(*) FROM namespaces WHERE cluster_id = $1"
	var totalCount int
	err := r.db.QueryRowContext(ctx, countQuery, clusterID).Scan(&totalCount)
	if err != nil {
		r.logger.Error("failed to count namespaces", zap.Error(err))
		return nil, fmt.Errorf("failed to count namespaces: %w", err)
	}

	// Get paginated results
	query := `
		SELECT id, cluster_id, name, status, phase, labels, annotations, created_at, updated_at
		FROM namespaces WHERE cluster_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, clusterID, params.PageSize, params.Offset())
	if err != nil {
		r.logger.Error("failed to query namespaces", zap.Error(err))
		return nil, fmt.Errorf("failed to query namespaces: %w", err)
	}
	defer rows.Close()

	var namespaces []*models.Namespace
	for rows.Next() {
		var namespace models.Namespace
		err := rows.Scan(
			&namespace.ID, &namespace.ClusterID, &namespace.Name, &namespace.Status, &namespace.Phase,
			&namespace.Labels, &namespace.Annotations, &namespace.CreatedAt, &namespace.UpdatedAt)
		if err != nil {
			r.logger.Error("failed to scan namespace", zap.Error(err))
			return nil, fmt.Errorf("failed to scan namespace: %w", err)
		}
		namespaces = append(namespaces, &namespace)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating namespace rows", zap.Error(err))
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return models.NewPaginatedResult(namespaces, totalCount, params), nil
}

// Update updates a namespace
func (r *namespaceRepository) Update(ctx context.Context, id string, updates models.NamespaceUpdateRequest) error {
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

	if len(setParts) == 0 {
		return fmt.Errorf("no fields to update")
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	args = append(args, id)

	query := fmt.Sprintf("UPDATE namespaces SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIndex)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("failed to update namespace", zap.Error(err), zap.String("id", id))
		return fmt.Errorf("failed to update namespace: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("namespace not found: %s", id)
	}

	r.logger.Info("namespace updated successfully", zap.String("id", id))
	return nil
}

// UpdateTx updates a namespace within a transaction
func (r *namespaceRepository) UpdateTx(ctx context.Context, tx *sql.Tx, id string, updates models.NamespaceUpdateRequest) error {
	return fmt.Errorf("UpdateTx not implemented yet")
}

// Delete deletes a namespace
func (r *namespaceRepository) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM namespaces WHERE id = $1"

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("failed to delete namespace", zap.Error(err), zap.String("id", id))
		return fmt.Errorf("failed to delete namespace: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("namespace not found: %s", id)
	}

	r.logger.Info("namespace deleted successfully", zap.String("id", id))
	return nil
}

// DeleteTx deletes a namespace within a transaction
func (r *namespaceRepository) DeleteTx(ctx context.Context, tx *sql.Tx, id string) error {
	query := "DELETE FROM namespaces WHERE id = $1"

	result, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("failed to delete namespace in transaction", zap.Error(err), zap.String("id", id))
		return fmt.Errorf("failed to delete namespace: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("namespace not found: %s", id)
	}

	return nil
}

// DeleteByCluster deletes all namespaces in a cluster
func (r *namespaceRepository) DeleteByCluster(ctx context.Context, clusterID string) error {
	query := "DELETE FROM namespaces WHERE cluster_id = $1"

	result, err := r.db.ExecContext(ctx, query, clusterID)
	if err != nil {
		r.logger.Error("failed to delete namespaces by cluster", zap.Error(err), zap.String("cluster_id", clusterID))
		return fmt.Errorf("failed to delete namespaces by cluster: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	r.logger.Info("namespaces deleted by cluster", zap.String("cluster_id", clusterID), zap.Int64("count", rowsAffected))
	return nil
}

// Count counts namespaces in a cluster
func (r *namespaceRepository) Count(ctx context.Context, clusterID string) (int, error) {
	query := "SELECT COUNT(*) FROM namespaces WHERE cluster_id = $1"

	var count int
	err := r.db.QueryRowContext(ctx, query, clusterID).Scan(&count)
	if err != nil {
		r.logger.Error("failed to count namespaces", zap.Error(err))
		return 0, fmt.Errorf("failed to count namespaces: %w", err)
	}

	return count, nil
}
