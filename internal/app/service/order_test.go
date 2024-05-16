package service

import (
	"context"
	"testing"

	"github.com/NikolosHGW/gophermart/internal/domain"
	"github.com/NikolosHGW/gophermart/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
)

type OrderRepository struct {
	mock.Mock
}

func (_m *OrderRepository) OrderExistsForUser(ctx context.Context, userID int, orderNumber string) (bool, error) {
	ret := _m.Called(ctx, userID, orderNumber)
	return ret.Get(0).(bool), ret.Error(1)
}

func (_m *OrderRepository) OrderClaimedByAnotherUser(
	ctx context.Context,
	userID int,
	orderNumber string,
) (bool, error) {
	ret := _m.Called(ctx, userID, orderNumber)
	return ret.Get(0).(bool), ret.Error(1)
}

func (_m *OrderRepository) AddOrder(ctx context.Context, userID int, orderNumber string) error {
	ret := _m.Called(ctx, userID, orderNumber)
	return ret.Error(0)
}

func (_m *OrderRepository) GetUserOrdersByID(ctx context.Context, userID int) ([]entity.Order, error) {
	return []entity.Order{}, nil
}

func TestOrderService_ProcessOrder(t *testing.T) {
	logger := zaptest.NewLogger(t)
	defer logger.Sync()

	ctx := context.Background()
	userID := 1
	orderNumber := "1234567890"

	tests := []struct {
		name          string
		mockSetup     func(*OrderRepository)
		expectedError error
	}{
		{
			name: "Положительный тест: загрузка нового номера",
			mockSetup: func(orderRepo *OrderRepository) {
				orderRepo.On("OrderExistsForUser", ctx, userID, orderNumber).Return(false, nil)
				orderRepo.On("OrderClaimedByAnotherUser", ctx, userID, orderNumber).Return(false, nil)
				orderRepo.On("AddOrder", ctx, userID, orderNumber).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "Отрицательный тест: номер уже существует у этого пользователя",
			mockSetup: func(orderRepo *OrderRepository) {
				orderRepo.On("OrderExistsForUser", ctx, userID, orderNumber).Return(true, nil)
			},
			expectedError: domain.ErrOrderAlreadyUploadedForThisUser,
		},
		{
			name: "Отрицательный тест: номер уже загружен для другого пользователя",
			mockSetup: func(orderRepo *OrderRepository) {
				orderRepo.On("OrderExistsForUser", ctx, userID, orderNumber).Return(false, nil)
				orderRepo.On("OrderClaimedByAnotherUser", ctx, userID, orderNumber).Return(true, nil)
			},
			expectedError: domain.ErrOrderAlreadyUploadedByAnotherUser,
		},
		{
			name: "Отрицательный тест: ошибка сервера",
			mockSetup: func(orderRepo *OrderRepository) {
				orderRepo.On("OrderExistsForUser", ctx, userID, orderNumber).Return(false, nil)
				orderRepo.On("OrderClaimedByAnotherUser", ctx, userID, orderNumber).Return(false, nil)
				orderRepo.On("AddOrder", ctx, userID, orderNumber).Return(assert.AnError)
			},
			expectedError: domain.ErrInternalServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orderRepo := new(OrderRepository)
			tt.mockSetup(orderRepo)
			orderService := NewOrderService(orderRepo, logger)
			err := orderService.ProcessOrder(ctx, userID, orderNumber)
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
