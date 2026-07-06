package k8s

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SecretInfo Secret 摘要（列表不返回值）
type SecretInfo struct {
	Name         string            `json:"name"`
	Namespace    string            `json:"namespace"`
	Type         string            `json:"type"`
	Labels       map[string]string `json:"labels,omitempty"`
	DataKeys     []string          `json:"dataKeys"`
	CreationTime string            `json:"creationTime"`
}

// SecretDetail Secret 详情（解码后的 data）
type SecretDetail struct {
	SecretInfo
	Data map[string]string `json:"data"`
}

// SecretService Secret 管理服务
type SecretService struct {
	clientManager *ClientManager
}

// NewSecretService 创建 Secret 服务
func NewSecretService(clientManager *ClientManager) *SecretService {
	return &SecretService{clientManager: clientManager}
}

func decodeSecretData(data map[string][]byte) map[string]string {
	result := make(map[string]string, len(data))
	for k, v := range data {
		result[k] = string(v)
	}
	return result
}

func toSecretInfo(sec corev1.Secret) SecretInfo {
	keys := make([]string, 0, len(sec.Data))
	for k := range sec.Data {
		keys = append(keys, k)
	}
	return SecretInfo{
		Name:         sec.Name,
		Namespace:    sec.Namespace,
		Type:         string(sec.Type),
		Labels:       sec.Labels,
		DataKeys:     keys,
		CreationTime: sec.CreationTimestamp.Format("2006-01-02 15:04:05"),
	}
}

// ListSecrets 获取集群所有 Secret
func (s *SecretService) ListSecrets(ctx context.Context, clusterName string) ([]SecretInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	list, err := client.CoreV1().Secrets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 Secret 列表失败: %w", err)
	}
	result := make([]SecretInfo, 0, len(list.Items))
	for _, sec := range list.Items {
		result = append(result, toSecretInfo(sec))
	}
	return result, nil
}

// ListSecretsByNamespace 按命名空间获取 Secret
func (s *SecretService) ListSecretsByNamespace(ctx context.Context, clusterName, namespace string) ([]SecretInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	list, err := client.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 Secret 列表失败: %w", err)
	}
	result := make([]SecretInfo, 0, len(list.Items))
	for _, sec := range list.Items {
		result = append(result, toSecretInfo(sec))
	}
	return result, nil
}

// GetSecret 获取 Secret 详情
func (s *SecretService) GetSecret(ctx context.Context, clusterName, namespace, name string) (*SecretDetail, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	sec, err := client.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 Secret 失败: %w", err)
	}
	info := toSecretInfo(*sec)
	return &SecretDetail{SecretInfo: info, Data: decodeSecretData(sec.Data)}, nil
}

// CreateSecretRequest 创建 Secret 请求
type CreateSecretRequest struct {
	Name       string            `json:"name"`
	Type       string            `json:"type,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
	StringData map[string]string `json:"stringData"`
}

// CreateSecret 创建 Secret
func (s *SecretService) CreateSecret(ctx context.Context, clusterName, namespace string, req CreateSecretRequest) (*SecretDetail, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("Secret 名称不能为空")
	}
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	secType := corev1.SecretTypeOpaque
	if req.Type != "" {
		secType = corev1.SecretType(req.Type)
	}
	sec := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: req.Name, Namespace: namespace, Labels: req.Labels},
		Type:       secType,
		StringData: req.StringData,
	}
	created, err := client.CoreV1().Secrets(namespace).Create(ctx, sec, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("创建 Secret 失败: %w", err)
	}
	info := toSecretInfo(*created)
	return &SecretDetail{SecretInfo: info, Data: decodeSecretData(created.Data)}, nil
}

// UpdateSecret 更新 Secret
func (s *SecretService) UpdateSecret(ctx context.Context, clusterName, namespace, name string, stringData map[string]string, labels map[string]string, secType string) (*SecretDetail, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	sec, err := client.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 Secret 失败: %w", err)
	}
	if stringData != nil {
		// Full replace: keys removed in the UI are deleted from the Secret
		newData := make(map[string][]byte, len(stringData))
		for k, v := range stringData {
			newData[k] = []byte(v)
		}
		sec.Data = newData
		sec.StringData = nil
	}
	if labels != nil {
		sec.Labels = labels
	}
	if secType != "" {
		sec.Type = corev1.SecretType(secType)
	}
	updated, err := client.CoreV1().Secrets(namespace).Update(ctx, sec, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("更新 Secret 失败: %w", err)
	}
	info := toSecretInfo(*updated)
	return &SecretDetail{SecretInfo: info, Data: decodeSecretData(updated.Data)}, nil
}

// DeleteSecret 删除 Secret
func (s *SecretService) DeleteSecret(ctx context.Context, clusterName, namespace, name string) error {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return err
	}
	if err := client.CoreV1().Secrets(namespace).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		return fmt.Errorf("删除 Secret 失败: %w", err)
	}
	return nil
}
