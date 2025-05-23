package rest

import (
	"net/http"
	"thanhnt208/vcs-sms/auth-service/internal/services"
	"thanhnt208/vcs-sms/auth-service/pkg/logger"
	"thanhnt208/vcs-sms/auth-service/utils"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService services.IAuthService
	logger      logger.ILogger
}

func NewAuthHandler(authService services.IAuthService, logger logger.ILogger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Role     string `json:"role" binding:"required,oneof=sudo admin user"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		h.logger.Error("Invalid request", "error", err)
		return
	}

	input := services.RegisterInput(req)
	userID, err := h.authService.Register(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		h.logger.Error("Failed to register user", "error", err)
		return
	}

	h.logger.Info("User registered successfully", "userId", userID)
	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully", "userId": userID})
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		h.logger.Error("Invalid request", "error", err)
		return
	}

	input := services.LoginInput(req)
	accessToken, refreshToken, err := h.authService.Login(input)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		h.logger.Error("Failed to login user", "username", req.Username, "error", err)
		return
	}

	h.logger.Info("User logged in successfully", "username", req.Username)
	c.JSON(http.StatusOK, gin.H{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
		"tokenType":    "Bearer",
		"expiresIn":    3600,
	})
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		h.logger.Error("Invalid request", "error", err)
		return
	}

	claims := c.MustGet("claims").(*utils.Claims)
	accessToken, err := h.authService.RefreshToken(claims.UserID, req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		h.logger.Error("Failed to refresh token", "error", err)
		return
	}

	h.logger.Info("Token refreshed successfully", "userId", claims.UserID)
	c.JSON(http.StatusOK, gin.H{
		"accessToken": accessToken,
		"tokenType":   "Bearer",
		"expiresIn":   3600,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	claims := c.MustGet("claims").(*utils.Claims)
	if err := h.authService.Logout(claims.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		h.logger.Error("Failed to logout user", "error", err)
		return
	}
	h.logger.Info("User logged out successfully", "userId", claims.UserID)
	c.JSON(http.StatusOK, gin.H{"message": "User logged out successfully"})
}
