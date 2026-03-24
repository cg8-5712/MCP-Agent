package handler

import (
	"net/http"

	"mcp-agent/internal/model"
	"mcp-agent/internal/service"

	"github.com/gin-gonic/gin"
)

type AgentHandler struct {
	agentSvc *service.AgentService
}

func NewAgentHandler(agentSvc *service.AgentService) *AgentHandler {
	return &AgentHandler{agentSvc: agentSvc}
}

func (h *AgentHandler) Execute(c *gin.Context) {
	var req model.AgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	uid, _ := userID.(int64)
	username, _ := c.Get("username")
	uname, _ := username.(string)

	result, err := h.agentSvc.Execute(req.Query, uid, uname)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": result})
}

func (h *AgentHandler) SearchTools(c *gin.Context) {
	var req model.ToolSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	if req.TopK <= 0 {
		req.TopK = 5
	}

	results, err := h.agentSvc.SearchTools(req.Query, req.TopK)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": results})
}

func (h *AgentHandler) IndexTools(c *gin.Context) {
	if err := h.agentSvc.IndexTools(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "tools indexed successfully"})
}
