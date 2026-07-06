package k8s

import (
	"context"
	"fmt"
	"sort"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PVCService PVC 管理服务
type PVCService struct {
	clientManager *ClientManager
}

// NewPVCService 创建 PVC 服务
func NewPVCService(clientManager *ClientManager) *PVCService {
	return &PVCService{clientManager: clientManager}
}

// PVCInfo PVC 摘要
type PVCInfo struct {
	Name             string            `json:"name"`
	Namespace        string            `json:"namespace"`
	Status           string            `json:"status"`
	VolumeName       string            `json:"volumeName,omitempty"`
	StorageClassName string            `json:"storageClassName,omitempty"`
	AccessModes      []string          `json:"accessModes,omitempty"`
	Capacity         string            `json:"capacity,omitempty"`
	CreationTime     time.Time         `json:"creationTime"`
	Labels           map[string]string `json:"labels,omitempty"`
}

// CreatePVCRequest 创建 PVC 请求
type CreatePVCRequest struct {
	Name             string            `json:"name" binding:"required"`
	Namespace        string            `json:"namespace"`
	Labels           map[string]string `json:"labels,omitempty"`
	StorageClassName string            `json:"storageClassName,omitempty"`
	AccessModes      []string          `json:"accessModes,omitempty"`
	Storage          string            `json:"storage" binding:"required"`
}

// ListPVCs 获取 PVC 列表
func (s *PVCService) ListPVCs(ctx context.Context, clusterName, namespace string) ([]PVCInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	var list *corev1.PersistentVolumeClaimList
	if namespace == "all" || namespace == "" {
		list, err = client.CoreV1().PersistentVolumeClaims("").List(ctx, metav1.ListOptions{})
	} else {
		list, err = client.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{})
	}
	if err != nil {
		return nil, fmt.Errorf("获取 PVC 列表失败: %w", err)
	}
	result := make([]PVCInfo, 0, len(list.Items))
	for _, pvc := range list.Items {
		result = append(result, convertPVCInfo(&pvc))
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreationTime.After(result[j].CreationTime)
	})
	return result, nil
}

// GetPVC 获取 PVC 详情
func (s *PVCService) GetPVC(ctx context.Context, clusterName, namespace, name string) (*PVCInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	pvc, err := client.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 PVC 失败: %w", err)
	}
	info := convertPVCInfo(pvc)
	return &info, nil
}

// CreatePVC 创建 PVC
func (s *PVCService) CreatePVC(ctx context.Context, clusterName, namespace string, req CreatePVCRequest) (*PVCInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	if req.Namespace != "" {
		namespace = req.Namespace
	}
	accessModes := []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
	for _, m := range req.AccessModes {
		accessModes = append(accessModes, corev1.PersistentVolumeAccessMode(m))
	}
	if len(req.AccessModes) > 0 {
		accessModes = make([]corev1.PersistentVolumeAccessMode, 0, len(req.AccessModes))
		for _, m := range req.AccessModes {
			accessModes = append(accessModes, corev1.PersistentVolumeAccessMode(m))
		}
	}
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: namespace,
			Labels:    req.Labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: accessModes,
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(req.Storage),
				},
			},
		},
	}
	if req.StorageClassName != "" {
		pvc.Spec.StorageClassName = &req.StorageClassName
	}
	created, err := client.CoreV1().PersistentVolumeClaims(namespace).Create(ctx, pvc, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("创建 PVC 失败: %w", err)
	}
	info := convertPVCInfo(created)
	return &info, nil
}

// DeletePVC 删除 PVC
func (s *PVCService) DeletePVC(ctx context.Context, clusterName, namespace, name string) error {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return err
	}
	return client.CoreV1().PersistentVolumeClaims(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func convertPVCInfo(pvc *corev1.PersistentVolumeClaim) PVCInfo {
	modes := make([]string, 0, len(pvc.Spec.AccessModes))
	for _, m := range pvc.Spec.AccessModes {
		modes = append(modes, string(m))
	}
	capacity := ""
	if qty, ok := pvc.Status.Capacity[corev1.ResourceStorage]; ok {
		capacity = qty.String()
	}
	scName := ""
	if pvc.Spec.StorageClassName != nil {
		scName = *pvc.Spec.StorageClassName
	}
	return PVCInfo{
		Name:             pvc.Name,
		Namespace:        pvc.Namespace,
		Status:           string(pvc.Status.Phase),
		VolumeName:       pvc.Spec.VolumeName,
		StorageClassName: scName,
		AccessModes:      modes,
		Capacity:         capacity,
		CreationTime:     pvc.CreationTimestamp.Time,
		Labels:           pvc.Labels,
	}
}
