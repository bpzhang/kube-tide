package api

import (
	"context"
	"net/http"

	"kube-tide/internal/core/k8s"

	"github.com/gin-gonic/gin"
)

type PVHandler struct {
	service *k8s.PVService
}

func NewPVHandler(service *k8s.PVService) *PVHandler {
	return &PVHandler{service: service}
}

func (h *PVHandler) ListPVs(c *gin.Context) {
	items, err := h.service.ListPVs(context.Background(), c.Param("cluster"))
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "pv.listFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"pvs": items})
}

func (h *PVHandler) GetPV(c *gin.Context) {
	item, err := h.service.GetPV(context.Background(), c.Param("cluster"), c.Param("pv"))
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "pv.getFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"pv": item})
}
