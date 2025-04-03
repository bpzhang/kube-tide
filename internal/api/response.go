package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response  Unified response structure
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ResponseSuccess Success response
func ResponseSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// ResponseError Error response
func ResponseError(c *gin.Context, code int, message string) {
	c.JSON(code, Response{
		Code:    code,
		Message: message,
	})
}

// Fail Returns a failure response
func Fail(c *gin.Context, httpCode int, message string) {
	c.JSON(httpCode, Response{
		Code:    httpCode,
		Message: message,
	})
}

// FailWithError Returns a failure response with error
func FailWithError(c *gin.Context, httpCode int, message string, err error) {
	if err != nil {
		message = message + ": " + err.Error()
	}
	Fail(c, httpCode, message)
}
