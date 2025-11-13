package repository

import (
	"context"
	"database/sql"

	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/domain"
)

// TeamRepository defines operations for managing teams
type TeamRepository interface {
	Create(ctx context.Context, team *domain.Team) error
	GetByName(ctx context.Context, teamName string) (*domain.Team, error)
	Exists(ctx context.Context, teamID string) (bool, error)
}

// UserRepository defines operations for managing users
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, userID string) (*domain.User, error)
	GetActiveTeamMembersIDs(ctx context.Context, teamID int64, excludeUserID string) ([]string, error)
}

// PullRequestRepository defines operations for managing pull requests and reviewers
type PullRequestRepository interface {
	Create(ctx context.Context, pr *domain.PullRequest) error
	Update(ctx context.Context, tx *sql.Tx, pr *domain.PullRequest) error
	GetByID(ctx context.Context, prID string) (*domain.PullRequest, error)
	GetByIDForUpdate(ctx context.Context, tx *sql.Tx, prID string) (*domain.PullRequest, error)
	AddReviewer(ctx context.Context, tx *sql.Tx, prID, userID string) error
	RemoveReviewer(ctx context.Context, tx *sql.Tx, prID, userID string) error
	GetPRsByReviewer(ctx context.Context, userID string) ([]*domain.PullRequest, error)
}
