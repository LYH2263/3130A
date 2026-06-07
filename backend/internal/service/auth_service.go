package service

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"label3130/backend/internal/auth"
	"label3130/backend/internal/dto"
	"label3130/backend/internal/models"
)

type AuthService struct {
	db     *gorm.DB
	tokens *auth.TokenManager
	log    *slog.Logger
}

type LoginResult struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

func NewAuthService(db *gorm.DB, tokens *auth.TokenManager, log *slog.Logger) *AuthService {
	return &AuthService{db: db, tokens: tokens, log: log}
}

func (s *AuthService) RegisterStudent(req dto.RegisterRequest) (*LoginResult, error) {
	username := strings.TrimSpace(strings.ToLower(req.Username))

	var existing models.User
	err := s.db.Where("username = ?", username).First(&existing).Error
	if err == nil {
		return nil, ErrUserExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("check existing user: %w", err)
	}

	var classRoom models.ClassRoom
	if err := s.db.First(&classRoom, req.ClassID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrClassNotFound
		}
		return nil, fmt.Errorf("find class: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := models.User{
		Username:     username,
		PasswordHash: string(hash),
		Role:         models.RoleStudent,
		ClassID:      &classRoom.ID,
	}
	if err := s.db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	token, err := s.tokens.Generate(user.ID, user.Role, user.ClassID)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	if err := s.db.Preload("ClassRoom").First(&user, user.ID).Error; err != nil {
		return nil, fmt.Errorf("reload user: %w", err)
	}

	s.log.Info("student registered", "userID", user.ID, "classID", classRoom.ID)
	return &LoginResult{Token: token, User: user}, nil
}

func (s *AuthService) Login(req dto.LoginRequest) (*LoginResult, error) {
	username := strings.TrimSpace(strings.ToLower(req.Username))

	var user models.User
	if err := s.db.Preload("ClassRoom").Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredential
		}
		return nil, fmt.Errorf("find user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredential
	}

	token, err := s.tokens.Generate(user.ID, user.Role, user.ClassID)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	s.log.Info("user login", "userID", user.ID, "role", user.Role)
	return &LoginResult{Token: token, User: user}, nil
}

func (s *AuthService) ListClasses() ([]models.ClassRoom, error) {
	var classes []models.ClassRoom
	if err := s.db.Order("name asc").Find(&classes).Error; err != nil {
		return nil, fmt.Errorf("list classes: %w", err)
	}
	return classes, nil
}

func (s *AuthService) GetUser(id uint) (*models.User, error) {
	var user models.User
	if err := s.db.Preload("ClassRoom").First(&user, id).Error; err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return &user, nil
}
