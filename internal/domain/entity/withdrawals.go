package entity

type Withdrawal struct {
	ID          string  `db:"id"`
	UserID      string  `db:"user_id"`
	OrderID     string  `db:"order_id"`
	ProcessedAt string  `db:"processed_at"`
	Sum         float64 `db:"sum"`
}
