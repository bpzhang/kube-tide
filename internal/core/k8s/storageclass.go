package k8s

import (
	"context"
	"fmt"
	"sort"
	"time"

	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StorageClassService StorageClass 只读服务
type StorageClassService struct {
	clientManager *ClientManager
}

// NewStorageClassService 创建 StorageClass 服务
func NewStorageClassService(clientManager *ClientManager) *StorageClassService {
	return &StorageClassService{clientManager: clientManager}
}

// StorageClassInfo StorageClass 摘要
type StorageClassInfo struct {
	Name                 string            `json:"name"`
	Provisioner          string            `json:"provisioner"`
	ReclaimPolicy        string            `json:"reclaimPolicy,omitempty"`
	VolumeBindingMode    string            `json:"volumeBindingMode,omitempty"`
	AllowVolumeExpansion bool              `json:"allowVolumeExpansion"`
	IsDefault            bool              `json:"isDefault"`
	CreationTime         time.Time         `json:"creationTime"`
	Labels               map[string]string `json:"labels,omitempty"`
}

// ListStorageClasses 获取 StorageClass 列表
func (s *StorageClassService) ListStorageClasses(ctx context.Context, clusterName string) ([]StorageClassInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	list, err := client.StorageV1().StorageClasses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 StorageClass 列表失败: %w", err)
	}
	result := make([]StorageClassInfo, 0, len(list.Items))
	for _, sc := range list.Items {
		result = append(result, convertStorageClassInfo(&sc))
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result, nil
}

// GetStorageClass 获取 StorageClass 详情
func (s *StorageClassService) GetStorageClass(ctx context.Context, clusterName, name string) (*StorageClassInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	sc, err := client.StorageV1().StorageClasses().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 StorageClass 失败: %w", err)
	}
	info := convertStorageClassInfo(sc)
	return &info, nil
}

func convertStorageClassInfo(sc *storagev1.StorageClass) StorageClassInfo {
	reclaim := ""
	if sc.ReclaimPolicy != nil {
		reclaim = string(*sc.ReclaimPolicy)
	}
	binding := ""
	if sc.VolumeBindingMode != nil {
		binding = string(*sc.VolumeBindingMode)
	}
	isDefault := sc.Annotations["storageclass.kubernetes.io/is-default-class"] == "true"
	allowExpansion := false
	if sc.AllowVolumeExpansion != nil {
		allowExpansion = *sc.AllowVolumeExpansion
	}
	return StorageClassInfo{
		Name:                 sc.Name,
		Provisioner:          sc.Provisioner,
		ReclaimPolicy:        reclaim,
		VolumeBindingMode:    binding,
		AllowVolumeExpansion: allowExpansion,
		IsDefault:            isDefault,
		CreationTime:         sc.CreationTimestamp.Time,
		Labels:               sc.Labels,
	}
}
