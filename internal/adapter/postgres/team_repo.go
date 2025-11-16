package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/domain"
	"go.uber.org/zap"
)

// TeamRepository handles database operations for teams
type TeamRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewTeamRepository creates a new instance of TeamRepository
func NewTeamRepository(db *sql.DB, logger *zap.Logger) *TeamRepository {
	return &TeamRepository{db: db, logger: logger}
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
		t.logger.Error("DB error on Team insert",
			zap.Error(err),
			zap.Int64("team_id", team.ID))
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
		t.logger.Error("DB error on Team select",
			zap.Error(err),
			zap.Int64("team_id", team.ID))
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

func (t *TeamRepository) GetTeamNameByID(ctx context.Context, teamID int64) (string, error) {
	query := `
			SELECT name
			FROM teams 
			WHERE id = $1`

	var teamName string
	err := t.db.QueryRowContext(ctx, query, teamID).Scan(&teamName)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", domain.ErrNotFound
		}
		t.logger.Error("DB error on Team select",
			zap.Error(err),
			zap.Int64("team_id", teamID))
		return "", err
	}

	return teamName, nil
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
	defer rows.Close() //nolint:errcheck

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
