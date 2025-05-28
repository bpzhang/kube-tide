package repository

import (
	"go.uber.org/zap"

	"kube-tide/internal/database"
)

// Repositories holds all repository instances
type Repositories struct {
	Cluster     ClusterRepository
	Node        NodeRepository
	Pod         PodRepository
	Namespace   NamespaceRepository
	Deployment  DeploymentRepository
	Service     ServiceRepository
	User        UserRepository
	Role        RoleRepository
	Permission  PermissionRepository
	UserSession UserSessionRepository
	AuditLog    AuditLogRepository
}

// NewRepositories creates a new repositories instance with all repositories
func NewRepositories(db *database.Database, logger *zap.Logger) *Repositories {
	return &Repositories{
		Cluster:     NewClusterRepository(db, logger),
		Node:        NewNodeRepository(db, logger),
		Pod:         NewPodRepository(db, logger),
		Namespace:   NewNamespaceRepository(db, logger),
		Deployment:  NewDeploymentRepository(db, logger),
		Service:     NewServiceRepository(db, logger),
		User:        NewUserRepository(db.DB(), logger),
		Role:        NewRoleRepository(db.DB(), logger),
		Permission:  NewPermissionRepository(db.DB(), logger),
		UserSession: NewUserSessionRepository(db.DB(), logger),
		AuditLog:    NewAuditLogRepository(db.DB(), logger),
	}
}
