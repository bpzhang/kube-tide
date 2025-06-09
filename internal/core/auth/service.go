package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Service provides authentication and authorization services
type Service struct {
	store      *Store
	jwtSecret  []byte
	jwtExpires time.Duration
}

// JWTClaims represents JWT claims for authentication
type JWTClaims struct {
	UserID   string   `json:"userId"`
	Username string   `json:"username"`
	RoleIDs  []string `json:"roleIds"`
	jwt.RegisteredClaims
}

// NewService creates a new auth service
func NewService(store *Store, jwtSecret string, jwtExpires time.Duration) *Service {
	if jwtExpires == 0 {
		jwtExpires = 24 * time.Hour // Default to 24 hours
	}

	return &Service{
		store:      store,
		jwtSecret:  []byte(jwtSecret),
		jwtExpires: jwtExpires,
	}
}

// Login authenticates a user and returns a JWT token
func (s *Service) Login(username, password string) (string, error) {
	user, err := s.store.AuthenticateUser(username, password)
	if err != nil {
		return "", err
	}

	// Update last login time
	user.LastLogin = time.Now()
	if err := s.store.UpdateUser(user); err != nil {
		return "", err
	}

	// Create JWT token
	token, err := s.generateJWT(user)
	if err != nil {
		return "", err
	}

	return token, nil
}

// ValidateToken validates a JWT token and returns the user claims
func (s *Service) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// CheckPermission checks if a user has permission to perform an action on a resource
func (s *Service) CheckPermission(userID, resource, action string) (bool, error) {
	// Get user roles
	roles, err := s.store.GetUserRoles(userID)
	if err != nil {
		return false, err
	}

	// Check permissions in each role
	for _, role := range roles {
		for _, permission := range role.Permissions {
			// Check for wildcard permissions
			if (permission.Resource == "*" || permission.Resource == resource) &&
				(permission.Action == "*" || permission.Action == action) {
				return true, nil
			}
		}
	}

	return false, nil
}

// generateJWT generates a JWT token for a user
func (s *Service) generateJWT(user *User) (string, error) {
	// Set expiration time
	expirationTime := time.Now().Add(s.jwtExpires)

	// Create claims
	claims := &JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		RoleIDs:  user.RoleIDs,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID,
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GetUserRoles gets all roles assigned to a user
func (s *Service) GetUserRoles(userID string) ([]*Role, error) {
	return s.store.GetUserRoles(userID)
}

// HasAdminRole checks if a user has the admin role
func (s *Service) HasAdminRole(userID string) (bool, error) {
	roles, err := s.store.GetUserRoles(userID)
	if err != nil {
		return false, err
	}

	for _, role := range roles {
		if role.ID == "admin" {
			return true, nil
		}
	}

	return false, nil
}

// GetUsers returns all users (only accessible to admins)
func (s *Service) GetUsers() ([]*User, error) {
	return s.store.GetUsers(), nil
}

// GetUser returns a user by ID
func (s *Service) GetUser(id string) (*User, error) {
	return s.store.GetUserByID(id)
}

// CreateUser creates a new user
func (s *Service) CreateUser(user *User) error {
	return s.store.AddUser(user)
}

// UpdateUser updates an existing user
func (s *Service) UpdateUser(user *User) error {
	return s.store.UpdateUser(user)
}

// DeleteUser deletes a user
func (s *Service) DeleteUser(id string) error {
	return s.store.DeleteUser(id)
}

// GetRoles returns all roles
func (s *Service) GetRoles() ([]*Role, error) {
	return s.store.GetRoles(), nil
}

// GetRole returns a role by ID
func (s *Service) GetRole(id string) (*Role, error) {
	return s.store.GetRoleByID(id)
}

// CreateRole creates a new role
func (s *Service) CreateRole(role *Role) error {
	return s.store.AddRole(role)
}

// UpdateRole updates an existing role
func (s *Service) UpdateRole(role *Role) error {
	return s.store.UpdateRole(role)
}

// DeleteRole deletes a role
func (s *Service) DeleteRole(id string) error {
	return s.store.DeleteRole(id)
}
