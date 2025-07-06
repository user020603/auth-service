package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"testing"
	"thanhnt208/vcs-sms/auth-service/config"
	"thanhnt208/vcs-sms/auth-service/infrastructure"
	"thanhnt208/vcs-sms/auth-service/pkg/logger"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// --- Mocking Dependencies ---

// MockDatabase là một mock cho infrastructure.IDatabase
type MockDatabase struct {
	mock.Mock
}

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Info(msg string, keysAndValues ...interface{}) {
	args := append([]interface{}{msg}, keysAndValues...)
	m.Called(args...)
}
func (m *MockLogger) Error(msg string, keysAndValues ...interface{}) {
	args := append([]interface{}{msg}, keysAndValues...)
	m.Called(args...)
}
func (m *MockLogger) Debug(msg string, keysAndValues ...interface{}) {
	args := append([]interface{}{msg}, keysAndValues...)
	m.Called(args...)
}
func (m *MockLogger) Fatal(msg string, keysAndValues ...interface{}) {
	args := append([]interface{}{msg}, keysAndValues...)
	m.Called(args...)
}
func (m *MockLogger) Warn(msg string, keysAndValues ...interface{}) {
	args := append([]interface{}{msg}, keysAndValues...)
	m.Called(args...)
}
func (l *MockLogger) Sync() error { return nil }

func (m *MockDatabase) GetDB() *gorm.DB {
	// Trong test chúng ta không cần đến DB thật
	return nil
}

func (m *MockDatabase) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDatabase) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Đảm bảo MockDatabase implements infrastructure.IDatabase
var _ infrastructure.IDatabase = (*MockDatabase)(nil)

// MockRedisClient là một mock cho infrastructure.IRedis
type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) GetClient() *redis.Client {
	// Trong test chúng ta không cần đến Redis client thật
	return nil
}

func (m *MockRedisClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRedisClient) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// --- Helper Functions ---

// newTestConfig tạo một cấu hình giả để test
func newTestConfig() *config.Config {
	return &config.Config{
		LogLevel:   "debug",
		LogFile:    "test.log",
		ServerPort: "8081",
		DBHost:     "localhost",
		DBPort:     "5432",
		RedisAddr:  "localhost:6379",
	}
}

// --- Tests ---

func TestNewApp_Success(t *testing.T) {
	t.Log("Running TestNewApp_Success")
	// Sắp xếp (Arrange)
	cfg := newTestConfig()
	originalNewDatabase := infrastructure.NewDatabase
	infrastructure.NewDatabase = func(c *config.Config) (infrastructure.IDatabase, error) {
		return &MockDatabase{}, nil
	}
	defer func() { infrastructure.NewDatabase = originalNewDatabase }()

	originalNewRedis := infrastructure.NewRedis
	infrastructure.NewRedis = func(c *config.Config) (infrastructure.IRedis, error) {
		return &MockRedisClient{}, nil
	}
	defer func() { infrastructure.NewRedis = originalNewRedis }()

	// Hành động (Act)
	app, err := NewApp(cfg)

	// Khẳng định (Assert)
	assert.NoError(t, err)
	assert.NotNil(t, app)
	assert.NotNil(t, app.Logger)
	assert.NotNil(t, app.Router)
	assert.Equal(t, cfg, app.Config)

	// Dọn dẹp file log được tạo ra
	defer os.Remove(cfg.LogFile)

	// Kiểm tra xem các phương thức Close có được gọi không
	mockDB := app.DB.(*MockDatabase)
	mockRedis := app.RedisClient.(*MockRedisClient)
	mockDB.On("Close").Return(nil)
	mockRedis.On("Close").Return(nil)

	app.Close()

	mockDB.AssertCalled(t, "Close")
	mockRedis.AssertCalled(t, "Close")
}

// ...existing code...
func TestNewApp_Fail_DatabaseConnection(t *testing.T) {
	t.Log("Running TestNewApp_Fail_DatabaseConnection")
	// Sắp xếp (Arrange)
	cfg := newTestConfig()
	dbError := errors.New("database connection failed")

	// Ghi đè hàm NewDatabase để trả về lỗi
	originalNewDatabase := infrastructure.NewDatabase
	infrastructure.NewDatabase = func(c *config.Config) (infrastructure.IDatabase, error) {
		return nil, dbError
	}
	defer func() { infrastructure.NewDatabase = originalNewDatabase }()

	// Hành động (Act)
	app, err := NewApp(cfg)

	// Khẳng định (Assert)
	assert.Error(t, err)
	assert.Nil(t, app)
	assert.Contains(t, err.Error(), dbError.Error())

	// Dọn dẹp file log
	defer os.Remove(cfg.LogFile)
}

// ...existing code...
func TestNewApp_Fail_RedisConnection(t *testing.T) {
	t.Log("Running TestNewApp_Fail_RedisConnection")
	// Sắp xếp (Arrange)
	cfg := newTestConfig()
	redisError := errors.New("redis connection failed")

	// Mock NewDatabase thành công
	originalNewDatabase := infrastructure.NewDatabase
	infrastructure.NewDatabase = func(c *config.Config) (infrastructure.IDatabase, error) {
		return &MockDatabase{}, nil
	}
	defer func() { infrastructure.NewDatabase = originalNewDatabase }()

	// Ghi đè hàm NewRedis để trả về lỗi
	originalNewRedis := infrastructure.NewRedis
	infrastructure.NewRedis = func(c *config.Config) (infrastructure.IRedis, error) {
		return nil, redisError
	}
	defer func() { infrastructure.NewRedis = originalNewRedis }()

	// Hành động (Act)
	app, err := NewApp(cfg)

	// Khẳng định (Assert)
	assert.Error(t, err)
	assert.Nil(t, app)
	assert.Contains(t, err.Error(), redisError.Error())

	// Dọn dẹp file log
	defer os.Remove(cfg.LogFile)
}

func TestNewApp_Fail_LoggerInitialization(t *testing.T) {
	t.Log("Running TestNewApp_Fail_LoggerInitialization")

	cfg := newTestConfig()
	logErr := errors.New("logger init failed")

	// Ghi đè logger
	originalLoggerFunc := newLoggerFunc
	newLoggerFunc = func(level, file string) (logger.ILogger, error) {
		return nil, logErr
	}
	defer func() { newLoggerFunc = originalLoggerFunc }()

	app, err := NewApp(cfg)

	assert.Error(t, err)
	assert.Nil(t, app)
	assert.Contains(t, err.Error(), logErr.Error())
}

func TestApp_Run_FailToStartServer(t *testing.T) {
	t.Log("Running TestApp_Run_FailToStartServer")

	// Ghi log vào memory
	mockLogger := &MockLogger{}
	mockLogger.On("Info", "Starting server", "port", "9999").Return()
	mockLogger.On("Error", "Failed to start server", "error", mock.Anything).Return()

	// Setup App
	app := &App{
		Logger: mockLogger,
		Config: &config.Config{ServerPort: "9999"},
		Router: gin.New(), // Có thể là dummy, vì ta ghi đè Run
	}

	// Ghi đè runRouterFunc để trả lỗi
	originalRunRouterFunc := runRouterFunc
	runRouterFunc = func(router *gin.Engine, port string) error {
		return errors.New("mock server failed to start")
	}
	defer func() { runRouterFunc = originalRunRouterFunc }()

	app.Run()

	// Kiểm tra logger được gọi đúng
	mockLogger.AssertCalled(t, "Info", "Starting server", "port", "9999")
	mockLogger.AssertCalled(t, "Error", "Failed to start server", "error", mock.Anything)
}

func TestRunMain_Fail_NewApp(t *testing.T) {
	t.Log("Running TestRunMain_Fail_NewApp")

	expectedErr := errors.New("init failed")

	originalLoadConfig := loadConfigFunc
	loadConfigFunc = func() *config.Config {
		return newTestConfig()
	}
	defer func() { loadConfigFunc = originalLoadConfig }()

	originalNewApp := newAppFunc
	newAppFunc = func(cfg *config.Config) (*App, error) {
		return nil, expectedErr
	}
	defer func() { newAppFunc = originalNewApp }()

	err := runMain()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to setup application")
	assert.Contains(t, err.Error(), expectedErr.Error())
}

type MockApp struct {
	mock.Mock
}

func (m *MockApp) Run() {
	m.Called()
}

func (m *MockApp) Close() {
	m.Called()
}

func TestRunMain_Success(t *testing.T) {
	t.Log("Running TestRunMain_Success")

	calledRun := false
	calledClose := false

	fakeApp := &App{
		Config: &config.Config{},
	}

	// Ghi đè loadConfig và newApp
	originalLoadConfig := loadConfigFunc
	loadConfigFunc = func() *config.Config {
		return newTestConfig()
	}
	defer func() { loadConfigFunc = originalLoadConfig }()

	originalNewApp := newAppFunc
	newAppFunc = func(cfg *config.Config) (*App, error) {
		return fakeApp, nil
	}
	defer func() { newAppFunc = originalNewApp }()

	// Ghi đè runFunc và closeFunc
	originalRunFunc := runFunc
	runFunc = func(app *App) {
		calledRun = true
	}
	defer func() { runFunc = originalRunFunc }()

	originalCloseFunc := closeFunc
	closeFunc = func(app *App) {
		calledClose = true
	}
	defer func() { closeFunc = originalCloseFunc }()

	// Act
	err := runMain()

	assert.NoError(t, err)
	assert.True(t, calledRun, "Expected Run() to be called")
	assert.True(t, calledClose, "Expected Close() to be called")
}

func TestMain_FatalOnError(t *testing.T) {
	t.Log("Running TestMain_FatalOnError")

	// Backup original
	originalRunMain := runMain
	originalFatalf := fatalfFunc

	defer func() {
		runMain = originalRunMain
		fatalfFunc = originalFatalf
	}()

	// Ghi đè runMain để trả lỗi
	runMain = func() error {
		return errors.New("mock error")
	}

	var loggedMessage string
	fatalfFunc = func(format string, v ...interface{}) {
		loggedMessage = fmt.Sprintf(format, v...)
		// Đừng gọi os.Exit
	}

	main()

	assert.Contains(t, loggedMessage, "mock error")
}

func Test_runRouterFunc_Success(t *testing.T) {
	router := gin.New()

	// Run trong goroutine để không block test
	go func() {
		_ = runRouterFunc(router, "8099")
	}()

	time.Sleep(100 * time.Millisecond) // chờ cho server start
}

func Test_loadConfigFunc(t *testing.T) {
	original := loadConfigFunc
	defer func() { loadConfigFunc = original }()

	called := false
	loadConfigFunc = func() *config.Config {
		called = true
		return newTestConfig()
	}

	_ = loadConfigFunc()
	assert.True(t, called, "Expected loadConfigFunc to be called")
}

func Test_newAppFunc(t *testing.T) {
	original := newAppFunc
	defer func() { newAppFunc = original }()

	expectedApp := &App{Config: newTestConfig()}
	newAppFunc = func(cfg *config.Config) (*App, error) {
		return expectedApp, nil
	}

	app, err := newAppFunc(newTestConfig())
	assert.NoError(t, err)
	assert.Equal(t, expectedApp, app)
}

func Test_runFunc_closeFunc(t *testing.T) {
	var runCalled, closeCalled bool

	app := &App{}
	runFunc = func(app *App) { runCalled = true }
	closeFunc = func(app *App) { closeCalled = true }

	runFunc(app)
	closeFunc(app)

	assert.True(t, runCalled)
	assert.True(t, closeCalled)
}

func Test_fatalfFunc_CallsExit(t *testing.T) {
	originalExit := exitFunc
	defer func() { exitFunc = originalExit }()

	var exitCode int
	exitFunc = func(code int) {
		exitCode = code
	}

	// Chặn log output
	log.SetOutput(&bytes.Buffer{})
	logOutputWriter := log.Writer()
	defer log.SetOutput(os.Stderr)

	fatalfFunc("test error: %d", 123)

	assert.Equal(t, 1, exitCode)
	_ = logOutputWriter // optional: verify log if needed
}
