package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// UserSession 用户会话模型
type UserSession struct {
	ID               string    `json:"id" db:"id"`
	UserID           string    `json:"user_id" db:"user_id"`
	Token            string    `json:"-" db:"token_hash"` // 实际存储的是 token hash
	TokenHash        string    `json:"-" db:"token_hash"`
	RefreshTokenHash *string   `json:"-" db:"refresh_token_hash"`
	IPAddress        *string   `json:"ip_address" db:"ip_address"`
	UserAgent        *string   `json:"user_agent" db:"user_agent"`
	ExpiresAt        time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	LastUsedAt       time.Time `json:"last_used_at" db:"last_used_at"`
}

// AuditLog 审计日志模型
type AuditLog struct {
	ID           string    `json:"id" db:"id"`
	UserID       string    `json:"user_id" db:"user_id"`
	Action       string    `json:"action" db:"action"`
	Resource     string    `json:"resource" db:"resource"`
	ResourceType *string   `json:"resource_type" db:"resource_type"`
	ResourceID   *string   `json:"resource_id" db:"resource_id"`
	ClusterName  *string   `json:"cluster_name" db:"cluster_name"`
	Namespace    *string   `json:"namespace" db:"namespace"`
	Result       string    `json:"result" db:"result"`
	Details      string    `json:"details" db:"details"`
	IPAddress    *string   `json:"ip_address" db:"ip_address"`
	UserAgent    *string   `json:"user_agent" db:"user_agent"`
	Status       string    `json:"status" db:"status"` // success, failed, denied
	Timestamp    time.Time `json:"timestamp" db:"timestamp"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`

	// 关联数据
	User *User `json:"user,omitempty"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username  string `json:"username" validate:"required"`
	Password  string `json:"password" validate:"required"`
	IPAddress string `json:"ip_address,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	User      *UserResponse `json:"user"`
	Token     string        `json:"token"`
	ExpiresAt time.Time     `json:"expires_at"`
}

// UserResponse 用户响应模型
type UserResponse struct {
	ID          string     `json:"id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	DisplayName *string    `json:"display_name"`
	AvatarURL   *string    `json:"avatar_url"`
	Status      string     `json:"status"`
	LastLoginAt *time.Time `json:"last_login_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// RefreshTokenRequest 刷新令牌请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// TokenClaims JWT 令牌声明
type TokenClaims struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	SessionID string `json:"session_id"`
}

// PermissionCheckRequest 权限检查请求
type PermissionCheckRequest struct {
	Resource    string  `json:"resource" validate:"required"`
	Action      string  `json:"action" validate:"required"`
	ClusterName *string `json:"cluster_name"`
	Namespace   *string `json:"namespace"`
}

// PermissionCheckResponse 权限检查响应
type PermissionCheckResponse struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason,omitempty"`
}

// AuditLogListFilters 审计日志列表过滤器
type AuditLogListFilters struct {
	UserID       string    `json:"user_id" form:"user_id"`
	Action       string    `json:"action" form:"action"`
	ResourceType string    `json:"resource_type" form:"resource_type"`
	ClusterName  string    `json:"cluster_name" form:"cluster_name"`
	Status       string    `json:"status" form:"status"`
	StartTime    time.Time `json:"start_time" form:"start_time"`
	EndTime      time.Time `json:"end_time" form:"end_time"`
}

// UserSessionListFilters 用户会话列表过滤器
type UserSessionListFilters struct {
	UserID    string `json:"user_id" form:"user_id"`
	IPAddress string `json:"ip_address" form:"ip_address"`
	Active    *bool  `json:"active" form:"active"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username    string `json:"username" validate:"required,min=3,max=50"`
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=8"`
	DisplayName string `json:"display_name" validate:"max=100"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

// JWTClaims JWT 声明
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}
