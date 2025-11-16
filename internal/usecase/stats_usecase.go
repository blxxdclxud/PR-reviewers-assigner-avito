package usecase

import (
	"context"

	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/domain"
	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/repository"
)

type StatsUseCase struct {
	statsRepo repository.StatsRepository
}

func NewStatsUseCase(statsRepo repository.StatsRepository) *StatsUseCase {
	return &StatsUseCase{
		statsRepo: statsRepo,
	}
}

func (u *StatsUseCase) GetStats(ctx context.Context) (*domain.Stats, error) {
	// Get general stats
	generalStats, err := u.statsRepo.GetGeneralStats(ctx)
	if err != nil {
		return nil, err
	}

	// Get reviewers
	topReviewers, err := u.statsRepo.GetReviewers(ctx)
	if err != nil {
		return nil, err
	}

	response := &domain.Stats{
		TotalTeams: generalStats["total_teams"].(int),
		TotalUsers: generalStats["total_users"].(int),
		TotalPRs:   generalStats["total_prs"].(int),
		OpenPRs:    generalStats["open_prs"].(int),
		MergedPRs:  generalStats["merged_prs"].(int),
		Reviewers:  topReviewers,
	}

	return response, nil
}
