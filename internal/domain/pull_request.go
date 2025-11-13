package domain

import "time"

// PRStatus represents the state of a pull request
type PRStatus string

const (
	StatusOpen   = PRStatus("OPEN")
	StatusMerged = PRStatus("MERGED")
)

// PullRequest represents a code review request.
// A PR can have up to MaxReviewersAmount assigned reviewers.
type PullRequest struct {
	ID           string
	Name         string
	AuthorID     string
	Status       PRStatus
	ReviewersIDs []string
	CreatedAt    time.Time
	MergedAt     *time.Time // nil if PR is not merged yet
}

// MaxReviewersAmount defines the maximum number of reviewers that can be assigned to a PR.
const MaxReviewersAmount = 2
