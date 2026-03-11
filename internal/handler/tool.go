package handler

import (
	"errors"
	"net/http"

	"mcp-agent/internal/model"
	"mcp-agent/internal/service"

	"github.com/gin-gonic/gin"
)

type ToolHandler struct {
	toolSvc *service.ToolService
}

func NewToolHandler(toolSvc *service.ToolService) *ToolHandler {
	return &ToolHandler{toolSvc: toolSvc}
}

func (h *ToolHandler) List(c *gin.Context) {
	tools, err := h.toolSvc.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": tools})
}

func (h *ToolHandler) Create(c *gin.Context) {
	var req model.CreateToolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	tool, err := h.toolSvc.Create(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "internal server error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": 201, "message": "success", "data": tool})
}

func (h *ToolHandler) Update(c *gin.Context) {
	name := c.Param("name")
	var req model.UpdateToolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	tool, err := h.toolSvc.Update(name, req)
	if err != nil {
		if errors.Is(err, service.ErrToolNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "tool not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": tool})
}

func (h *ToolHandler) Delete(c *gin.Context) {
	name := c.Param("name")
	if err := h.toolSvc.Delete(name); err != nil {
		if errors.Is(err, service.ErrToolNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "tool not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "deleted"})
}

func (h *ToolHandler) Call(c *gin.Context) {
	name := c.Param("name")
	var req model.CallToolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	uid, _ := userID.(int64)
	username, _ := c.Get("username")
	uname, _ := username.(string)

	result, err := h.toolSvc.CallTool(name, req.Arguments, uid, uname)
	if err != nil {
		if errors.Is(err, service.ErrToolNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "tool not found"})
			return
		}
		if errors.Is(err, service.ErrToolDisabled) {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "tool is disabled"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": result})
}
