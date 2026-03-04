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
	stats, err := u.statsRepo.GetGeneralStats(ctx)
	if err != nil {
		return nil, err
	}

	topReviewers, err := u.statsRepo.GetReviewers(ctx)
	if err != nil {
		return nil, err
	}

	stats.Reviewers = topReviewers

	return stats, nil
}
