package k8s

import (
	"fmt"
	"sync"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// ClientManager 管理k8s客户端连接
type ClientManager struct {
	clients map[string]*kubernetes.Clientset
	configs map[string]*rest.Config
	mutex   sync.RWMutex
}

// NewClientManager 创建客户端管理器
func NewClientManager() *ClientManager {
	return &ClientManager{
		clients: make(map[string]*kubernetes.Clientset),
		configs: make(map[string]*rest.Config),
	}
}

// AddCluster 添加集群
func (cm *ClientManager) AddCluster(clusterName, kubeconfigPath string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// 加载kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return fmt.Errorf("无法构建kubeconfig: %w", err)
	}

	// 创建客户端
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("无法创建kubernetes客户端: %w", err)
	}

	// 测试连接
	_, err = clientset.ServerVersion()
	if err != nil {
		return fmt.Errorf("无法连接到集群: %w", err)
	}

	// 存储客户端
	cm.clients[clusterName] = clientset
	cm.configs[clusterName] = config

	return nil
}

// RemoveCluster 移除集群
func (cm *ClientManager) RemoveCluster(clusterName string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	delete(cm.clients, clusterName)
	delete(cm.configs, clusterName)
}

// GetClient 获取指定集群的客户端
func (cm *ClientManager) GetClient(clusterName string) (*kubernetes.Clientset, error) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	client, exists := cm.clients[clusterName]
	if !exists {
		return nil, fmt.Errorf("集群 %s 未找到", clusterName)
	}
	return client, nil
}

// GetConfig 获取指定集群的配置
func (cm *ClientManager) GetConfig(clusterName string) (*rest.Config, error) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	config, exists := cm.configs[clusterName]
	if !exists {
		return nil, fmt.Errorf("集群 %s 未找到", clusterName)
	}
	return config, nil
}

// ListClusters 列出所有集群
func (cm *ClientManager) ListClusters() []string {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	clusters := make([]string, 0, len(cm.clients))
	for name := range cm.clients {
		clusters = append(clusters, name)
	}
	return clusters
}

// TestConnection 测试集群连接
func (cm *ClientManager) TestConnection(clusterName string) error {
	client, err := cm.GetClient(clusterName)
	if err != nil {
		return err
	}

	_, err = client.ServerVersion()
	if err != nil {
		return fmt.Errorf("连接测试失败: %w", err)
	}

	return nil
}
