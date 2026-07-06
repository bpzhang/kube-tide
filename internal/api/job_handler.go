package api

import (
	"context"
	"net/http"

	"kube-tide/internal/core/k8s"

	"github.com/gin-gonic/gin"
)

type JobHandler struct {
	service *k8s.JobService
}

func NewJobHandler(service *k8s.JobService) *JobHandler {
	return &JobHandler{service: service}
}

func (h *JobHandler) ListJobs(c *gin.Context) {
	namespace := namespaceFromRequest(c)
	items, err := h.service.ListJobs(context.Background(), c.Param("cluster"), namespace)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "job.listFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"jobs": items})
}

func (h *JobHandler) GetJob(c *gin.Context) {
	item, err := h.service.GetJob(context.Background(), c.Param("cluster"), c.Param("namespace"), c.Param("job"))
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "job.getFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"job": item})
}

func (h *JobHandler) CreateJob(c *gin.Context) {
	var req k8s.CreateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, "job.invalidRequest", err.Error())
		return
	}
	item, err := h.service.CreateJob(context.Background(), c.Param("cluster"), c.Param("namespace"), req)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "job.createFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"job": item})
}

func (h *JobHandler) DeleteJob(c *gin.Context) {
	if err := h.service.DeleteJob(context.Background(), c.Param("cluster"), c.Param("namespace"), c.Param("job")); err != nil {
		ResponseError(c, http.StatusInternalServerError, "job.deleteFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"message": "Job deleted successfully"})
}
