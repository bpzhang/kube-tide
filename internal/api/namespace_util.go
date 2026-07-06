package api

import "github.com/gin-gonic/gin"

// namespaceFromRequest resolves namespace from path param, then query param.
// Returns empty string for cluster-wide list routes (all namespaces).
func namespaceFromRequest(c *gin.Context) string {
	if ns := c.Param("namespace"); ns != "" {
		return ns
	}
	return c.Query("namespace")
}
