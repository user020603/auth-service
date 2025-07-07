package middlewares

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"thanhnt208/vcs-sms/auth-service/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var (
	originalParseJWT = utils.ParseJWT
)

func setupRouterWithJWT() *gin.Engine {
	r := gin.New()
	r.Use(JWTAuthMiddleware())
	r.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	return r
}

func TestJWTAuthMiddleware_MissingAuthHeader(t *testing.T) {
	router := setupRouterWithJWT()

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
	assert.Contains(t, resp.Body.String(), "Missing or invalid token")
}

func TestJWTAuthMiddleware_InvalidAuthHeader(t *testing.T) {
	router := setupRouterWithJWT()

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Token abc123")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
	assert.Contains(t, resp.Body.String(), "Missing or invalid token")
}

func TestJWTAuthMiddleware_InvalidToken(t *testing.T) {
	utils.ParseJWT = func(token string) (*utils.Claims, error) {
		return nil, errors.New("invalid token")
	}
	defer func() { utils.ParseJWT = originalParseJWT }()

	router := setupRouterWithJWT()

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalidtoken")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
	assert.Contains(t, resp.Body.String(), "Invalid or expired token")
}

func TestJWTAuthMiddleware_ValidToken(t *testing.T) {
	mockClaims := &utils.Claims{UserID: 1}
	utils.ParseJWT = func(token string) (*utils.Claims, error) {
		return mockClaims, nil
	}
	defer func() { utils.ParseJWT = originalParseJWT }()

	router := setupRouterWithJWT()

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer validtoken")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), "success")
}
