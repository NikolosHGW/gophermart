package usecase

import "context"

type BalanceUseCase interface {
	GetBalanceByUserID(ctx context.Context, userID int) (float64, float64, error)
}
