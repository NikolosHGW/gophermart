package repository

import (
	"context"

	"github.com/NikolosHGW/gophermart/internal/domain/entity"
)

type OrderRepository interface {
	OrderExistsForUser(ctx context.Context, userID int, orderNumber string) (bool, error)
	OrderClaimedByAnotherUser(ctx context.Context, userID int, orderNumber string) (bool, error)
	AddOrder(ctx context.Context, userID int, orderNumber string) error
	GetUserOrdersByID(ctx context.Context, userID int) ([]entity.Order, error)
}
