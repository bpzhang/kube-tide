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

// serviceRepository implements ServiceRepository interface
type serviceRepository struct {
	db     *database.Database
	logger *zap.Logger
}

// NewServiceRepository creates a new service repository
func NewServiceRepository(db *database.Database, logger *zap.Logger) ServiceRepository {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &serviceRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new service
func (r *serviceRepository) Create(ctx context.Context, service *models.Service) error {
	service.ID = uuid.New().String()
	service.CreatedAt = time.Now()
	service.UpdatedAt = time.Now()

	query := `
		INSERT INTO services (id, cluster_id, namespace, name, type, cluster_ip, external_ips,
			ports, selector, session_affinity, labels, annotations, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

	_, err := r.db.ExecContext(ctx, query,
		service.ID, service.ClusterID, service.Namespace, service.Name, service.Type,
		service.ClusterIP, service.ExternalIPs, service.Ports, service.Selector,
		service.SessionAffinity, service.Labels, service.Annotations,
		service.CreatedAt, service.UpdatedAt)

	if err != nil {
		r.logger.Error("failed to create service", zap.Error(err), zap.String("name", service.Name))
		return fmt.Errorf("failed to create service: %w", err)
	}

	r.logger.Info("service created successfully", zap.String("id", service.ID), zap.String("name", service.Name))
	return nil
}

// CreateTx creates a new service within a transaction
func (r *serviceRepository) CreateTx(ctx context.Context, tx *sql.Tx, service *models.Service) error {
	service.ID = uuid.New().String()
	service.CreatedAt = time.Now()
	service.UpdatedAt = time.Now()

	query := `
		INSERT INTO services (id, cluster_id, namespace, name, type, cluster_ip, external_ips,
			ports, selector, session_affinity, labels, annotations, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

	_, err := tx.ExecContext(ctx, query,
		service.ID, service.ClusterID, service.Namespace, service.Name, service.Type,
		service.ClusterIP, service.ExternalIPs, service.Ports, service.Selector,
		service.SessionAffinity, service.Labels, service.Annotations,
		service.CreatedAt, service.UpdatedAt)

	if err != nil {
		r.logger.Error("failed to create service in transaction", zap.Error(err), zap.String("name", service.Name))
		return fmt.Errorf("failed to create service: %w", err)
	}

	return nil
}

// GetByID retrieves a service by ID
func (r *serviceRepository) GetByID(ctx context.Context, id string) (*models.Service, error) {
	query := `
		SELECT id, cluster_id, namespace, name, type, cluster_ip, external_ips,
			ports, selector, session_affinity, labels, annotations, created_at, updated_at
		FROM services WHERE id = $1`

	var service models.Service
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&service.ID, &service.ClusterID, &service.Namespace, &service.Name, &service.Type,
		&service.ClusterIP, &service.ExternalIPs, &service.Ports, &service.Selector,
		&service.SessionAffinity, &service.Labels, &service.Annotations,
		&service.CreatedAt, &service.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("service not found: %s", id)
		}
		r.logger.Error("failed to get service by ID", zap.Error(err), zap.String("id", id))
		return nil, fmt.Errorf("failed to get service: %w", err)
	}

	return &service, nil
}

// GetByClusterNamespaceAndName retrieves a service by cluster ID, namespace, and name
func (r *serviceRepository) GetByClusterNamespaceAndName(ctx context.Context, clusterID, namespace, name string) (*models.Service, error) {
	query := `
		SELECT id, cluster_id, namespace, name, type, cluster_ip, external_ips,
			ports, selector, session_affinity, labels, annotations, created_at, updated_at
		FROM services WHERE cluster_id = $1 AND namespace = $2 AND name = $3`

	var service models.Service
	err := r.db.QueryRowContext(ctx, query, clusterID, namespace, name).Scan(
		&service.ID, &service.ClusterID, &service.Namespace, &service.Name, &service.Type,
		&service.ClusterIP, &service.ExternalIPs, &service.Ports, &service.Selector,
		&service.SessionAffinity, &service.Labels, &service.Annotations,
		&service.CreatedAt, &service.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("service not found: cluster=%s, namespace=%s, name=%s", clusterID, namespace, name)
		}
		r.logger.Error("failed to get service by cluster, namespace and name", zap.Error(err),
			zap.String("cluster_id", clusterID), zap.String("namespace", namespace), zap.String("name", name))
		return nil, fmt.Errorf("failed to get service: %w", err)
	}

	return &service, nil
}

// ListByCluster retrieves services by cluster ID with pagination
func (r *serviceRepository) ListByCluster(ctx context.Context, clusterID string, params models.PaginationParams) (*models.PaginatedResult, error) {
	// Count total records
	countQuery := "SELECT COUNT(*) FROM services WHERE cluster_id = $1"
	var totalCount int
	err := r.db.QueryRowContext(ctx, countQuery, clusterID).Scan(&totalCount)
	if err != nil {
		r.logger.Error("failed to count services", zap.Error(err))
		return nil, fmt.Errorf("failed to count services: %w", err)
	}

	// Get paginated results
	query := `
		SELECT id, cluster_id, namespace, name, type, cluster_ip, external_ips,
			ports, selector, session_affinity, labels, annotations, created_at, updated_at
		FROM services WHERE cluster_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, clusterID, params.PageSize, params.Offset())
	if err != nil {
		r.logger.Error("failed to query services", zap.Error(err))
		return nil, fmt.Errorf("failed to query services: %w", err)
	}
	defer rows.Close()

	var services []*models.Service
	for rows.Next() {
		var service models.Service
		err := rows.Scan(
			&service.ID, &service.ClusterID, &service.Namespace, &service.Name, &service.Type,
			&service.ClusterIP, &service.ExternalIPs, &service.Ports, &service.Selector,
			&service.SessionAffinity, &service.Labels, &service.Annotations,
			&service.CreatedAt, &service.UpdatedAt)
		if err != nil {
			r.logger.Error("failed to scan service", zap.Error(err))
			return nil, fmt.Errorf("failed to scan service: %w", err)
		}
		services = append(services, &service)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating service rows", zap.Error(err))
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return models.NewPaginatedResult(services, totalCount, params), nil
}

// ListByNamespace retrieves services by cluster ID and namespace with pagination
func (r *serviceRepository) ListByNamespace(ctx context.Context, clusterID, namespace string, params models.PaginationParams) (*models.PaginatedResult, error) {
	// Count total records
	countQuery := "SELECT COUNT(*) FROM services WHERE cluster_id = $1 AND namespace = $2"
	var totalCount int
	err := r.db.QueryRowContext(ctx, countQuery, clusterID, namespace).Scan(&totalCount)
	if err != nil {
		r.logger.Error("failed to count services", zap.Error(err))
		return nil, fmt.Errorf("failed to count services: %w", err)
	}

	// Get paginated results
	query := `
		SELECT id, cluster_id, namespace, name, type, cluster_ip, external_ips,
			ports, selector, session_affinity, labels, annotations, created_at, updated_at
		FROM services WHERE cluster_id = $1 AND namespace = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`

	rows, err := r.db.QueryContext(ctx, query, clusterID, namespace, params.PageSize, params.Offset())
	if err != nil {
		r.logger.Error("failed to query services", zap.Error(err))
		return nil, fmt.Errorf("failed to query services: %w", err)
	}
	defer rows.Close()

	var services []*models.Service
	for rows.Next() {
		var service models.Service
		err := rows.Scan(
			&service.ID, &service.ClusterID, &service.Namespace, &service.Name, &service.Type,
			&service.ClusterIP, &service.ExternalIPs, &service.Ports, &service.Selector,
			&service.SessionAffinity, &service.Labels, &service.Annotations,
			&service.CreatedAt, &service.UpdatedAt)
		if err != nil {
			r.logger.Error("failed to scan service", zap.Error(err))
			return nil, fmt.Errorf("failed to scan service: %w", err)
		}
		services = append(services, &service)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating service rows", zap.Error(err))
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return models.NewPaginatedResult(services, totalCount, params), nil
}

// Update updates a service
func (r *serviceRepository) Update(ctx context.Context, id string, updates models.ServiceUpdateRequest) error {
	var setParts []string
	var args []interface{}
	argIndex := 1

	if updates.Type != nil {
		setParts = append(setParts, fmt.Sprintf("type = $%d", argIndex))
		args = append(args, *updates.Type)
		argIndex++
	}

	if updates.ClusterIP != nil {
		setParts = append(setParts, fmt.Sprintf("cluster_ip = $%d", argIndex))
		args = append(args, *updates.ClusterIP)
		argIndex++
	}

	if updates.ExternalIPs != nil {
		setParts = append(setParts, fmt.Sprintf("external_ips = $%d", argIndex))
		args = append(args, *updates.ExternalIPs)
		argIndex++
	}

	if updates.Ports != nil {
		setParts = append(setParts, fmt.Sprintf("ports = $%d", argIndex))
		args = append(args, *updates.Ports)
		argIndex++
	}

	if updates.Selector != nil {
		setParts = append(setParts, fmt.Sprintf("selector = $%d", argIndex))
		args = append(args, *updates.Selector)
		argIndex++
	}

	if updates.SessionAffinity != nil {
		setParts = append(setParts, fmt.Sprintf("session_affinity = $%d", argIndex))
		args = append(args, *updates.SessionAffinity)
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

	query := fmt.Sprintf("UPDATE services SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIndex)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("failed to update service", zap.Error(err), zap.String("id", id))
		return fmt.Errorf("failed to update service: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("service not found: %s", id)
	}

	r.logger.Info("service updated successfully", zap.String("id", id))
	return nil
}

// UpdateTx updates a service within a transaction
func (r *serviceRepository) UpdateTx(ctx context.Context, tx *sql.Tx, id string, updates models.ServiceUpdateRequest) error {
	var setParts []string
	var args []interface{}
	argIndex := 1

	if updates.Type != nil {
		setParts = append(setParts, fmt.Sprintf("type = $%d", argIndex))
		args = append(args, *updates.Type)
		argIndex++
	}

	if updates.ClusterIP != nil {
		setParts = append(setParts, fmt.Sprintf("cluster_ip = $%d", argIndex))
		args = append(args, *updates.ClusterIP)
		argIndex++
	}

	if updates.ExternalIPs != nil {
		setParts = append(setParts, fmt.Sprintf("external_ips = $%d", argIndex))
		args = append(args, *updates.ExternalIPs)
		argIndex++
	}

	if updates.Ports != nil {
		setParts = append(setParts, fmt.Sprintf("ports = $%d", argIndex))
		args = append(args, *updates.Ports)
		argIndex++
	}

	if updates.Selector != nil {
		setParts = append(setParts, fmt.Sprintf("selector = $%d", argIndex))
		args = append(args, *updates.Selector)
		argIndex++
	}

	if updates.SessionAffinity != nil {
		setParts = append(setParts, fmt.Sprintf("session_affinity = $%d", argIndex))
		args = append(args, *updates.SessionAffinity)
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

	query := fmt.Sprintf("UPDATE services SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIndex)

	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("failed to update service in transaction", zap.Error(err), zap.String("id", id))
		return fmt.Errorf("failed to update service: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("service not found: %s", id)
	}

	return nil
}

// Delete deletes a service
func (r *serviceRepository) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM services WHERE id = $1"

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("failed to delete service", zap.Error(err), zap.String("id", id))
		return fmt.Errorf("failed to delete service: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("service not found: %s", id)
	}

	r.logger.Info("service deleted successfully", zap.String("id", id))
	return nil
}

// DeleteTx deletes a service within a transaction
func (r *serviceRepository) DeleteTx(ctx context.Context, tx *sql.Tx, id string) error {
	query := "DELETE FROM services WHERE id = $1"

	result, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("failed to delete service in transaction", zap.Error(err), zap.String("id", id))
		return fmt.Errorf("failed to delete service: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("service not found: %s", id)
	}

	return nil
}

// DeleteByCluster deletes all services in a cluster
func (r *serviceRepository) DeleteByCluster(ctx context.Context, clusterID string) error {
	query := "DELETE FROM services WHERE cluster_id = $1"

	result, err := r.db.ExecContext(ctx, query, clusterID)
	if err != nil {
		r.logger.Error("failed to delete services by cluster", zap.Error(err), zap.String("cluster_id", clusterID))
		return fmt.Errorf("failed to delete services by cluster: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	r.logger.Info("services deleted by cluster", zap.String("cluster_id", clusterID), zap.Int64("count", rowsAffected))
	return nil
}

// DeleteByNamespace deletes all services in a namespace
func (r *serviceRepository) DeleteByNamespace(ctx context.Context, clusterID, namespace string) error {
	query := "DELETE FROM services WHERE cluster_id = $1 AND namespace = $2"

	result, err := r.db.ExecContext(ctx, query, clusterID, namespace)
	if err != nil {
		r.logger.Error("failed to delete services by namespace", zap.Error(err),
			zap.String("cluster_id", clusterID), zap.String("namespace", namespace))
		return fmt.Errorf("failed to delete services by namespace: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	r.logger.Info("services deleted by namespace",
		zap.String("cluster_id", clusterID), zap.String("namespace", namespace), zap.Int64("count", rowsAffected))
	return nil
}

// Count counts services in a cluster
func (r *serviceRepository) Count(ctx context.Context, clusterID string) (int, error) {
	query := "SELECT COUNT(*) FROM services WHERE cluster_id = $1"

	var count int
	err := r.db.QueryRowContext(ctx, query, clusterID).Scan(&count)
	if err != nil {
		r.logger.Error("failed to count services", zap.Error(err))
		return 0, fmt.Errorf("failed to count services: %w", err)
	}

	return count, nil
}

// CountByNamespace counts services in a namespace
func (r *serviceRepository) CountByNamespace(ctx context.Context, clusterID, namespace string) (int, error) {
	query := "SELECT COUNT(*) FROM services WHERE cluster_id = $1 AND namespace = $2"

	var count int
	err := r.db.QueryRowContext(ctx, query, clusterID, namespace).Scan(&count)
	if err != nil {
		r.logger.Error("failed to count services by namespace", zap.Error(err))
		return 0, fmt.Errorf("failed to count services by namespace: %w", err)
	}

	return count, nil
}
