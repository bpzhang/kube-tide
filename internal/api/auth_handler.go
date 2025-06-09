package api

import (
	"kube-tide/internal/api/middleware"
	"kube-tide/internal/core/auth"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication and user/role management
type AuthHandler struct {
	authService *auth.Service
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(authService *auth.Service) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Login handles user login and returns a JWT token
func (h *AuthHandler) Login(c *gin.Context) {
	var loginRequest struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	token, err := h.authService.Login(loginRequest.Username, loginRequest.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

// GetCurrentUserInfo returns information about the current user
func (h *AuthHandler) GetCurrentUserInfo(c *gin.Context) {
	// Get user claims from context
	userClaims := middleware.GetCurrentUser(c)
	if userClaims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get detailed user info
	user, err := h.authService.GetUser(userClaims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
		return
	}

	// Get user roles
	roles, err := h.authService.GetUserRoles(userClaims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user roles"})
		return
	}

	// Format role names for response
	roleNames := make([]string, len(roles))
	for i, role := range roles {
		roleNames[i] = role.Name
	}

	c.JSON(http.StatusOK, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"fullName": user.FullName,
		"roles":    roleNames,
	})
}

// ListUsers returns a list of all users (admin only)
func (h *AuthHandler) ListUsers(c *gin.Context) {
	users, err := h.authService.GetUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get users"})
		return
	}

	c.JSON(http.StatusOK, users)
}

// GetUser returns details of a specific user (admin only)
func (h *AuthHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")
	user, err := h.authService.GetUser(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// CreateUser creates a new user (admin only)
func (h *AuthHandler) CreateUser(c *gin.Context) {
	var newUser auth.User
	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate required fields
	if newUser.Username == "" || newUser.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and password are required"})
		return
	}

	// Create a new user
	user := auth.NewUser(newUser.Username, newUser.Password, newUser.Email, newUser.FullName)
	user.RoleIDs = newUser.RoleIDs

	if err := h.authService.CreateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":       user.ID,
		"username": user.Username,
	})
}

// UpdateUser updates an existing user (admin only, or self-update)
func (h *AuthHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")

	// Get current user
	currentUser := middleware.GetCurrentUser(c)

	// Check if user is updating themselves or is an admin
	isAdmin, _ := h.authService.HasAdminRole(currentUser.UserID)
	isSelf := currentUser.UserID == userID

	if !isAdmin && !isSelf {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only update your own account unless you're an admin"})
		return
	}

	// Get existing user
	existingUser, err := h.authService.GetUser(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Bind update data
	var updateData struct {
		Email    string   `json:"email"`
		FullName string   `json:"fullName"`
		Password string   `json:"password"`
		RoleIDs  []string `json:"roleIds"`
		Active   *bool    `json:"active"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Update fields if provided
	if updateData.Email != "" {
		existingUser.Email = updateData.Email
	}
	if updateData.FullName != "" {
		existingUser.FullName = updateData.FullName
	}
	if updateData.Password != "" {
		existingUser.Password = updateData.Password
	}
	// Only allow admins to update roles
	if isAdmin && updateData.RoleIDs != nil {
		existingUser.RoleIDs = updateData.RoleIDs
	}
	// Only allow admins to update active status
	if isAdmin && updateData.Active != nil {
		existingUser.Active = *updateData.Active
	}

	if err := h.authService.UpdateUser(existingUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":       existingUser.ID,
		"username": existingUser.Username,
		"updated":  true,
	})
}

// DeleteUser deletes a user (admin only)
func (h *AuthHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")

	// Get current user
	currentUser := middleware.GetCurrentUser(c)

	// Don't allow users to delete themselves
	if currentUser.UserID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You cannot delete your own account"})
		return
	}

	if err := h.authService.DeleteUser(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"deleted": true,
	})
}

// ListRoles returns a list of all roles
func (h *AuthHandler) ListRoles(c *gin.Context) {
	roles, err := h.authService.GetRoles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get roles"})
		return
	}

	c.JSON(http.StatusOK, roles)
}

// GetRole returns details of a specific role
func (h *AuthHandler) GetRole(c *gin.Context) {
	roleID := c.Param("id")
	role, err := h.authService.GetRole(roleID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}

	c.JSON(http.StatusOK, role)
}

// CreateRole creates a new role (admin only)
func (h *AuthHandler) CreateRole(c *gin.Context) {
	var newRole auth.Role
	if err := c.ShouldBindJSON(&newRole); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate required fields
	if newRole.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Role name is required"})
		return
	}

	// Create a new role
	role := auth.NewRole(newRole.Name, newRole.Description, newRole.Permissions)

	if err := h.authService.CreateRole(role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create role"})
		return
	}

	c.JSON(http.StatusCreated, role)
}

// UpdateRole updates an existing role (admin only)
func (h *AuthHandler) UpdateRole(c *gin.Context) {
	roleID := c.Param("id")

	// Get existing role
	existingRole, err := h.authService.GetRole(roleID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}

	// Bind update data
	var updateData struct {
		Name        string            `json:"name"`
		Description string            `json:"description"`
		Permissions []auth.Permission `json:"permissions"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Update fields if provided
	if updateData.Name != "" {
		existingRole.Name = updateData.Name
	}
	if updateData.Description != "" {
		existingRole.Description = updateData.Description
	}
	if updateData.Permissions != nil {
		existingRole.Permissions = updateData.Permissions
	}

	if err := h.authService.UpdateRole(existingRole); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update role"})
		return
	}

	c.JSON(http.StatusOK, existingRole)
}

// DeleteRole deletes a role (admin only)
func (h *AuthHandler) DeleteRole(c *gin.Context) {
	roleID := c.Param("id")

	// Don't allow deleting the default roles
	if roleID == "admin" || roleID == "operator" || roleID == "viewer" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete default roles"})
		return
	}

	if err := h.authService.DeleteRole(roleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"deleted": true,
	})
}
