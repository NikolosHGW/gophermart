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
