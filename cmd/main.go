package main

import (
	"thanhnt208/vcs-sms/auth-service/api/routes"
	"thanhnt208/vcs-sms/auth-service/configs"
	"thanhnt208/vcs-sms/auth-service/infrastructure"
	"thanhnt208/vcs-sms/auth-service/internal/delivery/rest"
	"thanhnt208/vcs-sms/auth-service/internal/repositories"
	"thanhnt208/vcs-sms/auth-service/internal/services"
	"thanhnt208/vcs-sms/auth-service/pkg/logger"
)

func main() {
	cfg := configs.LoadConfig()

	logger, err := logger.NewLogger(cfg.LogLevel, cfg.LogFile)
	if err != nil {
		logger.Fatal("Failed to initialize logger", "error", err)
	}

	db, err := infrastructure.NewDatabase(cfg)
	if err != nil {
		logger.Fatal("Failed to connect to database", "error", err)
	}
	defer db.Close()
	logger.Info("Connected to database", "host", cfg.DBHost, "port", cfg.DBPort)

	redis, err := infrastructure.NewRedis(cfg)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", "error", err)
	}
	defer redis.Close()
	logger.Info("Connected to Redis", "address", cfg.RedisAddr)

	userRepo := repositories.NewUserRepository(db.GetDB())
	tokenRepo := repositories.NewTokenRepository(redis.GetClient())
	authService := services.NewAuthService(userRepo, tokenRepo, logger)
	authHandler := rest.NewAuthHandler(authService, logger)

	r := routes.SetupAuthRoutes(authHandler)

	port := cfg.ServerPort
	logger.Info("Starting server", "port", port)
	if err := r.Run(":" + port); err != nil {
		logger.Fatal("Failed to start server", "error", err)
	}
}
