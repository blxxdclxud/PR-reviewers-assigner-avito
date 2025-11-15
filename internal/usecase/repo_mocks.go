package usecase

import (
	"context"
	"database/sql"

	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/domain"
	"github.com/stretchr/testify/mock"
)

type TeamRepoMock struct {
	mock.Mock
}

func (m *TeamRepoMock) Create(ctx context.Context, tx *sql.Tx, team *domain.Team) error {
	args := m.Called(ctx, tx, team)
	return args.Error(0)
}

func (m *TeamRepoMock) GetByName(ctx context.Context, teamName string) (*domain.Team, error) {
	args := m.Called(ctx, teamName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Team), args.Error(1)
}

type UserRepoMock struct {
	mock.Mock
}

func (m *UserRepoMock) Create(ctx context.Context, tx *sql.Tx, user *domain.User) error {
	args := m.Called(ctx, tx, user)
	return args.Error(0)
}

func (m *UserRepoMock) Update(ctx context.Context, tx *sql.Tx, user *domain.User) error {
	args := m.Called(ctx, tx, user)
	return args.Error(0)
}

func (m *UserRepoMock) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *UserRepoMock) GetActiveTeamMembersIDs(ctx context.Context, teamID int64, excludeUserID string) ([]string, error) {
	args := m.Called(ctx, teamID, excludeUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

type PullRequestRepoMock struct {
	mock.Mock
}

func (m *PullRequestRepoMock) Create(ctx context.Context, tx *sql.Tx, pr *domain.PullRequest) error {
	args := m.Called(ctx, tx, pr)
	return args.Error(0)
}

func (m *PullRequestRepoMock) Update(ctx context.Context, tx *sql.Tx, pr *domain.PullRequest) error {
	args := m.Called(ctx, tx, pr)
	return args.Error(0)
}

func (m *PullRequestRepoMock) GetByID(ctx context.Context, prID string) (*domain.PullRequest, error) {
	args := m.Called(ctx, prID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PullRequest), args.Error(1)
}

func (m *PullRequestRepoMock) GetByIDForUpdate(ctx context.Context, tx *sql.Tx, prID string) (*domain.PullRequest, error) {
	args := m.Called(ctx, tx, prID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PullRequest), args.Error(1)
}

func (m *PullRequestRepoMock) AddReviewer(ctx context.Context, tx *sql.Tx, prID, userID string) error {
	args := m.Called(ctx, tx, prID, userID)
	return args.Error(0)
}

func (m *PullRequestRepoMock) RemoveReviewer(ctx context.Context, tx *sql.Tx, prID, userID string) error {
	args := m.Called(ctx, tx, prID, userID)
	return args.Error(0)
}

func (m *PullRequestRepoMock) GetPRsByReviewer(ctx context.Context, userID string) ([]*domain.PullRequest, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.PullRequest), args.Error(1)
}
