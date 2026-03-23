package services

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/veggieshop/backend/internal/models"
	"github.com/veggieshop/backend/internal/repositories"
	"github.com/veggieshop/backend/internal/utils"
)

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	User         *models.User `json:"user"`
}

type AuthService struct {
	userRepo   repositories.UserRepository
	jwtSecret  string
	accessExp  int
	refreshExp int
}

func NewAuthService(ur repositories.UserRepository, jwtSecret string, accessExp, refreshExp int) *AuthService {
	return &AuthService{
		userRepo:  ur,
		jwtSecret: jwtSecret,
		accessExp: accessExp,
		refreshExp: refreshExp,
	}
}

func (s *AuthService) RegisterUser(ctx context.Context, phone, password, firstName, lastName string) (*models.User, error) {
	phone = utils.NormalizePhone(phone)
	if !utils.ValidatePhone(phone) {
		return nil, utils.ErrInvalidInput
	}
	if ok, _ := utils.ValidatePassword(password); !ok {
		return nil, utils.ErrInvalidInput
	}
	hash, err := utils.HashPassword(password)
	if err != nil {
		slog.Error("password hash failed", "error", err)
		return nil, err
	}
	user := &models.User{
		Phone:        phone,
		PasswordHash: hash,
		FirstName:    firstName,
		LastName:     lastName,
		Role:         models.RoleCustomer,
		IsActive:     true,
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}
	slog.Info("user registered", "phone", phone)
	return user, nil
}

func (s *AuthService) LoginUser(ctx context.Context, phone, password string) (*TokenResponse, error) {
	phone = utils.NormalizePhone(phone)
	user, err := s.userRepo.GetByPhone(ctx, phone)
	if err != nil {
		return nil, utils.ErrUnauthorized
	}
	if !user.IsActive {
		return nil, utils.ErrUnauthorized
	}
	if !utils.CheckPassword(password, user.PasswordHash) {
		return nil, utils.ErrUnauthorized
	}
	return s.generateTokens(user)
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	claims, err := utils.ParseToken(refreshToken, s.jwtSecret)
	if err != nil {
		return nil, utils.ErrUnauthorized
	}
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil || !user.IsActive {
		return nil, utils.ErrUnauthorized
	}
	return s.generateTokens(user)
}

func (s *AuthService) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

func (s *AuthService) generateTokens(user *models.User) (*TokenResponse, error) {
	access, err := utils.GenerateAccessToken(user.ID, user.Phone, string(user.Role), s.jwtSecret, s.accessExp)
	if err != nil {
		return nil, err
	}
	refresh, err := utils.GenerateRefreshToken(user.ID, s.jwtSecret, s.refreshExp)
	if err != nil {
		return nil, err
	}
	return &TokenResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    s.accessExp * 60,
		User:         user,
	}, nil
}
