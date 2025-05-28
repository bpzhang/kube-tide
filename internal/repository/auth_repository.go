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

// UserSessionRepository 用户会话仓储接口
type UserSessionRepository interface {
	Create(ctx context.Context, session *models.UserSession) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*models.UserSession, error)
	GetByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (*models.UserSession, error)
	UpdateLastUsed(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
	DeleteByUserID(ctx context.Context, userID string) error
	DeleteExpired(ctx context.Context) error
	List(ctx context.Context, filters models.UserSessionListFilters, pagination utils.PaginationParams) (*utils.PaginatedResult[models.UserSession], error)
}

// AuditLogRepository 审计日志仓储接口
type AuditLogRepository interface {
	Create(ctx context.Context, log *models.AuditLog) error
	GetByID(ctx context.Context, id string) (*models.AuditLog, error)
	List(ctx context.Context, filters models.AuditLogListFilters, pagination utils.PaginationParams) (*utils.PaginatedResult[models.AuditLog], error)
	DeleteOldLogs(ctx context.Context, olderThan time.Time) error
}

// postgresUserSessionRepository PostgreSQL 用户会话仓储实现
type postgresUserSessionRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewUserSessionRepository 创建用户会话仓储
func NewUserSessionRepository(db *sql.DB, logger *zap.Logger) UserSessionRepository {
	return &postgresUserSessionRepository{
		db:     db,
		logger: logger,
	}
}

// Create 创建会话
func (r *postgresUserSessionRepository) Create(ctx context.Context, session *models.UserSession) error {
	query := `
		INSERT INTO user_sessions (id, user_id, token_hash, refresh_token_hash, ip_address, user_agent, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.ExecContext(ctx, query,
		session.ID, session.UserID, session.TokenHash, session.RefreshTokenHash,
		session.IPAddress, session.UserAgent, session.ExpiresAt)

	if err != nil {
		r.logger.Error("Failed to create user session", zap.Error(err), zap.String("userID", session.UserID))
		return fmt.Errorf("failed to create user session: %w", err)
	}

	return nil
}

// GetByTokenHash 根据令牌哈希获取会话
func (r *postgresUserSessionRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*models.UserSession, error) {
	query := `
		SELECT id, user_id, token_hash, refresh_token_hash, ip_address, user_agent,
		       expires_at, created_at, last_used_at
		FROM user_sessions 
		WHERE token_hash = $1 AND expires_at > NOW()`

	session := &models.UserSession{}
	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&session.ID, &session.UserID, &session.TokenHash, &session.RefreshTokenHash,
		&session.IPAddress, &session.UserAgent, &session.ExpiresAt,
		&session.CreatedAt, &session.LastUsedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrNotFound
		}
		r.logger.Error("Failed to get session by token hash", zap.Error(err))
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return session, nil
}

// GetByRefreshTokenHash 根据刷新令牌哈希获取会话
func (r *postgresUserSessionRepository) GetByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (*models.UserSession, error) {
	query := `
		SELECT id, user_id, token_hash, refresh_token_hash, ip_address, user_agent,
		       expires_at, created_at, last_used_at
		FROM user_sessions 
		WHERE refresh_token_hash = $1 AND expires_at > NOW()`

	session := &models.UserSession{}
	err := r.db.QueryRowContext(ctx, query, refreshTokenHash).Scan(
		&session.ID, &session.UserID, &session.TokenHash, &session.RefreshTokenHash,
		&session.IPAddress, &session.UserAgent, &session.ExpiresAt,
		&session.CreatedAt, &session.LastUsedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrNotFound
		}
		r.logger.Error("Failed to get session by refresh token hash", zap.Error(err))
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return session, nil
}

// UpdateLastUsed 更新最后使用时间
func (r *postgresUserSessionRepository) UpdateLastUsed(ctx context.Context, id string) error {
	query := `UPDATE user_sessions SET last_used_at = NOW() WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to update session last used", zap.Error(err), zap.String("id", id))
		return fmt.Errorf("failed to update session last used: %w", err)
	}

	return nil
}

// Delete 删除会话
func (r *postgresUserSessionRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM user_sessions WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to delete session", zap.Error(err), zap.String("id", id))
		return fmt.Errorf("failed to delete session: %w", err)
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

// DeleteByUserID 删除用户的所有会话
func (r *postgresUserSessionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	query := `DELETE FROM user_sessions WHERE user_id = $1`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		r.logger.Error("Failed to delete sessions by user ID", zap.Error(err), zap.String("userID", userID))
		return fmt.Errorf("failed to delete sessions: %w", err)
	}

	return nil
}

// DeleteExpired 删除过期会话
func (r *postgresUserSessionRepository) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM user_sessions WHERE expires_at <= NOW()`

	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		r.logger.Error("Failed to delete expired sessions", zap.Error(err))
		return fmt.Errorf("failed to delete expired sessions: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	r.logger.Info("Deleted expired sessions", zap.Int64("count", rowsAffected))
	return nil
}

// List 获取会话列表
func (r *postgresUserSessionRepository) List(ctx context.Context, filters models.UserSessionListFilters, pagination utils.PaginationParams) (*utils.PaginatedResult[models.UserSession], error) {
	whereParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if filters.UserID != "" {
		whereParts = append(whereParts, fmt.Sprintf("user_id = $%d", argIndex))
		args = append(args, filters.UserID)
		argIndex++
	}

	if filters.IPAddress != "" {
		whereParts = append(whereParts, fmt.Sprintf("ip_address = $%d", argIndex))
		args = append(args, filters.IPAddress)
		argIndex++
	}

	if filters.Active != nil {
		if *filters.Active {
			whereParts = append(whereParts, "expires_at > NOW()")
		} else {
			whereParts = append(whereParts, "expires_at <= NOW()")
		}
	}

	whereClause := ""
	if len(whereParts) > 0 {
		whereClause = "WHERE " + strings.Join(whereParts, " AND ")
	}

	// 获取总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM user_sessions %s", whereClause)
	var totalCount int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		r.logger.Error("Failed to count sessions", zap.Error(err))
		return nil, fmt.Errorf("failed to count sessions: %w", err)
	}

	// 获取数据
	offset := (pagination.Page - 1) * pagination.PageSize
	dataQuery := fmt.Sprintf(`
		SELECT id, user_id, token_hash, refresh_token_hash, ip_address, user_agent,
		       expires_at, created_at, last_used_at
		FROM user_sessions %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, pagination.PageSize, offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		r.logger.Error("Failed to list sessions", zap.Error(err))
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	defer rows.Close()

	sessions := []models.UserSession{}
	for rows.Next() {
		var session models.UserSession
		err := rows.Scan(
			&session.ID, &session.UserID, &session.TokenHash, &session.RefreshTokenHash,
			&session.IPAddress, &session.UserAgent, &session.ExpiresAt,
			&session.CreatedAt, &session.LastUsedAt)
		if err != nil {
			r.logger.Error("Failed to scan session", zap.Error(err))
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, session)
	}

	return utils.NewPaginatedResult(sessions, totalCount, pagination.Page, pagination.PageSize), nil
}

// postgresAuditLogRepository PostgreSQL 审计日志仓储实现
type postgresAuditLogRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewAuditLogRepository 创建审计日志仓储
func NewAuditLogRepository(db *sql.DB, logger *zap.Logger) AuditLogRepository {
	return &postgresAuditLogRepository{
		db:     db,
		logger: logger,
	}
}

// Create 创建审计日志
func (r *postgresAuditLogRepository) Create(ctx context.Context, log *models.AuditLog) error {
	query := `
		INSERT INTO audit_logs (id, user_id, action, resource_type, resource_id, cluster_name, 
		                       namespace, details, ip_address, user_agent, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := r.db.ExecContext(ctx, query,
		log.ID, log.UserID, log.Action, log.ResourceType, log.ResourceID,
		log.ClusterName, log.Namespace, log.Details, log.IPAddress, log.UserAgent, log.Status)

	if err != nil {
		r.logger.Error("Failed to create audit log", zap.Error(err), zap.String("action", log.Action))
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// GetByID 根据 ID 获取审计日志
func (r *postgresAuditLogRepository) GetByID(ctx context.Context, id string) (*models.AuditLog, error) {
	query := `
		SELECT al.id, al.user_id, al.action, al.resource_type, al.resource_id,
		       al.cluster_name, al.namespace, al.details, al.ip_address, al.user_agent,
		       al.status, al.created_at,
		       u.id, u.username, u.email, u.display_name
		FROM audit_logs al
		LEFT JOIN users u ON al.user_id = u.id
		WHERE al.id = $1`

	log := &models.AuditLog{}
	var user models.User
	var userID, username, email, displayName sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&log.ID, &log.UserID, &log.Action, &log.ResourceType, &log.ResourceID,
		&log.ClusterName, &log.Namespace, &log.Details, &log.IPAddress, &log.UserAgent,
		&log.Status, &log.CreatedAt,
		&userID, &username, &email, &displayName)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrNotFound
		}
		r.logger.Error("Failed to get audit log by ID", zap.Error(err), zap.String("id", id))
		return nil, fmt.Errorf("failed to get audit log: %w", err)
	}

	// 设置用户信息
	if userID.Valid {
		user.ID = userID.String
		user.Username = username.String
		user.Email = email.String
		if displayName.Valid {
			user.DisplayName = &displayName.String
		}
		log.User = &user
	}

	return log, nil
}

// List 获取审计日志列表
func (r *postgresAuditLogRepository) List(ctx context.Context, filters models.AuditLogListFilters, pagination utils.PaginationParams) (*utils.PaginatedResult[models.AuditLog], error) {
	whereParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if filters.UserID != "" {
		whereParts = append(whereParts, fmt.Sprintf("al.user_id = $%d", argIndex))
		args = append(args, filters.UserID)
		argIndex++
	}

	if filters.Action != "" {
		whereParts = append(whereParts, fmt.Sprintf("al.action ILIKE $%d", argIndex))
		args = append(args, "%"+filters.Action+"%")
		argIndex++
	}

	if filters.ResourceType != "" {
		whereParts = append(whereParts, fmt.Sprintf("al.resource_type = $%d", argIndex))
		args = append(args, filters.ResourceType)
		argIndex++
	}

	if filters.ClusterName != "" {
		whereParts = append(whereParts, fmt.Sprintf("al.cluster_name = $%d", argIndex))
		args = append(args, filters.ClusterName)
		argIndex++
	}

	if filters.Status != "" {
		whereParts = append(whereParts, fmt.Sprintf("al.status = $%d", argIndex))
		args = append(args, filters.Status)
		argIndex++
	}

	if !filters.StartTime.IsZero() {
		whereParts = append(whereParts, fmt.Sprintf("al.created_at >= $%d", argIndex))
		args = append(args, filters.StartTime)
		argIndex++
	}

	if !filters.EndTime.IsZero() {
		whereParts = append(whereParts, fmt.Sprintf("al.created_at <= $%d", argIndex))
		args = append(args, filters.EndTime)
		argIndex++
	}

	whereClause := ""
	if len(whereParts) > 0 {
		whereClause = "WHERE " + strings.Join(whereParts, " AND ")
	}

	// 获取总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM audit_logs al %s", whereClause)
	var totalCount int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		r.logger.Error("Failed to count audit logs", zap.Error(err))
		return nil, fmt.Errorf("failed to count audit logs: %w", err)
	}

	// 获取数据
	offset := (pagination.Page - 1) * pagination.PageSize
	dataQuery := fmt.Sprintf(`
		SELECT al.id, al.user_id, al.action, al.resource_type, al.resource_id,
		       al.cluster_name, al.namespace, al.details, al.ip_address, al.user_agent,
		       al.status, al.created_at,
		       u.id, u.username, u.email, u.display_name
		FROM audit_logs al
		LEFT JOIN users u ON al.user_id = u.id
		%s
		ORDER BY al.created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, pagination.PageSize, offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		r.logger.Error("Failed to list audit logs", zap.Error(err))
		return nil, fmt.Errorf("failed to list audit logs: %w", err)
	}
	defer rows.Close()

	logs := []models.AuditLog{}
	for rows.Next() {
		var log models.AuditLog
		var user models.User
		var userID, username, email, displayName sql.NullString

		err := rows.Scan(
			&log.ID, &log.UserID, &log.Action, &log.ResourceType, &log.ResourceID,
			&log.ClusterName, &log.Namespace, &log.Details, &log.IPAddress, &log.UserAgent,
			&log.Status, &log.CreatedAt,
			&userID, &username, &email, &displayName)
		if err != nil {
			r.logger.Error("Failed to scan audit log", zap.Error(err))
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}

		// 设置用户信息
		if userID.Valid {
			user.ID = userID.String
			user.Username = username.String
			user.Email = email.String
			if displayName.Valid {
				user.DisplayName = &displayName.String
			}
			log.User = &user
		}

		logs = append(logs, log)
	}

	return utils.NewPaginatedResult(logs, totalCount, pagination.Page, pagination.PageSize), nil
}

// DeleteOldLogs 删除旧的审计日志
func (r *postgresAuditLogRepository) DeleteOldLogs(ctx context.Context, olderThan time.Time) error {
	query := `DELETE FROM audit_logs WHERE created_at < $1`

	result, err := r.db.ExecContext(ctx, query, olderThan)
	if err != nil {
		r.logger.Error("Failed to delete old audit logs", zap.Error(err))
		return fmt.Errorf("failed to delete old audit logs: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	r.logger.Info("Deleted old audit logs", zap.Int64("count", rowsAffected), zap.Time("olderThan", olderThan))
	return nil
}
