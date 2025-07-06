package main

import (
	"fmt"
	"log"
	"os"
	"thanhnt208/vcs-sms/auth-service/api/routes"
	"thanhnt208/vcs-sms/auth-service/config"
	"thanhnt208/vcs-sms/auth-service/infrastructure"
	"thanhnt208/vcs-sms/auth-service/internal/delivery/rest"
	"thanhnt208/vcs-sms/auth-service/internal/repositories"
	"thanhnt208/vcs-sms/auth-service/internal/services"
	"thanhnt208/vcs-sms/auth-service/pkg/logger"

	"github.com/gin-gonic/gin"
)

// App chứa các phụ thuộc của ứng dụng
type App struct {
	Logger      logger.ILogger
	Config      *config.Config
	Router      *gin.Engine
	DB          infrastructure.IDatabase
	RedisClient infrastructure.IRedis
}

var newLoggerFunc = logger.NewLogger

var runRouterFunc = func(router *gin.Engine, port string) error {
	return router.Run(":" + port)
}

var loadConfigFunc = config.LoadConfig
var newAppFunc = NewApp

var (
	runFunc   = func(app *App) { app.Run() }
	closeFunc = func(app *App) { app.Close() }
)

var exitFunc = func(code int) {
	os.Exit(code)
}

var fatalfFunc = func(format string, v ...interface{}) {
	log.Printf(format, v...)
	exitFunc(1)
}

// NewApp khởi tạo và trả về một instance mới của App
func NewApp(cfg *config.Config) (*App, error) {
	// Khởi tạo Logger
	appLogger, err := newLoggerFunc(cfg.LogLevel, cfg.LogFile)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Kết nối Database
	db, err := infrastructure.NewDatabase(cfg)
	if err != nil {
		appLogger.Error("Failed to connect to database", "error", err)
		return nil, err
	}
	appLogger.Info("Connected to database", "host", cfg.DBHost, "port", cfg.DBPort)

	// Kết nối Redis
	redisClient, err := infrastructure.NewRedis(cfg)
	if err != nil {
		appLogger.Error("Failed to connect to Redis", "error", err)
		return nil, err
	}
	appLogger.Info("Connected to Redis", "address", cfg.RedisAddr)

	// Khởi tạo các Repositories
	userRepo := repositories.NewUserRepository(db.GetDB())
	tokenRepo := repositories.NewTokenRepository(redisClient.GetClient())

	// Khởi tạo Service và Handler
	authService := services.NewAuthService(userRepo, tokenRepo, appLogger)
	authHandler := rest.NewAuthHandler(authService, appLogger)

	// Thiết lập Routes
	router := routes.SetupAuthRoutes(authHandler)

	return &App{
		Logger:      appLogger,
		Config:      cfg,
		Router:      router,
		DB:          db,
		RedisClient: redisClient,
	}, nil
}

// Run khởi động server
func (a *App) Run() {
	port := a.Config.ServerPort
	a.Logger.Info("Starting server", "port", port)

	if err := runRouterFunc(a.Router, port); err != nil {
		a.Logger.Error("Failed to start server", "error", err)
	}
}

// Close giải phóng tài nguyên
func (a *App) Close() {
	if a.DB != nil {
		a.DB.Close()
	}
	if a.RedisClient != nil {
		a.RedisClient.Close()
	}
}

// runMain tách logic khỏi main để dễ test
var runMain = func() error {
	cfg := loadConfigFunc()

	app, err := newAppFunc(cfg)
	if err != nil {
		return fmt.Errorf("Failed to setup application: %w", err)
	}
	defer closeFunc(app)

	runFunc(app)
	return nil
}

func main() {
	if err := runMain(); err != nil {
		fatalfFunc("%v", err)
	}
}
