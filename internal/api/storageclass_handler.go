package api

import (
	"context"
	"net/http"

	"kube-tide/internal/core/k8s"

	"github.com/gin-gonic/gin"
)

type StorageClassHandler struct {
	service *k8s.StorageClassService
}

func NewStorageClassHandler(service *k8s.StorageClassService) *StorageClassHandler {
	return &StorageClassHandler{service: service}
}

func (h *StorageClassHandler) ListStorageClasses(c *gin.Context) {
	items, err := h.service.ListStorageClasses(context.Background(), c.Param("cluster"))
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "storageclass.listFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"storageclasses": items})
}

func (h *StorageClassHandler) GetStorageClass(c *gin.Context) {
	item, err := h.service.GetStorageClass(context.Background(), c.Param("cluster"), c.Param("storageclass"))
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "storageclass.getFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"storageclass": item})
}
