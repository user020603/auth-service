package services

import (
	"errors"
	"thanhnt208/vcs-sms/auth-service/internal/models"
	"thanhnt208/vcs-sms/auth-service/internal/repositories"
	"thanhnt208/vcs-sms/auth-service/pkg/logger"
	"thanhnt208/vcs-sms/auth-service/utils"
	"time"
)

type RegisterInput struct {
	Username string
	Password string
	Name     string
	Email    string
	Role     string
}

type LoginInput struct {
	Username string
	Password string
}

type IAuthService interface {
	Register(input RegisterInput) (uint, error)
	Login(input LoginInput) (string, string, error)
	RefreshToken(userID uint, refreshToken string) (string, error)
	Logout(userID uint) error
}

type authService struct {
	userRepo  repositories.IUserRepository
	tokenRepo repositories.ITokenRepository
	logger    logger.ILogger
}

func NewAuthService(userRepo repositories.IUserRepository, tokenRepo repositories.ITokenRepository, logger logger.ILogger) IAuthService {
	return &authService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		logger:    logger,
	}
}

func (s *authService) Register(input RegisterInput) (uint, error) {
	hash, err := utils.HashPassword(input.Password)
	if err != nil {
		s.logger.Error("Failed to hash password", "error", err)
		return 0, err
	}

	user := &models.User{
		Username: input.Username,
		Password: hash,
		Name:     input.Name,
		Email:    input.Email,
		Role:     input.Role,
	}

	err = s.userRepo.Create(user)
	if err != nil {
		s.logger.Error("Failed to create user", "username", input.Username, "error", err)
		return 0, err
	}

	return user.ID, nil
}

func (s *authService) Login(input LoginInput) (string, string, error) {
	user, err := s.userRepo.FindByUsername(input.Username)
	if err != nil {
		return "", "", errors.New("invalid username or password")
	}

	if !utils.CheckPasswordHash(input.Password, user.Password) {
		return "", "", errors.New("invalid username or password")
	}

	accessToken, err := utils.GenerateJWT(user.ID, user.Username, user.Role, time.Hour)
	if err != nil {
		s.logger.Error("Failed to generate access token", "username", user.Username, "error", err)
		return "", "", err
	}

	refreshToken, err := utils.GenerateJWT(user.ID, user.Username, user.Role, time.Hour*24*7)
	if err != nil {
		s.logger.Error("Failed to generate refresh token", "username", user.Username, "error", err)
		return "", "", err
	}

	err = s.tokenRepo.SetRefreshToken(user.ID, refreshToken, time.Hour*24*7)
	if err != nil {
		s.logger.Error("Failed to store refresh token", "userID", user.ID, "error", err)
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *authService) RefreshToken(userID uint, refreshToken string) (string, error) {
	storedToken, err := s.tokenRepo.GetRefreshToken(userID)
	if err != nil {
		s.logger.Error("Failed to get stored refresh token", "userID", userID, "error", err)
		return "", errors.New("invalid or expired refresh token")
	}
	if storedToken != refreshToken {
		s.logger.Warn("Refresh token mismatch", "userID", userID)
		return "", errors.New("invalid or expired refresh token")
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		s.logger.Error("Failed to find user during refresh", "userID", userID, "error", err)
		return "", err
	}

	newAccessToken, err := utils.GenerateJWT(user.ID, user.Username, user.Role, time.Hour)
	if err != nil {
		s.logger.Error("Failed to generate new access token", "userID", userID, "error", err)
		return "", err
	}

	return newAccessToken, nil
}

func (s *authService) Logout(userID uint) error {
	err := s.tokenRepo.DeleteRefreshToken(userID)
	if err != nil {
		s.logger.Error("Failed to delete refresh token on logout", "userID", userID, "error", err)
		return err
	}
	return nil
}
