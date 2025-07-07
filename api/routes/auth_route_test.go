package routes

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"thanhnt208/vcs-sms/auth-service/internal/delivery/rest"
	"thanhnt208/vcs-sms/auth-service/internal/services"
	"thanhnt208/vcs-sms/auth-service/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type MockAuthService struct {
	services.IAuthService
}

type MockLogger struct{}

func (m *MockAuthService) Register(input services.RegisterInput) (uint, error) {
	if input.Username == "fail" {
		return 0, errors.New("registration failed")
	}
	return 1, nil
}

func (m *MockAuthService) Login(input services.LoginInput) (string, string, error) {
	if input.Username == "fail" {
		return "", "", errors.New("login failed")
	}
	return "access-token", "refresh-token", nil
}

func (m *MockAuthService) RefreshToken(userID uint, token string) (string, error) {
	return "new-access-token", nil
}

func (m *MockAuthService) Logout(userID uint) error {
	return nil
}

func (l *MockLogger) Error(msg string, keyvals ...interface{}) {}
func (l *MockLogger) Info(msg string, keyvals ...interface{})  {}
func (l *MockLogger) Debug(msg string, keyvals ...interface{}) {}
func (l *MockLogger) Warn(msg string, keyvals ...interface{})  {}
func (l *MockLogger) Fatal(msg string, keyvals ...interface{}) {}
func (l *MockLogger) Sync() error                              { return nil }

func setupTestRouter() *gin.Engine {
	authService := &MockAuthService{}
	logger := &MockLogger{}
	handler := rest.NewAuthHandler(authService, logger)

	// Mock utils.ParseJWT for middleware testing
	utils.ParseJWT = func(token string) (*utils.Claims, error) {
		if token == "invalid" {
			return nil, errors.New("invalid token")
		}
		return &utils.Claims{UserID: 1}, nil
	}

	return SetupAuthRoutes(handler)
}

func TestHealthRoute(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), "ok")
}

func TestRegisterRoute(t *testing.T) {
	router := setupTestRouter()

	body := `{"username":"test","password":"pass","name":"Test","email":"test@example.com","role":"user"}`
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusCreated, resp.Code)
	assert.Contains(t, resp.Body.String(), "User registered successfully")
}

func TestLoginRoute(t *testing.T) {
	router := setupTestRouter()

	body := `{"username":"test","password":"pass"}`
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), "accessToken")
}

func TestRefreshTokenRoute_WithValidJWT(t *testing.T) {
	router := setupTestRouter()

	body := `{"refreshToken":"some-refresh-token"}`
	req := httptest.NewRequest(http.MethodPost, "/refresh-token", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), "new-access-token")
}

func TestRefreshTokenRoute_WithInvalidJWT(t *testing.T) {
	router := setupTestRouter()

	body := `{"refreshToken":"some-refresh-token"}`
	req := httptest.NewRequest(http.MethodPost, "/refresh-token", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer invalid")
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
	assert.Contains(t, resp.Body.String(), "Invalid or expired token")
}

func TestLogoutRoute(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), "User logged out successfully")
}

func TestLogoutRoute_WithoutJWT(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
	assert.Contains(t, resp.Body.String(), "Missing or invalid token")
}
