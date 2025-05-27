package repository

import (
	"go.uber.org/zap"

	"kube-tide/internal/database"
)

// Repositories holds all repository instances
type Repositories struct {
	Cluster    ClusterRepository
	Node       NodeRepository
	Pod        PodRepository
	Namespace  NamespaceRepository
	Deployment DeploymentRepository
	Service    ServiceRepository
}

// NewRepositories creates a new repositories instance with all repositories
func NewRepositories(db *database.Database, logger *zap.Logger) *Repositories {
	return &Repositories{
		Cluster:    NewClusterRepository(db, logger),
		Node:       NewNodeRepository(db, logger),
		Pod:        NewPodRepository(db, logger),
		Namespace:  NewNamespaceRepository(db, logger),
		Deployment: NewDeploymentRepository(db, logger),
		Service:    NewServiceRepository(db, logger),
	}
}
