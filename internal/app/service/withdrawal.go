package service

import (
	"context"
	"fmt"

	"github.com/NikolosHGW/gophermart/internal/app/repository"
	"github.com/NikolosHGW/gophermart/internal/domain"
	"github.com/NikolosHGW/gophermart/internal/domain/entity"
	"github.com/NikolosHGW/gophermart/internal/domain/usecase"
)

type WithdrawalService struct {
	withdrawalRepo repository.WithdrawalRepository
}

func NewWithdrawalService(withdrawalRepo repository.WithdrawalRepository) usecase.WithdrawalUseCase {
	return &WithdrawalService{
		withdrawalRepo: withdrawalRepo,
	}
}

func (s *WithdrawalService) ValidBalance(current, sum float64) bool {
	return current >= sum
}

func (s *WithdrawalService) WithdrawFunds(
	ctx context.Context,
	userID int,
	orderNumber string,
	sum float64,
) error {
	err := s.withdrawalRepo.WithdrawFunds(ctx, userID, orderNumber, sum)
	if err != nil {
		return domain.ErrInternalServer
	}
	return nil
}

func (s *WithdrawalService) GetWithdrawalsByUserID(ctx context.Context, userID int) ([]entity.Withdrawal, error) {
	withdrawals, err := s.withdrawalRepo.GetWithdrawalsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("не получилось достать все списания: %w", err)
	}

	return withdrawals, nil
}
