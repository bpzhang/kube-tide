package main

import (
	"context"
	"log"
	"time"

	"go.uber.org/zap"

	"kube-tide/internal/database"
	"kube-tide/internal/database/models"
	"kube-tide/internal/repository"
)

func main() {
	// 创建日志器
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Sync()

	// 创建数据库配置 - 使用 SQLite 进行演示
	config := &database.DatabaseConfig{
		Type:            database.SQLite,
		Database:        "kube_tide_example",
		SQLiteFilePath:  "./data/example.db",
		MaxOpenConns:    10,
		MaxIdleConns:    2,
		ConnMaxLifetime: 5 * time.Minute,
	}

	// 创建数据库服务
	dbService, err := database.NewService(config, logger)
	if err != nil {
		logger.Fatal("Failed to create database service", zap.Error(err))
	}
	defer dbService.Close()

	// 初始化数据库（运行迁移）
	ctx := context.Background()
	if err := dbService.Initialize(ctx); err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}

	logger.Info("Database initialized successfully")

	// 创建仓储
	repos := repository.NewRepositories(dbService.GetDatabase(), logger)

	// 演示集群操作
	if err := demonstrateClusterOperations(ctx, repos, logger); err != nil {
		logger.Error("Failed to demonstrate cluster operations", zap.Error(err))
	}

	// 演示节点操作
	if err := demonstrateNodeOperations(ctx, repos, logger); err != nil {
		logger.Error("Failed to demonstrate node operations", zap.Error(err))
	}

	// 演示 Pod 操作
	if err := demonstratePodOperations(ctx, repos, logger); err != nil {
		logger.Error("Failed to demonstrate pod operations", zap.Error(err))
	}

	logger.Info("Example completed successfully")
}

func demonstrateClusterOperations(ctx context.Context, repos *repository.Repositories, logger *zap.Logger) error {
	logger.Info("=== Demonstrating Cluster Operations ===")

	// 创建集群
	cluster := &models.Cluster{
		Name:        "example-cluster",
		Config:      "example-config",
		Status:      models.ClusterStatusActive,
		Description: "Example cluster for demonstration",
		Endpoint:    "https://api.example-cluster.com",
		Version:     "v1.28.0",
	}

	if err := repos.Cluster.Create(ctx, cluster); err != nil {
		return err
	}
	logger.Info("Created cluster", zap.String("id", cluster.ID), zap.String("name", cluster.Name))

	// 查询集群
	retrievedCluster, err := repos.Cluster.GetByID(ctx, cluster.ID)
	if err != nil {
		return err
	}
	logger.Info("Retrieved cluster", zap.String("name", retrievedCluster.Name))

	// 列出集群
	clusters, err := repos.Cluster.List(ctx, models.ClusterFilters{}, models.DefaultPaginationParams())
	if err != nil {
		return err
	}
	logger.Info("Listed clusters", zap.Int("count", clusters.TotalCount))

	// 更新集群
	updates := models.ClusterUpdateRequest{
		Description: stringPtr("Updated description"),
	}
	if err := repos.Cluster.Update(ctx, cluster.ID, updates); err != nil {
		return err
	}
	logger.Info("Updated cluster")

	return nil
}

func demonstrateNodeOperations(ctx context.Context, repos *repository.Repositories, logger *zap.Logger) error {
	logger.Info("=== Demonstrating Node Operations ===")

	// 首先获取一个集群
	clusters, err := repos.Cluster.List(ctx, models.ClusterFilters{}, models.DefaultPaginationParams())
	if err != nil {
		return err
	}
	if clusters.TotalCount == 0 {
		logger.Warn("No clusters found, skipping node operations")
		return nil
	}

	clusterList := clusters.Data.([]*models.Cluster)
	clusterID := clusterList[0].ID

	// 创建节点
	node := &models.Node{
		ClusterID:         clusterID,
		Name:              "example-node-1",
		Status:            models.NodeStatusReady,
		Roles:             "worker",
		Version:           "v1.28.0",
		InternalIP:        "10.0.1.100",
		OSImage:           "Ubuntu 22.04.3 LTS",
		KernelVersion:     "5.15.0-78-generic",
		ContainerRuntime:  "containerd://1.7.2",
		CPUCapacity:       "4",
		MemoryCapacity:    "8Gi",
		CPUAllocatable:    "3900m",
		MemoryAllocatable: "7.5Gi",
	}

	if err := repos.Node.Create(ctx, node); err != nil {
		return err
	}
	logger.Info("Created node", zap.String("id", node.ID), zap.String("name", node.Name))

	// 查询节点
	retrievedNode, err := repos.Node.GetByID(ctx, node.ID)
	if err != nil {
		return err
	}
	logger.Info("Retrieved node", zap.String("name", retrievedNode.Name))

	// 列出集群中的节点
	nodes, err := repos.Node.ListByCluster(ctx, clusterID, models.DefaultPaginationParams())
	if err != nil {
		return err
	}
	logger.Info("Listed nodes in cluster", zap.Int("count", nodes.TotalCount))

	return nil
}

func demonstratePodOperations(ctx context.Context, repos *repository.Repositories, logger *zap.Logger) error {
	logger.Info("=== Demonstrating Pod Operations ===")

	// 首先获取一个集群
	clusters, err := repos.Cluster.List(ctx, models.ClusterFilters{}, models.DefaultPaginationParams())
	if err != nil {
		return err
	}
	if clusters.TotalCount == 0 {
		logger.Warn("No clusters found, skipping pod operations")
		return nil
	}

	clusterList := clusters.Data.([]*models.Cluster)
	clusterID := clusterList[0].ID

	// 创建命名空间
	namespace := &models.Namespace{
		ClusterID: clusterID,
		Name:      "default",
		Status:    models.NamespaceStatusActive,
		Phase:     models.NamespacePhaseActive,
	}

	if err := repos.Namespace.Create(ctx, namespace); err != nil {
		logger.Warn("Failed to create namespace (might already exist)", zap.Error(err))
	} else {
		logger.Info("Created namespace", zap.String("id", namespace.ID), zap.String("name", namespace.Name))
	}

	// 创建 Pod
	pod := &models.Pod{
		ClusterID:       clusterID,
		Namespace:       "default",
		Name:            "example-pod",
		Status:          models.PodStatusRunning,
		Phase:           models.PodPhaseRunning,
		NodeName:        "example-node-1",
		PodIP:           "10.244.1.10",
		HostIP:          "10.0.1.100",
		RestartCount:    0,
		ReadyContainers: 1,
		TotalContainers: 1,
		CPURequests:     "100m",
		MemoryRequests:  "128Mi",
		CPULimits:       "500m",
		MemoryLimits:    "512Mi",
	}

	if err := repos.Pod.Create(ctx, pod); err != nil {
		return err
	}
	logger.Info("Created pod", zap.String("id", pod.ID), zap.String("name", pod.Name))

	// 查询 Pod
	retrievedPod, err := repos.Pod.GetByID(ctx, pod.ID)
	if err != nil {
		return err
	}
	logger.Info("Retrieved pod", zap.String("name", retrievedPod.Name))

	// 列出集群中的 Pod
	pods, err := repos.Pod.ListByCluster(ctx, clusterID, models.DefaultPaginationParams())
	if err != nil {
		return err
	}
	logger.Info("Listed pods in cluster", zap.Int("count", pods.TotalCount))

	// 列出命名空间中的 Pod
	namespacePods, err := repos.Pod.ListByNamespace(ctx, clusterID, "default", models.DefaultPaginationParams())
	if err != nil {
		return err
	}
	logger.Info("Listed pods in namespace", zap.Int("count", namespacePods.TotalCount))

	return nil
}

// 辅助函数
func stringPtr(s string) *string {
	return &s
}
