package repository

import "context"

type OrderRepository interface {
	OrderExistsForUser(ctx context.Context, userID int, orderNumber string) (bool, error)
	OrderClaimedByAnotherUser(ctx context.Context, userID int, orderNumber string) (bool, error)
	AddOrder(ctx context.Context, userID int, orderNumber string) error
}
