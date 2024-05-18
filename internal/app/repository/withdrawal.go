package repository

import "context"

type WithdrawalRepository interface {
	GetWithdrawalPoints(ctx context.Context, userID int) (float64, error)
}
