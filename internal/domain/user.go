package domain

type User struct {
	ID       string `json:"user_id"`
	Name     string `json:"username"`
	TeamID   int64
	IsActive bool `json:"is_active"`
}
