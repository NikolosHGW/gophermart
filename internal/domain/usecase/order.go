package usecase

import "context"

type OrderUseCase interface {
	ProcessOrder(ctx context.Context, userID int, orderNumber string) error
}
