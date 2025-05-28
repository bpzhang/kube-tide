package core

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"kube-tide/internal/database/models"
	"kube-tide/internal/repository"
	"kube-tide/internal/utils"
)

// UserService 用户管理服务
type UserService struct {
	userRepo  repository.UserRepository
	roleRepo  repository.RoleRepository
	auditRepo repository.AuditLogRepository
	logger    *zap.Logger
}

// NewUserService 创建用户服务
func NewUserService(
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	auditRepo repository.AuditLogRepository,
	logger *zap.Logger,
) *UserService {
	return &UserService{
		userRepo:  userRepo,
		roleRepo:  roleRepo,
		auditRepo: auditRepo,
		logger:    logger,
	}
}

// CreateUser 创建用户
func (s *UserService) CreateUser(ctx context.Context, req *models.UserCreateRequest, createdBy string) (*models.User, error) {
	// 检查用户名是否已存在
	if _, err := s.userRepo.GetByUsername(ctx, req.Username); err == nil {
		return nil, utils.ErrUsernameExists
	}

	// 检查邮箱是否已存在
	if _, err := s.userRepo.GetByEmail(ctx, req.Email); err == nil {
		return nil, utils.ErrEmailExists
	}

	// 生成密码哈希
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 创建用户
	user := &models.User{
		ID:           uuid.New().String(),
		Username:     req.Username,
		Email:        req.Email,
		DisplayName:  req.DisplayName,
		AvatarURL:    req.AvatarURL,
		PasswordHash: string(passwordHash),
		Status:       models.UserStatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		CreatedBy:    &createdBy,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// 分配默认角色
	if req.AssignDefaultRole {
		if err := s.assignDefaultRoles(ctx, user.ID, createdBy); err != nil {
			s.logger.Warn("Failed to assign default roles", zap.Error(err), zap.String("userID", user.ID))
		}
	}

	// 记录审计日志
	s.logAuditEvent(ctx, createdBy, "user_created", "success", fmt.Sprintf("Created user: %s", user.Username))

	return user, nil
}

// GetUser 获取用户
func (s *UserService) GetUser(ctx context.Context, id string) (*models.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

// GetUserByUsername 根据用户名获取用户
func (s *UserService) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	return s.userRepo.GetByUsername(ctx, username)
}

// UpdateUser 更新用户
func (s *UserService) UpdateUser(ctx context.Context, id string, req *models.UserUpdateRequest, updatedBy string) (*models.User, error) {
	// 检查用户是否存在
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// 检查邮箱是否被其他用户使用
	if req.Email != nil && *req.Email != user.Email {
		if existingUser, err := s.userRepo.GetByEmail(ctx, *req.Email); err == nil && existingUser.ID != id {
			return nil, utils.ErrEmailExists
		}
	}

	// 更新用户
	if err := s.userRepo.Update(ctx, id, *req); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// 获取更新后的用户
	updatedUser, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated user: %w", err)
	}

	// 记录审计日志
	s.logAuditEvent(ctx, updatedBy, "user_updated", "success", fmt.Sprintf("Updated user: %s", user.Username))

	return updatedUser, nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(ctx context.Context, id string, deletedBy string) error {
	// 检查用户是否存在
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// 删除用户
	if err := s.userRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// 记录审计日志
	s.logAuditEvent(ctx, deletedBy, "user_deleted", "success", fmt.Sprintf("Deleted user: %s", user.Username))

	return nil
}

// ListUsers 获取用户列表
func (s *UserService) ListUsers(ctx context.Context, filters models.UserListFilters, pagination utils.PaginationParams) (*utils.PaginatedResult[models.User], error) {
	return s.userRepo.List(ctx, filters, pagination)
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(ctx context.Context, userID string, req *models.ChangePasswordRequest) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// 验证当前密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		s.logAuditEvent(ctx, userID, "password_change_failed", "invalid_current_password", user.Username)
		return utils.ErrInvalidCredentials
	}

	// 生成新密码哈希
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// 更新密码
	if err := s.userRepo.UpdatePassword(ctx, userID, string(passwordHash)); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// 记录审计日志
	s.logAuditEvent(ctx, userID, "password_changed", "success", user.Username)

	return nil
}

// ResetPassword 重置密码（管理员操作）
func (s *UserService) ResetPassword(ctx context.Context, userID string, newPassword string, resetBy string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// 生成新密码哈希
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// 更新密码
	if err := s.userRepo.UpdatePassword(ctx, userID, string(passwordHash)); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// 记录审计日志
	s.logAuditEvent(ctx, resetBy, "password_reset", "success", fmt.Sprintf("Reset password for user: %s", user.Username))

	return nil
}

// GetUserRoles 获取用户角色
func (s *UserService) GetUserRoles(ctx context.Context, userID string, filters models.UserRoleListFilters) ([]models.UserRole, error) {
	return s.userRepo.GetUserRoles(ctx, userID, filters)
}

// AssignRole 分配角色
func (s *UserService) AssignRole(ctx context.Context, userID string, req *models.UserRoleAssignRequest, assignedBy string) error {
	// 检查用户是否存在
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// 检查角色是否存在
	role, err := s.roleRepo.GetByID(ctx, req.RoleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	// 创建用户角色关联
	userRole := &models.UserRole{
		ID:         uuid.New().String(),
		UserID:     userID,
		RoleID:     req.RoleID,
		ScopeType:  req.ScopeType,
		ScopeValue: req.ScopeValue,
		GrantedAt:  time.Now(),
		GrantedBy:  &assignedBy,
		ExpiresAt:  req.ExpiresAt,
	}

	if err := s.userRepo.AssignRole(ctx, userRole); err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	// 记录审计日志
	s.logAuditEvent(ctx, assignedBy, "role_assigned", "success",
		fmt.Sprintf("Assigned role %s to user %s", role.Name, user.Username))

	return nil
}

// RemoveRole 移除角色
func (s *UserService) RemoveRole(ctx context.Context, userID, roleID, scopeType string, scopeValue *string, removedBy string) error {
	// 检查用户是否存在
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// 检查角色是否存在
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	// 移除角色
	if err := s.userRepo.RemoveRole(ctx, userID, roleID, scopeType, scopeValue); err != nil {
		return fmt.Errorf("failed to remove role: %w", err)
	}

	// 记录审计日志
	s.logAuditEvent(ctx, removedBy, "role_removed", "success",
		fmt.Sprintf("Removed role %s from user %s", role.Name, user.Username))

	return nil
}

// assignDefaultRoles 分配默认角色
func (s *UserService) assignDefaultRoles(ctx context.Context, userID, assignedBy string) error {
	defaultRoles, err := s.roleRepo.GetDefaultRoles(ctx)
	if err != nil {
		return fmt.Errorf("failed to get default roles: %w", err)
	}

	for _, role := range defaultRoles {
		userRole := &models.UserRole{
			ID:        uuid.New().String(),
			UserID:    userID,
			RoleID:    role.ID,
			ScopeType: "global",
			GrantedAt: time.Now(),
			GrantedBy: &assignedBy,
		}

		if err := s.userRepo.AssignRole(ctx, userRole); err != nil {
			s.logger.Error("Failed to assign default role", zap.Error(err),
				zap.String("userID", userID), zap.String("roleID", role.ID))
		}
	}

	return nil
}

// logAuditEvent 记录审计日志
func (s *UserService) logAuditEvent(ctx context.Context, userID, action, result, details string) {
	auditLog := &models.AuditLog{
		ID:        uuid.New().String(),
		UserID:    userID,
		Action:    action,
		Resource:  "user",
		Result:    result,
		Details:   details,
		Timestamp: time.Now(),
	}

	if err := s.auditRepo.Create(ctx, auditLog); err != nil {
		s.logger.Error("Failed to create audit log", zap.Error(err))
	}
}
