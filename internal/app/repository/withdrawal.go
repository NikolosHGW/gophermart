package repository

import (
	"context"

	"github.com/NikolosHGW/gophermart/internal/domain/entity"
)

type WithdrawalRepository interface {
	GetWithdrawalPoints(ctx context.Context, userID int) (float64, error)
	WithdrawFunds(ctx context.Context, userID int, orderNumber string, sum float64) error
	GetWithdrawalsByUserID(ctx context.Context, userID int) ([]entity.Withdrawal, error)
}
