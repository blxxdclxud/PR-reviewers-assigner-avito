package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestSetUserIsActive_Success tests successful user activation update
func TestSetUserIsActive_Success(t *testing.T) {
	mockUserRepo := new(UserRepoMock)
	mockPRRepo := new(PullRequestRepoMock)
	mockTeamRepo := new(TeamRepoMock)

	db, dbMock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	userID := "u1"
	expectedUser := &domain.User{
		ID:       userID,
		Name:     "Alice",
		IsActive: true,
		TeamID:   1,
	}

	dbMock.ExpectBegin()

	// Mock expectations
	mockUserRepo.On("GetByID", ctx, userID).Return(expectedUser, nil)
	mockUserRepo.On("Update", ctx, mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.ID == userID && u.IsActive == false
	})).Return(nil)

	dbMock.ExpectCommit()

	// Execute
	uc := NewUserUseCase(mockUserRepo, mockPRRepo, mockTeamRepo, db)
	result, err := uc.SetUserIsActive(ctx, userID, false)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, userID, result.ID)
	assert.False(t, result.IsActive) // changed to false

	require.NoError(t, dbMock.ExpectationsWereMet())
	mockUserRepo.AssertExpectations(t)
}

// TestSetUserIsActive_UserNotFound tests error when user doesn't exist
func TestSetUserIsActive_UserNotFound(t *testing.T) {
	mockUserRepo := new(UserRepoMock)
	mockPRRepo := new(PullRequestRepoMock)
	mockTeamRepo := new(TeamRepoMock)

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	userID := "nonexistent"

	// Mock expectations
	mockUserRepo.On("GetByID", ctx, userID).Return(nil, domain.ErrNotFound)

	// Execute
	uc := NewUserUseCase(mockUserRepo, mockPRRepo, mockTeamRepo, db)
	result, err := uc.SetUserIsActive(ctx, userID, true)

	// Assert
	assert.ErrorIs(t, err, domain.ErrNotFound)
	assert.Nil(t, result)
	mockUserRepo.AssertExpectations(t)
}

// TestSetUserIsActive_UpdateError tests error during update
func TestSetUserIsActive_UpdateError(t *testing.T) {
	mockUserRepo := new(UserRepoMock)
	mockPRRepo := new(PullRequestRepoMock)
	mockTeamRepo := new(TeamRepoMock)

	db, dbMock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	userID := "u1"
	user := &domain.User{ID: userID, Name: "Alice", IsActive: true, TeamID: 1}
	updateErr := errors.New("database error")

	dbMock.ExpectBegin()

	mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)
	mockUserRepo.On("Update", ctx, mock.Anything, mock.Anything).Return(updateErr)

	dbMock.ExpectRollback()

	// Execute
	uc := NewUserUseCase(mockUserRepo, mockPRRepo, mockTeamRepo, db)
	result, err := uc.SetUserIsActive(ctx, userID, false)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, updateErr, err)
	assert.Nil(t, result)

	require.NoError(t, dbMock.ExpectationsWereMet())
	mockUserRepo.AssertExpectations(t)
}

// TestGetAssignedPRs_Success tests successful retrieval of assigned PRs
func TestGetAssignedPRs_Success(t *testing.T) {
	mockUserRepo := new(UserRepoMock)
	mockPRRepo := new(PullRequestRepoMock)
	mockTeamRepo := new(TeamRepoMock)

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	userID := "u1"
	expectedPRs := []*domain.PullRequest{
		{ID: "pr-1", Name: "Feature A", AuthorID: "u2", Status: domain.StatusOpen},
		{ID: "pr-2", Name: "Feature B", AuthorID: "u3", Status: domain.StatusMerged},
	}

	// Mock expectations
	mockPRRepo.On("GetPRsByReviewer", ctx, userID).Return(expectedPRs, nil)

	// Execute
	uc := NewUserUseCase(mockUserRepo, mockPRRepo, mockTeamRepo, db)
	result, err := uc.GetAssignedPRs(ctx, userID)

	// Assert
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "pr-1", result[0].ID)
	assert.Equal(t, "pr-2", result[1].ID)
	mockPRRepo.AssertExpectations(t)
}

// TestGetAssignedPRs_NoPRs tests when user has no assigned PRs
func TestGetAssignedPRs_NoPRs(t *testing.T) {
	mockUserRepo := new(UserRepoMock)
	mockPRRepo := new(PullRequestRepoMock)
	mockTeamRepo := new(TeamRepoMock)

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	userID := "u1"

	// Mock expectations - empty list
	mockPRRepo.On("GetPRsByReviewer", ctx, userID).Return([]*domain.PullRequest{}, nil)

	// Execute
	uc := NewUserUseCase(mockUserRepo, mockPRRepo, mockTeamRepo, db)
	result, err := uc.GetAssignedPRs(ctx, userID)

	// Assert
	require.NoError(t, err)
	assert.Empty(t, result)
	mockPRRepo.AssertExpectations(t)
}

// TestGetAssignedPRs_RepositoryError tests error from repository
func TestGetAssignedPRs_RepositoryError(t *testing.T) {
	mockUserRepo := new(UserRepoMock)
	mockPRRepo := new(PullRequestRepoMock)
	mockTeamRepo := new(TeamRepoMock)

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	userID := "u1"
	repoErr := errors.New("database connection lost")

	// Mock expectations
	mockPRRepo.On("GetPRsByReviewer", ctx, userID).Return(nil, repoErr)

	// Execute
	uc := NewUserUseCase(mockUserRepo, mockPRRepo, mockTeamRepo, db)
	result, err := uc.GetAssignedPRs(ctx, userID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, repoErr, err)
	assert.Nil(t, result)
	mockPRRepo.AssertExpectations(t)
}
