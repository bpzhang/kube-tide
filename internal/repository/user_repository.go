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

// UserRepository 用户仓储接口
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, id string, updates models.UserUpdateRequest) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filters models.UserListFilters, pagination utils.PaginationParams) (*utils.PaginatedResult[models.User], error)
	UpdatePassword(ctx context.Context, id string, passwordHash string) error
	UpdateLastLogin(ctx context.Context, id string) error
	GetUserRoles(ctx context.Context, userID string, filters models.UserRoleListFilters) ([]models.UserRole, error)
	AssignRole(ctx context.Context, userRole *models.UserRole) error
	RemoveRole(ctx context.Context, userID, roleID, scopeType string, scopeValue *string) error
}

// postgresUserRepository PostgreSQL 用户仓储实现
type postgresUserRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewUserRepository 创建用户仓储
func NewUserRepository(db *sql.DB, logger *zap.Logger) UserRepository {
	return &postgresUserRepository{
		db:     db,
		logger: logger,
	}
}

// Create 创建用户
func (r *postgresUserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, username, email, password_hash, display_name, avatar_url, status, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.Username, user.Email, user.PasswordHash,
		user.DisplayName, user.AvatarURL, user.Status, user.CreatedBy)

	if err != nil {
		r.logger.Error("Failed to create user", zap.Error(err), zap.String("username", user.Username))
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID 根据 ID 获取用户
func (r *postgresUserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, display_name, avatar_url, status,
		       last_login_at, created_at, updated_at, created_by
		FROM users WHERE id = $1`

	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.DisplayName, &user.AvatarURL, &user.Status,
		&user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt, &user.CreatedBy)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrNotFound
		}
		r.logger.Error("Failed to get user by ID", zap.Error(err), zap.String("id", id))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetByUsername 根据用户名获取用户
func (r *postgresUserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, display_name, avatar_url, status,
		       last_login_at, created_at, updated_at, created_by
		FROM users WHERE username = $1`

	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.DisplayName, &user.AvatarURL, &user.Status,
		&user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt, &user.CreatedBy)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrNotFound
		}
		r.logger.Error("Failed to get user by username", zap.Error(err), zap.String("username", username))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetByEmail 根据邮箱获取用户
func (r *postgresUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, display_name, avatar_url, status,
		       last_login_at, created_at, updated_at, created_by
		FROM users WHERE email = $1`

	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.DisplayName, &user.AvatarURL, &user.Status,
		&user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt, &user.CreatedBy)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrNotFound
		}
		r.logger.Error("Failed to get user by email", zap.Error(err), zap.String("email", email))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// Update 更新用户
func (r *postgresUserRepository) Update(ctx context.Context, id string, updates models.UserUpdateRequest) error {
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if updates.Email != nil {
		setParts = append(setParts, fmt.Sprintf("email = $%d", argIndex))
		args = append(args, *updates.Email)
		argIndex++
	}

	if updates.DisplayName != nil {
		setParts = append(setParts, fmt.Sprintf("display_name = $%d", argIndex))
		args = append(args, *updates.DisplayName)
		argIndex++
	}

	if updates.AvatarURL != nil {
		setParts = append(setParts, fmt.Sprintf("avatar_url = $%d", argIndex))
		args = append(args, *updates.AvatarURL)
		argIndex++
	}

	if updates.Status != nil {
		setParts = append(setParts, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *updates.Status)
		argIndex++
	}

	if len(setParts) == 0 {
		return fmt.Errorf("no fields to update")
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	args = append(args, id)

	query := fmt.Sprintf("UPDATE users SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIndex)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to update user", zap.Error(err), zap.String("id", id))
		return fmt.Errorf("failed to update user: %w", err)
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

// Delete 删除用户
func (r *postgresUserRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to delete user", zap.Error(err), zap.String("id", id))
		return fmt.Errorf("failed to delete user: %w", err)
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

// List 获取用户列表
func (r *postgresUserRepository) List(ctx context.Context, filters models.UserListFilters, pagination utils.PaginationParams) (*utils.PaginatedResult[models.User], error) {
	whereParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if filters.Status != "" {
		whereParts = append(whereParts, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, filters.Status)
		argIndex++
	}

	if filters.Username != "" {
		whereParts = append(whereParts, fmt.Sprintf("username ILIKE $%d", argIndex))
		args = append(args, "%"+filters.Username+"%")
		argIndex++
	}

	if filters.Email != "" {
		whereParts = append(whereParts, fmt.Sprintf("email ILIKE $%d", argIndex))
		args = append(args, "%"+filters.Email+"%")
		argIndex++
	}

	whereClause := ""
	if len(whereParts) > 0 {
		whereClause = "WHERE " + strings.Join(whereParts, " AND ")
	}

	// 获取总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users %s", whereClause)
	var totalCount int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		r.logger.Error("Failed to count users", zap.Error(err))
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	// 获取数据
	offset := (pagination.Page - 1) * pagination.PageSize
	dataQuery := fmt.Sprintf(`
		SELECT id, username, email, display_name, avatar_url, status,
		       last_login_at, created_at, updated_at, created_by
		FROM users %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, pagination.PageSize, offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		r.logger.Error("Failed to list users", zap.Error(err))
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	users := []models.User{}
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email,
			&user.DisplayName, &user.AvatarURL, &user.Status,
			&user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt, &user.CreatedBy)
		if err != nil {
			r.logger.Error("Failed to scan user", zap.Error(err))
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return utils.NewPaginatedResult(users, totalCount, pagination.Page, pagination.PageSize), nil
}

// UpdatePassword 更新密码
func (r *postgresUserRepository) UpdatePassword(ctx context.Context, id string, passwordHash string) error {
	query := `UPDATE users SET password_hash = $1, updated_at = $2 WHERE id = $3`

	result, err := r.db.ExecContext(ctx, query, passwordHash, time.Now(), id)
	if err != nil {
		r.logger.Error("Failed to update password", zap.Error(err), zap.String("id", id))
		return fmt.Errorf("failed to update password: %w", err)
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

// UpdateLastLogin 更新最后登录时间
func (r *postgresUserRepository) UpdateLastLogin(ctx context.Context, id string) error {
	query := `UPDATE users SET last_login_at = $1 WHERE id = $2`

	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		r.logger.Error("Failed to update last login", zap.Error(err), zap.String("id", id))
		return fmt.Errorf("failed to update last login: %w", err)
	}

	return nil
}

// GetUserRoles 获取用户角色
func (r *postgresUserRepository) GetUserRoles(ctx context.Context, userID string, filters models.UserRoleListFilters) ([]models.UserRole, error) {
	whereParts := []string{"ur.user_id = $1"}
	args := []interface{}{userID}
	argIndex := 2

	if filters.ScopeType != "" {
		whereParts = append(whereParts, fmt.Sprintf("ur.scope_type = $%d", argIndex))
		args = append(args, filters.ScopeType)
		argIndex++
	}

	if filters.ScopeValue != "" {
		whereParts = append(whereParts, fmt.Sprintf("ur.scope_value = $%d", argIndex))
		args = append(args, filters.ScopeValue)
		argIndex++
	}

	// 添加过期时间检查
	whereParts = append(whereParts, "(ur.expires_at IS NULL OR ur.expires_at > NOW())")

	whereClause := strings.Join(whereParts, " AND ")

	query := fmt.Sprintf(`
		SELECT ur.id, ur.user_id, ur.role_id, ur.scope_type, ur.scope_value,
		       ur.granted_at, ur.granted_by, ur.expires_at,
		       r.id, r.name, r.display_name, r.description, r.type, r.is_default,
		       r.created_at, r.updated_at, r.created_by
		FROM user_roles ur
		JOIN roles r ON ur.role_id = r.id
		WHERE %s
		ORDER BY ur.granted_at DESC`, whereClause)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to get user roles", zap.Error(err), zap.String("userID", userID))
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	defer rows.Close()

	userRoles := []models.UserRole{}
	for rows.Next() {
		var userRole models.UserRole
		var role models.Role

		err := rows.Scan(
			&userRole.ID, &userRole.UserID, &userRole.RoleID,
			&userRole.ScopeType, &userRole.ScopeValue,
			&userRole.GrantedAt, &userRole.GrantedBy, &userRole.ExpiresAt,
			&role.ID, &role.Name, &role.DisplayName, &role.Description,
			&role.Type, &role.IsDefault, &role.CreatedAt, &role.UpdatedAt, &role.CreatedBy)
		if err != nil {
			r.logger.Error("Failed to scan user role", zap.Error(err))
			return nil, fmt.Errorf("failed to scan user role: %w", err)
		}

		userRole.Role = &role
		userRoles = append(userRoles, userRole)
	}

	return userRoles, nil
}

// AssignRole 分配角色
func (r *postgresUserRepository) AssignRole(ctx context.Context, userRole *models.UserRole) error {
	query := `
		INSERT INTO user_roles (id, user_id, role_id, scope_type, scope_value, granted_by, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id, role_id, scope_type, scope_value) 
		DO UPDATE SET granted_at = NOW(), granted_by = $6, expires_at = $7`

	_, err := r.db.ExecContext(ctx, query,
		userRole.ID, userRole.UserID, userRole.RoleID,
		userRole.ScopeType, userRole.ScopeValue,
		userRole.GrantedBy, userRole.ExpiresAt)

	if err != nil {
		r.logger.Error("Failed to assign role", zap.Error(err),
			zap.String("userID", userRole.UserID), zap.String("roleID", userRole.RoleID))
		return fmt.Errorf("failed to assign role: %w", err)
	}

	return nil
}

// RemoveRole 移除角色
func (r *postgresUserRepository) RemoveRole(ctx context.Context, userID, roleID, scopeType string, scopeValue *string) error {
	query := `DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2 AND scope_type = $3`
	args := []interface{}{userID, roleID, scopeType}

	if scopeValue != nil {
		query += " AND scope_value = $4"
		args = append(args, *scopeValue)
	} else {
		query += " AND scope_value IS NULL"
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to remove role", zap.Error(err),
			zap.String("userID", userID), zap.String("roleID", roleID))
		return fmt.Errorf("failed to remove role: %w", err)
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
