package persistence

import (
	"context"

	"github.com/NikolosHGW/gophermart/internal/domain"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type SQLWithdrawalRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

func NewSQLWithdrawalRepository(
	db *sqlx.DB,
	logger *zap.Logger,
) *SQLWithdrawalRepository {
	return &SQLWithdrawalRepository{
		db:     db,
		logger: logger,
	}
}

func (r *SQLWithdrawalRepository) GetWithdrawalPoints(ctx context.Context, userID int) (float64, error) {
	var withdrawnPoints float64
	err := r.db.GetContext(
		ctx,
		&withdrawnPoints,
		`SELECT COALESCE(SUM(sum), 0) AS withdrawn_points FROM withdrawals WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		r.logger.Error("ошибка при получении суммы использованных баллов", zap.Error(err))
		return 0, domain.ErrInternalServer
	}
	return withdrawnPoints, nil
}

func (r *SQLWithdrawalRepository) WithdrawFunds(
	ctx context.Context,
	userID int,
	orderNumber string,
	sum float64,
) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		r.logger.Error("ошибка при запуске транзакции для списания", zap.Error(err))
		return domain.ErrInternalServer
	}

	insertLoyaltyPointsQuery := `INSERT INTO loyalty_points (user_id, spent_point, accrued_point) VALUES ($1, $2, 0)`
	_, err = tx.ExecContext(ctx, insertLoyaltyPointsQuery, userID, sum)
	if err != nil {
		errRollback := tx.Rollback()
		if errRollback != nil {
			r.logger.Error("ошибка при вызове tx.Rollback() 1", zap.Error(err))
		}
		r.logger.Error("ошибка при добавлении строки в loyalty_points", zap.Error(err))
		return domain.ErrInternalServer
	}

	insertWithdrawalQuery := `
	INSERT INTO withdrawals (user_id, order_id, sum) 
	VALUES ($1, (SELECT id FROM orders WHERE number = $2 AND user_id = $3), $4)`
	_, err = tx.ExecContext(ctx, insertWithdrawalQuery, userID, orderNumber, userID, sum)
	if err != nil {
		errRollback := tx.Rollback()
		if errRollback != nil {
			r.logger.Error("ошибка при вызове tx.Rollback() 2", zap.Error(err))
		}
		r.logger.Error("ошибка при добавлении строки в withdrawals", zap.Error(err))
		return domain.ErrInternalServer
	}

	err = tx.Commit()
	if err != nil {
		r.logger.Error("ошибка закрытии транзакции", zap.Error(err))
		return domain.ErrInternalServer
	}

	return nil
}