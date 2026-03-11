package service

import (
	"errors"
	"time"

	"mcp-agent/internal/config"
	"mcp-agent/internal/model"
	"mcp-agent/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo *repository.UserRepository
	jwtCfg   config.JWTConfig
}

func NewAuthService(userRepo *repository.UserRepository, jwtCfg config.JWTConfig) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		jwtCfg:   jwtCfg,
	}
}

func (s *AuthService) Login(req model.LoginRequest) (*model.LoginResponse, error) {
	user, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	expireHours := s.jwtCfg.ExpireHours
	if expireHours <= 0 {
		expireHours = 24
	}

	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Duration(expireHours) * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(s.jwtCfg.Secret))
	if err != nil {
		return nil, err
	}

	return &model.LoginResponse{
		Token:     tokenStr,
		ExpiresIn: expireHours * 3600,
	}, nil
}

func (s *AuthService) RefreshToken(userID int64) (*model.LoginResponse, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	expireHours := s.jwtCfg.ExpireHours
	if expireHours <= 0 {
		expireHours = 24
	}

	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Duration(expireHours) * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(s.jwtCfg.Secret))
	if err != nil {
		return nil, err
	}

	return &model.LoginResponse{
		Token:     tokenStr,
		ExpiresIn: expireHours * 3600,
	}, nil
}

func (s *AuthService) GetProfile(userID int64) (*model.User, error) {
	return s.userRepo.GetByID(userID)
}

func (s *AuthService) CreateUser(req model.CreateUserRequest) error {
	if _, err := s.userRepo.GetByUsername(req.Username); err == nil {
		return ErrUserExists
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &model.User{
		Username: req.Username,
		Password: string(hashed),
		Role:     req.Role,
	}
	return s.userRepo.Create(user)
}
