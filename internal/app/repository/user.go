package repository

import "github.com/NikolosHGW/gophermart/internal/domain/entity"

type UserRepository interface {
	Save(user *entity.User) error
	ExistsByLogin(string) bool
}
