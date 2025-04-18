package k8s

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// ClientManager Manages k8s client connections
type ClientManager struct {
	clients  map[string]*kubernetes.Clientset
	configs  map[string]*rest.Config
	addTypes map[string]string // 存储集群添加方式："path"或"content"
	mutex    sync.RWMutex
}

func (cm *ClientManager) ValidateKubeconfig(path string) error {
	// Load kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		return fmt.Errorf("failed to build kubeconfig: %w", err)
	}

	// Create client
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	// Test connection
	_, err = clientset.ServerVersion()
	if err != nil {
		return fmt.Errorf("failed to connect to cluster: %w", err)
	}

	return nil
}

// ValidateKubeconfigContent 验证kubeconfig内容是否有效
func (cm *ClientManager) ValidateKubeconfigContent(content string) error {
	// 创建临时文件
	tmpfile, err := os.CreateTemp("", "kubeconfig-*.yaml")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpfile.Name()) // 确保删除临时文件

	// 写入内容到临时文件
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		tmpfile.Close()
		return fmt.Errorf("failed to write to temporary file: %w", err)
	}
	tmpfile.Close()

	// 验证配置
	return cm.ValidateKubeconfig(tmpfile.Name())
}

type Cluster struct {
	Name              string `json:"name"`
	KubeconfigPath    string `json:"kubeconfigPath"`
	KubeconfigContent string `json:"kubeconfigContent,omitempty"`
	// 添加一个类型字段，标识用户通过哪种方式添加的集群
	AddType string `json:"addType,omitempty"` // "path" 或 "content"
}

// NewClientManager Create client manager
func NewClientManager() *ClientManager {
	return &ClientManager{
		clients:  make(map[string]*kubernetes.Clientset),
		configs:  make(map[string]*rest.Config),
		addTypes: make(map[string]string),
	}
}

// AddCluster Add cluster
func (cm *ClientManager) AddCluster(clusterName, kubeconfigPath string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// Load kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return fmt.Errorf("failed to build kubeconfig: %w", err)
	}

	// Create client
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	// Test connection
	_, err = clientset.ServerVersion()
	if err != nil {
		return fmt.Errorf("failed to connect to cluster: %w", err)
	}

	// Store client
	cm.clients[clusterName] = clientset
	cm.configs[clusterName] = config
	cm.addTypes[clusterName] = "path"

	return nil
}

// AddClusterWithContent 通过kubeconfig内容添加集群
func (cm *ClientManager) AddClusterWithContent(clusterName, content string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// 创建临时文件
	tmpDir := os.TempDir()
	kubeconfigPath := filepath.Join(tmpDir, fmt.Sprintf("kubeconfig-%s.yaml", clusterName))

	// 写入内容到临时文件
	if err := os.WriteFile(kubeconfigPath, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write kubeconfig content to file: %w", err)
	}

	// 加载kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		os.Remove(kubeconfigPath) // 清理临时文件
		return fmt.Errorf("failed to build kubeconfig: %w", err)
	}

	// 创建客户端
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		os.Remove(kubeconfigPath) // 清理临时文件
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	// 测试连接
	_, err = clientset.ServerVersion()
	if err != nil {
		os.Remove(kubeconfigPath) // 清理临时文件
		return fmt.Errorf("failed to connect to cluster: %w", err)
	}

	// 存储客户端
	cm.clients[clusterName] = clientset
	cm.configs[clusterName] = config
	cm.addTypes[clusterName] = "content"

	return nil
}

// RemoveCluster Remove cluster
func (cm *ClientManager) RemoveCluster(clusterName string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if _, exists := cm.clients[clusterName]; !exists {
		return fmt.Errorf("cluster %s not found", clusterName)
	}

	delete(cm.clients, clusterName)
	delete(cm.configs, clusterName)
	delete(cm.addTypes, clusterName)
	return nil
}

// GetClient Get client for specified cluster
func (cm *ClientManager) GetClient(clusterName string) (*kubernetes.Clientset, error) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	client, exists := cm.clients[clusterName]
	if !exists {
		return nil, fmt.Errorf("cluster %s not found", clusterName)
	}
	return client, nil
}

// GetConfig Get configuration for specified cluster
func (cm *ClientManager) GetConfig(clusterName string) (*rest.Config, error) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	config, exists := cm.configs[clusterName]
	if !exists {
		return nil, fmt.Errorf("cluster %s not found", clusterName)
	}
	return config, nil
}

// ListClusters List all clusters
func (cm *ClientManager) ListClusters() []string {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	clusters := make([]string, 0, len(cm.clients))
	for name := range cm.clients {
		clusters = append(clusters, name)
	}
	return clusters
}

// TestConnection Test cluster connection
func (cm *ClientManager) TestConnection(clusterName string) error {
	client, err := cm.GetClient(clusterName)
	if err != nil {
		return err
	}

	_, err = client.ServerVersion()
	if err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}

	return nil
}

// GetAddType 获取集群的添加方式
func (cm *ClientManager) GetAddType(clusterName string) string {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	addType, exists := cm.addTypes[clusterName]
	if !exists {
		return "unknown" // 如果找不到添加方式，返回"unknown"
	}
	return addType
}
