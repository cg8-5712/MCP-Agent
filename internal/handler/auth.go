package handler

import (
	"errors"
	"net/http"

	"mcp-agent/internal/model"
	"mcp-agent/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authSvc *service.AuthService
}

func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	resp, err := h.authSvc.Login(req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "invalid username or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": resp})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, _ := userID.(int64)

	resp, err := h.authSvc.RefreshToken(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": resp})
}

func (h *AuthHandler) Profile(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, _ := userID.(int64)

	user, err := h.authSvc.GetProfile(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": user})
}

func (h *AuthHandler) CreateUser(c *gin.Context) {
	var req model.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	if err := h.authSvc.CreateUser(req); err != nil {
		if errors.Is(err, service.ErrUserExists) {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "user already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": 201, "message": "user created"})
}
