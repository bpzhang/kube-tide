package api

import (
	"kube-tide/configs"
	"kube-tide/pkg/embed"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// App 应用程序结构体
type App struct {
	ClusterHandler     *ClusterHandler
	NodeHandler        *NodeHandler
	PodHandler         *PodHandler
	ServiceHandler     *ServiceHandler
	DeploymentHandler  *DeploymentHandler
	NodePoolHandler    *NodePoolHandler
	HealthHandler      *HealthCheckHandler
	PodTerminalHandler *PodTerminalHandler
	NamespaceHandler   *NamespaceHandler // 添加命名空间处理器
}

// InitRouter 初始化路由
func InitRouter(app *App) *gin.Engine {
	router := gin.Default()

	// 跨域配置
	router.Use(cors.Default())

	// 配置静态资源
	if configs.IsDevMode() {
		// 开发模式下从文件系统加载静态文件
		router.Static("/assets", "./web/assets")
		router.Static("/pages", "./web/pages")
		router.StaticFile("/", "./web/index.html")
		router.StaticFile("/favicon.ico", "./web/favicon.ico")
	} else {
		// 生产模式下使用嵌入的静态资源
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

	// API 版本分组
	v1 := router.Group("/api")
	{
		// 健康检查
		v1.GET("/health", app.HealthHandler.CheckHealth)
		// 集群管理
		v1.GET("/clusters", app.ClusterHandler.ListClusters)
		v1.POST("/clusters", app.ClusterHandler.AddCluster)
		v1.DELETE("/clusters/:cluster", app.ClusterHandler.RemoveCluster)
		// 修改为GET方法，与前端一致
		v1.GET("/clusters/:cluster/test", app.ClusterHandler.TestConnection)
		// 确保集群详情路由存在
		v1.GET("/clusters/:cluster", app.ClusterHandler.GetClusterDetails)
		// 集群监控指标
		v1.GET("/clusters/:cluster/metrics", app.ClusterHandler.GetClusterMetrics)
		// 集群事件
		v1.GET("/clusters/:cluster/events", app.ClusterHandler.GetClusterEvents)

		// 命名空间管理 - 使用专门的NamespaceHandler
		v1.GET("/clusters/:cluster/namespaces", app.NamespaceHandler.ListNamespaces)

		// 节点池管理
		v1.GET("/clusters/:cluster/nodepools", app.NodePoolHandler.ListNodePools)
		v1.POST("/clusters/:cluster/nodepools", app.NodePoolHandler.CreateNodePool)
		v1.GET("/clusters/:cluster/nodepools/:pool", app.NodePoolHandler.GetNodePool)
		v1.PUT("/clusters/:cluster/nodepools/:pool", app.NodePoolHandler.UpdateNodePool)
		v1.DELETE("/clusters/:cluster/nodepools/:pool", app.NodePoolHandler.DeleteNodePool)

		// 节点管理
		v1.GET("/clusters/:cluster/nodes", app.NodeHandler.ListNodes)
		v1.GET("/clusters/:cluster/nodes/:node", app.NodeHandler.GetNodeDetails)
		v1.GET("/clusters/:cluster/nodes/:node/metrics", app.NodeHandler.GetNodeMetrics)
		v1.POST("/clusters/:cluster/nodes/:node/drain", app.NodeHandler.DrainNode)
		v1.POST("/clusters/:cluster/nodes/:node/cordon", app.NodeHandler.CordonNode)
		v1.POST("/clusters/:cluster/nodes/:node/uncordon", app.NodeHandler.UncordonNode)
		// 节点操作接口
		v1.POST("/clusters/:cluster/nodes", app.NodeHandler.AddNode)
		v1.DELETE("/clusters/:cluster/nodes/:node", app.NodeHandler.RemoveNode)
		// 节点污点管理接口
		v1.GET("/clusters/:cluster/nodes/:node/taints", app.NodeHandler.GetNodeTaints)
		v1.POST("/clusters/:cluster/nodes/:node/taints", app.NodeHandler.AddNodeTaint)
		v1.DELETE("/clusters/:cluster/nodes/:node/taints", app.NodeHandler.RemoveNodeTaint)
		// 节点标签管理接口
		v1.GET("/clusters/:cluster/nodes/:node/labels", app.NodeHandler.GetNodeLabels)
		v1.POST("/clusters/:cluster/nodes/:node/labels", app.NodeHandler.AddNodeLabel)
		v1.DELETE("/clusters/:cluster/nodes/:node/labels", app.NodeHandler.RemoveNodeLabel)
		// 获取节点上的Pod列表接口
		v1.GET("/clusters/:cluster/nodes/:node/pods", app.NodeHandler.GetNodePods)

		// Pod管理
		v1.GET("/clusters/:cluster/pods", app.PodHandler.ListPods)
		v1.GET("/clusters/:cluster/namespaces/:namespace/pods", app.PodHandler.ListPodsByNamespace)
		v1.POST("/clusters/:cluster/namespaces/:namespace/pods/selector", app.PodHandler.GetPodsBySelector)
		v1.GET("/clusters/:cluster/namespaces/:namespace/pods/:pod", app.PodHandler.GetPodDetails)
		v1.GET("/clusters/:cluster/namespaces/:namespace/pods/:pod/events", app.PodHandler.GetPodEvents)
		v1.DELETE("/clusters/:cluster/namespaces/:namespace/pods/:pod", app.PodHandler.DeletePod)
		v1.GET("/clusters/:cluster/namespaces/:namespace/pods/:pod/logs", app.PodHandler.GetPodLogs)
		v1.GET("/clusters/:cluster/namespaces/:namespace/pods/:pod/logs/stream", app.PodHandler.StreamPodLogs)
		// Pod存在性检查API
		v1.GET("/clusters/:cluster/namespaces/:namespace/pods/:pod/exists", app.PodHandler.CheckPodExists)
		// Pod终端WebSocket路由
		v1.GET("/clusters/:cluster/namespaces/:namespace/pods/:pod/exec", app.PodTerminalHandler.HandleTerminal)

		// 服务管理
		v1.GET("/clusters/:cluster/services", app.ServiceHandler.ListServices)
		v1.GET("/clusters/:cluster/namespaces/:namespace/services", app.ServiceHandler.ListServicesByNamespace)
		v1.POST("/clusters/:cluster/namespaces/:namespace/services", app.ServiceHandler.CreateService)
		v1.GET("/clusters/:cluster/namespaces/:namespace/services/:service", app.ServiceHandler.GetServiceDetails)
		v1.DELETE("/clusters/:cluster/namespaces/:namespace/services/:service", app.ServiceHandler.DeleteService)

		// Deployment管理
		v1.GET("/clusters/:cluster/deployments", app.DeploymentHandler.ListDeployments)
		v1.GET("/clusters/:cluster/namespaces/:namespace/deployments", app.DeploymentHandler.ListDeploymentsByNamespace)
		v1.GET("/clusters/:cluster/namespaces/:namespace/deployments/:deployment", app.DeploymentHandler.GetDeploymentDetails)
		v1.GET("/clusters/:cluster/namespaces/:namespace/deployments/:deployment/events", app.DeploymentHandler.GetAllRelatedEvents)
		v1.GET("/clusters/:cluster/namespaces/:namespace/deployments/:deployment/all-events", app.DeploymentHandler.GetAllRelatedEvents)
		v1.POST("/clusters/:cluster/namespaces/:namespace/deployments", app.DeploymentHandler.CreateDeployment)
		v1.PUT("/clusters/:cluster/namespaces/:namespace/deployments/:deployment", app.DeploymentHandler.UpdateDeployment)
		v1.PUT("/clusters/:cluster/namespaces/:namespace/deployments/:deployment/scale", app.DeploymentHandler.ScaleDeployment)
		v1.POST("/clusters/:cluster/namespaces/:namespace/deployments/:deployment/restart", app.DeploymentHandler.RestartDeployment)
	}

	return router
}
