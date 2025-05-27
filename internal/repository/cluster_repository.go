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

// clusterRepository implements ClusterRepository interface
type clusterRepository struct {
	db     *database.Database
	logger *zap.Logger
}

// NewClusterRepository creates a new cluster repository
func NewClusterRepository(db *database.Database, logger *zap.Logger) ClusterRepository {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &clusterRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new cluster
func (r *clusterRepository) Create(ctx context.Context, cluster *models.Cluster) error {
	cluster.ID = uuid.New().String()
	cluster.CreatedAt = time.Now()
	cluster.UpdatedAt = time.Now()

	query := `
		INSERT INTO clusters (id, name, config, status, description, kubeconfig, endpoint, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.db.ExecContext(ctx, query,
		cluster.ID, cluster.Name, cluster.Config, cluster.Status,
		cluster.Description, cluster.Kubeconfig, cluster.Endpoint,
		cluster.Version, cluster.CreatedAt, cluster.UpdatedAt)

	if err != nil {
		r.logger.Error("failed to create cluster", zap.Error(err), zap.String("name", cluster.Name))
		return fmt.Errorf("failed to create cluster: %w", err)
	}

	r.logger.Info("cluster created successfully", zap.String("id", cluster.ID), zap.String("name", cluster.Name))
	return nil
}

// CreateTx creates a new cluster within a transaction
func (r *clusterRepository) CreateTx(ctx context.Context, tx *sql.Tx, cluster *models.Cluster) error {
	cluster.ID = uuid.New().String()
	cluster.CreatedAt = time.Now()
	cluster.UpdatedAt = time.Now()

	query := `
		INSERT INTO clusters (id, name, config, status, description, kubeconfig, endpoint, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := tx.ExecContext(ctx, query,
		cluster.ID, cluster.Name, cluster.Config, cluster.Status,
		cluster.Description, cluster.Kubeconfig, cluster.Endpoint,
		cluster.Version, cluster.CreatedAt, cluster.UpdatedAt)

	if err != nil {
		r.logger.Error("failed to create cluster in transaction", zap.Error(err), zap.String("name", cluster.Name))
		return fmt.Errorf("failed to create cluster: %w", err)
	}

	return nil
}

// GetByID retrieves a cluster by ID
func (r *clusterRepository) GetByID(ctx context.Context, id string) (*models.Cluster, error) {
	query := `
		SELECT id, name, config, status, description, kubeconfig, endpoint, version, created_at, updated_at
		FROM clusters WHERE id = $1`

	var cluster models.Cluster
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&cluster.ID, &cluster.Name, &cluster.Config, &cluster.Status,
		&cluster.Description, &cluster.Kubeconfig, &cluster.Endpoint,
		&cluster.Version, &cluster.CreatedAt, &cluster.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("cluster not found: %s", id)
		}
		r.logger.Error("failed to get cluster by ID", zap.Error(err), zap.String("id", id))
		return nil, fmt.Errorf("failed to get cluster: %w", err)
	}

	return &cluster, nil
}

// GetByName retrieves a cluster by name
func (r *clusterRepository) GetByName(ctx context.Context, name string) (*models.Cluster, error) {
	query := `
		SELECT id, name, config, status, description, kubeconfig, endpoint, version, created_at, updated_at
		FROM clusters WHERE name = $1`

	var cluster models.Cluster
	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&cluster.ID, &cluster.Name, &cluster.Config, &cluster.Status,
		&cluster.Description, &cluster.Kubeconfig, &cluster.Endpoint,
		&cluster.Version, &cluster.CreatedAt, &cluster.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("cluster not found: %s", name)
		}
		r.logger.Error("failed to get cluster by name", zap.Error(err), zap.String("name", name))
		return nil, fmt.Errorf("failed to get cluster: %w", err)
	}

	return &cluster, nil
}

// List retrieves clusters with filters and pagination
func (r *clusterRepository) List(ctx context.Context, filters models.ClusterFilters, params models.PaginationParams) (*models.PaginatedResult, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	// Build WHERE conditions
	if filters.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, filters.Status)
		argIndex++
	}

	if filters.Name != "" {
		conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", argIndex))
		args = append(args, "%"+filters.Name+"%")
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM clusters %s", whereClause)
	var totalCount int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		r.logger.Error("failed to count clusters", zap.Error(err))
		return nil, fmt.Errorf("failed to count clusters: %w", err)
	}

	// Get paginated results
	query := fmt.Sprintf(`
		SELECT id, name, config, status, description, kubeconfig, endpoint, version, created_at, updated_at
		FROM clusters %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, params.PageSize, params.Offset())

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("failed to query clusters", zap.Error(err))
		return nil, fmt.Errorf("failed to query clusters: %w", err)
	}
	defer rows.Close()

	var clusters []*models.Cluster
	for rows.Next() {
		var cluster models.Cluster
		err := rows.Scan(
			&cluster.ID, &cluster.Name, &cluster.Config, &cluster.Status,
			&cluster.Description, &cluster.Kubeconfig, &cluster.Endpoint,
			&cluster.Version, &cluster.CreatedAt, &cluster.UpdatedAt)
		if err != nil {
			r.logger.Error("failed to scan cluster", zap.Error(err))
			return nil, fmt.Errorf("failed to scan cluster: %w", err)
		}
		clusters = append(clusters, &cluster)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating cluster rows", zap.Error(err))
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return models.NewPaginatedResult(clusters, totalCount, params), nil
}

// Update updates a cluster
func (r *clusterRepository) Update(ctx context.Context, id string, updates models.ClusterUpdateRequest) error {
	var setParts []string
	var args []interface{}
	argIndex := 1

	if updates.Name != nil {
		setParts = append(setParts, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, *updates.Name)
		argIndex++
	}

	if updates.Config != nil {
		setParts = append(setParts, fmt.Sprintf("config = $%d", argIndex))
		args = append(args, *updates.Config)
		argIndex++
	}

	if updates.Status != nil {
		setParts = append(setParts, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *updates.Status)
		argIndex++
	}

	if updates.Description != nil {
		setParts = append(setParts, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, *updates.Description)
		argIndex++
	}

	if updates.Kubeconfig != nil {
		setParts = append(setParts, fmt.Sprintf("kubeconfig = $%d", argIndex))
		args = append(args, *updates.Kubeconfig)
		argIndex++
	}

	if updates.Endpoint != nil {
		setParts = append(setParts, fmt.Sprintf("endpoint = $%d", argIndex))
		args = append(args, *updates.Endpoint)
		argIndex++
	}

	if updates.Version != nil {
		setParts = append(setParts, fmt.Sprintf("version = $%d", argIndex))
		args = append(args, *updates.Version)
		argIndex++
	}

	if len(setParts) == 0 {
		return fmt.Errorf("no fields to update")
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	args = append(args, id)

	query := fmt.Sprintf("UPDATE clusters SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIndex)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("failed to update cluster", zap.Error(err), zap.String("id", id))
		return fmt.Errorf("failed to update cluster: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("cluster not found: %s", id)
	}

	r.logger.Info("cluster updated successfully", zap.String("id", id))
	return nil
}

// UpdateTx updates a cluster within a transaction
func (r *clusterRepository) UpdateTx(ctx context.Context, tx *sql.Tx, id string, updates models.ClusterUpdateRequest) error {
	var setParts []string
	var args []interface{}
	argIndex := 1

	if updates.Name != nil {
		setParts = append(setParts, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, *updates.Name)
		argIndex++
	}

	if updates.Config != nil {
		setParts = append(setParts, fmt.Sprintf("config = $%d", argIndex))
		args = append(args, *updates.Config)
		argIndex++
	}

	if updates.Status != nil {
		setParts = append(setParts, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *updates.Status)
		argIndex++
	}

	if updates.Description != nil {
		setParts = append(setParts, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, *updates.Description)
		argIndex++
	}

	if updates.Kubeconfig != nil {
		setParts = append(setParts, fmt.Sprintf("kubeconfig = $%d", argIndex))
		args = append(args, *updates.Kubeconfig)
		argIndex++
	}

	if updates.Endpoint != nil {
		setParts = append(setParts, fmt.Sprintf("endpoint = $%d", argIndex))
		args = append(args, *updates.Endpoint)
		argIndex++
	}

	if updates.Version != nil {
		setParts = append(setParts, fmt.Sprintf("version = $%d", argIndex))
		args = append(args, *updates.Version)
		argIndex++
	}

	if len(setParts) == 0 {
		return fmt.Errorf("no fields to update")
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	args = append(args, id)

	query := fmt.Sprintf("UPDATE clusters SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIndex)

	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("failed to update cluster in transaction", zap.Error(err), zap.String("id", id))
		return fmt.Errorf("failed to update cluster: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("cluster not found: %s", id)
	}

	return nil
}

// Delete deletes a cluster
func (r *clusterRepository) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM clusters WHERE id = $1"

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("failed to delete cluster", zap.Error(err), zap.String("id", id))
		return fmt.Errorf("failed to delete cluster: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("cluster not found: %s", id)
	}

	r.logger.Info("cluster deleted successfully", zap.String("id", id))
	return nil
}

// DeleteTx deletes a cluster within a transaction
func (r *clusterRepository) DeleteTx(ctx context.Context, tx *sql.Tx, id string) error {
	query := "DELETE FROM clusters WHERE id = $1"

	result, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("failed to delete cluster in transaction", zap.Error(err), zap.String("id", id))
		return fmt.Errorf("failed to delete cluster: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("cluster not found: %s", id)
	}

	return nil
}

// Count counts clusters with filters
func (r *clusterRepository) Count(ctx context.Context, filters models.ClusterFilters) (int, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	if filters.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, filters.Status)
		argIndex++
	}

	if filters.Name != "" {
		conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", argIndex))
		args = append(args, "%"+filters.Name+"%")
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	query := fmt.Sprintf("SELECT COUNT(*) FROM clusters %s", whereClause)

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		r.logger.Error("failed to count clusters", zap.Error(err))
		return 0, fmt.Errorf("failed to count clusters: %w", err)
	}

	return count, nil
}

// Exists checks if a cluster exists by ID
func (r *clusterRepository) Exists(ctx context.Context, id string) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM clusters WHERE id = $1)"

	var exists bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		r.logger.Error("failed to check cluster existence", zap.Error(err), zap.String("id", id))
		return false, fmt.Errorf("failed to check cluster existence: %w", err)
	}

	return exists, nil
}

// ExistsByName checks if a cluster exists by name
func (r *clusterRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM clusters WHERE name = $1)"

	var exists bool
	err := r.db.QueryRowContext(ctx, query, name).Scan(&exists)
	if err != nil {
		r.logger.Error("failed to check cluster existence by name", zap.Error(err), zap.String("name", name))
		return false, fmt.Errorf("failed to check cluster existence: %w", err)
	}

	return exists, nil
}
