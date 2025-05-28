package core

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"kube-tide/internal/database/models"
	"kube-tide/internal/repository"
	"kube-tide/internal/utils"
)

// RoleService 角色管理服务
type RoleService struct {
	roleRepo       repository.RoleRepository
	permissionRepo repository.PermissionRepository
	auditRepo      repository.AuditLogRepository
	logger         *zap.Logger
}

// NewRoleService 创建角色服务
func NewRoleService(
	roleRepo repository.RoleRepository,
	permissionRepo repository.PermissionRepository,
	auditRepo repository.AuditLogRepository,
	logger *zap.Logger,
) *RoleService {
	return &RoleService{
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
		auditRepo:      auditRepo,
		logger:         logger,
	}
}

// CreateRole 创建角色
func (s *RoleService) CreateRole(ctx context.Context, req *models.RoleCreateRequest, createdBy string) (*models.Role, error) {
	// 检查角色名是否已存在
	if _, err := s.roleRepo.GetByName(ctx, req.Name); err == nil {
		return nil, fmt.Errorf("role name already exists")
	}

	// 创建角色
	role := &models.Role{
		ID:          uuid.New().String(),
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Type:        "custom",
		IsDefault:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		CreatedBy:   &createdBy,
	}

	if req.IsDefault != nil {
		role.IsDefault = *req.IsDefault
	}

	if err := s.roleRepo.Create(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	// 记录审计日志
	s.logAuditEvent(ctx, createdBy, "role_created", "success", fmt.Sprintf("Created role: %s", role.Name))

	return role, nil
}

// GetRole 获取角色
func (s *RoleService) GetRole(ctx context.Context, id string) (*models.Role, error) {
	return s.roleRepo.GetByID(ctx, id)
}

// GetRoleByName 根据名称获取角色
func (s *RoleService) GetRoleByName(ctx context.Context, name string) (*models.Role, error) {
	return s.roleRepo.GetByName(ctx, name)
}

// UpdateRole 更新角色
func (s *RoleService) UpdateRole(ctx context.Context, id string, req *models.RoleUpdateRequest, updatedBy string) (*models.Role, error) {
	// 检查角色是否存在
	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("role not found: %w", err)
	}

	// 检查是否为系统角色
	if role.Type == "system" {
		return nil, fmt.Errorf("cannot update system role")
	}

	// 更新角色
	if err := s.roleRepo.Update(ctx, id, *req); err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	// 获取更新后的角色
	updatedRole, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated role: %w", err)
	}

	// 记录审计日志
	s.logAuditEvent(ctx, updatedBy, "role_updated", "success", fmt.Sprintf("Updated role: %s", role.Name))

	return updatedRole, nil
}

// DeleteRole 删除角色
func (s *RoleService) DeleteRole(ctx context.Context, id string, deletedBy string) error {
	// 检查角色是否存在
	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	// 检查是否为系统角色
	if role.Type == "system" {
		return fmt.Errorf("cannot delete system role")
	}

	// 删除角色
	if err := s.roleRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	// 记录审计日志
	s.logAuditEvent(ctx, deletedBy, "role_deleted", "success", fmt.Sprintf("Deleted role: %s", role.Name))

	return nil
}

// ListRoles 获取角色列表
func (s *RoleService) ListRoles(ctx context.Context, filters models.RoleListFilters, pagination utils.PaginationParams) (*utils.PaginatedResult[models.Role], error) {
	return s.roleRepo.List(ctx, filters, pagination)
}

// GetRolePermissions 获取角色权限
func (s *RoleService) GetRolePermissions(ctx context.Context, roleID string) ([]models.RolePermission, error) {
	return s.roleRepo.GetRolePermissions(ctx, roleID)
}

// GetRoleWithPermissions 获取包含权限的角色信息
func (s *RoleService) GetRoleWithPermissions(ctx context.Context, roleID string) (*models.RoleWithPermissions, error) {
	// 获取角色信息
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("role not found: %w", err)
	}

	// 获取角色权限
	permissions, err := s.roleRepo.GetRolePermissions(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}

	return &models.RoleWithPermissions{
		Role:        *role,
		Permissions: permissions,
	}, nil
}

// AssignPermissions 分配权限
func (s *RoleService) AssignPermissions(ctx context.Context, roleID string, req *models.RolePermissionAssignRequest, assignedBy string) error {
	// 检查角色是否存在
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	// 检查是否为系统角色
	if role.Type == "system" {
		return fmt.Errorf("cannot modify permissions for system role")
	}

	// 验证权限是否存在
	for _, permissionID := range req.PermissionIDs {
		if _, err := s.permissionRepo.GetByID(ctx, permissionID); err != nil {
			return fmt.Errorf("permission %s not found: %w", permissionID, err)
		}
	}

	// 分配权限
	if err := s.roleRepo.AssignPermissions(ctx, roleID, req.PermissionIDs); err != nil {
		return fmt.Errorf("failed to assign permissions: %w", err)
	}

	// 记录审计日志
	s.logAuditEvent(ctx, assignedBy, "permissions_assigned", "success",
		fmt.Sprintf("Assigned %d permissions to role: %s", len(req.PermissionIDs), role.Name))

	return nil
}

// RemovePermissions 移除权限
func (s *RoleService) RemovePermissions(ctx context.Context, roleID string, permissionIDs []string, removedBy string) error {
	// 检查角色是否存在
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	// 检查是否为系统角色
	if role.Type == "system" {
		return fmt.Errorf("cannot modify permissions for system role")
	}

	// 移除权限
	if err := s.roleRepo.RemovePermissions(ctx, roleID, permissionIDs); err != nil {
		return fmt.Errorf("failed to remove permissions: %w", err)
	}

	// 记录审计日志
	s.logAuditEvent(ctx, removedBy, "permissions_removed", "success",
		fmt.Sprintf("Removed %d permissions from role: %s", len(permissionIDs), role.Name))

	return nil
}

// GetDefaultRoles 获取默认角色
func (s *RoleService) GetDefaultRoles(ctx context.Context) ([]models.Role, error) {
	return s.roleRepo.GetDefaultRoles(ctx)
}

// logAuditEvent 记录审计日志
func (s *RoleService) logAuditEvent(ctx context.Context, userID, action, result, details string) {
	auditLog := &models.AuditLog{
		ID:        uuid.New().String(),
		UserID:    userID,
		Action:    action,
		Resource:  "role",
		Result:    result,
		Details:   details,
		Timestamp: time.Now(),
	}

	if err := s.auditRepo.Create(ctx, auditLog); err != nil {
		s.logger.Error("Failed to create audit log", zap.Error(err))
	}
}
