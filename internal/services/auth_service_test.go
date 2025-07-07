package services

import (
	"errors"
	"testing"
	"time"

	"thanhnt208/vcs-sms/auth-service/internal/models"
	"thanhnt208/vcs-sms/auth-service/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

type MockTokenRepository struct {
	mock.Mock
}

type MockLogger struct{}

func (m *MockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByUsername(username string) (*models.User, error) {
	args := m.Called(username)
	user, _ := args.Get(0).(*models.User)
	return user, args.Error(1)
}

func (m *MockUserRepository) FindByID(id uint) (*models.User, error) {
	args := m.Called(id)
	user, _ := args.Get(0).(*models.User)
	return user, args.Error(1)
}

func (m *MockTokenRepository) SetRefreshToken(userID uint, token string, expiration time.Duration) error {
	args := m.Called(userID, token, expiration)
	return args.Error(0)
}

func (m *MockTokenRepository) GetRefreshToken(userID uint) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}

func (m *MockTokenRepository) DeleteRefreshToken(userID uint) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (l *MockLogger) Fatal(msg string, keysAndValues ...interface{}) {
	panic("unimplemented")
}

func (l *MockLogger) Sync() error {
	panic("unimplemented")
}

func (l *MockLogger) Debug(message string, keyvals ...interface{}) {}
func (l *MockLogger) Info(message string, keyvals ...interface{})  {}
func (l *MockLogger) Warn(message string, keyvals ...interface{})  {}
func (l *MockLogger) Error(message string, keyvals ...interface{}) {}

func TestAuthService_Register(t *testing.T) {
	userRepo := new(MockUserRepository)
	tokenRepo := new(MockTokenRepository)
	logger := &MockLogger{}
	service := NewAuthService(userRepo, tokenRepo, logger)

	hashedPassword, _ := utils.HashPassword("password123")
	userToCreate := &models.User{
		Username: "testuser",
		Password: hashedPassword,
		Name:     "Test User",
		Email:    "test@example.com",
		Role:     "user",
	}

	userRepo.On("Create", mock.AnythingOfType("*models.User")).
		Run(func(args mock.Arguments) {
			u := args.Get(0).(*models.User)
			u.ID = 10
		}).
		Return(nil)

	id, err := service.Register(RegisterInput{
		Username: "testuser",
		Password: "password123",
		Name:     "Test User",
		Email:    "test@example.com",
		Role:     "user",
	})

	assert.NoError(t, err)
	assert.Equal(t, uint(10), id)
	userRepo.AssertCalled(t, "Create", mock.MatchedBy(func(u *models.User) bool {
		return u.Username == userToCreate.Username && u.Email == userToCreate.Email
	}))
}

func TestAuthService_Register_HashPasswordError(t *testing.T) {
	userRepo := new(MockUserRepository)
	tokenRepo := new(MockTokenRepository)
	logger := &MockLogger{}
	service := NewAuthService(userRepo, tokenRepo, logger)

	originalHashPassword := utils.HashPassword
	utils.HashPassword = func(password string) (string, error) {
		return "", errors.New("mock hash error")
	}
	defer func() {
		utils.HashPassword = originalHashPassword
	}()

	id, err := service.Register(RegisterInput{
		Username: "errorUser",
		Password: "somePassword",
	})

	assert.Error(t, err)
	assert.Equal(t, uint(0), id)
}

func TestAuthService_Register_CreateError(t *testing.T) {
	userRepo := new(MockUserRepository)
	tokenRepo := new(MockTokenRepository)
	logger := &MockLogger{}
	service := NewAuthService(userRepo, tokenRepo, logger)

	userRepo.On("Create", mock.AnythingOfType("*models.User")).Return(errors.New("create error"))

	_, err := service.Register(RegisterInput{
		Username: "testerror",
		Password: "pass",
	})
	assert.Error(t, err)
}

func TestAuthService_Login(t *testing.T) {
	userRepo := new(MockUserRepository)
	tokenRepo := new(MockTokenRepository)
	log := &MockLogger{}
	service := NewAuthService(userRepo, tokenRepo, log)

	hashedPassword, _ := utils.HashPassword("password123")
	storedUser := &models.User{
		ID:       20,
		Username: "testuser",
		Password: hashedPassword,
		Role:     "user",
	}

	userRepo.On("FindByUsername", "testuser").Return(storedUser, nil)
	tokenRepo.On("SetRefreshToken", uint(20), mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(nil)

	accessToken, refreshToken, err := service.Login(LoginInput{
		Username: "testuser",
		Password: "password123",
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
}

func TestAuthService_Login_InvalidUsername(t *testing.T) {
	userRepo := new(MockUserRepository)
	tokenRepo := new(MockTokenRepository)
	logger := &MockLogger{}
	service := NewAuthService(userRepo, tokenRepo, logger)

	userRepo.On("FindByUsername", "wronguser").Return(nil, nil)

	accessToken, refreshToken, err := service.Login(LoginInput{
		Username: "wronguser",
		Password: "password123",
	})

	assert.Error(t, err)
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	userRepo := new(MockUserRepository)
	tokenRepo := new(MockTokenRepository)
	logger := &MockLogger{}
	service := NewAuthService(userRepo, tokenRepo, logger)

	hashedPassword, _ := utils.HashPassword("password123")
	storedUser := &models.User{
		ID:       20,
		Username: "testuser",
		Password: hashedPassword,
		Role:     "user",
	}

	userRepo.On("FindByUsername", "testuser").Return(storedUser, nil)

	accessToken, refreshToken, err := service.Login(LoginInput{
		Username: "testuser",
		Password: "wrongpassword",
	})

	assert.Error(t, err)
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
}

func TestAuthService_Login_FindByUsernameError(t *testing.T) {
	userRepo := new(MockUserRepository)
	tokenRepo := new(MockTokenRepository)
	logger := &MockLogger{}
	service := NewAuthService(userRepo, tokenRepo, logger)

	userRepo.On("FindByUsername", "erroredUser").Return(nil, errors.New("db error"))

	accessToken, refreshToken, err := service.Login(LoginInput{
		Username: "erroredUser",
		Password: "somePassword",
	})

	assert.Error(t, err)
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
	assert.Contains(t, err.Error(), "invalid username or password")
}

func TestAuthService_Login_SetRefreshTokenError(t *testing.T) {
	userRepo := new(MockUserRepository)
	tokenRepo := new(MockTokenRepository)
	logger := &MockLogger{}
	service := NewAuthService(userRepo, tokenRepo, logger)

	hashedPassword, _ := utils.HashPassword("password123")
	storedUser := &models.User{
		ID:       20,
		Username: "testuser",
		Password: hashedPassword,
	}

	userRepo.On("FindByUsername", "testuser").Return(storedUser, nil)
	tokenRepo.On("SetRefreshToken", uint(20), mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).
		Return(errors.New("set token error"))

	accessToken, refreshToken, err := service.Login(LoginInput{
		Username: "testuser",
		Password: "password123",
	})

	assert.Error(t, err)
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
}

func TestAuthService_Login_GenerateJWTError(t *testing.T) {
	userRepo := new(MockUserRepository)
	tokenRepo := new(MockTokenRepository)
	logger := &MockLogger{}
	service := NewAuthService(userRepo, tokenRepo, logger)

	hashedPassword, _ := utils.HashPassword("password123")
	storedUser := &models.User{
		ID:       99,
		Username: "jwtErrorUser",
		Password: hashedPassword,
		Role:     "user",
	}

	userRepo.On("FindByUsername", "jwtErrorUser").Return(storedUser, nil)

	originalGenerateJWT := utils.GenerateJWT
	utils.GenerateJWT = func(id uint, username, role string, scopes []string, duration time.Duration) (string, error) {
		return "", errors.New("mocked error")
	}
	defer func() {
		utils.GenerateJWT = originalGenerateJWT
	}()

	accessToken, refreshToken, err := service.Login(LoginInput{
		Username: "jwtErrorUser",
		Password: "password123",
	})

	assert.Error(t, err)
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
}

func TestAuthService_Login_FailRefreshTokenGeneration(t *testing.T) {
	userRepo := new(MockUserRepository)
	tokenRepo := new(MockTokenRepository)
	logger := &MockLogger{}
	service := NewAuthService(userRepo, tokenRepo, logger)

	hashedPassword, _ := utils.HashPassword("password123")
	storedUser := &models.User{
		ID:       88,
		Username: "refreshTokenFailUser",
		Password: hashedPassword,
		Role:     "user",
	}

	userRepo.On("FindByUsername", "refreshTokenFailUser").Return(storedUser, nil)

	callCount := 0
	originalGenerateJWT := utils.GenerateJWT
	utils.GenerateJWT = func(id uint, username, role string, scopes []string, duration time.Duration) (string, error) {
		callCount++
		if callCount == 1 {
			return "mockedAccessToken", nil
		}
		return "", errors.New("failed to generate refresh token")
	}
	defer func() {
		utils.GenerateJWT = originalGenerateJWT
	}()

	accessToken, refreshToken, err := service.Login(LoginInput{
		Username: "refreshTokenFailUser",
		Password: "password123",
	})

	assert.Error(t, err)
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
}

func TestAuthService_RefreshToken(t *testing.T) {
	userRepo := new(MockUserRepository)
	tokenRepo := new(MockTokenRepository)
	log := &MockLogger{}
	service := NewAuthService(userRepo, tokenRepo, log)

	storedUser := &models.User{
		ID:       30,
		Username: "anotheruser",
		Password: "hashedPass",
		Role:     "user",
	}
	mockRefreshToken := "valid-refreshtoken"

	tokenRepo.On("GetRefreshToken", uint(30)).Return(mockRefreshToken, nil)
	userRepo.On("FindByID", uint(30)).Return(storedUser, nil)

	token, err := service.RefreshToken(30, mockRefreshToken)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestAuthService_RefreshToken_InvalidToken(t *testing.T) {
	userRepo := new(MockUserRepository)
	tokenRepo := new(MockTokenRepository)
	log := &MockLogger{}
	service := NewAuthService(userRepo, tokenRepo, log)

	tokenRepo.On("GetRefreshToken", uint(40)).Return("stored_token", nil)
	_, err := service.RefreshToken(40, "wrong_token")
	assert.Error(t, err)
}

func TestAuthService_RefreshToken_GetTokenError(t *testing.T) {
	userRepo := new(MockUserRepository)
	tokenRepo := new(MockTokenRepository)
	logger := &MockLogger{}
	service := NewAuthService(userRepo, tokenRepo, logger)

	tokenRepo.On("GetRefreshToken", uint(1)).Return("", errors.New("repo error"))

	_, err := service.RefreshToken(1, "anytoken")
	assert.Error(t, err)
}

func TestAuthService_RefreshToken_FindUserError(t *testing.T) {
	userRepo := new(MockUserRepository)
	tokenRepo := new(MockTokenRepository)
	logger := &MockLogger{}
	service := NewAuthService(userRepo, tokenRepo, logger)

	tokenRepo.On("GetRefreshToken", uint(2)).Return("stored_token", nil)
	userRepo.On("FindByID", uint(2)).Return(nil, errors.New("find error"))

	_, err := service.RefreshToken(2, "stored_token")
	assert.Error(t, err)
}

func TestAuthService_RefreshToken_FailAccessTokenGeneration(t *testing.T) {
	userRepo := new(MockUserRepository)
	tokenRepo := new(MockTokenRepository)
	logger := &MockLogger{}
	service := NewAuthService(userRepo, tokenRepo, logger)

	storedUser := &models.User{
		ID:       101,
		Username: "refreshFailUser",
		Password: "hashedPass",
		Role:     "admin",
	}
	tokenRepo.On("GetRefreshToken", uint(101)).Return("valid_token", nil)
	userRepo.On("FindByID", uint(101)).Return(storedUser, nil)

	originalGenerateJWT := utils.GenerateJWT
	utils.GenerateJWT = func(id uint, username, role string, scopes []string, duration time.Duration) (string, error) {
		return "", errors.New("mocked generation error")
	}
	defer func() {
		utils.GenerateJWT = originalGenerateJWT
	}()

	newAccessToken, err := service.RefreshToken(101, "valid_token")

	assert.Error(t, err)
	assert.Empty(t, newAccessToken)
}

func TestAuthService_Logout(t *testing.T) {
	userRepo := new(MockUserRepository)
	tokenRepo := new(MockTokenRepository)
	log := &MockLogger{}
	service := NewAuthService(userRepo, tokenRepo, log)

	tokenRepo.On("DeleteRefreshToken", uint(50)).Return(nil)

	err := service.Logout(50)
	assert.NoError(t, err)
	tokenRepo.AssertCalled(t, "DeleteRefreshToken", uint(50))
}

func TestAuthService_Logout_DeleteRefreshTokenError(t *testing.T) {
	userRepo := new(MockUserRepository)
	tokenRepo := new(MockTokenRepository)
	logger := &MockLogger{}
	service := NewAuthService(userRepo, tokenRepo, logger)

	tokenRepo.On("DeleteRefreshToken", uint(60)).Return(errors.New("delete token error"))

	err := service.Logout(60)
	assert.Error(t, err)
}
