package entity

type LoyaltyPoints struct {
	ID           string  `db:"id"`
	UserID       string  `db:"user_id"`
	AccruedPoint float64 `db:"accrued_point"`
	SpentPoint   float64 `db:"spent_point"`
}
