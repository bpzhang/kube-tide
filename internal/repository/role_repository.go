package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"kube-tide/internal/database/models"
	"kube-tide/internal/utils"

	"go.uber.org/zap"
)

// RoleRepository 角色仓储接口
type RoleRepository interface {
	Create(ctx context.Context, role *models.Role) error
	GetByID(ctx context.Context, id string) (*models.Role, error)
	GetByName(ctx context.Context, name string) (*models.Role, error)
	Update(ctx context.Context, id string, updates models.RoleUpdateRequest) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filters models.RoleListFilters, pagination utils.PaginationParams) (*utils.PaginatedResult[models.Role], error)
	GetRolePermissions(ctx context.Context, roleID string) ([]models.RolePermission, error)
	AssignPermissions(ctx context.Context, roleID string, permissionIDs []string) error
	RemovePermissions(ctx context.Context, roleID string, permissionIDs []string) error
	GetDefaultRoles(ctx context.Context) ([]models.Role, error)
	UserHasPermission(ctx context.Context, userID, permission, resource, scope string) (bool, error)
	GetUserPermissions(ctx context.Context, userID string) ([]*models.Permission, error)
}

// PermissionRepository 权限仓储接口
type PermissionRepository interface {
	GetByID(ctx context.Context, id string) (*models.Permission, error)
	GetByName(ctx context.Context, name string) (*models.Permission, error)
	List(ctx context.Context) ([]models.Permission, error)
	GetByResourceAndAction(ctx context.Context, resourceType, action string) (*models.Permission, error)
	GetUserPermissions(ctx context.Context, userID string, clusterName, namespace *string) ([]models.Permission, error)
}

// postgresRoleRepository PostgreSQL 角色仓储实现
type postgresRoleRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewRoleRepository 创建角色仓储
func NewRoleRepository(db *sql.DB, logger *zap.Logger) RoleRepository {
	return &postgresRoleRepository{
		db:     db,
		logger: logger,
	}
}

// Create 创建角色
func (r *postgresRoleRepository) Create(ctx context.Context, role *models.Role) error {
	query := `
		INSERT INTO roles (id, name, display_name, description, type, is_default, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.ExecContext(ctx, query,
		role.ID, role.Name, role.DisplayName, role.Description,
		role.Type, role.IsDefault, role.CreatedBy)

	if err != nil {
		r.logger.Error("Failed to create role", zap.Error(err), zap.String("name", role.Name))
		return fmt.Errorf("failed to create role: %w", err)
	}

	return nil
}

// GetByID 根据 ID 获取角色
func (r *postgresRoleRepository) GetByID(ctx context.Context, id string) (*models.Role, error) {
	query := `
		SELECT id, name, display_name, description, type, is_default,
		       created_at, updated_at, created_by
		FROM roles WHERE id = $1`

	role := &models.Role{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&role.ID, &role.Name, &role.DisplayName, &role.Description,
		&role.Type, &role.IsDefault, &role.CreatedAt, &role.UpdatedAt, &role.CreatedBy)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrNotFound
		}
		r.logger.Error("Failed to get role by ID", zap.Error(err), zap.String("id", id))
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	return role, nil
}

// GetByName 根据名称获取角色
func (r *postgresRoleRepository) GetByName(ctx context.Context, name string) (*models.Role, error) {
	query := `
		SELECT id, name, display_name, description, type, is_default,
		       created_at, updated_at, created_by
		FROM roles WHERE name = $1`

	role := &models.Role{}
	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&role.ID, &role.Name, &role.DisplayName, &role.Description,
		&role.Type, &role.IsDefault, &role.CreatedAt, &role.UpdatedAt, &role.CreatedBy)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrNotFound
		}
		r.logger.Error("Failed to get role by name", zap.Error(err), zap.String("name", name))
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	return role, nil
}

// Update 更新角色
func (r *postgresRoleRepository) Update(ctx context.Context, id string, updates models.RoleUpdateRequest) error {
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if updates.DisplayName != nil {
		setParts = append(setParts, fmt.Sprintf("display_name = $%d", argIndex))
		args = append(args, *updates.DisplayName)
		argIndex++
	}

	if updates.Description != nil {
		setParts = append(setParts, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, *updates.Description)
		argIndex++
	}

	if updates.IsDefault != nil {
		setParts = append(setParts, fmt.Sprintf("is_default = $%d", argIndex))
		args = append(args, *updates.IsDefault)
		argIndex++
	}

	if len(setParts) == 0 {
		return fmt.Errorf("no fields to update")
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	args = append(args, id)

	query := fmt.Sprintf("UPDATE roles SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIndex)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to update role", zap.Error(err), zap.String("id", id))
		return fmt.Errorf("failed to update role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return utils.ErrNotFound
	}

	return nil
}

// Delete 删除角色
func (r *postgresRoleRepository) Delete(ctx context.Context, id string) error {
	// 检查是否为系统角色
	var roleType string
	err := r.db.QueryRowContext(ctx, "SELECT type FROM roles WHERE id = $1", id).Scan(&roleType)
	if err != nil {
		if err == sql.ErrNoRows {
			return utils.ErrNotFound
		}
		return fmt.Errorf("failed to check role type: %w", err)
	}

	if roleType == "system" {
		return fmt.Errorf("cannot delete system role")
	}

	query := `DELETE FROM roles WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to delete role", zap.Error(err), zap.String("id", id))
		return fmt.Errorf("failed to delete role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return utils.ErrNotFound
	}

	return nil
}

// List 获取角色列表
func (r *postgresRoleRepository) List(ctx context.Context, filters models.RoleListFilters, pagination utils.PaginationParams) (*utils.PaginatedResult[models.Role], error) {
	whereParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if filters.Type != "" {
		whereParts = append(whereParts, fmt.Sprintf("type = $%d", argIndex))
		args = append(args, filters.Type)
		argIndex++
	}

	if filters.IsDefault != nil {
		whereParts = append(whereParts, fmt.Sprintf("is_default = $%d", argIndex))
		args = append(args, *filters.IsDefault)
		argIndex++
	}

	whereClause := ""
	if len(whereParts) > 0 {
		whereClause = "WHERE " + strings.Join(whereParts, " AND ")
	}

	// 获取总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM roles %s", whereClause)
	var totalCount int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		r.logger.Error("Failed to count roles", zap.Error(err))
		return nil, fmt.Errorf("failed to count roles: %w", err)
	}

	// 获取数据
	offset := (pagination.Page - 1) * pagination.PageSize
	dataQuery := fmt.Sprintf(`
		SELECT id, name, display_name, description, type, is_default,
		       created_at, updated_at, created_by
		FROM roles %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, pagination.PageSize, offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		r.logger.Error("Failed to list roles", zap.Error(err))
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}
	defer rows.Close()

	roles := []models.Role{}
	for rows.Next() {
		var role models.Role
		err := rows.Scan(
			&role.ID, &role.Name, &role.DisplayName, &role.Description,
			&role.Type, &role.IsDefault, &role.CreatedAt, &role.UpdatedAt, &role.CreatedBy)
		if err != nil {
			r.logger.Error("Failed to scan role", zap.Error(err))
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}
		roles = append(roles, role)
	}

	return utils.NewPaginatedResult(roles, totalCount, pagination.Page, pagination.PageSize), nil
}

// GetRolePermissions 获取角色权限
func (r *postgresRoleRepository) GetRolePermissions(ctx context.Context, roleID string) ([]models.RolePermission, error) {
	query := `
		SELECT rp.id, rp.role_id, rp.permission_id, rp.created_at,
		       p.id, p.name, p.display_name, p.description, p.resource_type, p.action, p.scope
		FROM role_permissions rp
		JOIN permissions p ON rp.permission_id = p.id
		WHERE rp.role_id = $1
		ORDER BY p.resource_type, p.action`

	rows, err := r.db.QueryContext(ctx, query, roleID)
	if err != nil {
		r.logger.Error("Failed to get role permissions", zap.Error(err), zap.String("roleID", roleID))
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}
	defer rows.Close()

	rolePermissions := []models.RolePermission{}
	for rows.Next() {
		var rp models.RolePermission
		var permission models.Permission

		err := rows.Scan(
			&rp.ID, &rp.RoleID, &rp.PermissionID, &rp.CreatedAt,
			&permission.ID, &permission.Name, &permission.DisplayName,
			&permission.Description, &permission.ResourceType, &permission.Action, &permission.Scope)
		if err != nil {
			r.logger.Error("Failed to scan role permission", zap.Error(err))
			return nil, fmt.Errorf("failed to scan role permission: %w", err)
		}

		rp.Permission = &permission
		rolePermissions = append(rolePermissions, rp)
	}

	return rolePermissions, nil
}

// AssignPermissions 分配权限
func (r *postgresRoleRepository) AssignPermissions(ctx context.Context, roleID string, permissionIDs []string) error {
	if len(permissionIDs) == 0 {
		return nil
	}

	// 使用事务
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 删除现有权限
	_, err = tx.ExecContext(ctx, "DELETE FROM role_permissions WHERE role_id = $1", roleID)
	if err != nil {
		return fmt.Errorf("failed to remove existing permissions: %w", err)
	}

	// 添加新权限
	for _, permissionID := range permissionIDs {
		_, err = tx.ExecContext(ctx,
			"INSERT INTO role_permissions (id, role_id, permission_id) VALUES (gen_random_uuid(), $1, $2)",
			roleID, permissionID)
		if err != nil {
			r.logger.Error("Failed to assign permission", zap.Error(err),
				zap.String("roleID", roleID), zap.String("permissionID", permissionID))
			return fmt.Errorf("failed to assign permission: %w", err)
		}
	}

	return tx.Commit()
}

// RemovePermissions 移除权限
func (r *postgresRoleRepository) RemovePermissions(ctx context.Context, roleID string, permissionIDs []string) error {
	if len(permissionIDs) == 0 {
		return nil
	}

	placeholders := make([]string, len(permissionIDs))
	args := []interface{}{roleID}
	for i, permissionID := range permissionIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args = append(args, permissionID)
	}

	query := fmt.Sprintf("DELETE FROM role_permissions WHERE role_id = $1 AND permission_id IN (%s)",
		strings.Join(placeholders, ","))

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to remove permissions", zap.Error(err), zap.String("roleID", roleID))
		return fmt.Errorf("failed to remove permissions: %w", err)
	}

	return nil
}

// GetDefaultRoles 获取默认角色
func (r *postgresRoleRepository) GetDefaultRoles(ctx context.Context) ([]models.Role, error) {
	query := `
		SELECT id, name, display_name, description, type, is_default,
		       created_at, updated_at, created_by
		FROM roles WHERE is_default = true
		ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		r.logger.Error("Failed to get default roles", zap.Error(err))
		return nil, fmt.Errorf("failed to get default roles: %w", err)
	}
	defer rows.Close()

	roles := []models.Role{}
	for rows.Next() {
		var role models.Role
		err := rows.Scan(
			&role.ID, &role.Name, &role.DisplayName, &role.Description,
			&role.Type, &role.IsDefault, &role.CreatedAt, &role.UpdatedAt, &role.CreatedBy)
		if err != nil {
			r.logger.Error("Failed to scan role", zap.Error(err))
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// UserHasPermission 检查用户是否有特定权限
func (r *postgresRoleRepository) UserHasPermission(ctx context.Context, userID, permission, resource, scope string) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = $1 
		  AND p.name = $2
		  AND p.resource_type = $3
		  AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
		  AND (
		    ur.scope_type = 'global' OR
		    (ur.scope_type = 'cluster' AND ur.scope_value = $4) OR
		    (ur.scope_type = 'namespace' AND ur.scope_value = $4)
		  )`

	var count int
	err := r.db.QueryRowContext(ctx, query, userID, permission, resource, scope).Scan(&count)
	if err != nil {
		r.logger.Error("Failed to check user permission", zap.Error(err),
			zap.String("userID", userID), zap.String("permission", permission))
		return false, fmt.Errorf("failed to check user permission: %w", err)
	}

	return count > 0, nil
}

// GetUserPermissions 获取用户所有权限
func (r *postgresRoleRepository) GetUserPermissions(ctx context.Context, userID string) ([]*models.Permission, error) {
	query := `
		SELECT DISTINCT p.id, p.name, p.display_name, p.description, p.resource_type, p.action, p.scope
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = $1
		  AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
		ORDER BY p.resource_type, p.action`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		r.logger.Error("Failed to get user permissions", zap.Error(err), zap.String("userID", userID))
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}
	defer rows.Close()

	var permissions []*models.Permission
	for rows.Next() {
		permission := &models.Permission{}
		err := rows.Scan(
			&permission.ID, &permission.Name, &permission.DisplayName,
			&permission.Description, &permission.ResourceType, &permission.Action, &permission.Scope)
		if err != nil {
			r.logger.Error("Failed to scan permission", zap.Error(err))
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		permissions = append(permissions, permission)
	}

	return permissions, nil
}

// postgresPermissionRepository PostgreSQL 权限仓储实现
type postgresPermissionRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewPermissionRepository 创建权限仓储
func NewPermissionRepository(db *sql.DB, logger *zap.Logger) PermissionRepository {
	return &postgresPermissionRepository{
		db:     db,
		logger: logger,
	}
}

// GetByID 根据 ID 获取权限
func (r *postgresPermissionRepository) GetByID(ctx context.Context, id string) (*models.Permission, error) {
	query := `
		SELECT id, name, display_name, description, resource_type, action, scope
		FROM permissions WHERE id = $1`

	permission := &models.Permission{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&permission.ID, &permission.Name, &permission.DisplayName,
		&permission.Description, &permission.ResourceType, &permission.Action, &permission.Scope)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrNotFound
		}
		r.logger.Error("Failed to get permission by ID", zap.Error(err), zap.String("id", id))
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	return permission, nil
}

// GetByName 根据名称获取权限
func (r *postgresPermissionRepository) GetByName(ctx context.Context, name string) (*models.Permission, error) {
	query := `
		SELECT id, name, display_name, description, resource_type, action, scope
		FROM permissions WHERE name = $1`

	permission := &models.Permission{}
	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&permission.ID, &permission.Name, &permission.DisplayName,
		&permission.Description, &permission.ResourceType, &permission.Action, &permission.Scope)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrNotFound
		}
		r.logger.Error("Failed to get permission by name", zap.Error(err), zap.String("name", name))
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	return permission, nil
}

// List 获取所有权限
func (r *postgresPermissionRepository) List(ctx context.Context) ([]models.Permission, error) {
	query := `
		SELECT id, name, display_name, description, resource_type, action, scope
		FROM permissions
		ORDER BY resource_type, action`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		r.logger.Error("Failed to list permissions", zap.Error(err))
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}
	defer rows.Close()

	permissions := []models.Permission{}
	for rows.Next() {
		var permission models.Permission
		err := rows.Scan(
			&permission.ID, &permission.Name, &permission.DisplayName,
			&permission.Description, &permission.ResourceType, &permission.Action, &permission.Scope)
		if err != nil {
			r.logger.Error("Failed to scan permission", zap.Error(err))
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		permissions = append(permissions, permission)
	}

	return permissions, nil
}

// GetByResourceAndAction 根据资源类型和操作获取权限
func (r *postgresPermissionRepository) GetByResourceAndAction(ctx context.Context, resourceType, action string) (*models.Permission, error) {
	query := `
		SELECT id, name, display_name, description, resource_type, action, scope
		FROM permissions WHERE resource_type = $1 AND action = $2`

	permission := &models.Permission{}
	err := r.db.QueryRowContext(ctx, query, resourceType, action).Scan(
		&permission.ID, &permission.Name, &permission.DisplayName,
		&permission.Description, &permission.ResourceType, &permission.Action, &permission.Scope)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrNotFound
		}
		r.logger.Error("Failed to get permission by resource and action", zap.Error(err),
			zap.String("resourceType", resourceType), zap.String("action", action))
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	return permission, nil
}

// GetUserPermissions 获取用户权限
func (r *postgresPermissionRepository) GetUserPermissions(ctx context.Context, userID string, clusterName, namespace *string) ([]models.Permission, error) {
	whereParts := []string{
		"ur.user_id = $1",
		"(ur.expires_at IS NULL OR ur.expires_at > NOW())",
	}
	args := []interface{}{userID}
	argIndex := 2

	// 根据作用域过滤
	if clusterName != nil && namespace != nil {
		// 命名空间级别：包括 global、cluster 和 namespace 权限
		whereParts = append(whereParts, fmt.Sprintf(`(
			ur.scope_type = 'global' OR 
			(ur.scope_type = 'cluster' AND ur.scope_value = $%d) OR
			(ur.scope_type = 'namespace' AND ur.scope_value = $%d)
		)`, argIndex, argIndex+1))
		args = append(args, *clusterName, fmt.Sprintf("%s/%s", *clusterName, *namespace))
		argIndex += 2
	} else if clusterName != nil {
		// 集群级别：包括 global 和 cluster 权限
		whereParts = append(whereParts, fmt.Sprintf(`(
			ur.scope_type = 'global' OR 
			(ur.scope_type = 'cluster' AND ur.scope_value = $%d)
		)`, argIndex))
		args = append(args, *clusterName)
		argIndex++
	} else {
		// 全局级别：只包括 global 权限
		whereParts = append(whereParts, "ur.scope_type = 'global'")
	}

	whereClause := strings.Join(whereParts, " AND ")

	query := fmt.Sprintf(`
		SELECT DISTINCT p.id, p.name, p.display_name, p.description, p.resource_type, p.action, p.scope
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE %s
		ORDER BY p.resource_type, p.action`, whereClause)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to get user permissions", zap.Error(err), zap.String("userID", userID))
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}
	defer rows.Close()

	permissions := []models.Permission{}
	for rows.Next() {
		var permission models.Permission
		err := rows.Scan(
			&permission.ID, &permission.Name, &permission.DisplayName,
			&permission.Description, &permission.ResourceType, &permission.Action, &permission.Scope)
		if err != nil {
			r.logger.Error("Failed to scan permission", zap.Error(err))
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		permissions = append(permissions, permission)
	}

	return permissions, nil
}
