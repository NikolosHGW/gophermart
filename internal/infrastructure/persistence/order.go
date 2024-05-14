package persistence

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/NikolosHGW/gophermart/internal/app/repository"
	"github.com/NikolosHGW/gophermart/internal/domain"
	"go.uber.org/zap"
)

type SQLOrderRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewSQLOrderRepository(db *sql.DB) repository.OrderRepository {
	return &SQLOrderRepository{db: db}
}

func (r *SQLOrderRepository) OrderExistsForUser(ctx context.Context, userID int, orderNumber string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM orders WHERE user_id = $1 AND number = $2)`
	err := r.db.QueryRowContext(ctx, query, userID, orderNumber).Scan(&exists)
	if err != nil {
		r.logger.Info("не получилось выполнить запрос на существования заказа", zap.Error(err))
		return exists, fmt.Errorf("внутренняя ошибка сервера")
	}
	return exists, nil
}

func (r *SQLOrderRepository) OrderClaimedByAnotherUser(
	ctx context.Context,
	userID int,
	orderNumber string,
) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM orders WHERE number = $1 AND user_id != $2)`
	err := r.db.QueryRowContext(ctx, query, orderNumber, userID).Scan(&exists)
	if err != nil {
		r.logger.Info("не получилось выполнить запрос на существования заказа", zap.Error(err))
		return exists, fmt.Errorf("внутренняя ошибка сервера")
	}
	return exists, nil
}

func (r *SQLOrderRepository) AddOrder(ctx context.Context, userID int, orderNumber string) error {
	query := `INSERT INTO orders (user_id, number, status) VALUES ($1, $2, $3)`
	_, err := r.db.ExecContext(ctx, query, userID, orderNumber, domain.StatusNew)
	if err != nil {
		r.logger.Info("не получилось добавить заказ", zap.Error(err))
		return fmt.Errorf("внутренняя ошибка сервера")
	}
	return nil
}
