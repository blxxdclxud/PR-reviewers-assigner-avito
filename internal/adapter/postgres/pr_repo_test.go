package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPullRequestRepository_Create(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := &PullRequestRepository{db: db}

	pr := &domain.PullRequest{ID: "test-pr-id", Name: "My-test-PR_wow", AuthorID: "ramadan", Status: "OPEN"}
	now := time.Now()

	// Correct insert
	mock.ExpectQuery(`INSERT INTO pull_requests`).
		WithArgs(pr.ID, pr.Name, pr.AuthorID, pr.Status).
		WillReturnRows(sqlmock.NewRows([]string{"created_at"}).AddRow(now))

	err := repo.Create(context.Background(), pr)
	require.NoError(t, err)
	assert.WithinDuration(t, now, pr.CreatedAt, time.Second)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Case with error
	mock.ExpectQuery(`INSERT INTO pull_requests`).
		WithArgs(pr.ID, pr.Name, pr.AuthorID, pr.Status).
		WillReturnError(errors.New("db error"))

	err = repo.Create(context.Background(), pr)
	assert.Error(t, err)
}

func TestPullRequestRepository_Update(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := &PullRequestRepository{db: db}

	mock.ExpectBegin()
	tx, err := db.Begin()
	require.NoError(t, err)
	require.NotNil(t, tx)

	pr := &domain.PullRequest{
		ID: "pr1", Status: "MERGED", MergedAt: &[]time.Time{time.Now()}[0],
	}

	// Successful update, only 1 updated row
	mock.ExpectExec(`UPDATE pull_requests`).
		WithArgs(pr.Status, pr.MergedAt, pr.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Update(context.Background(), tx, pr)
	require.NoError(t, err)

	// No rows (ErrNotFound)
	mock.ExpectExec(`UPDATE pull_requests`).
		WithArgs(pr.Status, pr.MergedAt, pr.ID).
		WillReturnResult(sqlmock.NewResult(1, 0))

	err = repo.Update(context.Background(), tx, pr)
	assert.ErrorIs(t, err, domain.ErrNotFound)

	// Query error
	mock.ExpectExec(`UPDATE pull_requests`).
		WithArgs(pr.Status, pr.MergedAt, pr.ID).
		WillReturnError(errors.New("update error"))

	err = repo.Update(context.Background(), tx, pr)
	assert.Error(t, err)

	mock.ExpectCommit()
	_ = tx.Commit()
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPullRequestRepository_GetByID(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := &PullRequestRepository{db: db}
	prID := "pr-1"

	// Case: row found, merged_at already not nil, reviewers returned
	// GetByID executes 2 queries
	mock.ExpectQuery(`SELECT id, name, author_id, status, created_at, merged_at FROM pull_requests`).
		WithArgs(prID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "author_id", "status", "created_at", "merged_at"}).
			AddRow(prID, "GetByID-PR", "admin-ramadan", "OPEN", time.Now(), time.Now()))
	mock.ExpectQuery(`SELECT user_id FROM pr_reviewers WHERE pr_id =`).
		WithArgs(prID).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow("junior-dev").AddRow("middle-dev"))

	pr, err := repo.GetByID(context.Background(), prID)
	require.NoError(t, err)
	assert.Equal(t, prID, pr.ID)
	assert.Equal(t, []string{"junior-dev", "middle-dev"}, pr.ReviewersIDs)
	assert.True(t, pr.MergedAt != nil)

	// Case: PR not found
	mock.ExpectQuery(`SELECT id, name, author_id, status, created_at, merged_at FROM pull_requests`).
		WithArgs("nil").WillReturnError(sql.ErrNoRows)

	pr, err = repo.GetByID(context.Background(), "nil")
	assert.ErrorIs(t, err, domain.ErrNotFound)
	assert.Nil(t, pr)

	// Error during fetching reviewers
	mock.ExpectQuery(`SELECT id, name, author_id, status, created_at, merged_at FROM pull_requests`).
		WithArgs(prID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "author_id", "status", "created_at", "merged_at"}).
			AddRow(prID, "PR", "u1", "OPEN", time.Now(), nil))
	mock.ExpectQuery(`SELECT user_id FROM pr_reviewers WHERE pr_id =`).
		WithArgs(prID).
		WillReturnError(errors.New("error reviewers"))

	pr, err = repo.GetByID(context.Background(), prID)
	assert.Error(t, err)
}

func TestPullRequestRepository_getReviewerIDs(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := &PullRequestRepository{db: db}
	prID := "pr-test"

	// reviewers correct
	mock.ExpectQuery(`SELECT user_id FROM pr_reviewers WHERE pr_id =`).
		WithArgs(prID).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow("user0").AddRow("user1"))

	ids, err := repo.getReviewerIDs(context.Background(), prID)
	require.NoError(t, err)
	assert.Equal(t, []string{"user0", "user1"}, ids)

	// no reviewers => just empty list
	mock.ExpectQuery(`SELECT user_id FROM pr_reviewers WHERE pr_id =`).
		WithArgs("empty").
		WillReturnError(sql.ErrNoRows)

	ids, err = repo.getReviewerIDs(context.Background(), "empty")
	require.NoError(t, err)
	assert.Equal(t, []string{}, ids)

	// Specific error during rows.Scan
	mock.ExpectQuery(`SELECT user_id FROM pr_reviewers WHERE pr_id =`).
		WithArgs(prID).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(nil))

	ids, err = repo.getReviewerIDs(context.Background(), prID)
	assert.Error(t, err)
}

func TestPRRepo_GetByIDForUpdate(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := &PullRequestRepository{db: db}
	prID := "pr-1"
	now := time.Now()

	mock.ExpectBegin()
	tx, _ := db.Begin()

	// PR found and reviewers returned
	mock.ExpectQuery(`SELECT id, name, author_id, status, created_at, merged_at FROM pull_requests WHERE id = \$1 FOR UPDATE`).
		WithArgs(prID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "author_id", "status", "created_at", "merged_at"}).AddRow(prID, "Feature", "user-1", "OPEN", now, nil))
	mock.ExpectQuery(`SELECT user_id FROM pr_reviewers WHERE pr_id =`).WithArgs(prID).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow("user-3"))

	pr, err := repo.GetByIDForUpdate(context.Background(), tx, prID)
	require.NoError(t, err)
	assert.Equal(t, []string{"user-3"}, pr.ReviewersIDs)

	// PR not found
	mock.ExpectQuery(`SELECT id, name, author_id, status, created_at, merged_at FROM pull_requests WHERE id = \$1 FOR UPDATE`).
		WithArgs("not-found").WillReturnError(sql.ErrNoRows)
	pr, err = repo.GetByIDForUpdate(context.Background(), tx, "not-found")
	assert.ErrorIs(t, err, domain.ErrNotFound)
	assert.Nil(t, pr)

	// getReviewerIDsTx error
	mock.ExpectQuery(`SELECT id, name, author_id, status, created_at, merged_at FROM pull_requests WHERE id = \$1 FOR UPDATE`).
		WithArgs(prID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "author_id", "status", "created_at", "merged_at"}).AddRow(prID, "Feature", "user-1", "OPEN", now, nil))
	mock.ExpectQuery(`SELECT user_id FROM pr_reviewers WHERE pr_id =`).
		WithArgs(prID).
		WillReturnError(errors.New("fail getReviewerIDsTx"))
	pr, err = repo.GetByIDForUpdate(context.Background(), tx, prID)
	assert.Error(t, err)
}

func TestPullRequestRepository_getReviewerIDsTx(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := &PullRequestRepository{db: db}
	prID := "pr-test"

	mock.ExpectBegin()
	tx, _ := db.Begin()

	// reviewers correct
	mock.ExpectQuery(`SELECT user_id FROM pr_reviewers WHERE pr_id =`).
		WithArgs(prID).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow("user0").AddRow("user1"))

	ids, err := repo.getReviewerIDsTx(context.Background(), tx, prID)
	require.NoError(t, err)
	assert.Equal(t, []string{"user0", "user1"}, ids)

	// no reviewers => just empty list
	mock.ExpectQuery(`SELECT user_id FROM pr_reviewers WHERE pr_id =`).
		WithArgs("empty").
		WillReturnError(sql.ErrNoRows)

	ids, err = repo.getReviewerIDsTx(context.Background(), tx, "empty")
	require.NoError(t, err)
	assert.Equal(t, []string{}, ids)

	// Specific error during rows.Scan
	mock.ExpectQuery(`SELECT user_id FROM pr_reviewers WHERE pr_id =`).
		WithArgs(prID).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(nil))

	ids, err = repo.getReviewerIDsTx(context.Background(), tx, prID)
	assert.Error(t, err)
}

func TestPRRepo_AddReviewer(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := &PullRequestRepository{db: db}

	mock.ExpectBegin()
	tx, _ := db.Begin()

	// Successful insert
	mock.ExpectExec(`INSERT INTO pr_reviewers`).WithArgs("pr-1", "user-1").WillReturnResult(sqlmock.NewResult(1, 1))
	err := repo.AddReviewer(context.Background(), tx, "pr-1", "user-1")
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Insert error
	mock.ExpectExec(`INSERT INTO pr_reviewers`).WithArgs("pr-1", "user-2").WillReturnError(errors.New("fail"))
	err = repo.AddReviewer(context.Background(), tx, "pr-1", "user-2")
	assert.Error(t, err)
}

func TestPRRepo_RemoveReviewer(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := &PullRequestRepository{db: db}

	mock.ExpectBegin()
	tx, _ := db.Begin()

	// successful delete
	mock.ExpectExec(`DELETE FROM pr_reviewers`).WithArgs("pr-1", "user-0").WillReturnResult(sqlmock.NewResult(1, 1))
	err := repo.RemoveReviewer(context.Background(), tx, "pr-1", "user-0")
	require.NoError(t, err)

	// No such reviewer
	mock.ExpectExec(`DELETE FROM pr_reviewers`).WithArgs("pr-1", "user-1").WillReturnResult(sqlmock.NewResult(1, 0))
	err = repo.RemoveReviewer(context.Background(), tx, "pr-1", "user-1")
	assert.ErrorIs(t, err, domain.ErrNotAssigned)

	// delete error
	mock.ExpectExec(`DELETE FROM pr_reviewers`).WithArgs("pr-1", "user-2").WillReturnError(errors.New("fail"))
	err = repo.RemoveReviewer(context.Background(), tx, "pr-1", "user-2")
	assert.Error(t, err)
}

func TestPRRepo_GetPRsByReviewer(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := &PullRequestRepository{db: db}
	userID := "user-777"
	now := time.Now()

	// Several PR returned
	mock.ExpectQuery(`SELECT id, name, author_id, status, created_at, merged_at FROM pull_requests`).WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "author_id", "status", "created_at", "merged_at"}).
			AddRow("pr-1", "pr-1", "user-1", "OPEN", now, nil).
			AddRow("pr-2", "pr-2", "user-3", "MERGED", now, now))

	prs, err := repo.GetPRsByReviewer(context.Background(), userID)
	require.NoError(t, err)
	assert.Len(t, prs, 2)
	assert.Equal(t, "pr-1", prs[0].ID)
	assert.Equal(t, "pr-2", prs[1].ID)

	// No such PR, sql.ErrNoRows error
	mock.ExpectQuery(`SELECT id, name, author_id, status, created_at, merged_at FROM pull_requests`).WithArgs("nil").
		WillReturnError(sql.ErrNoRows)
	prs, err = repo.GetPRsByReviewer(context.Background(), "nil")
	assert.ErrorIs(t, err, domain.ErrNotFound)
	assert.Nil(t, prs)

	// Query error
	mock.ExpectQuery(`SELECT id, name, author_id, status, created_at, merged_at FROM pull_requests`).WithArgs(userID).
		WillReturnError(errors.New("fail"))
	prs, err = repo.GetPRsByReviewer(context.Background(), userID)
	assert.Error(t, err)
}
