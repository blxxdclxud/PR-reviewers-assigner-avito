package domain

// User represents a team member who can author or review pull requests
type User struct {
	ID       string
	Name     string
	TeamID   int64
	IsActive bool // only active users can be assigned as reviewers
}
