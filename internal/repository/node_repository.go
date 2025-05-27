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

// nodeRepository implements NodeRepository interface
type nodeRepository struct {
	db     *database.Database
	logger *zap.Logger
}

// NewNodeRepository creates a new node repository
func NewNodeRepository(db *database.Database, logger *zap.Logger) NodeRepository {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &nodeRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new node
func (r *nodeRepository) Create(ctx context.Context, node *models.Node) error {
	node.ID = uuid.New().String()
	node.CreatedAt = time.Now()
	node.UpdatedAt = time.Now()

	query := `
		INSERT INTO nodes (id, cluster_id, name, status, roles, age, version, internal_ip, external_ip, 
			os_image, kernel_version, container_runtime, cpu_capacity, memory_capacity, 
			cpu_allocatable, memory_allocatable, conditions, labels, annotations, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)`

	_, err := r.db.ExecContext(ctx, query,
		node.ID, node.ClusterID, node.Name, node.Status, node.Roles, node.Age, node.Version,
		node.InternalIP, node.ExternalIP, node.OSImage, node.KernelVersion, node.ContainerRuntime,
		node.CPUCapacity, node.MemoryCapacity, node.CPUAllocatable, node.MemoryAllocatable,
		node.Conditions, node.Labels, node.Annotations, node.CreatedAt, node.UpdatedAt)

	if err != nil {
		r.logger.Error("failed to create node", zap.Error(err), zap.String("name", node.Name))
		return fmt.Errorf("failed to create node: %w", err)
	}

	r.logger.Info("node created successfully", zap.String("id", node.ID), zap.String("name", node.Name))
	return nil
}

// CreateTx creates a new node within a transaction
func (r *nodeRepository) CreateTx(ctx context.Context, tx *sql.Tx, node *models.Node) error {
	node.ID = uuid.New().String()
	node.CreatedAt = time.Now()
	node.UpdatedAt = time.Now()

	query := `
		INSERT INTO nodes (id, cluster_id, name, status, roles, age, version, internal_ip, external_ip, 
			os_image, kernel_version, container_runtime, cpu_capacity, memory_capacity, 
			cpu_allocatable, memory_allocatable, conditions, labels, annotations, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)`

	_, err := tx.ExecContext(ctx, query,
		node.ID, node.ClusterID, node.Name, node.Status, node.Roles, node.Age, node.Version,
		node.InternalIP, node.ExternalIP, node.OSImage, node.KernelVersion, node.ContainerRuntime,
		node.CPUCapacity, node.MemoryCapacity, node.CPUAllocatable, node.MemoryAllocatable,
		node.Conditions, node.Labels, node.Annotations, node.CreatedAt, node.UpdatedAt)

	if err != nil {
		r.logger.Error("failed to create node in transaction", zap.Error(err), zap.String("name", node.Name))
		return fmt.Errorf("failed to create node: %w", err)
	}

	return nil
}

// GetByID retrieves a node by ID
func (r *nodeRepository) GetByID(ctx context.Context, id string) (*models.Node, error) {
	query := `
		SELECT id, cluster_id, name, status, roles, age, version, internal_ip, external_ip, 
			os_image, kernel_version, container_runtime, cpu_capacity, memory_capacity, 
			cpu_allocatable, memory_allocatable, conditions, labels, annotations, created_at, updated_at
		FROM nodes WHERE id = $1`

	var node models.Node
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&node.ID, &node.ClusterID, &node.Name, &node.Status, &node.Roles, &node.Age, &node.Version,
		&node.InternalIP, &node.ExternalIP, &node.OSImage, &node.KernelVersion, &node.ContainerRuntime,
		&node.CPUCapacity, &node.MemoryCapacity, &node.CPUAllocatable, &node.MemoryAllocatable,
		&node.Conditions, &node.Labels, &node.Annotations, &node.CreatedAt, &node.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("node not found: %s", id)
		}
		r.logger.Error("failed to get node by ID", zap.Error(err), zap.String("id", id))
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	return &node, nil
}

// GetByClusterAndName retrieves a node by cluster ID and name
func (r *nodeRepository) GetByClusterAndName(ctx context.Context, clusterID, name string) (*models.Node, error) {
	query := `
		SELECT id, cluster_id, name, status, roles, age, version, internal_ip, external_ip, 
			os_image, kernel_version, container_runtime, cpu_capacity, memory_capacity, 
			cpu_allocatable, memory_allocatable, conditions, labels, annotations, created_at, updated_at
		FROM nodes WHERE cluster_id = $1 AND name = $2`

	var node models.Node
	err := r.db.QueryRowContext(ctx, query, clusterID, name).Scan(
		&node.ID, &node.ClusterID, &node.Name, &node.Status, &node.Roles, &node.Age, &node.Version,
		&node.InternalIP, &node.ExternalIP, &node.OSImage, &node.KernelVersion, &node.ContainerRuntime,
		&node.CPUCapacity, &node.MemoryCapacity, &node.CPUAllocatable, &node.MemoryAllocatable,
		&node.Conditions, &node.Labels, &node.Annotations, &node.CreatedAt, &node.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("node not found: cluster=%s, name=%s", clusterID, name)
		}
		r.logger.Error("failed to get node by cluster and name", zap.Error(err),
			zap.String("cluster_id", clusterID), zap.String("name", name))
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	return &node, nil
}

// ListByCluster retrieves nodes by cluster ID with pagination
func (r *nodeRepository) ListByCluster(ctx context.Context, clusterID string, params models.PaginationParams) (*models.PaginatedResult, error) {
	// Count total records
	countQuery := "SELECT COUNT(*) FROM nodes WHERE cluster_id = $1"
	var totalCount int
	err := r.db.QueryRowContext(ctx, countQuery, clusterID).Scan(&totalCount)
	if err != nil {
		r.logger.Error("failed to count nodes", zap.Error(err))
		return nil, fmt.Errorf("failed to count nodes: %w", err)
	}

	// Get paginated results
	query := `
		SELECT id, cluster_id, name, status, roles, age, version, internal_ip, external_ip, 
			os_image, kernel_version, container_runtime, cpu_capacity, memory_capacity, 
			cpu_allocatable, memory_allocatable, conditions, labels, annotations, created_at, updated_at
		FROM nodes WHERE cluster_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, clusterID, params.PageSize, params.Offset())
	if err != nil {
		r.logger.Error("failed to query nodes", zap.Error(err))
		return nil, fmt.Errorf("failed to query nodes: %w", err)
	}
	defer rows.Close()

	var nodes []*models.Node
	for rows.Next() {
		var node models.Node
		err := rows.Scan(
			&node.ID, &node.ClusterID, &node.Name, &node.Status, &node.Roles, &node.Age, &node.Version,
			&node.InternalIP, &node.ExternalIP, &node.OSImage, &node.KernelVersion, &node.ContainerRuntime,
			&node.CPUCapacity, &node.MemoryCapacity, &node.CPUAllocatable, &node.MemoryAllocatable,
			&node.Conditions, &node.Labels, &node.Annotations, &node.CreatedAt, &node.UpdatedAt)
		if err != nil {
			r.logger.Error("failed to scan node", zap.Error(err))
			return nil, fmt.Errorf("failed to scan node: %w", err)
		}
		nodes = append(nodes, &node)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating node rows", zap.Error(err))
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return models.NewPaginatedResult(nodes, totalCount, params), nil
}

// Update updates a node
func (r *nodeRepository) Update(ctx context.Context, id string, updates models.NodeUpdateRequest) error {
	var setParts []string
	var args []interface{}
	argIndex := 1

	if updates.Status != nil {
		setParts = append(setParts, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *updates.Status)
		argIndex++
	}

	if updates.Roles != nil {
		setParts = append(setParts, fmt.Sprintf("roles = $%d", argIndex))
		args = append(args, *updates.Roles)
		argIndex++
	}

	if updates.Age != nil {
		setParts = append(setParts, fmt.Sprintf("age = $%d", argIndex))
		args = append(args, *updates.Age)
		argIndex++
	}

	if updates.Version != nil {
		setParts = append(setParts, fmt.Sprintf("version = $%d", argIndex))
		args = append(args, *updates.Version)
		argIndex++
	}

	if updates.InternalIP != nil {
		setParts = append(setParts, fmt.Sprintf("internal_ip = $%d", argIndex))
		args = append(args, *updates.InternalIP)
		argIndex++
	}

	if updates.ExternalIP != nil {
		setParts = append(setParts, fmt.Sprintf("external_ip = $%d", argIndex))
		args = append(args, *updates.ExternalIP)
		argIndex++
	}

	if updates.OSImage != nil {
		setParts = append(setParts, fmt.Sprintf("os_image = $%d", argIndex))
		args = append(args, *updates.OSImage)
		argIndex++
	}

	if updates.KernelVersion != nil {
		setParts = append(setParts, fmt.Sprintf("kernel_version = $%d", argIndex))
		args = append(args, *updates.KernelVersion)
		argIndex++
	}

	if updates.ContainerRuntime != nil {
		setParts = append(setParts, fmt.Sprintf("container_runtime = $%d", argIndex))
		args = append(args, *updates.ContainerRuntime)
		argIndex++
	}

	if updates.CPUCapacity != nil {
		setParts = append(setParts, fmt.Sprintf("cpu_capacity = $%d", argIndex))
		args = append(args, *updates.CPUCapacity)
		argIndex++
	}

	if updates.MemoryCapacity != nil {
		setParts = append(setParts, fmt.Sprintf("memory_capacity = $%d", argIndex))
		args = append(args, *updates.MemoryCapacity)
		argIndex++
	}

	if updates.CPUAllocatable != nil {
		setParts = append(setParts, fmt.Sprintf("cpu_allocatable = $%d", argIndex))
		args = append(args, *updates.CPUAllocatable)
		argIndex++
	}

	if updates.MemoryAllocatable != nil {
		setParts = append(setParts, fmt.Sprintf("memory_allocatable = $%d", argIndex))
		args = append(args, *updates.MemoryAllocatable)
		argIndex++
	}

	if updates.Conditions != nil {
		setParts = append(setParts, fmt.Sprintf("conditions = $%d", argIndex))
		args = append(args, *updates.Conditions)
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

	query := fmt.Sprintf("UPDATE nodes SET %s WHERE id = $%d",
		strings.Join(setParts, ", "), argIndex)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("failed to update node", zap.Error(err), zap.String("id", id))
		return fmt.Errorf("failed to update node: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("node not found: %s", id)
	}

	r.logger.Info("node updated successfully", zap.String("id", id))
	return nil
}

// UpdateTx updates a node within a transaction
func (r *nodeRepository) UpdateTx(ctx context.Context, tx *sql.Tx, id string, updates models.NodeUpdateRequest) error {
	// Similar implementation to Update but using tx instead of r.db
	// Implementation omitted for brevity - would be similar to Update method
	return fmt.Errorf("UpdateTx not implemented yet")
}

// Delete deletes a node
func (r *nodeRepository) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM nodes WHERE id = $1"

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("failed to delete node", zap.Error(err), zap.String("id", id))
		return fmt.Errorf("failed to delete node: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("node not found: %s", id)
	}

	r.logger.Info("node deleted successfully", zap.String("id", id))
	return nil
}

// DeleteTx deletes a node within a transaction
func (r *nodeRepository) DeleteTx(ctx context.Context, tx *sql.Tx, id string) error {
	query := "DELETE FROM nodes WHERE id = $1"

	result, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("failed to delete node in transaction", zap.Error(err), zap.String("id", id))
		return fmt.Errorf("failed to delete node: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("node not found: %s", id)
	}

	return nil
}

// DeleteByCluster deletes all nodes in a cluster
func (r *nodeRepository) DeleteByCluster(ctx context.Context, clusterID string) error {
	query := "DELETE FROM nodes WHERE cluster_id = $1"

	result, err := r.db.ExecContext(ctx, query, clusterID)
	if err != nil {
		r.logger.Error("failed to delete nodes by cluster", zap.Error(err), zap.String("cluster_id", clusterID))
		return fmt.Errorf("failed to delete nodes by cluster: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	r.logger.Info("nodes deleted by cluster", zap.String("cluster_id", clusterID), zap.Int64("count", rowsAffected))
	return nil
}

// Count counts nodes in a cluster
func (r *nodeRepository) Count(ctx context.Context, clusterID string) (int, error) {
	query := "SELECT COUNT(*) FROM nodes WHERE cluster_id = $1"

	var count int
	err := r.db.QueryRowContext(ctx, query, clusterID).Scan(&count)
	if err != nil {
		r.logger.Error("failed to count nodes", zap.Error(err))
		return 0, fmt.Errorf("failed to count nodes: %w", err)
	}

	return count, nil
}
