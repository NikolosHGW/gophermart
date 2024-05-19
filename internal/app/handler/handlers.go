package handler

type Handlers struct {
	UserHandler       *UserHandler
	OrderHandler      *OrderHandler
	BalanceHandler    *BalanceHandler
	WithdrawalHandler *WithdrawalHandler
}
