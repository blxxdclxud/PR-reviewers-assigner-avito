package domain

type Team struct {
	ID      int64
	Name    string `json:"team_name"`
	Members []User `json:"members"`
}
