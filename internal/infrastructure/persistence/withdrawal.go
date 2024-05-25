package persistence

import (
	"context"

	"github.com/NikolosHGW/gophermart/internal/domain"
	"github.com/NikolosHGW/gophermart/internal/domain/entity"
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
		r.logger.Info("ошибка при получении суммы использованных баллов", zap.Error(err))
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
		r.logger.Info("ошибка при запуске транзакции для списания", zap.Error(err))
		return domain.ErrInternalServer
	}

	insertLoyaltyPointsQuery := `INSERT INTO loyalty_points (user_id, spent_point, accrued_point) VALUES ($1, $2, 0)`
	_, err = tx.ExecContext(ctx, insertLoyaltyPointsQuery, userID, sum)
	if err != nil {
		errRollback := tx.Rollback()
		if errRollback != nil {
			r.logger.Info("ошибка при вызове tx.Rollback() 1", zap.Error(err))
		}
		r.logger.Info("ошибка при добавлении строки в loyalty_points", zap.Error(err))
		return domain.ErrInternalServer
	}

	insertWithdrawalQuery := `
	INSERT INTO withdrawals (user_id, order_number, sum) VALUES ($1, $2, $3)`
	_, err = tx.ExecContext(ctx, insertWithdrawalQuery, userID, orderNumber, sum)
	if err != nil {
		errRollback := tx.Rollback()
		if errRollback != nil {
			r.logger.Info("ошибка при вызове tx.Rollback() 2", zap.Error(err))
		}
		r.logger.Info("ошибка при добавлении строки в withdrawals", zap.Error(err))
		return domain.ErrInternalServer
	}

	err = tx.Commit()
	if err != nil {
		r.logger.Info("ошибка закрытии транзакции", zap.Error(err))
		return domain.ErrInternalServer
	}

	return nil
}

func (r *SQLWithdrawalRepository) GetWithdrawalsByUserID(
	ctx context.Context,
	userID int,
) ([]entity.Withdrawal, error) {
	var withdrawals []entity.Withdrawal
	query := `
	SELECT order_number, sum, processed_at
	FROM withdrawals
	WHERE user_id = $1
	ORDER BY processed_at ASC`
	err := r.db.SelectContext(ctx, &withdrawals, query, userID)
	if err != nil {
		r.logger.Info("ошибка при выборке withdrawals", zap.Error(err))
		return nil, domain.ErrInternalServer
	}

	return withdrawals, nil
}
