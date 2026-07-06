package k8s

import (
	"context"
	"fmt"
	"sort"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PVService PV 只读服务
type PVService struct {
	clientManager *ClientManager
}

// NewPVService 创建 PV 服务
func NewPVService(clientManager *ClientManager) *PVService {
	return &PVService{clientManager: clientManager}
}

// PVInfo PV 摘要
type PVInfo struct {
	Name             string            `json:"name"`
	Status           string            `json:"status"`
	StorageClassName string            `json:"storageClassName,omitempty"`
	Capacity         string            `json:"capacity,omitempty"`
	AccessModes      []string          `json:"accessModes,omitempty"`
	ReclaimPolicy    string            `json:"reclaimPolicy,omitempty"`
	ClaimRef         string            `json:"claimRef,omitempty"`
	CreationTime     time.Time         `json:"creationTime"`
	Labels           map[string]string `json:"labels,omitempty"`
}

// ListPVs 获取 PV 列表
func (s *PVService) ListPVs(ctx context.Context, clusterName string) ([]PVInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	list, err := client.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 PV 列表失败: %w", err)
	}
	result := make([]PVInfo, 0, len(list.Items))
	for _, pv := range list.Items {
		result = append(result, convertPVInfo(&pv))
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreationTime.After(result[j].CreationTime)
	})
	return result, nil
}

// GetPV 获取 PV 详情
func (s *PVService) GetPV(ctx context.Context, clusterName, name string) (*PVInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	pv, err := client.CoreV1().PersistentVolumes().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 PV 失败: %w", err)
	}
	info := convertPVInfo(pv)
	return &info, nil
}

func convertPVInfo(pv *corev1.PersistentVolume) PVInfo {
	modes := make([]string, 0, len(pv.Spec.AccessModes))
	for _, m := range pv.Spec.AccessModes {
		modes = append(modes, string(m))
	}
	capacity := ""
	if qty, ok := pv.Spec.Capacity[corev1.ResourceStorage]; ok {
		capacity = qty.String()
	}
	claimRef := ""
	if pv.Spec.ClaimRef != nil {
		claimRef = pv.Spec.ClaimRef.Namespace + "/" + pv.Spec.ClaimRef.Name
	}
	return PVInfo{
		Name:             pv.Name,
		Status:           string(pv.Status.Phase),
		StorageClassName: pv.Spec.StorageClassName,
		Capacity:         capacity,
		AccessModes:      modes,
		ReclaimPolicy:    string(pv.Spec.PersistentVolumeReclaimPolicy),
		ClaimRef:         claimRef,
		CreationTime:     pv.CreationTimestamp.Time,
		Labels:           pv.Labels,
	}
}
