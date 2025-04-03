package k8s

import (
	"fmt"
	"sync"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// ClientManager Manages k8s client connections
type ClientManager struct {
	clients map[string]*kubernetes.Clientset
	configs map[string]*rest.Config
	mutex   sync.RWMutex
}

// NewClientManager Create client manager
func NewClientManager() *ClientManager {
	return &ClientManager{
		clients: make(map[string]*kubernetes.Clientset),
		configs: make(map[string]*rest.Config),
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

	return nil
}

// RemoveCluster Remove cluster
func (cm *ClientManager) RemoveCluster(clusterName string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	delete(cm.clients, clusterName)
	delete(cm.configs, clusterName)
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
