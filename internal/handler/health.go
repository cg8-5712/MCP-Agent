package handler

import (
	"net/http"

	"mcp-agent/internal/service"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	healthSvc *service.HealthService
}

func NewHealthHandler(healthSvc *service.HealthService) *HealthHandler {
	return &HealthHandler{healthSvc: healthSvc}
}

func (h *HealthHandler) CheckAll(c *gin.Context) {
	results := h.healthSvc.CheckAll()
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": results})
}

func (h *HealthHandler) CheckTool(c *gin.Context) {
	name := c.Param("name")
	result, err := h.healthSvc.CheckTool(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "tool not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": result})
}
