package models

type User struct {
	ID       int    `db:"id" json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Interaction struct {
	ID           int    `json:"id"`
	UserID       int    `json:"user_id"`
	Input        string `json:"input"`
	Response     string `json:"response"`
	UserDuration int    `json:"user_duration"`
}
