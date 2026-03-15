package handler

import (
	"net/http"

	"mcp-agent/internal/service"

	"github.com/gin-gonic/gin"
)

type StatsHandler struct {
	statsSvc *service.StatsService
}

func NewStatsHandler(statsSvc *service.StatsService) *StatsHandler {
	return &StatsHandler{statsSvc: statsSvc}
}

func (h *StatsHandler) GetToolStats(c *gin.Context) {
	toolName := c.Param("name")
	stats, err := h.statsSvc.GetToolStats(toolName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": stats})
}

func (h *StatsHandler) ListAllStats(c *gin.Context) {
	stats, err := h.statsSvc.ListAllStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": stats})
}
