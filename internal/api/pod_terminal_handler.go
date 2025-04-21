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
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

var upgradeOptions = websocket.AcceptOptions{
	InsecureSkipVerify: true, // It should be set appropriately in production environment
	// coder/websocket There is no direct HandshakeTimeout option
}

// TerminalMessage Define terminal message format
type TerminalMessage struct {
	Type    string      `json:"type"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// TerminalSession 
type TerminalSession struct {
	wsConn   *websocket.Conn
	sizeChan chan remotecommand.TerminalSize
	doneChan chan struct{}
	ctx      context.Context
}

// TerminalSize implementation of remotecommand.TerminalSize
// Next returns the next terminal size
func (t *TerminalSession) Next() *remotecommand.TerminalSize {
	select {
	case size := <-t.sizeChan:
		return &size
	case <-t.doneChan:
		return nil
	}
}

// Read data from websocket connection
// Read implements io.Reader interface
func (t *TerminalSession) Read(p []byte) (int, error) {
	messageType, message, err := t.wsConn.Read(t.ctx)
	if err != nil {
		logger.Errorf("Failed to read from websocket: %s", err.Error())
		return 0, err
	}

	// process ping message 
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
				// process heartbeat ping message, no further action needed
				return 0, nil
			}
		}
	}

	copy(p, message)
	return len(message), nil
}

// Write data to websocket connection
// Write implements io.Writer interface
// It sends binary data to the websocket connection
// The data is expected to be in the format of a byte slice
// The function returns the number of bytes written and any error encountered
// If the write operation fails, it returns an error
// If the write operation is successful, it returns the length of the byte slice
func (t *TerminalSession) Write(p []byte) (int, error) {
	err := t.wsConn.Write(t.ctx, websocket.MessageBinary, p)
	if err != nil {
		logger.Errorf("Failed to write to websocket: %s", err.Error())
		return 0, err
	}
	return len(p), nil
}

// Close closes the websocket connection
func (t *TerminalSession) Close() error {
	close(t.doneChan)
	return t.wsConn.Close(websocket.StatusNormalClosure, "Terminal session closed")
}

// PodTerminalHandler Pod terminal handler
type PodTerminalHandler struct {
	service *k8s.PodService
}

// NewPodTerminalHandler create a new PodTerminalHandler
func NewPodTerminalHandler(service *k8s.PodService) *PodTerminalHandler {
	return &PodTerminalHandler{
		service: service,
	}
}

// Send error message to WebSocket client
func sendErrorMessage(conn *websocket.Conn, ctx context.Context, errorType string, message string) {
	msg := TerminalMessage{
		Type:    "error",
		Data:    errorType,
		Message: message,
	}

	err := wsjson.Write(ctx, conn, msg)
	if err != nil {
		logger.Errorf("Failed to send error message: %v", err)
	}
}

// HandleTerminal handles the terminal connection
func (h *PodTerminalHandler) HandleTerminal(c *gin.Context) {
	clusterName := c.Param("cluster")
	namespace := c.Param("namespace")
	podName := c.Param("pod")
	containerName := c.Query("container")

	// Validate parameters
	if clusterName == "" || namespace == "" || podName == "" || containerName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": "Missing required parameters",
		})
		return
	}

	// Upgrade HTTP connection to WebSocket
	wsConn, err := websocket.Accept(c.Writer, c.Request, &upgradeOptions)
	if err != nil {
		logger.Errorf("WebSocket upgrade failed: %v", err)
		return
	}

	// Set context and close handling
	ctx, cancel := context.WithTimeout(c.Request.Context(), 1*time.Hour)
	defer cancel()
	defer wsConn.Close(websocket.StatusInternalError, "Connection closed")

	terminal := &TerminalSession{
		wsConn:   wsConn,
		sizeChan: make(chan remotecommand.TerminalSize),
		doneChan: make(chan struct{}),
		ctx:      ctx,
	}

	// Send connection success message
	err = wsConn.Write(ctx, websocket.MessageText, []byte("WebSocket connection successful, connecting to container terminal..."))
	if err != nil {
		logger.Errorf("Failed to send test message: %v", err)
		return
	}

	// Start executing terminal
	if err := h.service.ExecToPod(clusterName, namespace, podName, containerName, terminal); err != nil {
		logger.Errorf("Failed to connect to Pod terminal: %v", err)
		sendErrorMessage(wsConn, ctx, "exec_failed", "Can not connect to Pod terminal: "+err.Error())
		return
	}

	// Normal closure
	wsConn.Close(websocket.StatusNormalClosure, "Session ended")
}
