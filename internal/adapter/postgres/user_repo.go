package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/domain"
	"go.uber.org/zap"
)

// UserRepository handles database operations for users
type UserRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewUserRepository creates a new instance of UserRepository
func NewUserRepository(db *sql.DB, logger *zap.Logger) *UserRepository {
	return &UserRepository{db: db, logger: logger}
}

// Create inserts a new user into the database.
func (u *UserRepository) Create(ctx context.Context, tx *sql.Tx, user *domain.User) error {
	query := `
			INSERT INTO users (id, username, is_active, team_id)
			VALUES ($1, $2, $3, $4)`
	_, err := tx.ExecContext(ctx, query, user.ID, user.Name, user.IsActive, user.TeamID)

	if err != nil {
		u.logger.Error("DB error on User insert",
			zap.Error(err),
			zap.String("team_id", user.ID))
		return err
	}

	return nil
}

// Update modifies an existing user's information.
func (u *UserRepository) Update(ctx context.Context, tx *sql.Tx, user *domain.User) error {
	query := `
			UPDATE users
			SET username = $1, is_active = $2, team_id = $3
			WHERE id = $4`
	res, err := tx.ExecContext(ctx, query, user.Name, user.IsActive, user.TeamID, user.ID)
	if err != nil {
		u.logger.Error("DB error on User update",
			zap.Error(err),
			zap.String("team_id", user.ID))
		return err
	}

	// If no such user - return ErrNotFound
	if rows, err := res.RowsAffected(); err != nil {
		return err
	} else {
		if rows == 0 {
			return domain.ErrNotFound
		}
	}

	return err
}

// GetByID retrieves a user by their ID.
// Returns ErrNotFound if the user doesn't exist.
func (u *UserRepository) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	query := `
			SELECT id, username, is_active, team_id
			FROM users
			WHERE id = $1`

	var user domain.User
	err := u.db.QueryRowContext(ctx, query, userID).Scan(&user.ID, &user.Name, &user.IsActive, &user.TeamID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			u.logger.Error("DB error on User select",
				zap.Error(err),
				zap.String("team_id", userID))
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

// GetActiveTeamMembersIDs returns IDs of all active users in a team, excluding the specified user.
// This is useful for selecting reviewer candidates (excluding the PR author).
func (u *UserRepository) GetActiveTeamMembersIDs(ctx context.Context, teamID int64, excludeUserID string) ([]string, error) {
	query := `
			SELECT id
			FROM users
			WHERE team_id = $1
				AND is_active = true
				AND id != $2
			`

	rows, err := u.db.QueryContext(ctx, query, teamID, excludeUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []string{}, nil
		}
		return nil, err
	}

	var memberIDs []string
	for rows.Next() {
		var userID string
		err := rows.Scan(&userID)
		if err != nil {
			return nil, err
		}

		memberIDs = append(memberIDs, userID)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return memberIDs, nil
}
