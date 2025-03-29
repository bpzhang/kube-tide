package api

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"kube-tide/internal/core/k8s"

	"github.com/gin-gonic/gin"
)

// PodHandler Pod管理处理器
type PodHandler struct {
	service *k8s.PodService
}

// NewPodHandler 创建Pod管理处理器
func NewPodHandler(service *k8s.PodService) *PodHandler {
	return &PodHandler{
		service: service,
	}
}

// ListPods 获取所有Pod列表
func (h *PodHandler) ListPods(c *gin.Context) {
	clusterName := c.Param("cluster")
	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称不能为空")
		return
	}

	pods, err := h.service.GetPods(context.Background(), clusterName)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"pods": pods,
	})
}

// ListPodsByNamespace 获取指定命名空间的Pod列表
func (h *PodHandler) ListPodsByNamespace(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称不能为空")
		return
	}
	if namespace == "" {
		ResponseError(c, http.StatusBadRequest, "命名空间不能为空")
		return
	}

	pods, err := h.service.GetPodsByNamespace(context.Background(), clusterName, namespace)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"pods": pods,
	})
}

// GetPodDetails 获取Pod详情
func (h *PodHandler) GetPodDetails(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	podName := c.Param("pod")
	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称不能为空")
		return
	}
	if namespace == "" {
		ResponseError(c, http.StatusBadRequest, "命名空间不能为空")
		return
	}
	if podName == "" {
		ResponseError(c, http.StatusBadRequest, "Pod名称不能为空")
		return
	}

	pod, err := h.service.GetPodDetails(context.Background(), clusterName, namespace, podName)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"pod": pod,
	})
}

// DeletePod 删除Pod
func (h *PodHandler) DeletePod(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	podName := c.Param("pod")
	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称不能为空")
		return
	}
	if namespace == "" {
		ResponseError(c, http.StatusBadRequest, "命名空间不能为空")
		return
	}
	if podName == "" {
		ResponseError(c, http.StatusBadRequest, "Pod名称不能为空")
		return
	}

	err := h.service.DeletePod(context.Background(), clusterName, namespace, podName)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, nil)
}

// GetPodLogs 获取Pod日志
func (h *PodHandler) GetPodLogs(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	podName := c.Param("pod")
	containerName := c.Query("container")
	tailLines := c.Query("tailLines")

	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称不能为空")
		return
	}
	if namespace == "" {
		ResponseError(c, http.StatusBadRequest, "命名空间不能为空")
		return
	}
	if podName == "" {
		ResponseError(c, http.StatusBadRequest, "Pod名称不能为空")
		return
	}

	var lines int64 = 100 // 默认获取100行
	if tailLines != "" {
		if l, err := strconv.ParseInt(tailLines, 10, 64); err == nil {
			lines = l
		}
	}

	logs, err := h.service.GetPodLogs(context.Background(), clusterName, namespace, podName, containerName, lines)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"logs": logs,
	})
}

// StreamPodLogs 流式获取Pod日志（实时日志）
func (h *PodHandler) StreamPodLogs(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	podName := c.Param("pod")
	containerName := c.Query("container")
	tailLinesStr := c.Query("tailLines")
	followStr := c.DefaultQuery("follow", "true")

	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称不能为空")
		return
	}
	if namespace == "" {
		ResponseError(c, http.StatusBadRequest, "命名空间不能为空")
		return
	}
	if podName == "" {
		ResponseError(c, http.StatusBadRequest, "Pod名称不能为空")
		return
	}

	var tailLines int64 = 100 // 默认获取100行
	if tailLinesStr != "" {
		if l, err := strconv.ParseInt(tailLinesStr, 10, 64); err == nil {
			tailLines = l
		}
	}

	follow := followStr == "true"

	// 设置响应头，指示这是一个SSE流
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*") // 允许跨域
	c.Writer.Header().Set("X-Accel-Buffering", "no")          // 禁用Nginx缓冲，如果使用了Nginx

	// 立即刷新头信息
	c.Writer.Flush()

	// 获取日志流
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour) // 增加超时时间
	defer cancel()

	// 发送初始连接成功消息
	fmt.Fprintf(c.Writer, "data: %s\n\n", "连接已建立，开始接收日志流...")
	c.Writer.Flush()

	logStream, err := h.service.StreamPodLogs(ctx, clusterName, namespace, podName, containerName, tailLines, follow)
	if err != nil {
		errMsg := fmt.Sprintf("获取日志流失败: %s", err.Error())
		// 以SSE格式返回错误，而不是使用标准错误响应格式
		fmt.Fprintf(c.Writer, "data: {\"error\": \"%s\"}\n\n", errMsg)
		c.Writer.Flush()
		return
	}
	defer logStream.Close()

	// 使用bufio逐行读取日志并发送
	scanner := bufio.NewScanner(logStream)

	// 增加扫描器的缓冲区大小，以处理更长的日志行
	const maxScanTokenSize = 1024 * 1024 // 1MB
	scanBuf := make([]byte, maxScanTokenSize)
	scanner.Buffer(scanBuf, maxScanTokenSize)

	// 发送心跳以保持连接
	heartbeat := time.NewTicker(30 * time.Second)
	defer heartbeat.Stop()

	// 设置完成通道，用于通知心跳goroutine结束
	done := make(chan bool)
	defer close(done)

	// 启动心跳goroutine
	go func() {
		for {
			select {
			case <-heartbeat.C:
				// 发送注释行作为心跳
				fmt.Fprintf(c.Writer, ": heartbeat\n\n")
				c.Writer.Flush()
			case <-done:
				return
			}
		}
	}()

	c.Stream(func(w io.Writer) bool {
		if !scanner.Scan() {
			// 发送一个心跳确保客户端知道我们仍然在线
			fmt.Fprintf(w, ": heartbeat\n\n")
			// 等待一小段时间，避免在日志停止时立即返回false
			select {
			case <-time.After(500 * time.Millisecond):
				// 如果没有更多的日志，检查扫描器是否有错误
				if err := scanner.Err(); err != nil {
					fmt.Fprintf(w, "data: {\"error\": \"%s\"}\n\n", err.Error())
					return false
				}
				// 如果follow为true，而且没有错误，继续等待
				if follow {
					return true
				}
			case <-ctx.Done():
				// 上下文已取消
				fmt.Fprintf(w, "data: {\"status\": \"日志流已关闭\"}\n\n")
				return false
			}
			return false
		}

		line := scanner.Text()

		// 发送日志行作为SSE事件
		fmt.Fprintf(w, "data: %s\n\n", line)

		// 刷新缓冲区确保数据被发送
		c.Writer.Flush()

		return true
	})
}

// GetPodsBySelector 根据标签选择器获取Pod列表
func (h *PodHandler) GetPodsBySelector(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称不能为空")
		return
	}
	if namespace == "" {
		ResponseError(c, http.StatusBadRequest, "命名空间不能为空")
		return
	}

	var selector map[string]string
	if err := c.ShouldBindJSON(&selector); err != nil {
		ResponseError(c, http.StatusBadRequest, "无效的标签选择器")
		return
	}

	pods, err := h.service.GetPodsBySelector(context.Background(), clusterName, namespace, selector)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"pods": pods,
	})
}

// CheckPodExists 检查Pod是否存在及其状态
func (h *PodHandler) CheckPodExists(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	podName := c.Param("pod")

	if clusterName == "" {
		ResponseError(c, http.StatusBadRequest, "集群名称不能为空")
		return
	}
	if namespace == "" {
		ResponseError(c, http.StatusBadRequest, "命名空间不能为空")
		return
	}
	if podName == "" {
		ResponseError(c, http.StatusBadRequest, "Pod名称不能为空")
		return
	}

	// 尝试获取Pod详情
	pod, err := h.service.GetPodDetails(context.Background(), clusterName, namespace, podName)
	if err != nil {
		// 检查是否为"未找到"错误
		if k8s.IsNotFoundError(err) {
			ResponseSuccess(c, gin.H{
				"exists":  false,
				"message": "Pod不存在",
			})
			return
		}

		// 其他错误
		ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Pod存在，返回详情和状态
	ResponseSuccess(c, gin.H{
		"exists": true,
		"pod":    pod,
	})
}
