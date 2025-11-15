package usecase

import (
	"context"
	"math/rand"
	"slices"

	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/domain"
)

// getReviewersToAssign selects random active team members to be assigned as reviewers.
// Excludes users provided in excludeUserIDs (used to exclude the author and any user IDs).
// Returns up to domain.MaxReviewersAmount reviewers.
//
// Returns:
//   - []string: slice of user IDs to be assigned as reviewers (can be empty if no candidates)
//   - error: domain.ErrNotFound if author doesn't exist, or any database error
func (u *PRUseCase) getReviewersToAssign(ctx context.Context, authorID string, excludeUserIDs ...string) ([]string, error) {
	// Get author to extract his team ID
	author, err := u.userRepo.GetByID(ctx, authorID)
	if err != nil {
		return nil, err
	}

	// Get all active team members
	candidates, err := u.userRepo.GetActiveTeamMembersIDs(ctx, author.TeamID, authorID)
	if err != nil {
		return nil, err
	}

	// remove candidates that must be excluded
	for _, id := range excludeUserIDs {
		// remove current reviewer
		currIdx := slices.Index(candidates, id)
		if currIdx != -1 { // if it is in slice
			candidates = append(candidates[:currIdx], candidates[currIdx+1:]...)
		}
	}

	// Select random reviewers (up to domain.MaxReviewersAmount)
	reviewers := selectRandomReviewers(candidates, domain.MaxReviewersAmount)

	return reviewers, nil
}

// selectRandomReviewers randomly selects up to reviewersAmount users from candidates.
// If there are fewer candidates than requested, returns all candidates.
//
// Returns:
//   - []string: randomly shuffled slice of candidates' user IDs
func selectRandomReviewers(candidates []string, reviewersAmount int) []string {
	// If the amount of available candidates is less than needed amount, just select all of them
	if len(candidates) <= reviewersAmount {
		return candidates
	}

	// Shuffle candidates slice and return first `reviewersAmount` elements
	shuffled := make([]string, len(candidates))
	copy(shuffled, candidates)

	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	return shuffled[:reviewersAmount]
}
