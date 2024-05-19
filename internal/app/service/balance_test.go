package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockLoyaltyPointRepository struct {
	mock.Mock
}

func (_m *MockLoyaltyPointRepository) GetCurrentPoints(ctx context.Context, userID int) (float64, error) {
	ret := _m.Called(ctx, userID)
	return ret.Get(0).(float64), ret.Error(1)
}

type MockWithdrawalRepository struct {
	mock.Mock
}

func (_m *MockWithdrawalRepository) GetWithdrawalPoints(ctx context.Context, userID int) (float64, error) {
	ret := _m.Called(ctx, userID)
	return ret.Get(0).(float64), ret.Error(1)
}

func (_m *MockWithdrawalRepository) WithdrawFunds(
	ctx context.Context,
	userID int,
	orderNumber string,
	sum float64,
) error {
	return nil
}

func TestBalanceService_GetBalanceByUserID(t *testing.T) {
	mockLoyaltyPointRepo := new(MockLoyaltyPointRepository)
	mockWithdrawalRepo := new(MockWithdrawalRepository)

	mockLoyaltyPointRepo.On("GetCurrentPoints", context.Background(), 1).Return(500.5, nil)
	mockWithdrawalRepo.On("GetWithdrawalPoints", context.Background(), 1).Return(42.0, nil)

	logger, _ := zap.NewDevelopment()
	service := NewBalanceService(mockLoyaltyPointRepo, mockWithdrawalRepo, logger)

	current, withdrawn, err := service.GetBalanceByUserID(context.Background(), 1)

	assert.NoError(t, err)
	assert.Equal(t, 500.5, current)
	assert.Equal(t, 42.0, withdrawn)
}
