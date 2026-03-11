package middleware

import (
	"net/http"

	"mcp-agent/internal/permission"

	"github.com/gin-gonic/gin"
)

func PermissionCheck(pm *permission.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "no role in context"})
			c.Abort()
			return
		}

		toolName := c.Param("name")
		if toolName == "" {
			c.Next()
			return
		}

		roleStr, _ := role.(string)
		if !pm.CanAccess(roleStr, toolName) {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "permission denied for this tool"})
			c.Abort()
			return
		}

		c.Next()
	}
}
