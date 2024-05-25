package persistence

import (
	"context"
	"fmt"

	"github.com/NikolosHGW/gophermart/internal/app/repository"
	"github.com/NikolosHGW/gophermart/internal/domain"
	"github.com/NikolosHGW/gophermart/internal/domain/entity"
	"github.com/jmoiron/sqlx"
)

type SQLAccrualRepository struct {
	db *sqlx.DB
}

func NewSQLAccrualRepository(db *sqlx.DB) repository.AccrualRepository {
	return &SQLAccrualRepository{db: db}
}

func (r *SQLAccrualRepository) GetNonFinalOrders(
	ctx context.Context,
	limit int,
	prevOrderNumber string,
) ([]entity.Order, error) {
	orders := []entity.Order{}

	var lastOrderNumber string
	err := r.db.GetContext(ctx, &lastOrderNumber, "SELECT number FROM orders ORDER BY number DESC LIMIT 1")
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении последнего номера заказа: %w", err)
	}

	if prevOrderNumber == lastOrderNumber || prevOrderNumber == "" {
		prevOrderNumber = "0"
	}

	query := `
		SELECT number, status, uploaded_at
		FROM orders
		WHERE status NOT IN ($1, $2) AND CAST(number AS INTEGER) > CAST($3 AS INTEGER)
		LIMIT $4`
	err = r.db.SelectContext(
		ctx,
		&orders,
		query,
		domain.StatusInvalid,
		domain.StatusProcessed,
		prevOrderNumber,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка при запросе на получение незавершённых заказов: %w", err)
	}
	return orders, nil
}

func (r *SQLAccrualRepository) UpdateAccrual(
	ctx context.Context,
	orderNumber string,
	accrual float64,
	status string,
) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("ошибка при запуске транзакции: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			err = tx.Rollback()
			panic(p)
		} else if err != nil {
			err = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	query := `
		UPDATE orders 
		SET status = $1 
		WHERE number = $2`
	_, err = tx.ExecContext(ctx, query, status, orderNumber)
	if err != nil {
		return fmt.Errorf("ошибка при обновлении статуса: %w", err)
	}

	if status == "PROCESSED" && accrual > 0 {
		var userID int
		err = tx.GetContext(ctx, &userID, "SELECT user_id FROM orders WHERE number = $1", orderNumber)
		if err != nil {
			return fmt.Errorf("ошибка при запросе на получения user ID: %w", err)
		}

		_, err = tx.ExecContext(ctx, `
			INSERT INTO loyalty_points (user_id, accrued_point, spent_point)
			VALUES ($1, $2, 0)`, userID, accrual)
		if err != nil {
			return fmt.Errorf("ошибка при начислении баллов: %w", err)
		}
	}

	return nil
}
