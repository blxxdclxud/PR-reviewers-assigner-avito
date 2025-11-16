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

func (r *StatsRepository) GetGeneralStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total teams
	var totalTeams int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM teams").Scan(&totalTeams)
	if err != nil {
		return nil, err
	}
	stats["total_teams"] = totalTeams

	// Total users
	var totalUsers int
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&totalUsers)
	if err != nil {
		return nil, err
	}
	stats["total_users"] = totalUsers

	// Total PRs
	var totalPRs int
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM pull_requests").Scan(&totalPRs)
	if err != nil {
		return nil, err
	}
	stats["total_prs"] = totalPRs

	// Open PRs
	var openPRs int
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM pull_requests WHERE status = 'OPEN'").Scan(&openPRs)
	if err != nil {
		return nil, err
	}
	stats["open_prs"] = openPRs

	// Merged PRs
	var mergedPRs int
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM pull_requests WHERE status = 'MERGED'").Scan(&mergedPRs)
	if err != nil {
		return nil, err
	}
	stats["merged_prs"] = mergedPRs

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
	defer rows.Close()

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
