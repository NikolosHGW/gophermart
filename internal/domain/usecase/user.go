package usecase

import (
	"context"

	"github.com/NikolosHGW/gophermart/internal/domain/entity"
)

type UserUseCase interface {
	Register(ctx context.Context, login, password string) (*entity.User, error)
}
