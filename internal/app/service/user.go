package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/NikolosHGW/gophermart/internal/app/repository"
	"github.com/NikolosHGW/gophermart/internal/domain"
	"github.com/NikolosHGW/gophermart/internal/domain/entity"
	"github.com/NikolosHGW/gophermart/internal/domain/usecase"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo repository.UserRepository
	logger   *zap.Logger
}

func NewUserService(userRepo repository.UserRepository, logger *zap.Logger) usecase.UserUseCase {
	return &UserService{
		userRepo: userRepo,
		logger:   logger,
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
