package persistence

import (
	"context"

	"github.com/NikolosHGW/gophermart/internal/domain"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type SQLLoyaltyPointRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

func NewSQLLoyaltyPointRepository(db *sqlx.DB, logger *zap.Logger) *SQLLoyaltyPointRepository {
	return &SQLLoyaltyPointRepository{db: db, logger: logger}
}

func (r *SQLLoyaltyPointRepository) GetCurrentPoints(ctx context.Context, userID int) (float64, error) {
	var currentPoints float64
	query := `SELECT COALESCE(SUM(accrued_point) - SUM(spent_point), 0) AS current_points 
	FROM loyalty_points 
	WHERE user_id = $1`
	err := r.db.GetContext(ctx, &currentPoints, query, userID)
	if err != nil {
		r.logger.Error("ошибка при получении текущих баллов лояльности", zap.Error(err))
		return 0, domain.ErrInternalServer
	}
	return currentPoints, nil
}
