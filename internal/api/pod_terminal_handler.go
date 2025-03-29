package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"kube-tide/internal/core/k8s"
	"kube-tide/internal/utils/logger"

	"github.com/gin-gonic/gin"
	"k8s.io/client-go/tools/remotecommand"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

var upgradeOptions = websocket.AcceptOptions{
	InsecureSkipVerify: true, // 在生产环境中应该适当设置
	// nhooyr.io/websocket 没有直接的 HandshakeTimeout 选项
}

// TerminalMessage 定义终端消息格式
type TerminalMessage struct {
	Type    string      `json:"type"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// TerminalSession 包装了 websocket 连接和 终端大小信息
type TerminalSession struct {
	wsConn   *websocket.Conn
	sizeChan chan remotecommand.TerminalSize
	doneChan chan struct{}
	ctx      context.Context
}

// TerminalSize 实现 remotecommand.TerminalSizeQueue 接口的 Next 方法
// Next 返回终端的大小
func (t *TerminalSession) Next() *remotecommand.TerminalSize {
	select {
	case size := <-t.sizeChan:
		return &size
	case <-t.doneChan:
		return nil
	}
}

// Read 从 websocket 连接读取数据，实现 io.Reader 接口
func (t *TerminalSession) Read(p []byte) (int, error) {
	messageType, message, err := t.wsConn.Read(t.ctx)
	if err != nil {
		return 0, err
	}

	// 处理调整终端大小的消息
	if messageType == websocket.MessageText && len(message) > 1 && message[0] == '{' {
		var msg TerminalMessage
		if err := json.Unmarshal(message, &msg); err == nil {
			if msg.Type == "resize" {
				if data, ok := msg.Data.(map[string]interface{}); ok {
					cols, _ := data["cols"].(float64)
					rows, _ := data["rows"].(float64)

					t.sizeChan <- remotecommand.TerminalSize{
						Width:  uint16(cols),
						Height: uint16(rows),
					}
					return 0, nil
				}
			} else if msg.Type == "ping" {
				// 处理心跳ping消息，不需要进一步操作
				return 0, nil
			}
		}
	}

	copy(p, message)
	return len(message), nil
}

// Write 向 websocket 连接写入数据，实现 io.Writer 接口
func (t *TerminalSession) Write(p []byte) (int, error) {
	err := t.wsConn.Write(t.ctx, websocket.MessageBinary, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

// Close 关闭会话
func (t *TerminalSession) Close() error {
	close(t.doneChan)
	return t.wsConn.Close(websocket.StatusNormalClosure, "终端会话已关闭")
}

// PodTerminalHandler Pod终端处理器
type PodTerminalHandler struct {
	service *k8s.PodService
}

// NewPodTerminalHandler 创建Pod终端处理器
func NewPodTerminalHandler(service *k8s.PodService) *PodTerminalHandler {
	return &PodTerminalHandler{
		service: service,
	}
}

// 发送错误消息到WebSocket客户端
func sendErrorMessage(conn *websocket.Conn, ctx context.Context, errorType string, message string) {
	msg := TerminalMessage{
		Type:    "error",
		Data:    errorType,
		Message: message,
	}

	err := wsjson.Write(ctx, conn, msg)
	if err != nil {
		logger.Errorf("发送错误消息失败: %v", err)
	}
}

// HandleTerminal 处理终端请求
func (h *PodTerminalHandler) HandleTerminal(c *gin.Context) {
	// 获取参数
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	podName := c.Param("pod")
	containerName := c.Query("container")

	// 验证参数
	if clusterName == "" || namespace == "" || podName == "" || containerName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": "缺少必要参数",
		})
		return
	}

	// 升级HTTP连接为WebSocket
	wsConn, err := websocket.Accept(c.Writer, c.Request, &upgradeOptions)
	if err != nil {
		logger.Errorf("WebSocket升级失败: %v", err)
		return
	}

	// 设置上下文和关闭处理
	ctx, cancel := context.WithTimeout(c.Request.Context(), 1*time.Hour)
	defer cancel()
	defer wsConn.Close(websocket.StatusInternalError, "连接关闭")

	// 创建终端会话
	terminal := &TerminalSession{
		wsConn:   wsConn,
		sizeChan: make(chan remotecommand.TerminalSize),
		doneChan: make(chan struct{}),
		ctx:      ctx,
	}

	// 发送连接成功消息
	err = wsConn.Write(ctx, websocket.MessageText, []byte("WebSocket连接成功，正在连接到容器终端..."))
	if err != nil {
		logger.Errorf("发送测试消息失败: %v", err)
		return
	}

	// 启动执行终端
	if err := h.service.ExecToPod(clusterName, namespace, podName, containerName, terminal); err != nil {
		logger.Errorf("连接到Pod终端失败: %v", err)
		sendErrorMessage(wsConn, ctx, "exec_failed", "无法连接到Pod终端: "+err.Error())
		return
	}

	// 正常关闭
	wsConn.Close(websocket.StatusNormalClosure, "会话已结束")
}
