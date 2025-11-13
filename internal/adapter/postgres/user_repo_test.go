package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_Create(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := &UserRepository{db: db}
	user := &domain.User{ID: "user-1", Name: "Bob", IsActive: true, TeamID: int64(1)}

	// Success
	mock.ExpectExec("INSERT INTO users").WithArgs(user.ID, user.Name, user.IsActive, user.TeamID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	err := repo.Create(context.Background(), user)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	// insert error
	mock.ExpectExec("INSERT INTO users").WithArgs(user.ID, user.Name, user.IsActive, user.TeamID).
		WillReturnError(errors.New("insert error"))
	err = repo.Create(context.Background(), user)
	assert.Error(t, err)
}

func TestUserRepository_Update(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := &UserRepository{db: db}
	user := &domain.User{ID: "user-1", Name: "Bob", IsActive: false, TeamID: int64(1)}

	// Success
	mock.ExpectExec("UPDATE users").WithArgs(user.Name, user.IsActive, user.TeamID, user.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	err := repo.Update(context.Background(), user)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Update error
	mock.ExpectExec("UPDATE users").WithArgs(user.Name, user.IsActive, user.TeamID, user.ID).
		WillReturnError(errors.New("update error"))
	err = repo.Update(context.Background(), user)
	assert.Error(t, err)
}

func TestUserRepository_GetByID(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := &UserRepository{db: db}

	userID := "user-0"
	teamID := int64(1)

	// user found
	mock.ExpectQuery("SELECT id, username, is_active, team_id FROM users").WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "is_active", "team_id"}).
			AddRow(userID, "Bob", true, teamID))
	user, err := repo.GetByID(context.Background(), userID)
	require.NoError(t, err)
	assert.Equal(t, "Bob", user.Name)
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, true, user.IsActive)
	assert.Equal(t, teamID, user.TeamID)

	// User not found
	mock.ExpectQuery("SELECT id, username, is_active, team_id FROM users").WithArgs("user-1").
		WillReturnError(sql.ErrNoRows)
	res, err := repo.GetByID(context.Background(), "user-1")
	assert.ErrorIs(t, err, domain.ErrNotFound)
	assert.Nil(t, res)

	// Другая ошибка
	mock.ExpectQuery("SELECT id, username, is_active, team_id FROM users").WithArgs("user-500").
		WillReturnError(errors.New("db failed"))
	res, err = repo.GetByID(context.Background(), "user-500")
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestUserRepository_GetActiveTeamMembersIDs(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := &UserRepository{db: db}
	teamID := int64(1)
	exclude := "user-3"

	// Two members
	mock.ExpectQuery(`SELECT id FROM users WHERE team_id = \$1 AND is_active = true AND id != \$2`).
		WithArgs(teamID, exclude).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("user-1").AddRow("user-2"))
	ids, err := repo.GetActiveTeamMembersIDs(context.Background(), teamID, exclude)
	require.NoError(t, err)
	assert.Equal(t, []string{"user-1", "user-2"}, ids)

	// Empty members list
	mock.ExpectQuery(`SELECT id FROM users WHERE team_id = \$1 AND is_active = true AND id != \$2`).
		WithArgs(teamID, exclude).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	ids, err = repo.GetActiveTeamMembersIDs(context.Background(), teamID, exclude)
	require.NoError(t, err)
	assert.Empty(t, ids)

	// Scan error
	mock.ExpectQuery(`SELECT id FROM users WHERE team_id = \$1 AND is_active = true AND id != \$2`).
		WithArgs(teamID, exclude).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(nil))
	ids, err = repo.GetActiveTeamMembersIDs(context.Background(), teamID, exclude)
	assert.Error(t, err)
	assert.Nil(t, ids)

	// Query error
	mock.ExpectQuery(`SELECT id FROM users WHERE team_id = \$1 AND is_active = true AND id != \$2`).
		WithArgs(teamID, exclude).
		WillReturnError(errors.New("qfail"))
	ids, err = repo.GetActiveTeamMembersIDs(context.Background(), teamID, exclude)
	assert.Error(t, err)
	assert.Nil(t, ids)
}
