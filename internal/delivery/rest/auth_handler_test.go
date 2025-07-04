package rest

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"thanhnt208/vcs-sms/auth-service/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAuthService struct {
	mock.Mock
}

type MockLogger struct{}

func (m *MockAuthService) Register(input services.RegisterInput) (uint, error) {
	args := m.Called(input)
	return args.Get(0).(uint), args.Error(1)
}

func (m *MockAuthService) Login(input services.LoginInput) (string, string, error) {
	args := m.Called(input)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockAuthService) RefreshToken(userID uint, refreshToken string) (string, error) {
	args := m.Called(userID, refreshToken)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) Logout(userID uint) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (l *MockLogger) Error(msg string, keyvals ...interface{}) {}
func (l *MockLogger) Info(msg string, keyvals ...interface{})  {}
func (l *MockLogger) Debug(msg string, keyvals ...interface{}) {}
func (l *MockLogger) Warn(msg string, keyvals ...interface{})  {}
func (l *MockLogger) Fatal(msg string, keyvals ...interface{}) {}
func (l *MockLogger) Sync() error                              { return nil }

func setupRegisterRouter(handler *AuthHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/register", handler.Register)
	return r
}

func setupLoginRouter(handler *AuthHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/login", handler.Register)
	return r
}

// TEST CASE
func TestAuthHandler_Register_Success(t *testing.T) {
	mockService := new(MockAuthService)
	mockLogger := &MockLogger{}
	handler := NewAuthHandler(mockService, mockLogger)
	router := setupRegisterRouter(handler)

	input := services.RegisterInput{
		Username: "testuser",
		Password: "secret",
		Name:     "Test User",
		Email:    "test@example.com",
		Role:     "user",
	}
	mockService.On("Register", input).Return(uint(42), nil)

	reqBody := `{
		"username": "testuser",
		"password": "secret",
		"name": "Test User",
		"email": "test@example.com",
		"role": "user"
	}`
	req := httptest.NewRequest(
		http.MethodPost,
		"/register",
		bytes.NewBufferString(reqBody),
	)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), `"userId":42`)
}

func TestAuthHandler_Register_InvalidJSON(t *testing.T) {
	mockService := new(MockAuthService)
	mockLogger := &MockLogger{}
	handler := NewAuthHandler(mockService, mockLogger)
	router := setupRegisterRouter(handler)

	reqBody := `{
		"username": "testuser"
	}`
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `"error"`)
	mockService.AssertNotCalled(t, "Register")
}

func TestAuthHandler_Register_ServiceError(t *testing.T) {
	mockService := new(MockAuthService)
	mockLogger := &MockLogger{}
	handler := NewAuthHandler(mockService, mockLogger)
	router := setupRegisterRouter(handler)

	input := services.RegisterInput{
		Username: "testuser",
		Password: "secret",
		Name:     "Test User",
		Email:    "test@example.com",
		Role:     "user",
	}
	mockService.On("Register", input).Return(uint(0), errors.New("duplicate email"))

	reqBody := `{
		"username": "testuser",
		"password": "secret",
		"name": "Test User",
		"email": "test@example.com",
		"role": "user"
	}`
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `"error":"duplicate email"`)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	mockService := new(MockAuthService)
	mockLogger := &MockLogger{}
	handler := NewAuthHandler(mockService, mockLogger)
	router := setupLoginRouter(handler)

	input := services.LoginInput{
		Username: "testuser",
		Password: "secret",
	}
	mockService.On("Login", input).Return("access-token-123", "refresh-token-123", nil)

	reqBody := `
		"username": "testuser",
		"password": "secret"
	`
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"accessToken":"access-token-123"`)
	assert.Contains(t, w.Body.String(), `"refreshToken":"refresh-token-123"`)
}

func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
	mockService := new(MockAuthService)
	mockLogger := &MockLogger{}
	handler := NewAuthHandler(mockService, mockLogger)
	router := setupLoginRouter(handler)

	reqBody := `
		"username": "testuser"
	`
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `"error"`)
	mockService.AssertNotCalled(t, "Login")
}
