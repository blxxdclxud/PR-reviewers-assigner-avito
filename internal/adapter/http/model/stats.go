package model

import "github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/domain"

type StatsResponse struct {
	TotalTeams int                      `json:"total_teams"`
	TotalUsers int                      `json:"total_users"`
	TotalPRs   int                      `json:"total_prs"`
	OpenPRs    int                      `json:"open_prs"`
	MergedPRs  int                      `json:"merged_prs"`
	Reviewers  []domain.UserReviewStats `json:"top_reviewers"`
}

func StatsFromDomain(stats *domain.Stats) StatsResponse {
	return StatsResponse{
		TotalTeams: stats.TotalTeams,
		TotalUsers: stats.TotalUsers,
		TotalPRs:   stats.TotalPRs,
		OpenPRs:    stats.OpenPRs,
		MergedPRs:  stats.MergedPRs,
		Reviewers:  stats.Reviewers,
	}
}
