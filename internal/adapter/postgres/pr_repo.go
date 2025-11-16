package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/domain"
)

// PullRequestRepository handles database operations for pull requests
type PullRequestRepository struct {
	db *sql.DB
}

// NewPullRequestRepository creates a new instance of PullRequestRepository
func NewPullRequestRepository(db *sql.DB) *PullRequestRepository {
	return &PullRequestRepository{db: db}
}

// Create inserts a new pull request into the database and sets createdAt time value to PR object.
// Returns domain ErrPRExists if PR with given ID is present in DB
func (p *PullRequestRepository) Create(ctx context.Context, tx *sql.Tx, pr *domain.PullRequest) error {
	query := `
			INSERT INTO pull_requests (id, name, author_id, status)
			VALUES ($1, $2, $3, $4)
			RETURNING created_at`
	err := tx.QueryRowContext(ctx, query, pr.ID, pr.Name, pr.AuthorID, pr.Status).Scan(&pr.CreatedAt)

	// if PR already exists - return domain.ErrPRExists
	if err != nil {
		if isUniqueViolationError(err) {
			return domain.ErrPRExists
		}
		return err
	}

	return nil
}

// Update modifies an existing pull request within a transaction.
// Typically used to change status to MERGED and set merged_at timestamp.
// Returns ErrNotFound if the PR doesn't exist.
func (p *PullRequestRepository) Update(ctx context.Context, tx *sql.Tx, pr *domain.PullRequest) error {
	query := `
			UPDATE pull_requests
			SET status = $1, merged_at = $2
			WHERE id = $3`

	var mergedAt interface{}
	if pr.MergedAt != nil {
		mergedAt = *pr.MergedAt
	} else {
		mergedAt = nil
	}

	res, err := tx.ExecContext(ctx, query, pr.Status, mergedAt, pr.ID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// GetByID retrieves a pull request by ID including all assigned reviewers.
// Returns ErrNotFound if the PR doesn't exist.
func (p *PullRequestRepository) GetByID(ctx context.Context, prID string) (*domain.PullRequest, error) {
	query := `
			SELECT id, name, author_id, status, created_at, merged_at
			FROM pull_requests
			WHERE id = $1`

	var pr domain.PullRequest
	var mergedAt sql.NullTime
	err := p.db.QueryRowContext(ctx, query, prID).Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &mergedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
	}

	if mergedAt.Valid {
		pr.MergedAt = &mergedAt.Time
	}

	pr.ReviewersIDs, err = p.getReviewerIDs(ctx, prID)
	if err != nil {
		return nil, err
	}

	return &pr, nil
}

// getReviewerIDs retrieves all reviewer IDs assigned to a PR.
func (p *PullRequestRepository) getReviewerIDs(ctx context.Context, prID string) ([]string, error) {
	query := `
			SELECT user_id
			FROM pr_reviewers
			WHERE pr_id = $1`

	rows, err := p.db.QueryContext(ctx, query, prID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []string{}, nil
		}
		return nil, err
	}

	var reviewerIDs []string
	for rows.Next() {
		var id string
		err := rows.Scan(&id)
		if err != nil {
			return nil, err
		}

		reviewerIDs = append(reviewerIDs, id)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return reviewerIDs, nil
}

// GetByIDForUpdate retrieves a PR with a row-level lock within a transaction.
// Used before updating to prevent race conditions.
// Returns ErrNotFound if the PR doesn't exist.
func (p *PullRequestRepository) GetByIDForUpdate(ctx context.Context, tx *sql.Tx, prID string) (*domain.PullRequest, error) {
	query := `
			SELECT id, name, author_id, status, created_at, merged_at
			FROM pull_requests
			WHERE id = $1
			FOR UPDATE
			`

	var pr domain.PullRequest
	var mergedAt sql.NullTime
	err := tx.QueryRowContext(ctx, query, prID).Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &mergedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
	}

	if mergedAt.Valid {
		pr.MergedAt = &mergedAt.Time
	}

	pr.ReviewersIDs, err = p.getReviewerIDsTx(ctx, tx, prID)
	if err != nil {
		return nil, err
	}

	return &pr, nil
}

// getReviewerIDsTx retrieves reviewer IDs within a transaction.
func (p *PullRequestRepository) getReviewerIDsTx(ctx context.Context, tx *sql.Tx, prID string) ([]string, error) {
	query := `
			SELECT user_id
			FROM pr_reviewers
			WHERE pr_id = $1`

	rows, err := tx.QueryContext(ctx, query, prID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []string{}, nil
		}
		return nil, err
	}

	var reviewerIDs []string
	for rows.Next() {
		var id string
		err := rows.Scan(&id)
		if err != nil {
			return nil, err
		}

		reviewerIDs = append(reviewerIDs, id)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return reviewerIDs, nil
}

// AddReviewer assigns a reviewer to a PR within a transaction.
// If the reviewer is already assigned, the operation is idempotent (no error on duplicate).
func (p *PullRequestRepository) AddReviewer(ctx context.Context, tx *sql.Tx, prID, userID string) error {
	query := `
			INSERT INTO pr_reviewers (pr_id, user_id)
			VALUES ($1, $2)
			ON CONFLICT (pr_id, user_id) DO NOTHING
			`

	_, err := tx.ExecContext(ctx, query, prID, userID)
	return err
}

// RemoveReviewer removes a reviewer from a PR within a transaction.
// Returns ErrNotAssigned if the user was not assigned as a reviewer.
func (p *PullRequestRepository) RemoveReviewer(ctx context.Context, tx *sql.Tx, prID, userID string) error {
	query := `
			DELETE FROM pr_reviewers
			WHERE pr_id = $1 AND user_id = $2
			`

	res, err := tx.ExecContext(ctx, query, prID, userID)
	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return domain.ErrNotAssigned
	}
	return err
}

// GetPRsByReviewer retrieves all pull requests assigned to a specific reviewer.
// Returns ErrNotFound if no PRs are found for the reviewer.
func (p *PullRequestRepository) GetPRsByReviewer(ctx context.Context, userID string) ([]*domain.PullRequest, error) {
	query := `
			SELECT id, name, author_id, status, created_at, merged_at
			FROM pull_requests as pr
			JOIN pr_reviewers as r ON r.pr_id = pr.id
			WHERE r.user_id = $1
			`

	rows, err := p.db.QueryContext(ctx, query, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	var prs []*domain.PullRequest
	for rows.Next() {
		var pr domain.PullRequest
		var mergedAt sql.NullTime
		err := rows.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &mergedAt)
		if err != nil {
			return nil, err
		}

		if mergedAt.Valid {
			pr.MergedAt = &mergedAt.Time
		}

		prs = append(prs, &pr)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return prs, nil
}
