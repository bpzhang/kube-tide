package api

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"kube-tide/internal/utils/logger"
	"net/http"
	"strconv"
	"time"

	"kube-tide/internal/core/k8s"

	"github.com/gin-gonic/gin"
)

// PodHandler pod management handler
type PodHandler struct {
	service *k8s.PodService
}

// NewPodHandler create pod management handler
func NewPodHandler(service *k8s.PodService) *PodHandler {
	return &PodHandler{
		service: service,
	}
}

// ListPods get all pods list
func (h *PodHandler) ListPods(c *gin.Context) {
	clusterName := c.Param("cluster")
	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "cluster name cannot be empty")
		return
	}

	pods, err := h.service.GetPods(context.Background(), clusterName)
	if err != nil {
		logger.Errorf("Failed to get pods: %s", err.Error())
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"pods": pods,
	})
}

// ListPodsByNamespace get pods list by specified namespace
func (h *PodHandler) ListPodsByNamespace(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "cluster name cannot be empty")
		return
	}
	if namespace == "" {
		ResponseError(c, http.StatusBadRequest, "namespace cannot be empty")
		logger.Errorf("Failed to list pods by namespace: namespace cannot be empty")
		return
	}

	pods, err := h.service.GetPodsByNamespace(context.Background(), clusterName, namespace)
	if err != nil {
		logger.Errorf("Failed to get pods by namespace: %s", err.Error())
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"pods": pods,
	})
}

// GetPodDetails get pod details
func (h *PodHandler) GetPodDetails(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	podName := c.Param("pod")
	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "cluster.clusterNameEmpty")
		return
	}
	if namespace == "" {
		ResponseError(c, http.StatusBadRequest, "namespace.namespaceNameEmpty")
		logger.Errorf("Failed to get pod details: namespace cannot be empty")
		return
	}
	if podName == "" {
		ResponseError(c, http.StatusBadRequest, "pod.podNameEmpty")
		logger.Errorf("Failed to get pod details: pod name cannot be empty")
		return
	}

	pod, err := h.service.GetPodDetails(context.Background(), clusterName, namespace, podName)
	if err != nil {
		if k8s.IsNotFoundError(err) {
			ResponseError(c, http.StatusNotFound, "pod.notFound")
			return
		}
		FailWithError(c, http.StatusInternalServerError, "pod.fetchFailed", err)
		return
	}

	ResponseSuccess(c, gin.H{
		"pod": pod,
	})
}

// DeletePod Delete a Pod
func (h *PodHandler) DeletePod(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	podName := c.Param("pod")

	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "cluster.clusterNameEmpty")
		return
	}
	if namespace == "" {
		ResponseError(c, http.StatusBadRequest, "namespace.namespaceNameEmpty")
		return
	}
	if podName == "" {
		ResponseError(c, http.StatusBadRequest, "pod.podNameEmpty")
		return
	}

	// Run deletion operation
	err := h.service.DeletePod(context.Background(), clusterName, namespace, podName)
	if err != nil {
		logger.Errorf("Failed to delete pod %s/%s: %v", namespace, podName, err)
		FailWithError(c, http.StatusInternalServerError, "pod.deleteFailed", err)
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "pod.deleteSuccess",
	})
}

// GetPodLogs Get Pod logs
func (h *PodHandler) GetPodLogs(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	podName := c.Param("pod")
	container := c.DefaultQuery("container", "")
	tailLines := c.DefaultQuery("tailLines", "100")

	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "cluster.clusterNameEmpty")
		return
	}
	if namespace == "" {
		ResponseError(c, http.StatusBadRequest, "namespace.namespaceNameEmpty")
		return
	}
	if podName == "" {
		ResponseError(c, http.StatusBadRequest, "pod.podNameEmpty")
		return
	}

	// Convert tailLines to int
	tailInt, _ := strconv.ParseInt(tailLines, 10, 64)

	// Get logs
	logs, err := h.service.GetPodLogs(context.Background(), clusterName, namespace, podName, container, tailInt)
	if err != nil {
		logger.Errorf("Failed to get pod logs %s/%s: %v", namespace, podName, err)
		FailWithError(c, http.StatusInternalServerError, "pod.logFailed", err)
		return
	}

	ResponseSuccess(c, gin.H{
		"logs": logs,
	})
}

// StreamPodLogs stream pod logs (real-time logs)
func (h *PodHandler) StreamPodLogs(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	podName := c.Param("pod")
	containerName := c.Query("container")
	tailLinesStr := c.Query("tailLines")
	followStr := c.DefaultQuery("follow", "true")

	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "cluster name cannot be empty")
		return
	}
	if namespace == "" {
		ResponseError(c, http.StatusBadRequest, "namespace cannot be empty")
		logger.Errorf("Failed to stream pod logs: namespace cannot be empty")
		return
	}
	if podName == "" {
		ResponseError(c, http.StatusBadRequest, "pod name cannot be empty")
		logger.Errorf("Failed to stream pod logs: pod name cannot be empty")
		return
	}

	var tailLines int64 = 100 // default to 100 lines
	if tailLinesStr != "" {
		if l, err := strconv.ParseInt(tailLinesStr, 10, 64); err == nil {
			tailLines = l
		}
	}

	follow := followStr == "true"

	// Set response headers to indicate this is an SSE stream
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*") // Allow cross-origin
	c.Writer.Header().Set("X-Accel-Buffering", "no")          // Disable Nginx buffering if using Nginx

	// Immediately flush headers
	c.Writer.Flush()

	// get the context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour) // increase timeout to 1 hour
	defer cancel()

	// send initial message to the client
	fmt.Fprintf(c.Writer, "data: %s\n\n", "logs are starting...")
	c.Writer.Flush()

	logStream, err := h.service.StreamPodLogs(ctx, clusterName, namespace, podName, containerName, tailLines, follow)
	if err != nil {
		errMsg := fmt.Sprintf("Get pod logs failed: %s", err.Error())
		// send error message to the client
		fmt.Fprintf(c.Writer, "data: {\"error\": \"%s\"}\n\n", errMsg)
		c.Writer.Flush()
		return
	}
	defer logStream.Close()

	// Use bufio to read the log line by line and send
	scanner := bufio.NewScanner(logStream)

	// Increase the buffer size of the scanner to handle longer log lines
	const maxScanTokenSize = 1024 * 1024 // 1MB
	scanBuf := make([]byte, maxScanTokenSize)
	scanner.Buffer(scanBuf, maxScanTokenSize)

	// Send heartbeat to keep connected
	heartbeat := time.NewTicker(30 * time.Second)
	defer heartbeat.Stop()

	// Set done channel to notify heartbeat goroutine to finish
	done := make(chan bool)
	defer close(done)

	// Start heartbeat goroutine
	go func() {
		for {
			select {
			case <-heartbeat.C:
				// Send comment line as heartbeat
				fmt.Fprintf(c.Writer, ": heartbeat\n\n")
				c.Writer.Flush()
			case <-done:
				return
			}
		}
	}()

	c.Stream(func(w io.Writer) bool {
		if !scanner.Scan() {
			// Send a heartbeat to ensure the client knows we are still online
			fmt.Fprintf(w, ": heartbeat\n\n")
			// Wait a short time to avoid returning false immediately when logs stop
			select {
			case <-time.After(500 * time.Millisecond):
				// If no more logs, check if the scanner has errors
				if err := scanner.Err(); err != nil {
					fmt.Fprintf(w, "data: {\"error\": \"%s\"}\n\n", err.Error())
					return false
				}
				// If follow is true and no errors, continue waiting
				if follow {
					return true
				}
			case <-ctx.Done():
				// Context has been canceled
				fmt.Fprintf(w, "data: {\"status\": \"Log stream closed\"}\n\n")
				return false
			}
			return false
		}

		line := scanner.Text()

		// Send log line as SSE event
		fmt.Fprintf(w, "data: %s\n\n", line)

		// Flush the buffer to ensure data is sent
		c.Writer.Flush()

		return true
	})
}

// GetPodsBySelector
func (h *PodHandler) GetPodsBySelector(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	logger.Infof("Get pods by selector for cluster: %s, namespace: %s", clusterName, namespace)
	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "cluster name cannot be empty")
		logger.Errorf("Failed to get pods by selector: cluster name cannot be empty")
		return
	}
	if namespace == "" {
		ResponseError(c, http.StatusBadRequest, "namespace cannot be empty")
		logger.Errorf("Failed to get pods by selector: namespace cannot be empty")
		return
	}

	var selector map[string]string
	if err := c.ShouldBindJSON(&selector); err != nil {
		ResponseError(c, http.StatusBadRequest, "invalid label selector")
		return
	}

	pods, err := h.service.GetPodsBySelector(context.Background(), clusterName, namespace, selector)
	if err != nil {
		logger.Errorf("Failed to get pods by selector: %s", err.Error())
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"pods": pods,
	})
}

// CheckPodExists Check if a Pod exists
func (h *PodHandler) CheckPodExists(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	podName := c.Param("pod")

	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "cluster.clusterNameEmpty")
		return
	}
	if namespace == "" {
		ResponseError(c, http.StatusBadRequest, "namespace.namespaceNameEmpty")
		return
	}
	if podName == "" {
		ResponseError(c, http.StatusBadRequest, "pod.podNameEmpty")
		return
	}

	_, exists, err := h.service.CheckPodExists(context.Background(), clusterName, namespace, podName)
	if err != nil {
		FailWithError(c, http.StatusInternalServerError, "pod.fetchFailed", err)
		return
	}

	ResponseSuccess(c, gin.H{
		"exists": exists,
	})
}

// GetPodEvents
func (h *PodHandler) GetPodEvents(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	podName := c.Param("pod")
	logger.Infof("Get pod events for cluster: %s, namespace: %s, pod: %s", clusterName, namespace, podName)
	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "cluster name cannot be empty")
		return
	}
	if namespace == "" {
		ResponseError(c, http.StatusBadRequest, "namespace cannot be empty")
		return
	}
	if podName == "" {
		ResponseError(c, http.StatusBadRequest, "Pod name cannot be empty")
		logger.Errorf("Failed to get pod events: pod name cannot be empty")
		return
	}

	events, err := h.service.GetPodEvents(context.Background(), clusterName, namespace, podName)
	if err != nil {
		logger.Errorf("Failed to get pod events: %s", err.Error())
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"events": events,
	})
}
