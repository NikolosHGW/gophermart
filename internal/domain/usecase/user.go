package usecase

import (
	"context"

	"github.com/NikolosHGW/gophermart/internal/domain/entity"
)

type UserUseCase interface {
	Register(ctx context.Context, login, password string) (*entity.User, error)
	GenerateJWT(user *entity.User) (string, error)
	Authenticate(ctx context.Context, login, password string) (*entity.User, error)
}
