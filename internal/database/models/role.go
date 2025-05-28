package models

import (
	"time"
)

// Role 角色模型
type Role struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name" validate:"required,min=2,max=50"`
	DisplayName string    `json:"display_name" db:"display_name" validate:"required,max=100"`
	Description *string   `json:"description" db:"description"`
	Type        string    `json:"type" db:"type" validate:"oneof=system custom"`
	IsDefault   bool      `json:"is_default" db:"is_default"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	CreatedBy   *string   `json:"created_by" db:"created_by"`
}

// Permission 权限模型
type Permission struct {
	ID           string  `json:"id" db:"id"`
	Name         string  `json:"name" db:"name" validate:"required,max=100"`
	DisplayName  string  `json:"display_name" db:"display_name" validate:"required,max=100"`
	Description  *string `json:"description" db:"description"`
	ResourceType string  `json:"resource_type" db:"resource_type" validate:"required,max=50"`
	Action       string  `json:"action" db:"action" validate:"required,max=50"`
	Scope        string  `json:"scope" db:"scope" validate:"oneof=global cluster namespace"`
}

// UserRole 用户角色关联模型
type UserRole struct {
	ID         string     `json:"id" db:"id"`
	UserID     string     `json:"user_id" db:"user_id"`
	RoleID     string     `json:"role_id" db:"role_id"`
	ScopeType  string     `json:"scope_type" db:"scope_type" validate:"oneof=global cluster namespace"`
	ScopeValue *string    `json:"scope_value" db:"scope_value"`
	GrantedAt  time.Time  `json:"granted_at" db:"granted_at"`
	GrantedBy  *string    `json:"granted_by" db:"granted_by"`
	ExpiresAt  *time.Time `json:"expires_at" db:"expires_at"`

	// 关联数据
	Role *Role `json:"role,omitempty"`
}

// RolePermission 角色权限关联模型
type RolePermission struct {
	ID           string    `json:"id" db:"id"`
	RoleID       string    `json:"role_id" db:"role_id"`
	PermissionID string    `json:"permission_id" db:"permission_id"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`

	// 关联数据
	Permission *Permission `json:"permission,omitempty"`
}

// RoleCreateRequest 创建角色请求
type RoleCreateRequest struct {
	Name        string  `json:"name" validate:"required,min=2,max=50"`
	DisplayName string  `json:"display_name" validate:"required,max=100"`
	Description *string `json:"description" validate:"omitempty,max=500"`
	IsDefault   *bool   `json:"is_default"`
}

// RoleUpdateRequest 更新角色请求
type RoleUpdateRequest struct {
	DisplayName *string `json:"display_name" validate:"omitempty,max=100"`
	Description *string `json:"description" validate:"omitempty,max=500"`
	IsDefault   *bool   `json:"is_default"`
}

// UserRoleAssignRequest 分配用户角色请求
type UserRoleAssignRequest struct {
	RoleID     string     `json:"role_id" validate:"required"`
	ScopeType  string     `json:"scope_type" validate:"oneof=global cluster namespace"`
	ScopeValue *string    `json:"scope_value"`
	ExpiresAt  *time.Time `json:"expires_at"`
}

// RolePermissionAssignRequest 分配角色权限请求
type RolePermissionAssignRequest struct {
	PermissionIDs []string `json:"permission_ids" validate:"required,min=1"`
}

// RoleWithPermissions 包含权限信息的角色
type RoleWithPermissions struct {
	Role
	Permissions []RolePermission `json:"permissions"`
}

// UserRoleListFilters 用户角色列表过滤器
type UserRoleListFilters struct {
	ScopeType  string `json:"scope_type" form:"scope_type"`
	ScopeValue string `json:"scope_value" form:"scope_value"`
}

// RoleListFilters 角色列表过滤器
type RoleListFilters struct {
	Type      string `json:"type" form:"type"`
	IsDefault *bool  `json:"is_default" form:"is_default"`
}
