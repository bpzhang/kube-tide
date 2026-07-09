package api

import (
	"context"
	"net/http"

	"kube-tide/internal/core/k8s"

	"github.com/gin-gonic/gin"
)

type TrafficTopologyHandler struct {
	service *k8s.TrafficTopologyService
}

func NewTrafficTopologyHandler(service *k8s.TrafficTopologyService) *TrafficTopologyHandler {
	return &TrafficTopologyHandler{service: service}
}

func (h *TrafficTopologyHandler) GetTrafficTopology(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := namespaceFromRequest(c)
	topology, err := h.service.GetTrafficTopology(context.Background(), clusterName, namespace)
	if err != nil {
		FailWithError(c, http.StatusInternalServerError, "trafficTopology.fetchFailed", err)
		return
	}
	ResponseSuccess(c, topology)
}
