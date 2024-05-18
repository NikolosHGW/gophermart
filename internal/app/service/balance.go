package service

import (
	"context"

	"github.com/NikolosHGW/gophermart/internal/app/repository"
	"github.com/NikolosHGW/gophermart/internal/domain"
	"github.com/NikolosHGW/gophermart/internal/domain/usecase"
	"go.uber.org/zap"
)

type BalanceService struct {
	loyaltyPointRepo repository.LoyaltyPointRepository
	withdrawalRepo   repository.WithdrawalRepository
	logger           *zap.Logger
}

func NewBalanceService(
	loyaltyPointRepo repository.LoyaltyPointRepository,
	withdrawalRepo repository.WithdrawalRepository,
	logger *zap.Logger,
) usecase.BalanceUseCase {
	return &BalanceService{
		loyaltyPointRepo: loyaltyPointRepo,
		withdrawalRepo:   withdrawalRepo,
		logger:           logger,
	}
}

func (s *BalanceService) GetBalanceByUserID(ctx context.Context, userID int) (float64, float64, error) {
	current, err := s.loyaltyPointRepo.GetCurrentPoints(ctx, userID)
	if err != nil {
		return 0, 0, domain.ErrInternalServer
	}
	withdrawn, err := s.withdrawalRepo.GetWithdrawalPoints(ctx, userID)
	if err != nil {
		return 0, 0, domain.ErrInternalServer
	}
	return current, withdrawn, nil
}
