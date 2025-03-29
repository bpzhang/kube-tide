package k8s

import (
	"context"
	"fmt"

	"kube-tide/configs"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ServiceManager Service服务管理
type ServiceManager struct {
	clientManager *ClientManager
}

// NewServiceManager 创建Service管理器
func NewServiceManager(clientManager *ClientManager) *ServiceManager {
	return &ServiceManager{
		clientManager: clientManager,
	}
}

// GetServices 获取所有命名空间的Service列表
func (s *ServiceManager) GetServices(ctx context.Context, clusterName string) ([]corev1.Service, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	serviceList, err := client.CoreV1().Services("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取Service列表失败: %w", err)
	}

	return serviceList.Items, nil
}

// GetServicesByNamespace 获取指定命名空间的Service列表
func (s *ServiceManager) GetServicesByNamespace(ctx context.Context, clusterName, namespace string) ([]corev1.Service, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	serviceList, err := client.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取命名空间 %s 的Service列表失败: %w", namespace, err)
	}

	return serviceList.Items, nil
}

// GetServiceDetails 获取Service详情
func (s *ServiceManager) GetServiceDetails(ctx context.Context, clusterName, namespace, serviceName string) (*corev1.Service, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	service, err := client.CoreV1().Services(namespace).Get(ctx, serviceName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取Service详情失败: %w", err)
	}

	return service, nil
}

// GetServiceEndpoints 获取Service关联的Endpoints
func (s *ServiceManager) GetServiceEndpoints(ctx context.Context, clusterName, namespace, serviceName string) (*corev1.Endpoints, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}

	endpoints, err := client.CoreV1().Endpoints(namespace).Get(ctx, serviceName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取Service的Endpoints失败: %w", err)
	}

	return endpoints, nil
}

// CreateService 创建Service
func (s *ServiceManager) CreateService(ctx context.Context, clusterName string, service *corev1.Service) error {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return err
	}

	_, err = client.CoreV1().Services(service.Namespace).Create(ctx, service, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("创建Service失败: %w", err)
	}

	return nil
}

// UpdateService 更新Service
func (s *ServiceManager) UpdateService(ctx context.Context, clusterName string, service *corev1.Service) error {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return err
	}

	_, err = client.CoreV1().Services(service.Namespace).Update(ctx, service, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("更新Service失败: %w", err)
	}

	return nil
}

// DeleteService 删除Service
func (s *ServiceManager) DeleteService(ctx context.Context, clusterName, namespace, serviceName string) error {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return err
	}

	err = client.CoreV1().Services(namespace).Delete(ctx, serviceName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("删除Service失败: %w", err)
	}

	return nil
}

// GetServiceType 获取Service类型
func (s *ServiceManager) GetServiceType(service *corev1.Service) string {
	return string(service.Spec.Type)
}

// GetServiceClusterIP 获取Service的ClusterIP
func (s *ServiceManager) GetServiceClusterIP(service *corev1.Service) string {
	return service.Spec.ClusterIP
}

// GetServiceExternalIPs 获取Service的外部IP列表
func (s *ServiceManager) GetServiceExternalIPs(service *corev1.Service) []string {
	return service.Spec.ExternalIPs
}

// GetServicePorts 获取Service的端口映射
func (s *ServiceManager) GetServicePorts(service *corev1.Service) []corev1.ServicePort {
	return service.Spec.Ports
}

// ServiceService Service服务
type ServiceService struct {
	clientManager *ClientManager
}

// NewServiceService 创建Service服务
func NewServiceService(clientManager *ClientManager) *ServiceService {
	return &ServiceService{
		clientManager: clientManager,
	}
}

// ClusterService 集群服务
type ClusterService struct {
	clientManager *ClientManager
	config        *configs.Config
}

// NewClusterService 创建集群服务
func NewClusterService(clientManager *ClientManager, config *configs.Config) *ClusterService {
	return &ClusterService{
		clientManager: clientManager,
		config:        config,
	}
}
