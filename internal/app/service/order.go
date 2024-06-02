package service

import (
	"context"

	"github.com/NikolosHGW/gophermart/internal/app/repository"
	"github.com/NikolosHGW/gophermart/internal/domain"
	"github.com/NikolosHGW/gophermart/internal/domain/entity"
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

func (s *OrderService) GetUserOrdersByID(ctx context.Context, userID int) ([]entity.Order, error) {
	orders, err := s.orderRepo.GetUserOrdersByID(ctx, userID)
	if err != nil {
		return orders, domain.ErrInternalServer
	}
	return orders, nil
}

func (s *OrderService) OrderExists(ctx context.Context, userID int, orderNumber string) (bool, error) {
	exists, err := s.orderRepo.OrderExistsForUser(ctx, userID, orderNumber)
	if err != nil {
		return exists, domain.ErrInternalServer
	}
	return exists, nil
}
