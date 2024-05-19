package repository

import "context"

type WithdrawalRepository interface {
	GetWithdrawalPoints(ctx context.Context, userID int) (float64, error)
	WithdrawFunds(ctx context.Context, userID int, orderNumber string, sum float64) error
}
