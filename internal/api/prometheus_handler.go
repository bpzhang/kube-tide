package api

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"kube-tide/internal/core/k8s"

	"github.com/gin-gonic/gin"
)

type PrometheusHandler struct {
	service *k8s.PrometheusService
}

func NewPrometheusHandler(service *k8s.PrometheusService) *PrometheusHandler {
	return &PrometheusHandler{service: service}
}

func (h *PrometheusHandler) QueryRange(c *gin.Context) {
	clusterName := c.Param("cluster")
	params := k8s.QueryRangeParams{
		Query: c.Query("query"),
		Start: c.Query("start"),
		End:   c.Query("end"),
		Step:  c.Query("step"),
	}
	if params.Query == "" || params.Start == "" || params.End == "" || params.Step == "" {
		var body k8s.QueryRangeParams
		if err := c.ShouldBindJSON(&body); err == nil {
			params = body
		}
	}
	if params.Query == "" {
		ResponseError(c, http.StatusBadRequest, "prometheus.queryRequired")
		return
	}
	if len(params.Query) > k8s.MaxPrometheusQueryLen() {
		ResponseError(c, http.StatusBadRequest, "prometheus.queryTooLong")
		return
	}
	if params.Start == "" || params.End == "" || params.Step == "" {
		ResponseError(c, http.StatusBadRequest, "prometheus.paramsRequired")
		return
	}

	timeout := 30 * time.Second
	if t := c.Query("timeout"); t != "" {
		if seconds, err := strconv.Atoi(t); err == nil && seconds > 0 {
			timeout = time.Duration(seconds) * time.Second
		}
	}

	result, err := h.service.QueryRange(context.Background(), clusterName, params, timeout)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "prometheus.queryFailed", err.Error())
		return
	}
	c.Data(http.StatusOK, "application/json", result)
}
