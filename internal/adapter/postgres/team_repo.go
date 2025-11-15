package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/domain"
)

// TeamRepository handles database operations for teams
type TeamRepository struct {
	db *sql.DB
}

// NewTeamRepository creates a new instance of TeamRepository
func NewTeamRepository(db *sql.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

// Create inserts a new team into the database.
// Returns ErrTeamExists if a team with the same name already exists.
func (t *TeamRepository) Create(ctx context.Context, tx *sql.Tx, team *domain.Team) error {
	query := "INSERT INTO teams (name) VALUES ($1) RETURNING id"

	err := tx.QueryRowContext(ctx, query, team.Name).Scan(&team.ID)
	if err != nil {
		if isUniqueViolationError(err) {
			return domain.ErrTeamExists
		}
		return err
	}
	return nil
}

// GetByName retrieves a team by name including all team members.
// Returns ErrNotFound if the team doesn't exist.
func (t *TeamRepository) GetByName(ctx context.Context, teamName string) (*domain.Team, error) {
	query := `
			SELECT id, name
			FROM teams 
			WHERE name = $1`

	var team domain.Team
	err := t.db.QueryRowContext(ctx, query, teamName).Scan(&team.ID, &team.Name)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	// Get all team members
	members, err := t.getTeamMembers(ctx, team.ID)
	if err != nil {
		return nil, err
	}

	team.Members = members

	return &team, nil
}

// getTeamMembers retrieves all members of a team by team ID.
func (t *TeamRepository) getTeamMembers(ctx context.Context, teamID int64) ([]domain.User, error) {
	query := `
			SELECT id, username, is_active, team_id 
			FROM users
			WHERE team_id = $1
			`

	rows, err := t.db.QueryContext(ctx, query, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		err := rows.Scan(&user.ID, &user.Name, &user.IsActive, &user.TeamID)
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// Exists checks if a team with the given ID exists.
// Returns true if the team exists, false otherwise.
func (t *TeamRepository) Exists(ctx context.Context, teamID string) (bool, error) {
	//TODO implement me
	panic("implement me")
}
