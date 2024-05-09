package entity

type User struct {
	Login    string `json:"login" db:"login"`
	Password string `json:"password" db:"password"`
	ID       int    `json:"id" db:"id"`
}
