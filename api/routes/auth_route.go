package routes

import (
	"thanhnt208/vcs-sms/auth-service/api/middlewares"
	"thanhnt208/vcs-sms/auth-service/internal/delivery/rest"

	"github.com/gin-gonic/gin"
)

func SetupAuthRoutes(h *rest.AuthHandler) *gin.Engine {
	router := gin.Default()
	router.POST("/register", h.Register)
	router.POST("/login", h.Login)
	router.POST("/refresh-token", middlewares.JWTAuthMiddleware(), h.RefreshToken)
	router.POST("/logout", middlewares.JWTAuthMiddleware(), h.Logout)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	return router
}
