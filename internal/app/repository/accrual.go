package repository

import (
	"context"

	"github.com/NikolosHGW/gophermart/internal/domain/entity"
)

type AccrualRepository interface {
	GetNonFinalOrders(ctx context.Context) ([]entity.Order, error)
	UpdateAccrual(ctx context.Context, orderNumber string, accrual float64, status string) error
}
