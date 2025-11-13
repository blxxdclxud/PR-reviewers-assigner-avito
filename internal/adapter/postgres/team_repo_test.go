package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/domain"
	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeamRepository_Create(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := &TeamRepository{db: db}

	team := &domain.Team{
		ID:      0,
		Name:    "team-1",
		Members: nil,
	}

	// Correct insert
	mock.ExpectQuery(`INSERT INTO teams`).
		WithArgs(team.Name).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err := repo.Create(context.Background(), team)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Case with error
	mock.ExpectQuery(`INSERT INTO teams`).
		WithArgs(team.Name).
		WillReturnError(errors.New("db error"))

	err = repo.Create(context.Background(), team)
	assert.Error(t, err)

	// Case with duplicated team name, should return domain.ErrTeamExists
	uniqErr := &pq.Error{Code: pgerrcode.UniqueViolation}

	mock.ExpectQuery(`INSERT INTO teams`).
		WithArgs(team.Name).
		WillReturnError(uniqErr)

	err = repo.Create(context.Background(), team)
	assert.ErrorIs(t, err, domain.ErrTeamExists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTeamRepository_GetByName(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := &TeamRepository{db: db}

	teamID := int64(42)
	teamName := "team-1"

	// Team found
	mock.ExpectQuery(`SELECT id, name FROM teams`).
		WithArgs(teamName).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(teamID, teamName))

	// Two members
	mock.ExpectQuery(`SELECT id, username, is_active, team_id FROM users`).
		WithArgs(teamID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "is_active", "team_id"}).
			AddRow("user-1", "A", true, teamID).
			AddRow("user-2", "B", false, teamID))

	result, err := repo.GetByName(context.Background(), teamName)
	require.NoError(t, err)
	assert.Equal(t, teamID, result.ID)
	assert.Equal(t, "team-1", result.Name)
	assert.Len(t, result.Members, 2)
	assert.Equal(t, "user-1", result.Members[0].ID)
	assert.Equal(t, "B", result.Members[1].Name)

	// Team not found
	mock.ExpectQuery(`SELECT id, name FROM teams`).WithArgs("missing-team").WillReturnError(sql.ErrNoRows)
	res, err := repo.GetByName(context.Background(), "missing-team")
	assert.ErrorIs(t, err, domain.ErrNotFound)
	assert.Nil(t, res)

	// Query error
	mock.ExpectQuery(`SELECT id, name FROM teams`).WithArgs("fail-team").WillReturnError(errors.New("DB fail"))
	res, err = repo.GetByName(context.Background(), "fail-team")
	assert.Error(t, err)
	assert.Nil(t, res)

	// Error in getTeamMembers
	mock.ExpectQuery(`SELECT id, name FROM teams`).WithArgs("error-mem").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(teamID, "error-mem"))
	mock.ExpectQuery(`SELECT id, username, is_active, team_id FROM users`).WithArgs(teamID).WillReturnError(errors.New("user query error"))
	res, err = repo.GetByName(context.Background(), "error-mem")
	assert.Error(t, err)
	assert.Nil(t, res)

	// No members in the team
	mock.ExpectQuery(`SELECT id, name FROM teams`).
		WithArgs("lonely-team").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(100, "lonely-team"))
	mock.ExpectQuery(`SELECT id, username, is_active, team_id FROM users`).
		WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "is_active", "team_id"}))
	res, err = repo.GetByName(context.Background(), "lonely-team")
	require.NoError(t, err)
	assert.Equal(t, int64(100), res.ID)
	assert.Empty(t, res.Members)

	assert.NoError(t, mock.ExpectationsWereMet())
}
