package postgres

import (
	"context"
	"database/sql"

	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/domain"
)

type StatsRepository struct {
	db *sql.DB
}

func NewStatsRepository(db *sql.DB) *StatsRepository {
	return &StatsRepository{db: db}
}

func (r *StatsRepository) GetGeneralStats(ctx context.Context) (*domain.Stats, error) {
	stats := &domain.Stats{}

	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM teams").Scan(&stats.TotalTeams)
	if err != nil {
		return nil, err
	}

	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&stats.TotalUsers)
	if err != nil {
		return nil, err
	}

	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM pull_requests").Scan(&stats.TotalPRs)
	if err != nil {
		return nil, err
	}

	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM pull_requests WHERE status = 'OPEN'").Scan(&stats.OpenPRs)
	if err != nil {
		return nil, err
	}

	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM pull_requests WHERE status = 'MERGED'").Scan(&stats.MergedPRs)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (r *StatsRepository) GetReviewers(ctx context.Context) ([]domain.UserReviewStats, error) {
	query := `
        SELECT u.id, u.username, COUNT(pr.pr_id) as review_count
        FROM users u
        LEFT JOIN pr_reviewers pr ON u.id = pr.user_id
        GROUP BY u.id, u.username
        HAVING COUNT(pr.pr_id) > 0
        ORDER BY review_count DESC
    `

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	var reviewers []domain.UserReviewStats
	for rows.Next() {
		var r domain.UserReviewStats
		if err := rows.Scan(&r.UserID, &r.Username, &r.ReviewCount); err != nil {
			return nil, err
		}
		reviewers = append(reviewers, r)
	}

	return reviewers, rows.Err()
}
