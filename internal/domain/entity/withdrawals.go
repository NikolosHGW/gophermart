package entity

type Withdrawal struct {
	Order       string  `db:"number" json:"order"`
	ProcessedAt string  `db:"processed_at" json:"processed_at"`
	Sum         float64 `db:"sum" json:"sum"`
}
