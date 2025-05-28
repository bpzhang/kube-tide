package models

import (
	"time"
)

// 用户状态常量
const (
	UserStatusActive    = "active"
	UserStatusInactive  = "inactive"
	UserStatusSuspended = "suspended"
)

// User 用户模型
type User struct {
	ID           string     `json:"id" db:"id"`
	Username     string     `json:"username" db:"username" validate:"required,min=3,max=50"`
	Email        string     `json:"email" db:"email" validate:"required,email"`
	PasswordHash string     `json:"-" db:"password_hash"`
	DisplayName  *string    `json:"display_name" db:"display_name"`
	AvatarURL    *string    `json:"avatar_url" db:"avatar_url"`
	Status       string     `json:"status" db:"status" validate:"oneof=active inactive suspended"`
	LastLoginAt  *time.Time `json:"last_login_at" db:"last_login_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	CreatedBy    *string    `json:"created_by" db:"created_by"`
}

// UserCreateRequest 创建用户请求
type UserCreateRequest struct {
	Username          string  `json:"username" validate:"required,min=3,max=50"`
	Email             string  `json:"email" validate:"required,email"`
	Password          string  `json:"password" validate:"required,min=8"`
	DisplayName       *string `json:"display_name" validate:"omitempty,max=100"`
	AvatarURL         *string `json:"avatar_url" validate:"omitempty,url"`
	AssignDefaultRole bool    `json:"assign_default_role"`
}

// UserUpdateRequest 更新用户请求
type UserUpdateRequest struct {
	Email       *string   `json:"email" validate:"omitempty,email"`
	DisplayName *string   `json:"display_name" validate:"omitempty,max=100"`
	AvatarURL   *string   `json:"avatar_url" validate:"omitempty,url"`
	Status      *string   `json:"status" validate:"omitempty,oneof=active inactive suspended"`
	UpdateTime  time.Time `json:"update_time" validate:"required"`
	Password    *string   `json:"password" validate:"omitempty,min=8"`
}

// UserPasswordUpdateRequest 更新密码请求
type UserPasswordUpdateRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// UserListFilters 用户列表过滤器
type UserListFilters struct {
	Status   string `json:"status" form:"status"`
	Username string `json:"username" form:"username"`
	Email    string `json:"email" form:"email"`
}

// UserWithRoles 包含角色信息的用户
type UserWithRoles struct {
	User
	Roles []UserRole `json:"roles"`
}
