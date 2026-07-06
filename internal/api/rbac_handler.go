package api

import (
	"context"
	"net/http"

	"kube-tide/internal/core/k8s"

	"github.com/gin-gonic/gin"
)

type RBACHandler struct {
	service *k8s.RBACService
}

func NewRBACHandler(service *k8s.RBACService) *RBACHandler {
	return &RBACHandler{service: service}
}

func (h *RBACHandler) ListRoles(c *gin.Context) {
	namespace := namespaceFromRequest(c)
	items, err := h.service.ListRoles(context.Background(), c.Param("cluster"), namespace)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "rbac.listRolesFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"roles": items})
}

func (h *RBACHandler) GetRole(c *gin.Context) {
	item, err := h.service.GetRole(context.Background(), c.Param("cluster"), c.Param("namespace"), c.Param("role"))
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "rbac.getRoleFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"role": item})
}

func (h *RBACHandler) ListClusterRoles(c *gin.Context) {
	items, err := h.service.ListClusterRoles(context.Background(), c.Param("cluster"))
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "rbac.listClusterRolesFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"clusterroles": items})
}

func (h *RBACHandler) GetClusterRole(c *gin.Context) {
	item, err := h.service.GetClusterRole(context.Background(), c.Param("cluster"), c.Param("clusterrole"))
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "rbac.getClusterRoleFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"clusterrole": item})
}

func (h *RBACHandler) ListRoleBindings(c *gin.Context) {
	namespace := namespaceFromRequest(c)
	items, err := h.service.ListRoleBindings(context.Background(), c.Param("cluster"), namespace)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "rbac.listRoleBindingsFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"rolebindings": items})
}

func (h *RBACHandler) GetRoleBinding(c *gin.Context) {
	item, err := h.service.GetRoleBinding(context.Background(), c.Param("cluster"), c.Param("namespace"), c.Param("rolebinding"))
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "rbac.getRoleBindingFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"rolebinding": item})
}

func (h *RBACHandler) CreateRoleBinding(c *gin.Context) {
	var req k8s.CreateRoleBindingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, "rbac.invalidRequest", err.Error())
		return
	}
	item, err := h.service.CreateRoleBinding(context.Background(), c.Param("cluster"), c.Param("namespace"), req)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "rbac.createRoleBindingFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"rolebinding": item})
}

func (h *RBACHandler) DeleteRoleBinding(c *gin.Context) {
	if err := h.service.DeleteRoleBinding(context.Background(), c.Param("cluster"), c.Param("namespace"), c.Param("rolebinding")); err != nil {
		ResponseError(c, http.StatusInternalServerError, "rbac.deleteRoleBindingFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"message": "RoleBinding deleted successfully"})
}

func (h *RBACHandler) ListClusterRoleBindings(c *gin.Context) {
	items, err := h.service.ListClusterRoleBindings(context.Background(), c.Param("cluster"))
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "rbac.listClusterRoleBindingsFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"clusterrolebindings": items})
}

func (h *RBACHandler) GetClusterRoleBinding(c *gin.Context) {
	item, err := h.service.GetClusterRoleBinding(context.Background(), c.Param("cluster"), c.Param("clusterrolebinding"))
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "rbac.getClusterRoleBindingFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"clusterrolebinding": item})
}

func (h *RBACHandler) CreateClusterRoleBinding(c *gin.Context) {
	var req k8s.CreateClusterRoleBindingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, "rbac.invalidRequest", err.Error())
		return
	}
	item, err := h.service.CreateClusterRoleBinding(context.Background(), c.Param("cluster"), req)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "rbac.createClusterRoleBindingFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"clusterrolebinding": item})
}

func (h *RBACHandler) DeleteClusterRoleBinding(c *gin.Context) {
	if err := h.service.DeleteClusterRoleBinding(context.Background(), c.Param("cluster"), c.Param("clusterrolebinding")); err != nil {
		ResponseError(c, http.StatusInternalServerError, "rbac.deleteClusterRoleBindingFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"message": "ClusterRoleBinding deleted successfully"})
}
