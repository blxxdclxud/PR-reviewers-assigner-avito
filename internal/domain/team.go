package domain

// Team represents a development team with its members.
type Team struct {
	ID      int64
	Name    string
	Members []User
}
