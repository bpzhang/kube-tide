package k8s

import (
	"context"
	"fmt"
	"strings"

	authv1 "k8s.io/api/authentication/v1"
	authorizationv1 "k8s.io/api/authorization/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const kubeTideClusterRoleName = "kube-tide"

// PermissionCheck 单项权限检查结果
type PermissionCheck struct {
	Name    string `json:"name"`
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason,omitempty"`
}

// ClusterPermissionReport 集群权限检查报告
type ClusterPermissionReport struct {
	Identity   string            `json:"identity"`
	AllGranted bool              `json:"allGranted"`
	Checks     []PermissionCheck `json:"checks"`
	RBACApplied bool             `json:"rbacApplied,omitempty"`
	Message    string            `json:"message,omitempty"`
}

var requiredMonitoringChecks = []struct {
	name      string
	group     string
	resource  string
	verb      string
	subresource string
}{
	{name: "metrics-server pods", group: "metrics.k8s.io", resource: "pods", verb: "get"},
	{name: "metrics-server nodes", group: "metrics.k8s.io", resource: "nodes", verb: "list"},
	{name: "kubelet stats (nodes/proxy)", group: "", resource: "nodes", verb: "get", subresource: "proxy"},
	{name: "read pods", group: "", resource: "pods", verb: "list"},
	{name: "read PVC", group: "", resource: "persistentvolumeclaims", verb: "get"},
}

func kubeTideClusterRoleRules() []rbacv1.PolicyRule {
	return []rbacv1.PolicyRule{
		{APIGroups: []string{""}, Resources: []string{"namespaces", "nodes", "pods", "pods/log", "services", "endpoints", "events", "configmaps", "secrets", "persistentvolumeclaims", "persistentvolumes", "resourcequotas", "limitranges"}, Verbs: []string{"get", "list", "watch"}},
		{APIGroups: []string{""}, Resources: []string{"namespaces"}, Verbs: []string{"create", "update", "patch", "delete"}},
		{APIGroups: []string{""}, Resources: []string{"configmaps", "secrets"}, Verbs: []string{"create", "update", "patch", "delete"}},
		{APIGroups: []string{""}, Resources: []string{"nodes"}, Verbs: []string{"update", "patch", "delete"}},
		{APIGroups: []string{""}, Resources: []string{"pods"}, Verbs: []string{"create", "update", "patch", "delete"}},
		{APIGroups: []string{""}, Resources: []string{"pods/eviction"}, Verbs: []string{"create"}},
		{APIGroups: []string{""}, Resources: []string{"services", "endpoints"}, Verbs: []string{"create", "update", "patch", "delete"}},
		{APIGroups: []string{""}, Resources: []string{"pods/exec"}, Verbs: []string{"create"}},
		{APIGroups: []string{""}, Resources: []string{"persistentvolumeclaims"}, Verbs: []string{"create", "delete"}},
		{APIGroups: []string{""}, Resources: []string{"resourcequotas", "limitranges"}, Verbs: []string{"create", "update", "patch", "delete"}},
		{APIGroups: []string{""}, Resources: []string{"nodes/proxy"}, Verbs: []string{"get"}},
		{APIGroups: []string{"metrics.k8s.io"}, Resources: []string{"pods", "nodes"}, Verbs: []string{"get", "list"}},
		{APIGroups: []string{"apps"}, Resources: []string{"deployments", "replicasets", "statefulsets", "daemonsets"}, Verbs: []string{"get", "list", "watch", "create", "update", "patch", "delete"}},
		{APIGroups: []string{"batch"}, Resources: []string{"jobs", "cronjobs"}, Verbs: []string{"get", "list", "watch", "create", "update", "patch", "delete"}},
		{APIGroups: []string{"networking.k8s.io"}, Resources: []string{"ingresses", "networkpolicies"}, Verbs: []string{"get", "list", "watch", "create", "update", "patch", "delete"}},
		{APIGroups: []string{"autoscaling"}, Resources: []string{"horizontalpodautoscalers"}, Verbs: []string{"get", "list", "watch", "create", "update", "patch", "delete"}},
		{APIGroups: []string{"policy"}, Resources: []string{"poddisruptionbudgets"}, Verbs: []string{"get", "list", "watch", "create", "update", "patch", "delete"}},
		{APIGroups: []string{"storage.k8s.io"}, Resources: []string{"storageclasses"}, Verbs: []string{"get", "list", "watch"}},
		{APIGroups: []string{"rbac.authorization.k8s.io"}, Resources: []string{"roles", "rolebindings", "clusterroles", "clusterrolebindings"}, Verbs: []string{"get", "list", "watch"}},
		{APIGroups: []string{"rbac.authorization.k8s.io"}, Resources: []string{"rolebindings", "clusterrolebindings"}, Verbs: []string{"create", "delete"}},
		{APIGroups: []string{"rbac.authorization.k8s.io"}, Resources: []string{"clusterroles", "clusterrolebindings"}, Verbs: []string{"create", "update", "patch"}},
		{APIGroups: []string{""}, Resources: []string{"serviceaccounts"}, Verbs: []string{"get", "list", "watch", "create", "update", "patch"}},
		{APIGroups: []string{"coordination.k8s.io"}, Resources: []string{"leases"}, Verbs: []string{"get", "list", "watch", "create", "update", "patch"}},
	}
}

// CheckClusterPermissions 检查磁盘/指标监控所需权限
func CheckClusterPermissions(ctx context.Context, client *kubernetes.Clientset) *ClusterPermissionReport {
	report := &ClusterPermissionReport{Checks: make([]PermissionCheck, 0, len(requiredMonitoringChecks))}

	review, err := client.AuthenticationV1().SelfSubjectReviews().Create(ctx, &authv1.SelfSubjectReview{}, metav1.CreateOptions{})
	if err == nil && review.Status.UserInfo.Username != "" {
		report.Identity = review.Status.UserInfo.Username
	}

	allGranted := true
	for _, check := range requiredMonitoringChecks {
		allowed, reason := canAccess(ctx, client, check.group, check.resource, check.verb, check.subresource)
		report.Checks = append(report.Checks, PermissionCheck{
			Name:    check.name,
			Allowed: allowed,
			Reason:  reason,
		})
		if !allowed {
			allGranted = false
		}
	}
	report.AllGranted = allGranted
	if !allGranted {
		report.Message = "缺少监控所需 RBAC，请对目标集群执行 deployments/k8s/kube-tide-rbac.yaml，或使用具备 cluster-admin 的 kubeconfig 添加集群以自动补齐"
	}
	return report
}

func canAccess(ctx context.Context, client *kubernetes.Clientset, group, resource, verb, subresource string) (bool, string) {
	review := &authorizationv1.SelfSubjectAccessReview{
		Spec: authorizationv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authorizationv1.ResourceAttributes{
				Group:       group,
				Resource:    resource,
				Subresource: subresource,
				Verb:        verb,
			},
		},
	}
	result, err := client.AuthorizationV1().SelfSubjectAccessReviews().Create(ctx, review, metav1.CreateOptions{})
	if err != nil {
		return false, err.Error()
	}
	if result.Status.Allowed {
		return true, ""
	}
	if result.Status.Reason != "" {
		return false, result.Status.Reason
	}
	return false, "forbidden"
}

// EnsureKubeTideRBAC 尝试在集群中创建/更新 kube-tide ClusterRole 并绑定当前 ServiceAccount
func EnsureKubeTideRBAC(ctx context.Context, client *kubernetes.Clientset) (bool, error) {
	subject, err := currentServiceAccountSubject(ctx, client)
	if err != nil {
		return false, err
	}
	if subject == nil {
		return false, fmt.Errorf("当前 kubeconfig 不是 ServiceAccount，无法自动绑定 RBAC")
	}

	if err := ensureServiceAccount(ctx, client, subject.Namespace, subject.Name); err != nil {
		return false, err
	}
	if err := upsertKubeTideClusterRole(ctx, client); err != nil {
		return false, err
	}
	if err := upsertKubeTideClusterRoleBinding(ctx, client, *subject); err != nil {
		return false, err
	}
	return true, nil
}

func currentServiceAccountSubject(ctx context.Context, client *kubernetes.Clientset) (*rbacv1.Subject, error) {
	review, err := client.AuthenticationV1().SelfSubjectReviews().Create(ctx, &authv1.SelfSubjectReview{}, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取当前身份失败: %w", err)
	}
	username := review.Status.UserInfo.Username
	const prefix = "system:serviceaccount:"
	if !strings.HasPrefix(username, prefix) {
		return nil, nil
	}
	parts := strings.Split(strings.TrimPrefix(username, prefix), ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("无法解析 ServiceAccount 身份: %s", username)
	}
	return &rbacv1.Subject{
		Kind:      rbacv1.ServiceAccountKind,
		Namespace: parts[0],
		Name:      parts[1],
	}, nil
}

func ensureServiceAccount(ctx context.Context, client *kubernetes.Clientset, namespace, name string) error {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	_, err := client.CoreV1().ServiceAccounts(namespace).Create(ctx, sa, metav1.CreateOptions{})
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("创建 ServiceAccount 失败: %w", err)
	}
	return nil
}

func upsertKubeTideClusterRole(ctx context.Context, client *kubernetes.Clientset) error {
	rules := kubeTideClusterRoleRules()
	existing, err := client.RbacV1().ClusterRoles().Get(ctx, kubeTideClusterRoleName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = client.RbacV1().ClusterRoles().Create(ctx, &rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{Name: kubeTideClusterRoleName},
			Rules:      rules,
		}, metav1.CreateOptions{})
		return err
	}
	if err != nil {
		return err
	}
	existing.Rules = rules
	_, err = client.RbacV1().ClusterRoles().Update(ctx, existing, metav1.UpdateOptions{})
	return err
}

func upsertKubeTideClusterRoleBinding(ctx context.Context, client *kubernetes.Clientset, subject rbacv1.Subject) error {
	binding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: kubeTideClusterRoleName},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     kubeTideClusterRoleName,
		},
		Subjects: []rbacv1.Subject{subject},
	}

	existing, err := client.RbacV1().ClusterRoleBindings().Get(ctx, kubeTideClusterRoleName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = client.RbacV1().ClusterRoleBindings().Create(ctx, binding, metav1.CreateOptions{})
		return err
	}
	if err != nil {
		return err
	}
	existing.RoleRef = binding.RoleRef
	existing.Subjects = binding.Subjects
	_, err = client.RbacV1().ClusterRoleBindings().Update(ctx, existing, metav1.UpdateOptions{})
	return err
}

// PrepareClusterAccess 添加集群后检查权限，并在可能时自动补齐 RBAC
func PrepareClusterAccess(ctx context.Context, client *kubernetes.Clientset) *ClusterPermissionReport {
	report := CheckClusterPermissions(ctx, client)
	if report.AllGranted {
		report.Message = "监控所需权限已满足"
		return report
	}

	canManageRBAC, _ := canAccess(ctx, client, rbacv1.GroupName, "clusterroles", "create", "")
	if !canManageRBAC {
		return report
	}

	applied, err := EnsureKubeTideRBAC(ctx, client)
	if err != nil {
		report.Message = fmt.Sprintf("%s；自动补齐 RBAC 失败: %v", report.Message, err)
		return report
	}
	if !applied {
		return report
	}

	report.RBACApplied = true
	recheck := CheckClusterPermissions(ctx, client)
	report.Checks = recheck.Checks
	report.AllGranted = recheck.AllGranted
	if recheck.AllGranted {
		report.Message = "已自动创建/更新 kube-tide RBAC，监控权限已满足"
	} else {
		report.Message = "已尝试自动补齐 RBAC，部分权限仍不足，请确认 metrics-server 已安装"
	}
	return report
}
