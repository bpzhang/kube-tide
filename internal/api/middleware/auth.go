package middleware

import (
	"kube-tide/internal/core/auth"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	// AuthUserKey is the key used to store the authenticated user in the context
	AuthUserKey = "authUser"
	// AuthTokenKey is the key used to store the auth token in the context
	AuthTokenKey = "authToken"
)

// AuthRequired middleware ensures that a valid JWT token is provided
func AuthRequired(authService *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the token from the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			return
		}

		// The Authorization header should be in the format "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header format must be Bearer <token>",
			})
			return
		}

		tokenString := parts[1]

		// Validate the token
		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			return
		}

		// Store the token and user claims in the context for later use
		c.Set(AuthTokenKey, tokenString)
		c.Set(AuthUserKey, claims)

		c.Next()
	}
}

// RequirePermission middleware ensures that the user has the required permission
func RequirePermission(authService *auth.Service, resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user claims from context
		claims, exists := c.Get(AuthUserKey)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "User not authenticated",
			})
			return
		}

		userClaims := claims.(*auth.JWTClaims)

		// Check if user has permission
		hasPermission, err := authService.CheckPermission(userClaims.UserID, resource, action)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to check permissions",
			})
			return
		}

		if !hasPermission {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Insufficient permissions",
			})
			return
		}

		c.Next()
	}
}

// RequireAdmin middleware ensures that the user has admin role
func RequireAdmin(authService *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user claims from context
		claims, exists := c.Get(AuthUserKey)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "User not authenticated",
			})
			return
		}

		userClaims := claims.(*auth.JWTClaims)

		// Check if user has admin role
		isAdmin, err := authService.HasAdminRole(userClaims.UserID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to check admin status",
			})
			return
		}

		if !isAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Admin privileges required",
			})
			return
		}

		c.Next()
	}
}

// GetCurrentUser returns the current authenticated user from the context
func GetCurrentUser(c *gin.Context) *auth.JWTClaims {
	claims, exists := c.Get(AuthUserKey)
	if !exists {
		return nil
	}

	return claims.(*auth.JWTClaims)
}
