package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/NikolosHGW/gophermart/internal/domain"
	"github.com/NikolosHGW/gophermart/internal/domain/entity"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type SQLUserRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

func NewSQLUserRepository(db *sqlx.DB, logger *zap.Logger) *SQLUserRepository {
	return &SQLUserRepository{db: db, logger: logger}
}

func (r *SQLUserRepository) Save(ctx context.Context, user *entity.User) error {
	query := `INSERT INTO users (login, password) VALUES ($1, $2)`
	_, err := r.db.ExecContext(ctx, query, user.Login, user.Password)
	if err != nil {
		return fmt.Errorf("ошибка при сохранении пользователя: %w", err)
	}

	return nil
}

func (r *SQLUserRepository) ExistsByLogin(ctx context.Context, login string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE login=$1)`
	err := r.db.QueryRowxContext(ctx, query, login).Scan(&exists)
	if err != nil {
		r.logger.Info("не получилось записать результат запроса в переменную", zap.Error(err))
		return false, fmt.Errorf("временная ошибка сервиса, попробуйте ещё раз позже")
	}
	return exists, nil
}

func (r *SQLUserRepository) FindByLogin(ctx context.Context, login string) (*entity.User, error) {
	var user entity.User
	query := `SELECT id, login, password FROM users WHERE login = $1`
	err := r.db.GetContext(ctx, &user, query, login)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrInvalidCredentials
		}

		r.logger.Info("ошибка при поиске пользователя: ", zap.Error(err))

		return nil, fmt.Errorf("ошибка при поиске пользователя")
	}
	return &user, nil
}
