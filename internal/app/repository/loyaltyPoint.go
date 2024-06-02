package repository

import "context"

type LoyaltyPointRepository interface {
	GetCurrentPoints(ctx context.Context, userID int) (float64, error)
}
