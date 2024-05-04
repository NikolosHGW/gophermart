package service

import (
	"fmt"

	"github.com/NikolosHGW/gophermart/internal/domain/entity"
	"github.com/NikolosHGW/gophermart/internal/domain/repository"
	"github.com/NikolosHGW/gophermart/internal/domain/usecase"
	"golang.org/x/crypto/bcrypt"
)

// ErrLoginAlreadyExists заглушка для ошибки.
var ErrLoginAlreadyExists = fmt.Errorf("qq")

type UserService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) usecase.UserUseCase {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) Register(login, password string) (*entity.User, error) {
	if s.userRepo.ExistsByLogin(login) {
		return nil, ErrLoginAlreadyExists
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("ошибка при хэшировании пароля: %w", err)
	}

	user := &entity.User{
		Login:    login,
		Password: string(passwordHash),
	}

	if err := s.userRepo.Save(user); err != nil {
		return nil, fmt.Errorf("ошибка при сохранении пользователя: %w", err)
	}

	return user, nil
}
