package k8s

import (
	"context"
	"fmt"
	"sort"
	"time"

	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// PDBService PodDisruptionBudget 管理服务
type PDBService struct {
	clientManager *ClientManager
}

// NewPDBService 创建 PDB 服务
func NewPDBService(clientManager *ClientManager) *PDBService {
	return &PDBService{clientManager: clientManager}
}

// PDBInfo PDB 摘要
type PDBInfo struct {
	Name               string            `json:"name"`
	Namespace          string            `json:"namespace"`
	MinAvailable       string            `json:"minAvailable,omitempty"`
	MaxUnavailable     string            `json:"maxUnavailable,omitempty"`
	Selector           map[string]string `json:"selector,omitempty"`
	CurrentHealthy     int32             `json:"currentHealthy"`
	DesiredHealthy     int32             `json:"desiredHealthy"`
	DisruptionsAllowed int32             `json:"disruptionsAllowed"`
	CreationTime       time.Time         `json:"creationTime"`
	Labels             map[string]string `json:"labels,omitempty"`
}

// CreatePDBRequest 创建 PDB 请求
type CreatePDBRequest struct {
	Name           string            `json:"name" binding:"required"`
	Namespace      string            `json:"namespace"`
	Labels         map[string]string `json:"labels,omitempty"`
	Selector       map[string]string `json:"selector" binding:"required"`
	MinAvailable   string            `json:"minAvailable,omitempty"`
	MaxUnavailable string            `json:"maxUnavailable,omitempty"`
}

// UpdatePDBRequest 更新 PDB 请求
type UpdatePDBRequest struct {
	Labels         map[string]string `json:"labels,omitempty"`
	MinAvailable   *string           `json:"minAvailable,omitempty"`
	MaxUnavailable *string           `json:"maxUnavailable,omitempty"`
	Selector       map[string]string `json:"selector,omitempty"`
}

// ListPDBs 获取 PDB 列表
func (s *PDBService) ListPDBs(ctx context.Context, clusterName, namespace string) ([]PDBInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	var list *policyv1.PodDisruptionBudgetList
	if namespace == "all" || namespace == "" {
		list, err = client.PolicyV1().PodDisruptionBudgets("").List(ctx, metav1.ListOptions{})
	} else {
		list, err = client.PolicyV1().PodDisruptionBudgets(namespace).List(ctx, metav1.ListOptions{})
	}
	if err != nil {
		return nil, fmt.Errorf("获取 PDB 列表失败: %w", err)
	}
	result := make([]PDBInfo, 0, len(list.Items))
	for _, pdb := range list.Items {
		result = append(result, convertPDBInfo(&pdb))
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreationTime.After(result[j].CreationTime)
	})
	return result, nil
}

// GetPDB 获取 PDB 详情
func (s *PDBService) GetPDB(ctx context.Context, clusterName, namespace, name string) (*PDBInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	pdb, err := client.PolicyV1().PodDisruptionBudgets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 PDB 失败: %w", err)
	}
	info := convertPDBInfo(pdb)
	return &info, nil
}

// CreatePDB 创建 PDB
func (s *PDBService) CreatePDB(ctx context.Context, clusterName, namespace string, req CreatePDBRequest) (*PDBInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	if req.Namespace != "" {
		namespace = req.Namespace
	}
	spec := policyv1.PodDisruptionBudgetSpec{
		Selector: &metav1.LabelSelector{MatchLabels: req.Selector},
	}
	if req.MinAvailable != "" {
		v := intstrFromString(req.MinAvailable)
		spec.MinAvailable = &v
	}
	if req.MaxUnavailable != "" {
		v := intstrFromString(req.MaxUnavailable)
		spec.MaxUnavailable = &v
	}
	pdb := &policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: namespace,
			Labels:    req.Labels,
		},
		Spec: spec,
	}
	created, err := client.PolicyV1().PodDisruptionBudgets(namespace).Create(ctx, pdb, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("创建 PDB 失败: %w", err)
	}
	info := convertPDBInfo(created)
	return &info, nil
}

// UpdatePDB 更新 PDB
func (s *PDBService) UpdatePDB(ctx context.Context, clusterName, namespace, name string, req UpdatePDBRequest) (*PDBInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	pdb, err := client.PolicyV1().PodDisruptionBudgets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 PDB 失败: %w", err)
	}
	if req.Labels != nil {
		pdb.Labels = req.Labels
	}
	if req.Selector != nil {
		pdb.Spec.Selector = &metav1.LabelSelector{MatchLabels: req.Selector}
	}
	if req.MinAvailable != nil {
		v := intstrFromString(*req.MinAvailable)
		pdb.Spec.MinAvailable = &v
		pdb.Spec.MaxUnavailable = nil
	}
	if req.MaxUnavailable != nil {
		v := intstrFromString(*req.MaxUnavailable)
		pdb.Spec.MaxUnavailable = &v
		pdb.Spec.MinAvailable = nil
	}
	updated, err := client.PolicyV1().PodDisruptionBudgets(namespace).Update(ctx, pdb, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("更新 PDB 失败: %w", err)
	}
	info := convertPDBInfo(updated)
	return &info, nil
}

// DeletePDB 删除 PDB
func (s *PDBService) DeletePDB(ctx context.Context, clusterName, namespace, name string) error {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return err
	}
	return client.PolicyV1().PodDisruptionBudgets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func convertPDBInfo(pdb *policyv1.PodDisruptionBudget) PDBInfo {
	info := PDBInfo{
		Name:               pdb.Name,
		Namespace:          pdb.Namespace,
		CurrentHealthy:     pdb.Status.CurrentHealthy,
		DesiredHealthy:     pdb.Status.DesiredHealthy,
		DisruptionsAllowed: pdb.Status.DisruptionsAllowed,
		CreationTime:       pdb.CreationTimestamp.Time,
		Labels:             pdb.Labels,
	}
	if pdb.Spec.Selector != nil {
		info.Selector = pdb.Spec.Selector.MatchLabels
	}
	if pdb.Spec.MinAvailable != nil {
		info.MinAvailable = pdb.Spec.MinAvailable.String()
	}
	if pdb.Spec.MaxUnavailable != nil {
		info.MaxUnavailable = pdb.Spec.MaxUnavailable.String()
	}
	return info
}

func intstrFromString(s string) intstr.IntOrString {
	if i, err := parseInt32(s); err == nil {
		return intstr.FromInt(int(i))
	}
	return intstr.FromString(s)
}

func parseInt32(s string) (int32, error) {
	var v int32
	_, err := fmt.Sscanf(s, "%d", &v)
	return v, err
}
