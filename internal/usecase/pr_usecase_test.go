package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/domain"
)

// ---- CreatePRAndSetReviewers tests ----

func TestPRUseCase_CreatePRAndSetReviewers_Success_TwoReviewers(t *testing.T) {
	mockUserRepo := new(UserRepoMock)
	mockPRRepo := new(PullRequestRepoMock)

	db, dbMock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	pr := domain.PullRequest{
		ID:       "pr-1001",
		Name:     "Add feature",
		AuthorID: "u1",
	}

	author := &domain.User{ID: "u1", Name: "Alice", TeamID: 1}
	candidates := []string{"u2", "u3", "u4"} // 3 candidates, will select 2

	dbMock.ExpectBegin()

	// Mock expectations
	mockUserRepo.On("GetByID", ctx, "u1").Return(author, nil)
	mockUserRepo.On("GetActiveTeamMembersIDs", ctx, int64(1), "u1").Return(candidates, nil)

	mockPRRepo.On("Create", ctx, mock.Anything, mock.MatchedBy(func(p *domain.PullRequest) bool {
		return p.ID == "pr-1001" && p.Status == domain.StatusOpen
	})).Return(nil)

	// Expect 2 reviewers to be added
	mockPRRepo.On("AddReviewer", ctx, mock.Anything, "pr-1001", mock.AnythingOfType("string")).Return(nil).Twice()

	dbMock.ExpectCommit()

	// Execute
	uc := NewPRUseCase(mockUserRepo, mockPRRepo, db)
	result, err := uc.CreatePRAndSetReviewers(ctx, pr)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "pr-1001", result.ID)
	assert.Equal(t, domain.StatusOpen, result.Status)
	assert.Len(t, result.ReviewersIDs, 2) // must be 2 reviewers

	require.NoError(t, dbMock.ExpectationsWereMet())
	mockUserRepo.AssertExpectations(t)
	mockPRRepo.AssertExpectations(t)
}

func TestPRUseCase_CreatePRAndSetReviewers_Success_OneReviewer(t *testing.T) {
	mockUserRepo := new(UserRepoMock)
	mockPRRepo := new(PullRequestRepoMock)

	db, dbMock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	pr := domain.PullRequest{
		ID:       "pr-1002",
		Name:     "Fix bug",
		AuthorID: "u1",
	}

	author := &domain.User{ID: "u1", Name: "Alice", TeamID: 1}
	candidates := []string{"u2"} // only 1 candidate

	dbMock.ExpectBegin()

	mockUserRepo.On("GetByID", ctx, "u1").Return(author, nil)
	mockUserRepo.On("GetActiveTeamMembersIDs", ctx, int64(1), "u1").Return(candidates, nil)

	mockPRRepo.On("Create", ctx, mock.Anything, mock.Anything).Return(nil)
	mockPRRepo.On("AddReviewer", ctx, mock.Anything, "pr-1002", "u2").Return(nil)

	dbMock.ExpectCommit()

	// Execute
	uc := NewPRUseCase(mockUserRepo, mockPRRepo, db)
	result, err := uc.CreatePRAndSetReviewers(ctx, pr)

	// Assert
	require.NoError(t, err)
	assert.Len(t, result.ReviewersIDs, 1) // only 1 reviewer assigned

	require.NoError(t, dbMock.ExpectationsWereMet())
	mockUserRepo.AssertExpectations(t)
	mockPRRepo.AssertExpectations(t)
}

func TestPRUseCase_CreatePRAndSetReviewers_Success_NoReviewers(t *testing.T) {
	mockUserRepo := new(UserRepoMock)
	mockPRRepo := new(PullRequestRepoMock)

	db, dbMock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	pr := domain.PullRequest{
		ID:       "pr-1003",
		Name:     "Solo work",
		AuthorID: "u1",
	}

	author := &domain.User{ID: "u1", Name: "Alice", TeamID: 1}
	var candidates []string // only author if the team

	dbMock.ExpectBegin()

	mockUserRepo.On("GetByID", ctx, "u1").Return(author, nil)
	mockUserRepo.On("GetActiveTeamMembersIDs", ctx, int64(1), "u1").Return(candidates, nil)

	mockPRRepo.On("Create", ctx, mock.Anything, mock.Anything).Return(nil)

	dbMock.ExpectCommit()

	// Execute
	uc := NewPRUseCase(mockUserRepo, mockPRRepo, db)
	result, err := uc.CreatePRAndSetReviewers(ctx, pr)

	// Assert
	require.NoError(t, err)
	assert.Empty(t, result.ReviewersIDs) // no reviewers

	require.NoError(t, dbMock.ExpectationsWereMet())
	mockUserRepo.AssertExpectations(t)
	mockPRRepo.AssertExpectations(t)
}

func TestPRUseCase_CreatePRAndSetReviewers_AuthorNotFound(t *testing.T) {
	mockUserRepo := new(UserRepoMock)
	mockPRRepo := new(PullRequestRepoMock)

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	pr := domain.PullRequest{
		ID:       "pr-1004",
		Name:     "Feature",
		AuthorID: "nonexistent",
	}

	mockUserRepo.On("GetByID", ctx, "nonexistent").Return(nil, domain.ErrNotFound)

	// Execute
	uc := NewPRUseCase(mockUserRepo, mockPRRepo, db)
	result, err := uc.CreatePRAndSetReviewers(ctx, pr)

	// Assert
	assert.ErrorIs(t, err, domain.ErrNotFound)
	assert.Nil(t, result)

	mockUserRepo.AssertExpectations(t)
	mockPRRepo.AssertNotCalled(t, "Create")
}

func TestPRUseCase_CreatePRAndSetReviewers_PRAlreadyExists(t *testing.T) {
	mockUserRepo := new(UserRepoMock)
	mockPRRepo := new(PullRequestRepoMock)

	db, dbMock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	pr := domain.PullRequest{
		ID:       "pr-1001",
		Name:     "Duplicate",
		AuthorID: "u1",
	}

	author := &domain.User{ID: "u1", Name: "Alice", TeamID: 1}
	candidates := []string{"u2"}

	dbMock.ExpectBegin()

	mockUserRepo.On("GetByID", ctx, "u1").Return(author, nil)
	mockUserRepo.On("GetActiveTeamMembersIDs", ctx, int64(1), "u1").Return(candidates, nil)
	mockPRRepo.On("Create", ctx, mock.Anything, mock.Anything).Return(domain.ErrPRExists)

	dbMock.ExpectRollback()

	// Execute
	uc := NewPRUseCase(mockUserRepo, mockPRRepo, db)
	result, err := uc.CreatePRAndSetReviewers(ctx, pr)

	// Assert
	assert.ErrorIs(t, err, domain.ErrPRExists)
	assert.Nil(t, result)

	require.NoError(t, dbMock.ExpectationsWereMet())
	mockUserRepo.AssertExpectations(t)
	mockPRRepo.AssertExpectations(t)
}

func TestPRUseCase_CreatePRAndSetReviewers_AddReviewerError(t *testing.T) {
	mockUserRepo := new(UserRepoMock)
	mockPRRepo := new(PullRequestRepoMock)

	db, dbMock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	pr := domain.PullRequest{
		ID:       "pr-1005",
		Name:     "Feature",
		AuthorID: "u1",
	}

	author := &domain.User{ID: "u1", Name: "Alice", TeamID: 1}
	candidates := []string{"u2"}

	dbMock.ExpectBegin()

	mockUserRepo.On("GetByID", ctx, "u1").Return(author, nil)
	mockUserRepo.On("GetActiveTeamMembersIDs", ctx, int64(1), "u1").Return(candidates, nil)
	mockPRRepo.On("Create", ctx, mock.Anything, mock.Anything).Return(nil)
	mockPRRepo.On("AddReviewer", ctx, mock.Anything, "pr-1005", "u2").Return(errors.New("db error"))

	dbMock.ExpectRollback()

	// Execute
	uc := NewPRUseCase(mockUserRepo, mockPRRepo, db)
	result, err := uc.CreatePRAndSetReviewers(ctx, pr)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	require.NoError(t, dbMock.ExpectationsWereMet())
	mockUserRepo.AssertExpectations(t)
	mockPRRepo.AssertExpectations(t)
}

// ---- MergePR tests ----

func TestPRUseCase_MergePR_Success(t *testing.T) {
	mockUserRepo := new(UserRepoMock)
	mockPRRepo := new(PullRequestRepoMock)

	db, dbMock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	prID := "pr-1001"

	pr := &domain.PullRequest{
		ID:           prID,
		Name:         "Feature",
		AuthorID:     "u1",
		Status:       domain.StatusOpen,
		ReviewersIDs: []string{"u2", "u3"},
	}

	dbMock.ExpectBegin()

	mockPRRepo.On("GetByIDForUpdate", ctx, mock.Anything, prID).Return(pr, nil)
	mockPRRepo.On("Update", ctx, mock.Anything, mock.MatchedBy(func(p *domain.PullRequest) bool {
		return p.ID == prID && p.Status == domain.StatusMerged
	})).Return(nil)

	dbMock.ExpectCommit()

	// Execute
	uc := NewPRUseCase(mockUserRepo, mockPRRepo, db)
	result, err := uc.MergePR(ctx, prID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, domain.StatusMerged, result.Status)

	require.NoError(t, dbMock.ExpectationsWereMet())
	mockPRRepo.AssertExpectations(t)
}

func TestPRUseCase_MergePR_Idempotent(t *testing.T) {
	mockUserRepo := new(UserRepoMock)
	mockPRRepo := new(PullRequestRepoMock)

	db, mockDb, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	prID := "pr-1001"

	pr := &domain.PullRequest{
		ID:           prID,
		Status:       domain.StatusMerged,
		ReviewersIDs: []string{"u2"},
	}

	mockPRRepo.On("GetByIDForUpdate", ctx, mock.Anything, prID).Return(pr, nil)

	mockDb.ExpectBegin()
	mockDb.ExpectCommit()

	// Execute
	uc := NewPRUseCase(mockUserRepo, mockPRRepo, db)
	result, err := uc.MergePR(ctx, prID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, domain.StatusMerged, result.Status)

	require.NoError(t, mockDb.ExpectationsWereMet())
	mockPRRepo.AssertExpectations(t)
	mockPRRepo.AssertNotCalled(t, "Update")
}

func TestPRUseCase_MergePR_NotFound(t *testing.T) {
	mockUserRepo := new(UserRepoMock)
	mockPRRepo := new(PullRequestRepoMock)

	db, mockDb, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	prID := "nonexistent"

	mockPRRepo.On("GetByIDForUpdate", ctx, mock.Anything, prID).Return(nil, domain.ErrNotFound)

	mockDb.ExpectBegin()
	mockDb.ExpectRollback()

	// Execute
	uc := NewPRUseCase(mockUserRepo, mockPRRepo, db)
	result, err := uc.MergePR(ctx, prID)

	// Assert
	assert.ErrorIs(t, err, domain.ErrNotFound)
	assert.Nil(t, result)

	require.NoError(t, mockDb.ExpectationsWereMet())

	mockPRRepo.AssertExpectations(t)
}

func TestPRUseCase_MergePR_UpdateError(t *testing.T) {
	mockUserRepo := new(UserRepoMock)
	mockPRRepo := new(PullRequestRepoMock)

	db, dbMock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	prID := "pr-1001"

	pr := &domain.PullRequest{
		ID:     prID,
		Status: domain.StatusOpen,
	}

	dbMock.ExpectBegin()

	mockPRRepo.On("GetByIDForUpdate", ctx, mock.Anything, prID).Return(pr, nil)
	mockPRRepo.On("Update", ctx, mock.Anything, mock.Anything).Return(errors.New("db error"))

	dbMock.ExpectRollback()

	// Execute
	uc := NewPRUseCase(mockUserRepo, mockPRRepo, db)
	result, err := uc.MergePR(ctx, prID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	require.NoError(t, dbMock.ExpectationsWereMet())
	mockPRRepo.AssertExpectations(t)
}

// ---- ReassignReviewer tests ----

func TestPRUseCase_ReassignReviewer_Success(t *testing.T) {
	mockUserRepo := new(UserRepoMock)
	mockPRRepo := new(PullRequestRepoMock)

	db, dbMock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	prID := "pr-1001"
	oldReviewerID := "u2"

	pr := &domain.PullRequest{
		ID:           prID,
		Name:         "Feature",
		AuthorID:     "u1",
		Status:       domain.StatusOpen,
		ReviewersIDs: []string{"u2", "u3"},
	}

	author := &domain.User{ID: "u1", TeamID: 1}
	candidate := "u4"

	dbMock.ExpectBegin()

	mockPRRepo.On("GetByID", ctx, prID).Return(pr, nil)
	mockUserRepo.On("GetByID", ctx, "u1").Return(author, nil)
	mockUserRepo.On("GetActiveTeamMembersIDs", ctx, int64(1), "u1").Return([]string{"u2", "u3", "u4"}, nil)

	mockPRRepo.On("RemoveReviewer", ctx, mock.Anything, prID, oldReviewerID).Return(nil)
	mockPRRepo.On("AddReviewer", ctx, mock.Anything, prID, candidate).Return(nil)

	dbMock.ExpectCommit()

	// Execute
	uc := NewPRUseCase(mockUserRepo, mockPRRepo, db)
	resultPR, newReviewerID, err := uc.ReassignReviewer(ctx, prID, oldReviewerID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "u4", newReviewerID)
	assert.Contains(t, resultPR.ReviewersIDs, "u4")
	assert.NotContains(t, resultPR.ReviewersIDs, "u2")
	assert.Contains(t, resultPR.ReviewersIDs, "u3") // u3 left

	require.NoError(t, dbMock.ExpectationsWereMet())
	mockUserRepo.AssertExpectations(t)
	mockPRRepo.AssertExpectations(t)
}

func TestPRUseCase_ReassignReviewer_NotAssigned(t *testing.T) {
	mockUserRepo := new(UserRepoMock)
	mockPRRepo := new(PullRequestRepoMock)

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	prID := "pr-1001"
	oldReviewerID := "u5" // not assigned

	pr := &domain.PullRequest{
		ID:           prID,
		Status:       domain.StatusOpen,
		ReviewersIDs: []string{"u2", "u3"},
	}

	mockPRRepo.On("GetByID", ctx, prID).Return(pr, nil)

	// Execute
	uc := NewPRUseCase(mockUserRepo, mockPRRepo, db)
	resultPR, newReviewerID, err := uc.ReassignReviewer(ctx, prID, oldReviewerID)

	// Assert
	assert.ErrorIs(t, err, domain.ErrNotAssigned)
	assert.Empty(t, newReviewerID)
	assert.Nil(t, resultPR)

	mockPRRepo.AssertExpectations(t)
}

func TestPRUseCase_ReassignReviewer_PRMerged(t *testing.T) {
	mockUserRepo := new(UserRepoMock)
	mockPRRepo := new(PullRequestRepoMock)

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	prID := "pr-1001"
	oldReviewerID := "u2"

	pr := &domain.PullRequest{
		ID:           prID,
		Status:       domain.StatusMerged,
		ReviewersIDs: []string{"u2", "u3"},
	}

	mockPRRepo.On("GetByID", ctx, prID).Return(pr, nil)

	// Execute
	uc := NewPRUseCase(mockUserRepo, mockPRRepo, db)
	resultPR, newReviewerID, err := uc.ReassignReviewer(ctx, prID, oldReviewerID)

	// Assert
	assert.ErrorIs(t, err, domain.ErrPRMerged)
	assert.Empty(t, newReviewerID)
	assert.Nil(t, resultPR)

	mockPRRepo.AssertExpectations(t)
}

func TestPRUseCase_ReassignReviewer_NoCandidate(t *testing.T) {
	mockUserRepo := new(UserRepoMock)
	mockPRRepo := new(PullRequestRepoMock)

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	prID := "pr-1001"
	oldReviewerID := "u2"

	pr := &domain.PullRequest{
		ID:           prID,
		AuthorID:     "u1",
		Status:       domain.StatusOpen,
		ReviewersIDs: []string{"u2", "u3"},
	}

	author := &domain.User{ID: "u1", TeamID: 1}
	// Team contains an author and current reviewers only - no candidates
	candidates := []string{"u2", "u3"} // only current reviewers

	mockPRRepo.On("GetByID", ctx, prID).Return(pr, nil)
	mockUserRepo.On("GetByID", ctx, "u1").Return(author, nil)
	mockUserRepo.On("GetActiveTeamMembersIDs", ctx, int64(1), "u1").Return(candidates, nil)

	// Execute
	uc := NewPRUseCase(mockUserRepo, mockPRRepo, db)
	resultPR, newReviewerID, err := uc.ReassignReviewer(ctx, prID, oldReviewerID)

	// Assert
	assert.ErrorIs(t, err, domain.ErrNoCandidate)
	assert.Empty(t, newReviewerID)
	assert.Nil(t, resultPR)

	mockUserRepo.AssertExpectations(t)
	mockPRRepo.AssertExpectations(t)
}

func TestPRUseCase_ReassignReviewer_PRNotFound(t *testing.T) {
	mockUserRepo := new(UserRepoMock)
	mockPRRepo := new(PullRequestRepoMock)

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	prID := "nonexistent"
	oldReviewerID := "u2"

	mockPRRepo.On("GetByID", ctx, prID).Return(nil, domain.ErrNotFound)

	// Execute
	uc := NewPRUseCase(mockUserRepo, mockPRRepo, db)
	resultPR, newReviewerID, err := uc.ReassignReviewer(ctx, prID, oldReviewerID)

	// Assert
	assert.ErrorIs(t, err, domain.ErrNotFound)
	assert.Empty(t, newReviewerID)
	assert.Nil(t, resultPR)

	mockPRRepo.AssertExpectations(t)
}

func TestPRUseCase_ReassignReviewer_RemoveReviewerError(t *testing.T) {
	mockUserRepo := new(UserRepoMock)
	mockPRRepo := new(PullRequestRepoMock)

	db, dbMock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	prID := "pr-1001"
	oldReviewerID := "u2"

	pr := &domain.PullRequest{
		ID:           prID,
		AuthorID:     "u1",
		Status:       domain.StatusOpen,
		ReviewersIDs: []string{"u2", "u3"},
	}

	author := &domain.User{ID: "u1", TeamID: 1}
	candidates := []string{"u2", "u3", "u4"}

	dbMock.ExpectBegin()

	mockPRRepo.On("GetByID", ctx, prID).Return(pr, nil)
	mockUserRepo.On("GetByID", ctx, "u1").Return(author, nil)
	mockUserRepo.On("GetActiveTeamMembersIDs", ctx, int64(1), "u1").Return(candidates, nil)
	mockPRRepo.On("RemoveReviewer", ctx, mock.Anything, prID, oldReviewerID).Return(errors.New("db error"))

	dbMock.ExpectRollback()

	// Execute
	uc := NewPRUseCase(mockUserRepo, mockPRRepo, db)
	resultPR, newReviewerID, err := uc.ReassignReviewer(ctx, prID, oldReviewerID)

	// Assert
	assert.Error(t, err)
	assert.Empty(t, newReviewerID)
	assert.Nil(t, resultPR)

	require.NoError(t, dbMock.ExpectationsWereMet())
	mockUserRepo.AssertExpectations(t)
	mockPRRepo.AssertExpectations(t)
}
