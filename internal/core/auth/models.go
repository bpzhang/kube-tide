package auth

import (
	"time"

	"github.com/google/uuid"
)

// User represents a system user
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"` // Password is not exposed in JSON
	Email     string    `json:"email"`
	FullName  string    `json:"fullName"`
	RoleIDs   []string  `json:"roleIds"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	LastLogin time.Time `json:"lastLogin,omitempty"`
}

// NewUser creates a new user with default values
func NewUser(username, password, email, fullName string) *User {
	now := time.Now()
	return &User{
		ID:        uuid.New().String(),
		Username:  username,
		Password:  password, // Note: This should be hashed before storing
		Email:     email,
		FullName:  fullName,
		RoleIDs:   []string{},
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Role represents a user role with specific permissions
type Role struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Permissions []Permission `json:"permissions"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
}

// NewRole creates a new role with default values
func NewRole(name, description string, permissions []Permission) *Role {
	now := time.Now()
	return &Role{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		Permissions: permissions,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Permission represents a specific action that can be performed on a resource
type Permission struct {
	Resource string `json:"resource"` // e.g., "cluster", "node", "pod"
	Action   string `json:"action"`   // e.g., "create", "read", "update", "delete"
}

// DefaultRoles provides predefined roles for the system
func DefaultRoles() []*Role {
	now := time.Now()

	// Admin role with all permissions
	adminRole := &Role{
		ID:          "admin",
		Name:        "Administrator",
		Description: "Full access to all system features",
		Permissions: []Permission{
			{Resource: "*", Action: "*"}, // Wildcard permission for all resources and actions
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Operator role with management permissions but no user/role management
	operatorRole := &Role{
		ID:          "operator",
		Name:        "Operator",
		Description: "Can manage all Kubernetes resources but not users and roles",
		Permissions: []Permission{
			{Resource: "cluster", Action: "*"},
			{Resource: "node", Action: "*"},
			{Resource: "pod", Action: "*"},
			{Resource: "service", Action: "*"},
			{Resource: "deployment", Action: "*"},
			{Resource: "statefulset", Action: "*"},
			{Resource: "nodepool", Action: "*"},
			{Resource: "namespace", Action: "*"},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Viewer role with read-only permissions
	viewerRole := &Role{
		ID:          "viewer",
		Name:        "Viewer",
		Description: "Read-only access to all Kubernetes resources",
		Permissions: []Permission{
			{Resource: "cluster", Action: "read"},
			{Resource: "node", Action: "read"},
			{Resource: "pod", Action: "read"},
			{Resource: "service", Action: "read"},
			{Resource: "deployment", Action: "read"},
			{Resource: "statefulset", Action: "read"},
			{Resource: "nodepool", Action: "read"},
			{Resource: "namespace", Action: "read"},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	return []*Role{adminRole, operatorRole, viewerRole}
}
