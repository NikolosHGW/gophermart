package usecase

import (
	"context"

	"github.com/NikolosHGW/gophermart/internal/domain/entity"
)

type WithdrawalUseCase interface {
	ValidBalance(float64, float64) bool
	WithdrawFunds(context.Context, int, string, float64) error
	GetWithdrawalsByUserID(context.Context, int) ([]entity.Withdrawal, error)
}
