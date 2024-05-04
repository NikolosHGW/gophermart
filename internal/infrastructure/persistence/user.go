package persistence

import (
	"fmt"

	"github.com/NikolosHGW/gophermart/internal/domain/entity"
	"github.com/jmoiron/sqlx"
)

type SQLUserRepository struct {
	db *sqlx.DB
}

func NewSQLUserRepository(db *sqlx.DB) *SQLUserRepository {
	return &SQLUserRepository{db: db}
}

func (r *SQLUserRepository) Save(user *entity.User) error {
	// Реализация сохранения пользователя в базу данных
	return fmt.Errorf("qq")
}

func (r *SQLUserRepository) ExistsByLogin(login string) bool {
	return true
}
