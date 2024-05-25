package persistence

import (
	"context"

	"github.com/NikolosHGW/gophermart/internal/app/repository"
	"github.com/NikolosHGW/gophermart/internal/domain"
	"github.com/NikolosHGW/gophermart/internal/domain/entity"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type SQLOrderRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

func NewSQLOrderRepository(db *sqlx.DB, logger *zap.Logger) repository.OrderRepository {
	return &SQLOrderRepository{db: db, logger: logger}
}

func (r *SQLOrderRepository) OrderExistsForUser(ctx context.Context, userID int, orderNumber string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM orders WHERE user_id = $1 AND number = $2)`
	err := r.db.QueryRowxContext(ctx, query, userID, orderNumber).Scan(&exists)
	if err != nil {
		r.logger.Info("не получилось выполнить запрос на существования заказа", zap.Error(err))
		return exists, domain.ErrInternalServer
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
	err := r.db.QueryRowxContext(ctx, query, orderNumber, userID).Scan(&exists)
	if err != nil {
		r.logger.Info("не получилось выполнить запрос на существования заказа", zap.Error(err))
		return exists, domain.ErrInternalServer
	}
	return exists, nil
}

func (r *SQLOrderRepository) AddOrder(ctx context.Context, userID int, orderNumber string) error {
	query := `INSERT INTO orders (user_id, number, status) VALUES ($1, $2, $3)`
	_, err := r.db.ExecContext(ctx, query, userID, orderNumber, domain.StatusNew)
	if err != nil {
		r.logger.Info("не получилось добавить заказ", zap.Error(err))
		return domain.ErrInternalServer
	}
	return nil
}

func (r *SQLOrderRepository) GetUserOrdersByID(ctx context.Context, userID int) ([]entity.Order, error) {
	var orders []entity.Order
	query := `
	SELECT o.number,
		o.status,
		to_char(o.uploaded_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SSZ') as uploaded_at,
		lp.accrued_point as accrual
	FROM orders o
	LEFT JOIN loyalty_points lp ON o.number = lp.order_number
	WHERE o.user_id = $1
	ORDER BY o.uploaded_at ASC`
	err := r.db.SelectContext(ctx, &orders, query, userID)
	if err != nil {
		r.logger.Info("не получилось получить список заказов", zap.Error(err))
		return nil, domain.ErrInternalServer
	}
	return orders, nil
}
