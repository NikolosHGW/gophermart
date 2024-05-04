package usecase

import "github.com/NikolosHGW/gophermart/internal/domain/entity"

type UserUseCase interface {
	Register(login, password string) (*entity.User, error)
}
