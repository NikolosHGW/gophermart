package service

import (
	"context"

	"github.com/NikolosHGW/gophermart/internal/app/repository"
	"github.com/NikolosHGW/gophermart/internal/domain"
	"github.com/NikolosHGW/gophermart/internal/domain/usecase"
	"go.uber.org/zap"
)

type OrderService struct {
	orderRepo repository.OrderRepository
	logger    *zap.Logger
}

func NewOrderService(orderRepo repository.OrderRepository, logger *zap.Logger) usecase.OrderUseCase {
	return &OrderService{
		orderRepo: orderRepo,
		logger:    logger,
	}
}

func (s *OrderService) ProcessOrder(ctx context.Context, userID int, orderNumber string) error {
	exists, err := s.orderRepo.OrderExistsForUser(ctx, userID, orderNumber)
	if err != nil {
		s.logger.Info("ошибка при проверке существования заказа", zap.Error(err))
		return domain.ErrInternalServer
	}
	if exists {
		return domain.ErrOrderAlreadyUploadedForThisUser
	}

	claimed, err := s.orderRepo.OrderClaimedByAnotherUser(ctx, userID, orderNumber)
	if err != nil {
		s.logger.Info("ошибка при проверке заказа у других пользователей", zap.Error(err))
		return domain.ErrInternalServer
	}
	if claimed {
		return domain.ErrOrderAlreadyUploadedByAnotherUser
	}

	err = s.orderRepo.AddOrder(ctx, userID, orderNumber)
	if err != nil {
		s.logger.Info("ошибка при добавлении заказа", zap.Error(err))
		return domain.ErrInternalServer
	}

	return nil
}
