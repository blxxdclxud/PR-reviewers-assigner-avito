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

func TestTeamUseCase_CreateTeam_Success_NewUsers(t *testing.T) {
	// Setup
	mockTeamRepo := new(TeamRepoMock)
	mockUserRepo := new(UserRepoMock)

	ctx := context.Background()
	team := domain.Team{
		Name: "backend",
		Members: []domain.User{
			{ID: "u1", Name: "User", IsActive: true},
			{ID: "u2", Name: "Admin", IsActive: true},
		},
	}

	// create sql.DB dbMock
	db, dbMock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbMock.ExpectBegin()

	// Expectations
	mockTeamRepo.On("Create", ctx, mock.Anything, &team).Return(nil).Run(func(args mock.Arguments) {
		team := args.Get(2).(*domain.Team)
		team.ID = 123
	})

	// Both users are new, so GetByID should return ErrNotFound
	mockUserRepo.On("GetByID", ctx, "u1").Return(nil, domain.ErrNotFound)
	mockUserRepo.On("Create", ctx, mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.ID == "u1" && u.TeamID == 123
	})).Return(nil)

	mockUserRepo.On("GetByID", ctx, "u2").Return(nil, domain.ErrNotFound)
	mockUserRepo.On("Create", ctx, mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.ID == "u2" && u.TeamID == 123
	})).Return(nil)

	dbMock.ExpectCommit()

	// perform tests
	uc := NewTeamUseCase(mockTeamRepo, mockUserRepo, db)

	result, err := uc.CreateTeam(ctx, team)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "backend", result.Name)
	assert.Equal(t, int64(123), result.ID)
	assert.Len(t, result.Members, 2)

	require.NoError(t, dbMock.ExpectationsWereMet())

	mockTeamRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestTeamUseCase_CreateTeam_Success_ExistingUsers(t *testing.T) {
	// Setup
	mockTeamRepo := new(TeamRepoMock)
	mockUserRepo := new(UserRepoMock)

	ctx := context.Background()
	team := domain.Team{
		Name: "frontend",
		Members: []domain.User{
			{ID: "u1", Name: "User Updated", IsActive: false},
		},
	}

	// Expectations
	mockTeamRepo.On("Create", ctx, mock.Anything, &team).Return(nil).Run(func(args mock.Arguments) {
		team := args.Get(2).(*domain.Team)
		team.ID = 456
	})

	// User already exists
	existingUser := &domain.User{
		ID:       "u1",
		Name:     "User Old",
		IsActive: true,
		TeamID:   1,
	}
	mockUserRepo.On("GetByID", ctx, "u1").Return(existingUser, nil)

	// Expect update with new values
	mockUserRepo.On("Update", ctx, mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.ID == "u1" &&
			u.Name == "User Updated" &&
			u.IsActive == false &&
			u.TeamID == 456
	})).Return(nil)

	// Create sql.DB dbMock
	db, dbMock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbMock.ExpectBegin()
	dbMock.ExpectCommit()

	// Perform tests
	uc := NewTeamUseCase(mockTeamRepo, mockUserRepo, db)

	result, err := uc.CreateTeam(ctx, team)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "frontend", result.Name)
	assert.Equal(t, int64(456), result.ID)
	assert.Len(t, result.Members, 1)

	require.NoError(t, dbMock.ExpectationsWereMet())
	mockTeamRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestTeamUseCase_CreateTeam_TeamAlreadyExists(t *testing.T) {
	// Setup
	mockTeamRepo := new(TeamRepoMock)
	mockUserRepo := new(UserRepoMock)

	ctx := context.Background()
	team := domain.Team{
		Name:    "backend",
		Members: []domain.User{{ID: "u1", Name: "Admin", IsActive: true}},
	}

	// Expectations
	mockTeamRepo.On("Create", ctx, mock.Anything, &team).Return(domain.ErrTeamExists)

	// Create sql.DB dbMock
	db, dbMock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbMock.ExpectBegin()
	dbMock.ExpectRollback()

	// Perform tests
	uc := NewTeamUseCase(mockTeamRepo, mockUserRepo, db)

	result, err := uc.CreateTeam(ctx, team)

	// Assert
	assert.ErrorIs(t, err, domain.ErrTeamExists)
	assert.Nil(t, result)

	require.NoError(t, dbMock.ExpectationsWereMet())
	mockTeamRepo.AssertExpectations(t)
	mockUserRepo.AssertNotCalled(t, "GetByID")
	mockUserRepo.AssertNotCalled(t, "Create")
	mockUserRepo.AssertNotCalled(t, "Update")
}

func TestTeamUseCase_CreateTeam_FailedToCreateUser(t *testing.T) {
	// Setup
	mockTeamRepo := new(TeamRepoMock)
	mockUserRepo := new(UserRepoMock)

	ctx := context.Background()
	team := domain.Team{
		Name: "backend",
		Members: []domain.User{
			{ID: "u1", Name: "Admin", IsActive: true},
		},
	}

	// Expectations
	mockTeamRepo.On("Create", ctx, mock.Anything, &team).Return(nil).Run(func(args mock.Arguments) {
		team := args.Get(2).(*domain.Team)
		team.ID = 123
	})

	mockUserRepo.On("GetByID", ctx, "u1").Return(nil, domain.ErrNotFound)
	mockUserRepo.On("Create", ctx, mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.ID == "u1" && u.TeamID == 123
	})).Return(errors.New("database error"))

	// Create sql.DB dbMock
	db, dbMock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbMock.ExpectBegin()
	dbMock.ExpectRollback()

	// Perform tests
	uc := NewTeamUseCase(mockTeamRepo, mockUserRepo, db)

	result, err := uc.CreateTeam(ctx, team)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "database error", err.Error())
	assert.Nil(t, result)

	require.NoError(t, dbMock.ExpectationsWereMet())
	mockTeamRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestTeamUseCase_CreateTeam_FailedToUpdateUser(t *testing.T) {
	// Setup
	mockTeamRepo := new(TeamRepoMock)
	mockUserRepo := new(UserRepoMock)

	ctx := context.Background()
	team := domain.Team{
		Name: "backend",
		Members: []domain.User{
			{ID: "u1", Name: "Admin", IsActive: true},
		},
	}

	// Expectations
	mockTeamRepo.On("Create", ctx, mock.Anything, &team).Return(nil).Run(func(args mock.Arguments) {
		team := args.Get(2).(*domain.Team)
		team.ID = 123
	})

	existingUser := &domain.User{
		ID:       "u1",
		Name:     "Old Name",
		IsActive: false,
		TeamID:   1,
	}
	mockUserRepo.On("GetByID", ctx, "u1").Return(existingUser, nil)
	mockUserRepo.On("Update", ctx, mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.ID == "u1" && u.TeamID == 123
	})).Return(errors.New("update failed"))

	// Create sql.DB dbMock
	db, dbMock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbMock.ExpectBegin()
	dbMock.ExpectRollback()

	// Perform tests
	uc := NewTeamUseCase(mockTeamRepo, mockUserRepo, db)

	result, err := uc.CreateTeam(ctx, team)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "update failed", err.Error())
	assert.Nil(t, result)

	require.NoError(t, dbMock.ExpectationsWereMet())
	mockTeamRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestTeamUseCase_CreateTeam_GetUserReturnsUnexpectedError(t *testing.T) {
	// Setup
	mockTeamRepo := new(TeamRepoMock)
	mockUserRepo := new(UserRepoMock)

	ctx := context.Background()
	team := domain.Team{
		Name: "backend",
		Members: []domain.User{
			{ID: "u1", Name: "Admin", IsActive: true},
		},
	}

	// Expectations
	mockTeamRepo.On("Create", ctx, mock.Anything, &team).Return(nil).Run(func(args mock.Arguments) {
		team := args.Get(2).(*domain.Team)
		team.ID = 123
	})

	mockUserRepo.On("GetByID", ctx, "u1").Return(nil, errors.New("database connection lost"))

	// Create sql.DB dbMock
	db, dbMock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbMock.ExpectBegin()
	dbMock.ExpectRollback()

	// Perform tests
	uc := NewTeamUseCase(mockTeamRepo, mockUserRepo, db)

	result, err := uc.CreateTeam(ctx, team)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "database connection lost", err.Error())
	assert.Nil(t, result)

	require.NoError(t, dbMock.ExpectationsWereMet())
	mockTeamRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockUserRepo.AssertNotCalled(t, "Create")
	mockUserRepo.AssertNotCalled(t, "Update")
}
func TestTeamUseCase_CreateTeam_MultipleMembers_MixedScenario(t *testing.T) {
	// Setup
	mockTeamRepo := new(TeamRepoMock)
	mockUserRepo := new(UserRepoMock)

	ctx := context.Background()
	team := domain.Team{
		Name: "fullstack",
		Members: []domain.User{
			{ID: "u1", Name: "Admin", IsActive: true},   // new
			{ID: "u2", Name: "Junior", IsActive: false}, // existing
			{ID: "u3", Name: "Middle", IsActive: true},  // new
		},
	}

	// Expectations
	mockTeamRepo.On("Create", ctx, mock.Anything, &team).Return(nil).Run(func(args mock.Arguments) {
		team := args.Get(2).(*domain.Team)
		team.ID = 789
	})

	// u1 - new
	mockUserRepo.On("GetByID", ctx, "u1").Return(nil, domain.ErrNotFound)
	mockUserRepo.On("Create", ctx, mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.ID == "u1" && u.TeamID == 789
	})).Return(nil)

	// u2 - existing
	existingU2 := &domain.User{
		ID:       "u2",
		Name:     "Junior Old",
		IsActive: true,
		TeamID:   1,
	}
	mockUserRepo.On("GetByID", ctx, "u2").Return(existingU2, nil)
	mockUserRepo.On("Update", ctx, mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.ID == "u2" &&
			u.Name == "Junior" &&
			u.IsActive == false &&
			u.TeamID == 789
	})).Return(nil)

	// u3 - new
	mockUserRepo.On("GetByID", ctx, "u3").Return(nil, domain.ErrNotFound)
	mockUserRepo.On("Create", ctx, mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.ID == "u3" && u.TeamID == 789
	})).Return(nil)

	// Create sql.DB dbMock
	db, dbMock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbMock.ExpectBegin()
	dbMock.ExpectCommit()

	// Perform tests
	uc := NewTeamUseCase(mockTeamRepo, mockUserRepo, db)

	result, err := uc.CreateTeam(ctx, team)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "fullstack", result.Name)
	assert.Equal(t, int64(789), result.ID)
	assert.Len(t, result.Members, 3)

	require.NoError(t, dbMock.ExpectationsWereMet())
	mockTeamRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockUserRepo.AssertNumberOfCalls(t, "GetByID", 3)
	mockUserRepo.AssertNumberOfCalls(t, "Create", 2)
	mockUserRepo.AssertNumberOfCalls(t, "Update", 1)
}

func TestTeamUseCase_GetTeam_Success(t *testing.T) {
	// Setup
	mockTeamRepo := new(TeamRepoMock)
	mockUserRepo := new(UserRepoMock) // not used, but needed for constructor
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	teamName := "backend"
	expectedTeam := &domain.Team{
		ID:   123,
		Name: teamName,
		Members: []domain.User{
			{ID: "u1", Name: "User", IsActive: true},
			{ID: "u2", Name: "Admin", IsActive: true},
		},
	}

	// Expectations
	mockTeamRepo.On("GetByName", ctx, teamName).Return(expectedTeam, nil)

	// Execute
	uc := NewTeamUseCase(mockTeamRepo, mockUserRepo, db)
	result, err := uc.GetTeam(ctx, teamName)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedTeam, result)
	mockTeamRepo.AssertExpectations(t)
}

func TestTeamUseCase_GetTeam_NotFound(t *testing.T) {
	mockTeamRepo := new(TeamRepoMock)
	mockUserRepo := new(UserRepoMock)
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	teamName := "nonexistent"

	// Expectations: return domain.ErrNotFound
	mockTeamRepo.On("GetByName", ctx, teamName).Return(nil, domain.ErrNotFound)

	// Execute
	uc := NewTeamUseCase(mockTeamRepo, mockUserRepo, db)
	result, err := uc.GetTeam(ctx, teamName)

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrNotFound)
	assert.Nil(t, result)
	mockTeamRepo.AssertExpectations(t)
}

func TestTeamUseCase_GetTeam_RepoError(t *testing.T) {
	mockTeamRepo := new(TeamRepoMock)
	mockUserRepo := new(UserRepoMock)
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	teamName := "backend"
	repoErr := errors.New("some db error")

	// Expectations: return unexpected error
	mockTeamRepo.On("GetByName", ctx, teamName).Return(nil, repoErr)

	// Execute
	uc := NewTeamUseCase(mockTeamRepo, mockUserRepo, db)
	result, err := uc.GetTeam(ctx, teamName)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, repoErr, err)
	assert.Nil(t, result)
	mockTeamRepo.AssertExpectations(t)
}
