package entity

type Order struct {
	ID         string `json:"id" db:"id"`
	UserID     string `json:"user_id" db:"user_id"`
	Status     string `json:"status" db:"status"`
	UploadedAt string `json:"uploaded_at" db:"uploaded_at"`
	Number     int64  `json:"number" db:"number"`
}
