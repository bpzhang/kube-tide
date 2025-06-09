package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	// DefaultStorageDir is the default directory for storing auth data
	DefaultStorageDir = "./data/auth"
	// UsersFilename is the filename for users data
	UsersFilename = "users.json"
	// RolesFilename is the filename for roles data
	RolesFilename = "roles.json"
)

// Store provides storage and retrieval of auth data
type Store struct {
	storageDir string
	users      []*User
	roles      []*Role
	mu         sync.RWMutex
}

// NewStore creates a new auth store
func NewStore(storageDir string) (*Store, error) {
	if storageDir == "" {
		storageDir = DefaultStorageDir
	}

	store := &Store{
		storageDir: storageDir,
		users:      []*User{},
		roles:      []*Role{},
	}

	// Ensure storage directory exists
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Load existing data or initialize with defaults
	if err := store.loadOrInitialize(); err != nil {
		return nil, err
	}

	return store, nil
}

// loadOrInitialize loads existing auth data or initializes with defaults
func (s *Store) loadOrInitialize() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Try to load roles
	rolesFile := filepath.Join(s.storageDir, RolesFilename)
	if _, err := os.Stat(rolesFile); os.IsNotExist(err) {
		// Initialize with default roles
		s.roles = DefaultRoles()
		if err := s.saveRoles(); err != nil {
			return fmt.Errorf("failed to save default roles: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to check roles file: %w", err)
	} else {
		// Load existing roles
		data, err := os.ReadFile(rolesFile)
		if err != nil {
			return fmt.Errorf("failed to read roles file: %w", err)
		}
		if err := json.Unmarshal(data, &s.roles); err != nil {
			return fmt.Errorf("failed to parse roles data: %w", err)
		}
	}

	// Try to load users
	usersFile := filepath.Join(s.storageDir, UsersFilename)
	if _, err := os.Stat(usersFile); os.IsNotExist(err) {
		// Initialize with default admin user
		adminPassword, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash default admin password: %w", err)
		}

		admin := NewUser("admin", string(adminPassword), "admin@example.com", "Administrator")
		admin.RoleIDs = []string{"admin"}
		s.users = []*User{admin}

		if err := s.saveUsers(); err != nil {
			return fmt.Errorf("failed to save default users: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to check users file: %w", err)
	} else {
		// Load existing users
		data, err := os.ReadFile(usersFile)
		if err != nil {
			return fmt.Errorf("failed to read users file: %w", err)
		}
		if err := json.Unmarshal(data, &s.users); err != nil {
			return fmt.Errorf("failed to parse users data: %w", err)
		}
	}

	return nil
}

// saveUsers saves users data to file
func (s *Store) saveUsers() error {
	data, err := json.MarshalIndent(s.users, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal users data: %w", err)
	}

	usersFile := filepath.Join(s.storageDir, UsersFilename)
	if err := os.WriteFile(usersFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write users file: %w", err)
	}

	return nil
}

// saveRoles saves roles data to file
func (s *Store) saveRoles() error {
	data, err := json.MarshalIndent(s.roles, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal roles data: %w", err)
	}

	rolesFile := filepath.Join(s.storageDir, RolesFilename)
	if err := os.WriteFile(rolesFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write roles file: %w", err)
	}

	return nil
}

// GetUsers returns all users
func (s *Store) GetUsers() []*User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to avoid race conditions
	usersCopy := make([]*User, len(s.users))
	copy(usersCopy, s.users)
	return usersCopy
}

// GetUserByID returns a user by ID
func (s *Store) GetUserByID(id string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.users {
		if user.ID == id {
			return user, nil
		}
	}

	return nil, errors.New("user not found")
}

// GetUserByUsername returns a user by username
func (s *Store) GetUserByUsername(username string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.users {
		if user.Username == username {
			return user, nil
		}
	}

	return nil, errors.New("user not found")
}

// AddUser adds a new user
func (s *Store) AddUser(user *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if username already exists
	for _, existingUser := range s.users {
		if existingUser.Username == user.Username {
			return errors.New("username already exists")
		}
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	user.Password = string(hashedPassword)

	s.users = append(s.users, user)
	return s.saveUsers()
}

// UpdateUser updates an existing user
func (s *Store) UpdateUser(user *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existingUser := range s.users {
		if existingUser.ID == user.ID {
			// Don't update password if it's empty
			if user.Password != "" {
				hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
				if err != nil {
					return fmt.Errorf("failed to hash password: %w", err)
				}
				user.Password = string(hashedPassword)
			} else {
				user.Password = existingUser.Password
			}

			user.UpdatedAt = time.Now()
			s.users[i] = user
			return s.saveUsers()
		}
	}

	return errors.New("user not found")
}

// DeleteUser deletes a user by ID
func (s *Store) DeleteUser(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, user := range s.users {
		if user.ID == id {
			// Remove user from slice
			s.users = append(s.users[:i], s.users[i+1:]...)
			return s.saveUsers()
		}
	}

	return errors.New("user not found")
}

// AuthenticateUser authenticates a user by username and password
func (s *Store) AuthenticateUser(username, password string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.users {
		if user.Username == username && user.Active {
			// Compare password hash
			if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err == nil {
				return user, nil
			}
			return nil, errors.New("invalid password")
		}
	}

	return nil, errors.New("user not found or inactive")
}

// GetRoles returns all roles
func (s *Store) GetRoles() []*Role {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to avoid race conditions
	rolesCopy := make([]*Role, len(s.roles))
	copy(rolesCopy, s.roles)
	return rolesCopy
}

// GetRoleByID returns a role by ID
func (s *Store) GetRoleByID(id string) (*Role, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, role := range s.roles {
		if role.ID == id {
			return role, nil
		}
	}

	return nil, errors.New("role not found")
}

// AddRole adds a new role
func (s *Store) AddRole(role *Role) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if role ID already exists
	for _, existingRole := range s.roles {
		if existingRole.ID == role.ID {
			return errors.New("role ID already exists")
		}
	}

	s.roles = append(s.roles, role)
	return s.saveRoles()
}

// UpdateRole updates an existing role
func (s *Store) UpdateRole(role *Role) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existingRole := range s.roles {
		if existingRole.ID == role.ID {
			role.UpdatedAt = time.Now()
			s.roles[i] = role
			return s.saveRoles()
		}
	}

	return errors.New("role not found")
}

// DeleteRole deletes a role by ID
func (s *Store) DeleteRole(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if role is in use by any user
	for _, user := range s.users {
		for _, roleID := range user.RoleIDs {
			if roleID == id {
				return errors.New("role is assigned to users and cannot be deleted")
			}
		}
	}

	for i, role := range s.roles {
		if role.ID == id {
			// Remove role from slice
			s.roles = append(s.roles[:i], s.roles[i+1:]...)
			return s.saveRoles()
		}
	}

	return errors.New("role not found")
}

// GetUserRoles returns all roles assigned to a user
func (s *Store) GetUserRoles(userID string) ([]*Role, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var user *User
	for _, u := range s.users {
		if u.ID == userID {
			user = u
			break
		}
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	userRoles := []*Role{}
	for _, roleID := range user.RoleIDs {
		for _, role := range s.roles {
			if role.ID == roleID {
				userRoles = append(userRoles, role)
				break
			}
		}
	}

	return userRoles, nil
}
