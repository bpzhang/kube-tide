package middleware

import (
	"context"
	"net/http"
	"strings"

	"kube-tide/internal/core"
	"kube-tide/internal/database/models"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	UserContextKey = "user"
	UserIDKey      = "user_id"
)

// AuthMiddleware 认证中间件
type AuthMiddleware struct {
	authService *core.AuthService
	logger      *zap.Logger
}

// NewAuthMiddleware 创建认证中间件
func NewAuthMiddleware(authService *core.AuthService, logger *zap.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		logger:      logger,
	}
}

// RequireAuth 需要认证的中间件
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "Missing authorization token",
			})
			c.Abort()
			return
		}

		user, err := m.authService.ValidateToken(c.Request.Context(), token)
		if err != nil {
			m.logger.Warn("Token validation failed", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set(UserContextKey, user)
		c.Set(UserIDKey, user.ID)
		c.Next()
	}
}

// RequirePermission 需要特定权限的中间件
func (m *AuthMiddleware) RequirePermission(permission, resource string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetCurrentUser(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "User not authenticated",
			})
			c.Abort()
			return
		}

		// 从路径参数中获取作用域信息
		clusterName := c.Param("cluster")
		namespace := c.Param("namespace")

		// 检查权限
		hasPermission, err := m.checkUserPermission(c.Request.Context(), user.ID, permission, resource, clusterName, namespace)
		if err != nil {
			m.logger.Error("Permission check failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    http.StatusInternalServerError,
				"message": "Permission check failed",
			})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    http.StatusForbidden,
				"message": "Insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole 需要特定角色的中间件
func (m *AuthMiddleware) RequireRole(roleNames ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetCurrentUser(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "User not authenticated",
			})
			c.Abort()
			return
		}

		// 检查用户是否具有指定角色
		hasRole, err := m.checkUserRole(c.Request.Context(), user.ID, roleNames...)
		if err != nil {
			m.logger.Error("Role check failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    http.StatusInternalServerError,
				"message": "Role check failed",
			})
			c.Abort()
			return
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    http.StatusForbidden,
				"message": "Insufficient role permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// OptionalAuth 可选认证中间件（不强制要求认证）
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token != "" {
			user, err := m.authService.ValidateToken(c.Request.Context(), token)
			if err == nil {
				c.Set(UserContextKey, user)
				c.Set(UserIDKey, user.ID)
			}
		}
		c.Next()
	}
}

// checkUserPermission 检查用户权限
func (m *AuthMiddleware) checkUserPermission(ctx context.Context, userID, permission, resource, clusterName, namespace string) (bool, error) {
	// 构建作用域
	scope := "global"
	if clusterName != "" {
		if namespace != "" {
			scope = "namespace"
		} else {
			scope = "cluster"
		}
	}

	scopeValue := ""
	if scope == "cluster" {
		scopeValue = clusterName
	} else if scope == "namespace" {
		scopeValue = clusterName + "/" + namespace
	}

	return m.authService.HasPermission(ctx, userID, permission, resource, scopeValue)
}

// checkUserRole 检查用户角色
func (m *AuthMiddleware) checkUserRole(ctx context.Context, userID string, roleNames ...string) (bool, error) {
	// 这里需要实现角色检查逻辑
	// 暂时返回 true，后续可以扩展
	return true, nil
}

// extractToken 从请求中提取令牌
func extractToken(c *gin.Context) string {
	// 从 Authorization header 中提取
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}

	// 从查询参数中提取
	token := c.Query("token")
	if token != "" {
		return token
	}

	// 从 Cookie 中提取
	cookie, err := c.Cookie("token")
	if err == nil && cookie != "" {
		return cookie
	}

	return ""
}

// GetCurrentUser 从上下文中获取当前用户
func GetCurrentUser(c *gin.Context) *models.User {
	if user, exists := c.Get(UserContextKey); exists {
		if u, ok := user.(*models.User); ok {
			return u
		}
	}
	return nil
}

// GetCurrentUserID 从上下文中获取当前用户ID
func GetCurrentUserID(c *gin.Context) string {
	if userID, exists := c.Get(UserIDKey); exists {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return ""
}

// MustGetCurrentUser 从上下文中获取当前用户（必须存在）
func MustGetCurrentUser(c *gin.Context) *models.User {
	user := GetCurrentUser(c)
	if user == nil {
		panic("user not found in context")
	}
	return user
}

// MustGetCurrentUserID 从上下文中获取当前用户ID（必须存在）
func MustGetCurrentUserID(c *gin.Context) string {
	userID := GetCurrentUserID(c)
	if userID == "" {
		panic("user ID not found in context")
	}
	return userID
}
