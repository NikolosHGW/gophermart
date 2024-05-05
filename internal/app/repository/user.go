package repository

import (
	"context"

	"github.com/NikolosHGW/gophermart/internal/domain/entity"
)

type UserRepository interface {
	Save(context.Context, *entity.User) error
	ExistsByLogin(context.Context, string) (bool, error)
}
