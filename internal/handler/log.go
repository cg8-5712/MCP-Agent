package handler

import (
	"net/http"

	"mcp-agent/internal/model"
	"mcp-agent/internal/service"

	"github.com/gin-gonic/gin"
)

type LogHandler struct {
	logSvc *service.LogService
}

func NewLogHandler(logSvc *service.LogService) *LogHandler {
	return &LogHandler{logSvc: logSvc}
}

func (h *LogHandler) Query(c *gin.Context) {
	var req model.LogQueryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	result, err := h.logSvc.Query(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": result})
}
