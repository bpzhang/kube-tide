package core

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"kube-tide/internal/database/models"
	"kube-tide/internal/repository"
)

// PermissionService 权限管理服务
type PermissionService struct {
	permissionRepo repository.PermissionRepository
	logger         *zap.Logger
}

// NewPermissionService 创建权限服务
func NewPermissionService(
	permissionRepo repository.PermissionRepository,
	logger *zap.Logger,
) *PermissionService {
	return &PermissionService{
		permissionRepo: permissionRepo,
		logger:         logger,
	}
}

// GetPermission 获取权限
func (s *PermissionService) GetPermission(ctx context.Context, id string) (*models.Permission, error) {
	return s.permissionRepo.GetByID(ctx, id)
}

// GetPermissionByName 根据名称获取权限
func (s *PermissionService) GetPermissionByName(ctx context.Context, name string) (*models.Permission, error) {
	return s.permissionRepo.GetByName(ctx, name)
}

// ListPermissions 获取权限列表
func (s *PermissionService) ListPermissions(ctx context.Context) ([]models.Permission, error) {
	return s.permissionRepo.List(ctx)
}

// GetPermissionsByResource 根据资源类型获取权限
func (s *PermissionService) GetPermissionsByResource(ctx context.Context, resourceType string) ([]models.Permission, error) {
	allPermissions, err := s.permissionRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions: %w", err)
	}

	var filteredPermissions []models.Permission
	for _, permission := range allPermissions {
		if permission.ResourceType == resourceType {
			filteredPermissions = append(filteredPermissions, permission)
		}
	}

	return filteredPermissions, nil
}

// GetUserPermissions 获取用户权限
func (s *PermissionService) GetUserPermissions(ctx context.Context, userID string, clusterName, namespace *string) ([]models.Permission, error) {
	return s.permissionRepo.GetUserPermissions(ctx, userID, clusterName, namespace)
}

// CheckUserPermission 检查用户是否有特定权限
func (s *PermissionService) CheckUserPermission(ctx context.Context, userID, resourceType, action string, clusterName, namespace *string) (bool, error) {
	// 获取用户在指定作用域的权限
	permissions, err := s.permissionRepo.GetUserPermissions(ctx, userID, clusterName, namespace)
	if err != nil {
		return false, fmt.Errorf("failed to get user permissions: %w", err)
	}

	// 检查是否有匹配的权限
	for _, permission := range permissions {
		if permission.ResourceType == resourceType && permission.Action == action {
			return true, nil
		}

		// 检查通配符权限
		if permission.ResourceType == resourceType && permission.Action == "*" {
			return true, nil
		}

		if permission.ResourceType == "*" && permission.Action == action {
			return true, nil
		}

		if permission.ResourceType == "*" && permission.Action == "*" {
			return true, nil
		}
	}

	return false, nil
}

// CheckPermission 检查权限（带详细信息）
func (s *PermissionService) CheckPermission(ctx context.Context, req *models.PermissionCheckRequest, userID string) (*models.PermissionCheckResponse, error) {
	allowed, err := s.CheckUserPermission(ctx, userID, req.Resource, req.Action, req.ClusterName, req.Namespace)
	if err != nil {
		return &models.PermissionCheckResponse{
			Allowed: false,
			Reason:  fmt.Sprintf("Error checking permission: %v", err),
		}, nil
	}

	response := &models.PermissionCheckResponse{
		Allowed: allowed,
	}

	if !allowed {
		scope := "global"
		if req.ClusterName != nil && req.Namespace != nil {
			scope = fmt.Sprintf("namespace %s/%s", *req.ClusterName, *req.Namespace)
		} else if req.ClusterName != nil {
			scope = fmt.Sprintf("cluster %s", *req.ClusterName)
		}

		response.Reason = fmt.Sprintf("User does not have permission to %s %s in %s scope",
			req.Action, req.Resource, scope)
	}

	return response, nil
}

// GetPermissionsByScope 根据作用域获取权限
func (s *PermissionService) GetPermissionsByScope(ctx context.Context, scope string) ([]models.Permission, error) {
	allPermissions, err := s.permissionRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions: %w", err)
	}

	var filteredPermissions []models.Permission
	for _, permission := range allPermissions {
		if permission.Scope == scope {
			filteredPermissions = append(filteredPermissions, permission)
		}
	}

	return filteredPermissions, nil
}

// GroupPermissionsByResource 按资源类型分组权限
func (s *PermissionService) GroupPermissionsByResource(ctx context.Context) (map[string][]models.Permission, error) {
	permissions, err := s.permissionRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions: %w", err)
	}

	grouped := make(map[string][]models.Permission)
	for _, permission := range permissions {
		grouped[permission.ResourceType] = append(grouped[permission.ResourceType], permission)
	}

	return grouped, nil
}

// ValidatePermissionName 验证权限名称格式
func (s *PermissionService) ValidatePermissionName(name string) error {
	parts := strings.Split(name, ":")
	if len(parts) != 2 {
		return fmt.Errorf("permission name must be in format 'resource:action'")
	}

	resource := parts[0]
	action := parts[1]

	if resource == "" {
		return fmt.Errorf("resource type cannot be empty")
	}

	if action == "" {
		return fmt.Errorf("action cannot be empty")
	}

	// 验证资源类型格式
	if !isValidResourceType(resource) {
		return fmt.Errorf("invalid resource type: %s", resource)
	}

	// 验证操作格式
	if !isValidAction(action) {
		return fmt.Errorf("invalid action: %s", action)
	}

	return nil
}

// isValidResourceType 验证资源类型是否有效
func isValidResourceType(resourceType string) bool {
	validResources := []string{
		"cluster", "node", "namespace", "deployment", "service", "pod",
		"user", "role", "audit", "*",
	}

	for _, valid := range validResources {
		if resourceType == valid {
			return true
		}
	}

	return false
}

// isValidAction 验证操作是否有效
func isValidAction(action string) bool {
	validActions := []string{
		"create", "read", "update", "delete", "list",
		"scale", "restart", "drain", "cordon", "logs", "exec",
		"*",
	}

	for _, valid := range validActions {
		if action == valid {
			return true
		}
	}

	return false
}
