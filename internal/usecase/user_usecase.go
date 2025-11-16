package usecase

import (
	"context"
	"database/sql"

	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/domain"
	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/repository"
)

type UserUseCase struct {
	userRepo repository.UserRepository
	prRepo   repository.PullRequestRepository
	teamRepo repository.TeamRepository
	db       *sql.DB
}

func NewUserUseCase(
	userRepo repository.UserRepository,
	prRepo repository.PullRequestRepository,
	teamRepo repository.TeamRepository,
	db *sql.DB) *UserUseCase {
	return &UserUseCase{
		userRepo: userRepo,
		prRepo:   prRepo,
		teamRepo: teamRepo,
		db:       db,
	}
}

// SetUserIsActive updates the isActive flag for the specified user.
//
// Returns:
//   - *domain.User: updated user with the new isActive value
//   - error: domain.ErrNotFound if user doesn't exist, or any database error
func (u *UserUseCase) SetUserIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error) {
	// Try to get user by ID
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// If exists - update isActive field in the domain and call repo method
	user.IsActive = isActive
	tx, _ := u.db.Begin()
	defer tx.Rollback()

	// Update user in database
	err = u.userRepo.Update(ctx, tx, user)
	if err != nil {
		return nil, err
	}

	teamName, err := u.teamRepo.GetTeamNameById(ctx, user.TeamID)
	if err != nil {
		return nil, err
	}

	// Commit
	if err = tx.Commit(); err != nil {
		return nil, err
	}

	user.TeamName = teamName

	return user, nil
}

// GetAssignedPRs gets all pull requests where the given user is assigned as a reviewer.
// Returns both OPEN and MERGED PRs.
//
// Returns:
//   - []*domain.PullRequest: slice of PRs where user is a reviewer (empty if no PRs found)
//   - error: domain.ErrNotFound if user does not exist, or any database error.
func (u *UserUseCase) GetAssignedPRs(ctx context.Context, userID string) ([]*domain.PullRequest, error) {
	prs, err := u.prRepo.GetPRsByReviewer(ctx, userID)
	if err != nil {
		return nil, err // err can be domain.NotFound if user does not exist
	}

	return prs, nil
}
