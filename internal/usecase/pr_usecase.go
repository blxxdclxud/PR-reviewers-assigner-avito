package usecase

import (
	"context"
	"database/sql"
	"slices"
	"time"

	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/domain"
	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/repository"
)

type PRUseCase struct {
	userRepo repository.UserRepository
	prRepo   repository.PullRequestRepository
	db       *sql.DB
}

func NewPRUseCase(
	userRepo repository.UserRepository,
	prRepo repository.PullRequestRepository,
	db *sql.DB) *PRUseCase {
	return &PRUseCase{
		userRepo: userRepo,
		prRepo:   prRepo,
		db:       db,
	}
}

// CreatePRAndSetReviewers creates a new pull request with the given details and automatically
// assigns up to 2 random reviewers from the author's team (excluding the author).
// PR is created with status OPEN.
//
// Returns:
//   - *domain.PullRequest: created PR with assigned reviewers in ReviewersIDs field
//   - error: domain.ErrNotFound if author doesn't exist, domain.ErrPRExists if PR ID already exists,
//     or any database error
func (u *PRUseCase) CreatePRAndSetReviewers(ctx context.Context, pr domain.PullRequest) (*domain.PullRequest, error) {
	// set PR status to open
	pr.Status = domain.StatusOpen

	// Select reviewers
	reviewers, err := u.getReviewersToAssign(ctx, pr.AuthorID)
	if err != nil { // err can be domain.ErrNotFound if author or his team do not exist
		return nil, err
	}

	// Start transaction
	tx, err := u.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Create PR (sets createdAt time value implicitly). Can return domain.ErrPRExist
	err = u.prRepo.Create(ctx, tx, &pr)
	if err != nil {
		return nil, err
	}

	// Assign each reviewer in repo
	for _, revID := range reviewers {
		err = u.prRepo.AddReviewer(ctx, tx, pr.ID, revID)
		if err != nil {
			return nil, err
		}
	}

	// Commit changes
	if err = tx.Commit(); err != nil {
		return nil, err
	}

	// Update PR model with assigned reviewers
	pr.ReviewersIDs = reviewers

	return &pr, nil
}

// MergePR marks the given pull request as MERGED. This operation is idempotent -
// if PR is already merged, it returns the PR without modifications.
//
// Returns:
//   - *domain.PullRequest: PR with status MERGED and mergedAt timestamp set
//   - error: domain.ErrNotFound if PR doesn't exist, or any database error
func (u *PRUseCase) MergePR(ctx context.Context, prID string) (*domain.PullRequest, error) {
	tx, err := u.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Get PR object from DB
	pr, err := u.prRepo.GetByIDForUpdate(ctx, tx, prID)
	if err != nil {
		return nil, err // err can be domain.ErrNotFound
	}

	// Set PR status as Merged.
	// Operation in idempotent - even if it is already merged, keep it so
	// But update in DB only if it is not merged yet
	if pr.Status == domain.StatusOpen {
		pr.Status = domain.StatusMerged
		now := time.Now()
		pr.MergedAt = &now

		// Update PR in DB
		err = u.prRepo.Update(ctx, tx, pr)
		if err != nil {
			return nil, err
		}
	}

	// Commit changes
	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return pr, nil
}

// ReassignReviewer replaces an existing reviewer with a new random reviewer from the same team.
// The new reviewer must be active, not the PR author, and not already assigned to the PR.
//
// Returns:
//   - *domain.PullRequest: PR with updated list of assigned reviewers
//   - string: user_id of the newly assigned reviewer
//   - error: domain.ErrNotFound if PR doesn't exist, domain.ErrNotAssigned if oldReviewerID
//     is not assigned to the PR, domain.ErrPRMerged if PR is already merged,
//     domain.ErrNoCandidate if no suitable replacement found in the team, or any database error
func (u *PRUseCase) ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (*domain.PullRequest, string, error) {
	// Get PR
	pr, err := u.prRepo.GetByID(ctx, prID)
	if err != nil {
		return nil, "", err // err can be domain.ErrNotFound
	}

	// Check rules before reassigning:
	// - oldReviewerID is not real reviewer for given PR
	oldIdx := slices.Index(pr.ReviewersIDs, oldReviewerID)
	if oldIdx == -1 {
		return nil, "", domain.ErrNotAssigned
	}

	// - PR with status "MERGED" cannot be modified
	if pr.Status == domain.StatusMerged {
		return nil, "", domain.ErrPRMerged
	}

	// - Cannot reassign reviewer if there are no other candidates
	// Get candidates, excluding all current reviewers
	reviewers, err := u.getReviewersToAssign(ctx, pr.AuthorID, pr.ReviewersIDs...)
	if err != nil { // err can be domain.ErrNotFound if author or his team do not exist
		return nil, "", err
	}

	// Assert that there are any candidate
	if len(reviewers) == 0 {
		return nil, "", domain.ErrNoCandidate
	}

	// select first reviewer, since slice is already shuffled
	newReviewerID := reviewers[0]

	tx, err := u.db.Begin()
	if err != nil {
		return nil, "", err
	}
	defer tx.Rollback()

	// Remove old reviewer
	err = u.prRepo.RemoveReviewer(ctx, tx, prID, oldReviewerID)
	if err != nil {
		return nil, "", err
	}

	// Assign new reviewer
	err = u.prRepo.AddReviewer(ctx, tx, prID, newReviewerID)
	if err != nil {
		return nil, "", err
	}

	// Commit changes
	if err = tx.Commit(); err != nil {
		return nil, "", err
	}

	// update PR model
	pr.ReviewersIDs = append(pr.ReviewersIDs[:oldIdx], pr.ReviewersIDs[oldIdx+1:]...)
	pr.ReviewersIDs = append(pr.ReviewersIDs, newReviewerID)

	return pr, newReviewerID, nil
}
