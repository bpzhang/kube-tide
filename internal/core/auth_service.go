package core

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"kube-tide/internal/database/models"
	"kube-tide/internal/repository"
	"kube-tide/internal/utils"
)

// AuthService handles authentication and authorization
type AuthService struct {
	userRepo    repository.UserRepository
	sessionRepo repository.UserSessionRepository
	auditRepo   repository.AuditLogRepository
	roleRepo    repository.RoleRepository
	jwtSecret   string
	logger      *zap.Logger
}

// NewAuthService creates a new authentication service
func NewAuthService(
	userRepo repository.UserRepository,
	sessionRepo repository.UserSessionRepository,
	auditRepo repository.AuditLogRepository,
	roleRepo repository.RoleRepository,
	jwtSecret string,
	logger *zap.Logger,
) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		auditRepo:   auditRepo,
		roleRepo:    roleRepo,
		jwtSecret:   jwtSecret,
		logger:      logger,
	}
}

// Login authenticates a user and returns a JWT token
func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	// Find user by username or email
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		// Try by email if username not found
		if user, err = s.userRepo.GetByEmail(ctx, req.Username); err != nil {
			s.logAuditEvent(ctx, "", "login_failed", "user_not_found", req.Username)
			return nil, utils.ErrInvalidCredentials
		}
	}

	// Check if user is active
	if user.Status != models.UserStatusActive {
		s.logAuditEvent(ctx, user.ID, "login_failed", "user_inactive", req.Username)
		return nil, utils.ErrUserInactive
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		s.logAuditEvent(ctx, user.ID, "login_failed", "invalid_password", req.Username)
		return nil, utils.ErrInvalidCredentials
	}

	// Generate JWT token
	token, expiresAt, err := s.generateJWTToken(user)
	if err != nil {
		s.logger.Error("Failed to generate JWT token", zap.Error(err))
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Create session
	session := &models.UserSession{
		ID:        generateSessionID(),
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: expiresAt,
		IPAddress: &req.IPAddress,
		UserAgent: &req.UserAgent,
		CreatedAt: time.Now(),
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		s.logger.Error("Failed to create session", zap.Error(err))
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Update last login time
	if err := s.userRepo.Update(ctx, user.ID, models.UserUpdateRequest{
		UpdateTime: session.CreatedAt,
	}); err != nil {
		s.logger.Warn("Failed to update last login time", zap.Error(err))
	}

	// Log successful login
	s.logAuditEvent(ctx, user.ID, "login_success", "user_logged_in", req.Username)

	return &models.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User: &models.UserResponse{
			ID:          user.ID,
			Username:    user.Username,
			Email:       user.Email,
			DisplayName: user.DisplayName,
			Status:      user.Status,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
		},
	}, nil
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, req *models.RegisterRequest) (*models.UserResponse, error) {
	// Check if username already exists
	if _, err := s.userRepo.GetByUsername(ctx, req.Username); err == nil {
		return nil, utils.ErrUsernameExists
	}

	// Check if email already exists
	if _, err := s.userRepo.GetByEmail(ctx, req.Email); err == nil {
		return nil, utils.ErrEmailExists
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		ID:           generateUserID(),
		Username:     req.Username,
		Email:        req.Email,
		DisplayName:  &req.DisplayName,
		PasswordHash: string(passwordHash),
		Status:       models.UserStatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Log user registration
	s.logAuditEvent(ctx, user.ID, "user_registered", "new_user_created", req.Username)

	return &models.UserResponse{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Status:      user.Status,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}, nil
}

// Logout invalidates a user session
func (s *AuthService) Logout(ctx context.Context, token string) error {
	session, err := s.sessionRepo.GetByTokenHash(ctx, token)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	if err := s.sessionRepo.Delete(ctx, session.ID); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	// Log logout
	s.logAuditEvent(ctx, session.UserID, "logout", "user_logged_out", "")

	return nil
}

// ValidateToken validates a JWT token and returns the user
func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (*models.User, error) {
	// Parse JWT token
	token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*models.JWTClaims)
	if !ok || !token.Valid {
		return nil, utils.ErrInvalidToken
	}

	// Check if session exists and is valid
	session, err := s.sessionRepo.GetByTokenHash(ctx, tokenString)
	if err != nil {
		return nil, utils.ErrSessionNotFound
	}

	if session.ExpiresAt.Before(time.Now()) {
		// Clean up expired session
		s.sessionRepo.Delete(ctx, session.ID)
		return nil, utils.ErrSessionExpired
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if user.Status != models.UserStatusActive {
		return nil, utils.ErrUserInactive
	}

	return user, nil
}

// ChangePassword changes a user's password
func (s *AuthService) ChangePassword(ctx context.Context, userID string, req *models.ChangePasswordRequest) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		s.logAuditEvent(ctx, userID, "password_change_failed", "invalid_current_password", user.Username)
		return utils.ErrInvalidCredentials
	}

	// Hash new password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	user.PasswordHash = string(passwordHash)
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user.ID, models.UserUpdateRequest{
		Password: &user.PasswordHash,
		UpdateTime: time.Now(),
	}); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Invalidate all existing sessions for this user
	if err := s.sessionRepo.DeleteByUserID(ctx, userID); err != nil {
		s.logger.Warn("Failed to invalidate user sessions", zap.Error(err))
	}

	// Log password change
	s.logAuditEvent(ctx, userID, "password_changed", "user_password_updated", user.Username)

	return nil
}

// HasPermission checks if a user has a specific permission
func (s *AuthService) HasPermission(ctx context.Context, userID, permission, resource, scope string) (bool, error) {
	return s.roleRepo.UserHasPermission(ctx, userID, permission, resource, scope)
}

// GetUserPermissions returns all permissions for a user
func (s *AuthService) GetUserPermissions(ctx context.Context, userID string) ([]*models.Permission, error) {
	return s.roleRepo.GetUserPermissions(ctx, userID)
}

// generateJWTToken generates a JWT token for a user
func (s *AuthService) generateJWTToken(user *models.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(24 * time.Hour) // Token expires in 24 hours

	claims := &models.JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "kube-tide",
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// logAuditEvent logs an audit event
func (s *AuthService) logAuditEvent(ctx context.Context, userID, action, result, details string) {
	auditLog := &models.AuditLog{
		ID:        generateAuditID(),
		UserID:    userID,
		Action:    action,
		Resource:  "auth",
		Result:    result,
		Details:   details,
		Timestamp: time.Now(),
	}

	if err := s.auditRepo.Create(ctx, auditLog); err != nil {
		s.logger.Error("Failed to create audit log", zap.Error(err))
	}
}

// generateSessionID generates a random session ID
func generateSessionID() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// generateUserID generates a random user ID
func generateUserID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// generateAuditID generates a random audit log ID
func generateAuditID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
} 