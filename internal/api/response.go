package api

import (
	"net/http"

	"kube-tide/internal/api/middleware"
	"kube-tide/internal/utils/i18n"

	"github.com/gin-gonic/gin"
)

// Response  Unified response structure
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    any        `json:"data,omitempty"`
}

// ResponseSuccess Success response
func ResponseSuccess(c *gin.Context, data any) {
	// Get language from context
	lang := middleware.GetLanguage(c)

	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: i18n.GetInstance().Translate(lang, "common.success"),
		Data:    data,
	})
}

// ResponseError Error response
func ResponseError(c *gin.Context, code int, messageKey string, args ...any) {
	// Get language from context
	lang := middleware.GetLanguage(c)

	c.JSON(code, Response{
		Code:    code,
		Message: i18n.GetInstance().Translate(lang, messageKey, args...),
	})
}

// Fail Returns a failure response
func Fail(c *gin.Context, httpCode int, messageKey string, args ...any) {
	// Get language from context
	lang := middleware.GetLanguage(c)

	c.JSON(httpCode, Response{
		Code:    httpCode,
		Message: i18n.GetInstance().Translate(lang, messageKey, args...),
	})
}

// FailWithError Returns a failure response with error
func FailWithError(c *gin.Context, httpCode int, messageKey string, err error, args ...any) {
	// Get language from context
	lang := middleware.GetLanguage(c)

	message := i18n.GetInstance().Translate(lang, messageKey, args...)
	if err != nil {
		message = message + ": " + err.Error()
	}

	c.JSON(httpCode, Response{
		Code:    httpCode,
		Message: message,
	})
}
