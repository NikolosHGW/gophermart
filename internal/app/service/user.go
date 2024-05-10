package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/NikolosHGW/gophermart/internal/app/repository"
	"github.com/NikolosHGW/gophermart/internal/domain"
	"github.com/NikolosHGW/gophermart/internal/domain/entity"
	"github.com/NikolosHGW/gophermart/internal/domain/usecase"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const TokenExp = time.Hour * 5

type UserService struct {
	userRepo  repository.UserRepository
	logger    *zap.Logger
	secretKey string
}

func NewUserService(userRepo repository.UserRepository, logger *zap.Logger, secretKey string) usecase.UserUseCase {
	return &UserService{
		userRepo:  userRepo,
		logger:    logger,
		secretKey: secretKey,
	}
}

func (s *UserService) Register(ctx context.Context, login, password string) (*entity.User, error) {
	isLoginExist, err := s.userRepo.ExistsByLogin(ctx, login)
	if err != nil {
		return nil, fmt.Errorf("ошибка сервера: %w", err)
	}
	if isLoginExist {
		return nil, domain.ErrLoginAlreadyExists
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Info("ошибка при хэшировании пароля: ", zap.Error(err))
		return nil, errors.New("временная ошибка сервиса, попробуйте ещё раз позже")
	}

	user := &entity.User{
		Login:    login,
		Password: string(passwordHash),
	}

	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, fmt.Errorf("ошибка при сохранении пользователя: %w", err)
	}

	return user, nil
}

func (s *UserService) GenerateJWT(user *entity.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
		},
		UserID: user.ID,
	})

	if s.secretKey == "" {
		s.logger.Info("для создании подписи токена секретный ключ пустой")
		return "", fmt.Errorf("ошибки при создании токена")
	}
	tokenString, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		s.logger.Info("ошибки при создании подписи токена: ", zap.Error(err))
		return "", fmt.Errorf("ошибки при создании подписи токена")
	}

	return tokenString, nil
}

type Claims struct {
	jwt.RegisteredClaims
	UserID int
}
