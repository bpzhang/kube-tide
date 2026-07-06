package api

import (
	"context"
	"net/http"

	"kube-tide/internal/core/k8s"

	"github.com/gin-gonic/gin"
)

type PVCHandler struct {
	service *k8s.PVCService
}

func NewPVCHandler(service *k8s.PVCService) *PVCHandler {
	return &PVCHandler{service: service}
}

func (h *PVCHandler) ListPVCs(c *gin.Context) {
	namespace := namespaceFromRequest(c)
	items, err := h.service.ListPVCs(context.Background(), c.Param("cluster"), namespace)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "pvc.listFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"pvcs": items})
}

func (h *PVCHandler) GetPVC(c *gin.Context) {
	item, err := h.service.GetPVC(context.Background(), c.Param("cluster"), c.Param("namespace"), c.Param("pvc"))
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "pvc.getFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"pvc": item})
}

func (h *PVCHandler) CreatePVC(c *gin.Context) {
	var req k8s.CreatePVCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, "pvc.invalidRequest", err.Error())
		return
	}
	item, err := h.service.CreatePVC(context.Background(), c.Param("cluster"), c.Param("namespace"), req)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "pvc.createFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"pvc": item})
}

func (h *PVCHandler) DeletePVC(c *gin.Context) {
	if err := h.service.DeletePVC(context.Background(), c.Param("cluster"), c.Param("namespace"), c.Param("pvc")); err != nil {
		ResponseError(c, http.StatusInternalServerError, "pvc.deleteFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"message": "PVC deleted successfully"})
}
