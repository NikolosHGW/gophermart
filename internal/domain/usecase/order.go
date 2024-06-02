package usecase

import (
	"context"

	"github.com/NikolosHGW/gophermart/internal/domain/entity"
)

type OrderUseCase interface {
	ProcessOrder(ctx context.Context, userID int, orderNumber string) error
	GetUserOrdersByID(ctx context.Context, userID int) ([]entity.Order, error)
	OrderExists(ctx context.Context, userID int, orderNumber string) (bool, error)
}
