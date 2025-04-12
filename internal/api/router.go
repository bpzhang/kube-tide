package api

import (
	"kube-tide/configs"
	"kube-tide/internal/api/middleware"
	"kube-tide/pkg/embed"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// App Application structure
type App struct {
	ClusterHandler     *ClusterHandler
	NodeHandler        *NodeHandler
	PodHandler         *PodHandler
	ServiceHandler     *ServiceHandler
	DeploymentHandler  *DeploymentHandler
	StatefulSetHandler *StatefulSetHandler // 添加StatefulSet处理器
	NodePoolHandler    *NodePoolHandler
	HealthHandler      *HealthCheckHandler
	PodTerminalHandler *PodTerminalHandler
	NamespaceHandler   *NamespaceHandler // Add namespace handler
}

// InitRouter Initialize router
func InitRouter(app *App) *gin.Engine {
	router := gin.Default()

	// Cross-origin configuration
	router.Use(cors.Default())

	// Add language detection middleware
	router.Use(middleware.DetectLanguage())

	// Configure static resources
	if configs.IsDevMode() {
		// Load static files from the file system in development mode
		router.Static("/assets", "./web/assets")
		router.Static("/pages", "./web/pages")
		router.StaticFile("/", "./web/index.html")
		router.StaticFile("/favicon.ico", "./web/favicon.ico")
	} else {
		// Use embedded static resources in production mode
		staticHandler := embed.StaticHandler("/")
		router.GET("/", func(c *gin.Context) {
			staticHandler.ServeHTTP(c.Writer, c.Request)
		})
		router.GET("/assets/*filepath", func(c *gin.Context) {
			staticHandler.ServeHTTP(c.Writer, c.Request)
		})
		router.GET("/pages/*filepath", func(c *gin.Context) {
			staticHandler.ServeHTTP(c.Writer, c.Request)
		})
		router.GET("/favicon.ico", func(c *gin.Context) {
			staticHandler.ServeHTTP(c.Writer, c.Request)
		})
	}

	// API version grouping
	v1 := router.Group("/api")
	{
		// Health check
		v1.GET("/health", app.HealthHandler.CheckHealth)
		// Cluster management
		v1.GET("/clusters", app.ClusterHandler.ListClusters)
		v1.POST("/clusters", app.ClusterHandler.AddCluster)
		v1.DELETE("/clusters/:cluster", app.ClusterHandler.RemoveCluster)
		// Change to GET method to match the frontend
		v1.GET("/clusters/:cluster/test", app.ClusterHandler.TestConnection)
		// Ensure cluster details route exists
		v1.GET("/clusters/:cluster", app.ClusterHandler.GetClusterDetails)
		// Cluster monitoring metrics
		v1.GET("/clusters/:cluster/metrics", app.ClusterHandler.GetClusterMetrics)
		// Cluster events
		v1.GET("/clusters/:cluster/events", app.ClusterHandler.GetClusterEvents)
		// Get cluster add type information
		v1.GET("/clusters/:cluster/add-type", app.ClusterHandler.GetClusterAddType)

		// Namespace management - Use dedicated NamespaceHandler
		v1.GET("/clusters/:cluster/namespaces", app.NamespaceHandler.ListNamespaces)

		// Node pool management
		v1.GET("/clusters/:cluster/nodepools", app.NodePoolHandler.ListNodePools)
		v1.POST("/clusters/:cluster/nodepools", app.NodePoolHandler.CreateNodePool)
		v1.GET("/clusters/:cluster/nodepools/:pool", app.NodePoolHandler.GetNodePool)
		v1.PUT("/clusters/:cluster/nodepools/:pool", app.NodePoolHandler.UpdateNodePool)
		v1.DELETE("/clusters/:cluster/nodepools/:pool", app.NodePoolHandler.DeleteNodePool)

		// Node management
		v1.GET("/clusters/:cluster/nodes", app.NodeHandler.ListNodes)
		v1.GET("/clusters/:cluster/nodes/:node", app.NodeHandler.GetNodeDetails)
		v1.GET("/clusters/:cluster/nodes/:node/metrics", app.NodeHandler.GetNodeMetrics)
		v1.POST("/clusters/:cluster/nodes/:node/drain", app.NodeHandler.DrainNode)
		v1.POST("/clusters/:cluster/nodes/:node/cordon", app.NodeHandler.CordonNode)
		v1.POST("/clusters/:cluster/nodes/:node/uncordon", app.NodeHandler.UncordonNode)
		// Node operation interface
		v1.POST("/clusters/:cluster/nodes", app.NodeHandler.AddNode)
		v1.DELETE("/clusters/:cluster/nodes/:node", app.NodeHandler.RemoveNode)
		// Node taint management interface
		v1.GET("/clusters/:cluster/nodes/:node/taints", app.NodeHandler.GetNodeTaints)
		v1.POST("/clusters/:cluster/nodes/:node/taints", app.NodeHandler.AddNodeTaint)
		v1.DELETE("/clusters/:cluster/nodes/:node/taints", app.NodeHandler.RemoveNodeTaint)
		// Node label management interface
		v1.GET("/clusters/:cluster/nodes/:node/labels", app.NodeHandler.GetNodeLabels)
		v1.POST("/clusters/:cluster/nodes/:node/labels", app.NodeHandler.AddNodeLabel)
		v1.DELETE("/clusters/:cluster/nodes/:node/labels", app.NodeHandler.RemoveNodeLabel)
		// Get the list of Pods on the node interface
		v1.GET("/clusters/:cluster/nodes/:node/pods", app.NodeHandler.GetNodePods)

		// Pod management
		v1.GET("/clusters/:cluster/pods", app.PodHandler.ListPods)
		v1.GET("/clusters/:cluster/namespaces/:namespace/pods", app.PodHandler.ListPodsByNamespace)
		v1.POST("/clusters/:cluster/namespaces/:namespace/pods/selector", app.PodHandler.GetPodsBySelector)
		v1.GET("/clusters/:cluster/namespaces/:namespace/pods/:pod", app.PodHandler.GetPodDetails)
		v1.GET("/clusters/:cluster/namespaces/:namespace/pods/:pod/events", app.PodHandler.GetPodEvents)
		v1.DELETE("/clusters/:cluster/namespaces/:namespace/pods/:pod", app.PodHandler.DeletePod)
		v1.GET("/clusters/:cluster/namespaces/:namespace/pods/:pod/logs", app.PodHandler.GetPodLogs)
		v1.GET("/clusters/:cluster/namespaces/:namespace/pods/:pod/logs/stream", app.PodHandler.StreamPodLogs)
		// Pod existence check API
		v1.GET("/clusters/:cluster/namespaces/:namespace/pods/:pod/exists", app.PodHandler.CheckPodExists)
		// Pod terminal WebSocket route
		v1.GET("/clusters/:cluster/namespaces/:namespace/pods/:pod/exec", app.PodTerminalHandler.HandleTerminal)

		// Service management
		v1.GET("/clusters/:cluster/services", app.ServiceHandler.ListServices)
		v1.GET("/clusters/:cluster/namespaces/:namespace/services", app.ServiceHandler.ListServicesByNamespace)
		v1.POST("/clusters/:cluster/namespaces/:namespace/services", app.ServiceHandler.CreateService)
		v1.GET("/clusters/:cluster/namespaces/:namespace/services/:service", app.ServiceHandler.GetServiceDetails)
		v1.DELETE("/clusters/:cluster/namespaces/:namespace/services/:service", app.ServiceHandler.DeleteService)

		// Deployment management
		v1.GET("/clusters/:cluster/deployments", app.DeploymentHandler.ListDeployments)
		v1.GET("/clusters/:cluster/namespaces/:namespace/deployments", app.DeploymentHandler.ListDeploymentsByNamespace)
		v1.GET("/clusters/:cluster/namespaces/:namespace/deployments/:deployment", app.DeploymentHandler.GetDeploymentDetails)
		v1.GET("/clusters/:cluster/namespaces/:namespace/deployments/:deployment/events", app.DeploymentHandler.GetAllRelatedEvents)
		v1.GET("/clusters/:cluster/namespaces/:namespace/deployments/:deployment/all-events", app.DeploymentHandler.GetAllRelatedEvents)
		v1.POST("/clusters/:cluster/namespaces/:namespace/deployments", app.DeploymentHandler.CreateDeployment)
		v1.PUT("/clusters/:cluster/namespaces/:namespace/deployments/:deployment", app.DeploymentHandler.UpdateDeployment)
		v1.PUT("/clusters/:cluster/namespaces/:namespace/deployments/:deployment/scale", app.DeploymentHandler.ScaleDeployment)
		v1.POST("/clusters/:cluster/namespaces/:namespace/deployments/:deployment/restart", app.DeploymentHandler.RestartDeployment)
		v1.DELETE("/clusters/:cluster/namespaces/:namespace/deployments/:deployment", app.DeploymentHandler.DeleteDeployment)

		// StatefulSet management
		v1.GET("/clusters/:cluster/statefulsets", app.StatefulSetHandler.ListStatefulSets)
		v1.GET("/clusters/:cluster/namespaces/:namespace/statefulsets", app.StatefulSetHandler.ListStatefulSets)
		v1.GET("/clusters/:cluster/namespaces/:namespace/statefulsets/:statefulset", app.StatefulSetHandler.GetStatefulSetDetails)
		v1.GET("/clusters/:cluster/namespaces/:namespace/statefulsets/:statefulset/events", app.StatefulSetHandler.GetStatefulSetEvents)
		v1.GET("/clusters/:cluster/namespaces/:namespace/statefulsets/:statefulset/all-events", app.StatefulSetHandler.GetAllStatefulSetEvents)
		v1.GET("/clusters/:cluster/namespaces/:namespace/statefulsets/:statefulset/pods", app.StatefulSetHandler.GetStatefulSetPods)
		v1.POST("/clusters/:cluster/namespaces/:namespace/statefulsets", app.StatefulSetHandler.CreateStatefulSet)
		v1.PUT("/clusters/:cluster/namespaces/:namespace/statefulsets/:statefulset", app.StatefulSetHandler.UpdateStatefulSet)
		v1.PUT("/clusters/:cluster/namespaces/:namespace/statefulsets/:statefulset/scale", app.StatefulSetHandler.ScaleStatefulSet)
		v1.POST("/clusters/:cluster/namespaces/:namespace/statefulsets/:statefulset/restart", app.StatefulSetHandler.RestartStatefulSet)
		v1.DELETE("/clusters/:cluster/namespaces/:namespace/statefulsets/:statefulset", app.StatefulSetHandler.DeleteStatefulSet)
	}

	return router
}
