package routes

import (
	"thanhnt208/vcs-sms/auth-service/api/middlewares"
	"thanhnt208/vcs-sms/auth-service/internal/delivery/rest"

	"github.com/gin-gonic/gin"
)

func SetupAuthRoutes(h *rest.AuthHandler) *gin.Engine {
	router := gin.Default()
	router.POST("/auth/register", h.Register)
	router.POST("/auth/login", h.Login)
	router.POST("/auth/refresh-token", middlewares.JWTAuthMiddleware(), h.RefreshToken)
	router.POST("/auth/logout", middlewares.JWTAuthMiddleware(), h.Logout)
	return router
}
