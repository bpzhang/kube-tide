package api

import (
	"context"
	"net/http"

	"kube-tide/internal/core/k8s"

	"github.com/gin-gonic/gin"
)

type CronJobHandler struct {
	service *k8s.CronJobService
}

func NewCronJobHandler(service *k8s.CronJobService) *CronJobHandler {
	return &CronJobHandler{service: service}
}

func (h *CronJobHandler) ListCronJobs(c *gin.Context) {
	namespace := namespaceFromRequest(c)
	items, err := h.service.ListCronJobs(context.Background(), c.Param("cluster"), namespace)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "cronjob.listFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"cronjobs": items})
}

func (h *CronJobHandler) GetCronJob(c *gin.Context) {
	item, err := h.service.GetCronJob(context.Background(), c.Param("cluster"), c.Param("namespace"), c.Param("cronjob"))
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "cronjob.getFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"cronjob": item})
}

func (h *CronJobHandler) CreateCronJob(c *gin.Context) {
	var req k8s.CreateCronJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, "cronjob.invalidRequest", err.Error())
		return
	}
	item, err := h.service.CreateCronJob(context.Background(), c.Param("cluster"), c.Param("namespace"), req)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "cronjob.createFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"cronjob": item})
}

func (h *CronJobHandler) UpdateCronJob(c *gin.Context) {
	var req k8s.UpdateCronJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, "cronjob.invalidRequest", err.Error())
		return
	}
	item, err := h.service.UpdateCronJob(context.Background(), c.Param("cluster"), c.Param("namespace"), c.Param("cronjob"), req)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "cronjob.updateFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"cronjob": item})
}

func (h *CronJobHandler) SuspendCronJob(c *gin.Context) {
	var req struct {
		Suspend bool `json:"suspend"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, "cronjob.invalidRequest", err.Error())
		return
	}
	item, err := h.service.SuspendCronJob(context.Background(), c.Param("cluster"), c.Param("namespace"), c.Param("cronjob"), req.Suspend)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "cronjob.suspendFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"cronjob": item})
}

func (h *CronJobHandler) DeleteCronJob(c *gin.Context) {
	if err := h.service.DeleteCronJob(context.Background(), c.Param("cluster"), c.Param("namespace"), c.Param("cronjob")); err != nil {
		ResponseError(c, http.StatusInternalServerError, "cronjob.deleteFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"message": "CronJob deleted successfully"})
}
