package entity

type Order struct {
	Status     string `json:"status" db:"status"`
	UploadedAt string `json:"uploaded_at" db:"uploaded_at"`
	Number     string `json:"number" db:"number"`
}
