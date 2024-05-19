package usecase

import "context"

type WithdrawalUseCase interface {
	ValidBalance(float64, float64) bool
	WithdrawFunds(context.Context, int, string, float64) error
}
