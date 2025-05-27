package api

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"kube-tide/internal/core"
)

// DBRouter handles database API routes
type DBRouter struct {
	coreService *core.Service
	logger      *zap.Logger
}

// NewDBRouter creates a new database router
func NewDBRouter(coreService *core.Service, logger *zap.Logger) *DBRouter {
	return &DBRouter{
		coreService: coreService,
		logger:      logger,
	}
}

// SetupRoutes sets up all database API routes
func (r *DBRouter) SetupRoutes(router *gin.Engine) {
	// API v1 group for database operations
	v1 := router.Group("/api/v1/db")
	
	// Setup deployment routes
	r.setupDeploymentRoutes(v1)
	
	// Setup service routes
	r.setupServiceRoutes(v1)
}

// setupDeploymentRoutes sets up deployment-related routes
func (r *DBRouter) setupDeploymentRoutes(v1 *gin.RouterGroup) {
	deploymentHandler := NewDBDeploymentHandler(r.coreService.DeploymentService(), r.logger)
	deployments := v1.Group("/deployments")
	{
		deployments.POST("", deploymentHandler.CreateDeployment)
		deployments.GET("/:id", deploymentHandler.GetDeployment)
		deployments.PUT("/:id", deploymentHandler.UpdateDeployment)
		deployments.DELETE("/:id", deploymentHandler.DeleteDeployment)
	}
	
	// Cluster-specific deployment routes
	clusterDeployments := v1.Group("/clusters/:cluster_id/deployments")
	{
		clusterDeployments.GET("", deploymentHandler.ListDeploymentsByCluster)
		clusterDeployments.DELETE("", deploymentHandler.DeleteDeploymentsByCluster)
		clusterDeployments.GET("/count", deploymentHandler.CountDeploymentsByCluster)
	}
	
	// Namespace-specific deployment routes
	namespaceDeployments := v1.Group("/clusters/:cluster_id/namespaces/:namespace/deployments")
	{
		namespaceDeployments.GET("", deploymentHandler.ListDeploymentsByNamespace)
		namespaceDeployments.GET("/:name", deploymentHandler.GetDeploymentByName)
		namespaceDeployments.DELETE("", deploymentHandler.DeleteDeploymentsByNamespace)
		namespaceDeployments.GET("/count", deploymentHandler.CountDeploymentsByNamespace)
	}
}

// setupServiceRoutes sets up service-related routes
func (r *DBRouter) setupServiceRoutes(v1 *gin.RouterGroup) {
	serviceHandler := NewDBServiceHandler(r.coreService.ServiceService(), r.logger)
	services := v1.Group("/services")
	{
		services.POST("", serviceHandler.CreateService)
		services.GET("/:id", serviceHandler.GetService)
		services.PUT("/:id", serviceHandler.UpdateService)
		services.DELETE("/:id", serviceHandler.DeleteService)
	}
	
	// Cluster-specific service routes
	clusterServices := v1.Group("/clusters/:cluster_id/services")
	{
		clusterServices.GET("", serviceHandler.ListServicesByCluster)
		clusterServices.DELETE("", serviceHandler.DeleteServicesByCluster)
		clusterServices.GET("/count", serviceHandler.CountServicesByCluster)
	}
	
	// Namespace-specific service routes
	namespaceServices := v1.Group("/clusters/:cluster_id/namespaces/:namespace/services")
	{
		namespaceServices.GET("", serviceHandler.ListServicesByNamespace)
		namespaceServices.GET("/:name", serviceHandler.GetServiceByName)
		namespaceServices.DELETE("", serviceHandler.DeleteServicesByNamespace)
		namespaceServices.GET("/count", serviceHandler.CountServicesByNamespace)
	}
} 