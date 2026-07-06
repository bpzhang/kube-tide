package k8s

import (
	"context"
	"fmt"
	"sort"
	"time"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RBACService RBAC 管理服务
type RBACService struct {
	clientManager *ClientManager
}

// NewRBACService 创建 RBAC 服务
func NewRBACService(clientManager *ClientManager) *RBACService {
	return &RBACService{clientManager: clientManager}
}

// RoleInfo Role 摘要
type RoleInfo struct {
	Name         string            `json:"name"`
	Namespace    string            `json:"namespace,omitempty"`
	RuleCount    int               `json:"ruleCount"`
	CreationTime time.Time         `json:"creationTime"`
	Labels       map[string]string `json:"labels,omitempty"`
}

// RoleBindingInfo RoleBinding 摘要
type RoleBindingInfo struct {
	Name         string            `json:"name"`
	Namespace    string            `json:"namespace,omitempty"`
	RoleRef      RBACRoleRef       `json:"roleRef"`
	SubjectCount int               `json:"subjectCount"`
	Subjects     []RBACSubject     `json:"subjects,omitempty"`
	CreationTime time.Time         `json:"creationTime"`
	Labels       map[string]string `json:"labels,omitempty"`
}

// RBACRoleRef 角色引用
type RBACRoleRef struct {
	Kind     string `json:"kind"`
	Name     string `json:"name"`
	APIGroup string `json:"apiGroup,omitempty"`
}

// RBACSubject 绑定主体
type RBACSubject struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

// CreateRoleBindingRequest 创建 RoleBinding 请求
type CreateRoleBindingRequest struct {
	Name      string        `json:"name" binding:"required"`
	Namespace string        `json:"namespace"`
	Labels    map[string]string `json:"labels,omitempty"`
	RoleRef   RBACRoleRef   `json:"roleRef" binding:"required"`
	Subjects  []RBACSubject `json:"subjects" binding:"required"`
}

// CreateClusterRoleBindingRequest 创建 ClusterRoleBinding 请求
type CreateClusterRoleBindingRequest struct {
	Name     string        `json:"name" binding:"required"`
	Labels   map[string]string `json:"labels,omitempty"`
	RoleRef  RBACRoleRef   `json:"roleRef" binding:"required"`
	Subjects []RBACSubject `json:"subjects" binding:"required"`
}

// ListRoles 获取 Role 列表
func (s *RBACService) ListRoles(ctx context.Context, clusterName, namespace string) ([]RoleInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	var list *rbacv1.RoleList
	if namespace == "all" || namespace == "" {
		list, err = client.RbacV1().Roles("").List(ctx, metav1.ListOptions{})
	} else {
		list, err = client.RbacV1().Roles(namespace).List(ctx, metav1.ListOptions{})
	}
	if err != nil {
		return nil, fmt.Errorf("获取 Role 列表失败: %w", err)
	}
	result := make([]RoleInfo, 0, len(list.Items))
	for _, role := range list.Items {
		result = append(result, convertRoleInfo(&role))
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreationTime.After(result[j].CreationTime)
	})
	return result, nil
}

// GetRole 获取 Role 详情
func (s *RBACService) GetRole(ctx context.Context, clusterName, namespace, name string) (*rbacv1.Role, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	return client.RbacV1().Roles(namespace).Get(ctx, name, metav1.GetOptions{})
}

// ListClusterRoles 获取 ClusterRole 列表
func (s *RBACService) ListClusterRoles(ctx context.Context, clusterName string) ([]RoleInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	list, err := client.RbacV1().ClusterRoles().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 ClusterRole 列表失败: %w", err)
	}
	result := make([]RoleInfo, 0, len(list.Items))
	for _, role := range list.Items {
		result = append(result, RoleInfo{
			Name:         role.Name,
			RuleCount:    len(role.Rules),
			CreationTime: role.CreationTimestamp.Time,
			Labels:       role.Labels,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result, nil
}

// GetClusterRole 获取 ClusterRole 详情
func (s *RBACService) GetClusterRole(ctx context.Context, clusterName, name string) (*rbacv1.ClusterRole, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	return client.RbacV1().ClusterRoles().Get(ctx, name, metav1.GetOptions{})
}

// ListRoleBindings 获取 RoleBinding 列表
func (s *RBACService) ListRoleBindings(ctx context.Context, clusterName, namespace string) ([]RoleBindingInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	var list *rbacv1.RoleBindingList
	if namespace == "all" || namespace == "" {
		list, err = client.RbacV1().RoleBindings("").List(ctx, metav1.ListOptions{})
	} else {
		list, err = client.RbacV1().RoleBindings(namespace).List(ctx, metav1.ListOptions{})
	}
	if err != nil {
		return nil, fmt.Errorf("获取 RoleBinding 列表失败: %w", err)
	}
	result := make([]RoleBindingInfo, 0, len(list.Items))
	for _, rb := range list.Items {
		result = append(result, convertRoleBindingInfo(&rb))
	}
	return result, nil
}

// GetRoleBinding 获取 RoleBinding 详情
func (s *RBACService) GetRoleBinding(ctx context.Context, clusterName, namespace, name string) (*RoleBindingInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	rb, err := client.RbacV1().RoleBindings(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 RoleBinding 失败: %w", err)
	}
	info := convertRoleBindingInfo(rb)
	return &info, nil
}

// CreateRoleBinding 创建 RoleBinding
func (s *RBACService) CreateRoleBinding(ctx context.Context, clusterName, namespace string, req CreateRoleBindingRequest) (*RoleBindingInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	if req.Namespace != "" {
		namespace = req.Namespace
	}
	rb := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: namespace,
			Labels:    req.Labels,
		},
		RoleRef:  buildRoleRef(req.RoleRef),
		Subjects: buildRBACSubjects(req.Subjects),
	}
	created, err := client.RbacV1().RoleBindings(namespace).Create(ctx, rb, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("创建 RoleBinding 失败: %w", err)
	}
	info := convertRoleBindingInfo(created)
	return &info, nil
}

// DeleteRoleBinding 删除 RoleBinding
func (s *RBACService) DeleteRoleBinding(ctx context.Context, clusterName, namespace, name string) error {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return err
	}
	return client.RbacV1().RoleBindings(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// ListClusterRoleBindings 获取 ClusterRoleBinding 列表
func (s *RBACService) ListClusterRoleBindings(ctx context.Context, clusterName string) ([]RoleBindingInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	list, err := client.RbacV1().ClusterRoleBindings().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 ClusterRoleBinding 列表失败: %w", err)
	}
	result := make([]RoleBindingInfo, 0, len(list.Items))
	for _, rb := range list.Items {
		result = append(result, convertClusterRoleBindingInfo(&rb))
	}
	return result, nil
}

// GetClusterRoleBinding 获取 ClusterRoleBinding 详情
func (s *RBACService) GetClusterRoleBinding(ctx context.Context, clusterName, name string) (*RoleBindingInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	rb, err := client.RbacV1().ClusterRoleBindings().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取 ClusterRoleBinding 失败: %w", err)
	}
	info := convertClusterRoleBindingInfo(rb)
	return &info, nil
}

// CreateClusterRoleBinding 创建 ClusterRoleBinding
func (s *RBACService) CreateClusterRoleBinding(ctx context.Context, clusterName string, req CreateClusterRoleBindingRequest) (*RoleBindingInfo, error) {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return nil, err
	}
	rb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   req.Name,
			Labels: req.Labels,
		},
		RoleRef:  buildRoleRef(req.RoleRef),
		Subjects: buildRBACSubjects(req.Subjects),
	}
	created, err := client.RbacV1().ClusterRoleBindings().Create(ctx, rb, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("创建 ClusterRoleBinding 失败: %w", err)
	}
	info := convertClusterRoleBindingInfo(created)
	return &info, nil
}

// DeleteClusterRoleBinding 删除 ClusterRoleBinding
func (s *RBACService) DeleteClusterRoleBinding(ctx context.Context, clusterName, name string) error {
	client, err := s.clientManager.GetClient(clusterName)
	if err != nil {
		return err
	}
	return client.RbacV1().ClusterRoleBindings().Delete(ctx, name, metav1.DeleteOptions{})
}

func convertRoleInfo(role *rbacv1.Role) RoleInfo {
	return RoleInfo{
		Name:         role.Name,
		Namespace:    role.Namespace,
		RuleCount:    len(role.Rules),
		CreationTime: role.CreationTimestamp.Time,
		Labels:       role.Labels,
	}
}

func convertRoleBindingInfo(rb *rbacv1.RoleBinding) RoleBindingInfo {
	return RoleBindingInfo{
		Name:         rb.Name,
		Namespace:    rb.Namespace,
		RoleRef:      RBACRoleRef{Kind: rb.RoleRef.Kind, Name: rb.RoleRef.Name, APIGroup: rb.RoleRef.APIGroup},
		SubjectCount: len(rb.Subjects),
		Subjects:     convertRBACSubjects(rb.Subjects),
		CreationTime: rb.CreationTimestamp.Time,
		Labels:       rb.Labels,
	}
}

func convertClusterRoleBindingInfo(rb *rbacv1.ClusterRoleBinding) RoleBindingInfo {
	return RoleBindingInfo{
		Name:         rb.Name,
		RoleRef:      RBACRoleRef{Kind: rb.RoleRef.Kind, Name: rb.RoleRef.Name, APIGroup: rb.RoleRef.APIGroup},
		SubjectCount: len(rb.Subjects),
		Subjects:     convertRBACSubjects(rb.Subjects),
		CreationTime: rb.CreationTimestamp.Time,
		Labels:       rb.Labels,
	}
}

func buildRoleRef(ref RBACRoleRef) rbacv1.RoleRef {
	apiGroup := ref.APIGroup
	if apiGroup == "" {
		apiGroup = rbacv1.GroupName
	}
	return rbacv1.RoleRef{APIGroup: apiGroup, Kind: ref.Kind, Name: ref.Name}
}

func buildRBACSubjects(subjects []RBACSubject) []rbacv1.Subject {
	result := make([]rbacv1.Subject, 0, len(subjects))
	for _, s := range subjects {
		result = append(result, rbacv1.Subject{Kind: s.Kind, Name: s.Name, Namespace: s.Namespace})
	}
	return result
}

func convertRBACSubjects(subjects []rbacv1.Subject) []RBACSubject {
	result := make([]RBACSubject, 0, len(subjects))
	for _, s := range subjects {
		result = append(result, RBACSubject{Kind: s.Kind, Name: s.Name, Namespace: s.Namespace})
	}
	return result
}
