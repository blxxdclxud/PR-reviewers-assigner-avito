//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

// IntegrationTestSuite — set of integration tests
type IntegrationTestSuite struct {
	suite.Suite
	db       *sql.DB
	userRepo *UserRepository
	teamRepo *TeamRepository
	prRepo   *PullRequestRepository
}

func (s *IntegrationTestSuite) SetupSuite() {
	dsn := "host=localhost port=5433 user=test_user password=test_pass dbname=test_db sslmode=disable"

	db, err := sql.Open("postgres", dsn)
	require.NoError(s.T(), err)

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	require.NoError(s.T(), err, "failed to connect to test database")

	logger := zap.NewNop()

	s.db = db
	s.teamRepo = NewTeamRepository(db, logger)
	s.userRepo = NewUserRepository(db, logger)
	s.prRepo = NewPullRequestRepository(db, logger)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}
}

func (s *IntegrationTestSuite) SetupTest() {
	s.cleanupTables()
}

func (s *IntegrationTestSuite) TearDownTest() {
	s.cleanupTables()
}

func (s *IntegrationTestSuite) cleanupTables() {
	tables := []string{"pr_reviewers", "pull_requests", "users", "teams"}
	for _, table := range tables {
		_, err := s.db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		require.NoError(s.T(), err)
	}
}

// ==== TeamRepository tests ====
func (s *IntegrationTestSuite) TestTeamCreate_Success() {
	team := &domain.Team{Name: "backend-team"}

	tx, _ := s.db.Begin()
	err := s.teamRepo.Create(context.Background(), tx, team)
	require.NoError(s.T(), tx.Commit())

	require.NoError(s.T(), err)
	assert.NotZero(s.T(), team.ID)
}

func (s *IntegrationTestSuite) TestTeamCreate_DuplicateName() {
	team1 := &domain.Team{Name: "backend-team"}
	team2 := &domain.Team{Name: "backend-team"}

	tx, _ := s.db.Begin()

	err := s.teamRepo.Create(context.Background(), tx, team1)
	require.NoError(s.T(), err)

	// Try to create not unique team
	err = s.teamRepo.Create(context.Background(), tx, team2)
	require.NoError(s.T(), tx.Rollback())
	assert.ErrorIs(s.T(), err, domain.ErrTeamExists)
}

func (s *IntegrationTestSuite) TestTeamGetByName_Found() {
	// Create team and users
	team := &domain.Team{Name: "frontend-team"}
	tx, _ := s.db.Begin()
	err := s.teamRepo.Create(context.Background(), tx, team)
	require.NoError(s.T(), err)

	user1 := &domain.User{ID: "user-1", Name: "A", IsActive: true, TeamID: team.ID}
	user2 := &domain.User{ID: "user-2", Name: "B", IsActive: false, TeamID: team.ID}

	s.userRepo.Create(context.Background(), tx, user1)
	s.userRepo.Create(context.Background(), tx, user2)
	require.NoError(s.T(), tx.Commit())

	// Get team with members
	result, err := s.teamRepo.GetByName(context.Background(), "frontend-team")

	require.NoError(s.T(), err)
	assert.Equal(s.T(), "frontend-team", result.Name)
	assert.Len(s.T(), result.Members, 2)
	assert.Equal(s.T(), "A", result.Members[0].Name)
}

func (s *IntegrationTestSuite) TestTeamGetByName_NotFound() {
	_, err := s.teamRepo.GetByName(context.Background(), "nonexistent-team")
	assert.ErrorIs(s.T(), err, domain.ErrNotFound)
}

func (s *IntegrationTestSuite) TestTeamGetByName_EmptyMembers() {
	team := &domain.Team{Name: "empty-team"}
	tx, _ := s.db.Begin()
	s.teamRepo.Create(context.Background(), tx, team)
	require.NoError(s.T(), tx.Commit())

	result, err := s.teamRepo.GetByName(context.Background(), "empty-team")

	require.NoError(s.T(), err)
	assert.Empty(s.T(), result.Members)
}

// ==== UserRepository tests ====
func (s *IntegrationTestSuite) TestUserCreate_Success() {
	team := &domain.Team{Name: "team-1"}
	tx, _ := s.db.Begin()
	s.teamRepo.Create(context.Background(), tx, team)

	user := &domain.User{
		ID:       "user-1",
		Name:     "Charlie",
		IsActive: true,
		TeamID:   team.ID,
	}

	err := s.userRepo.Create(context.Background(), tx, user)
	require.NoError(s.T(), tx.Commit())
	assert.NoError(s.T(), err)
}

func (s *IntegrationTestSuite) TestUserCreate_InvalidTeamID() {
	user := &domain.User{
		ID:       "user-1",
		Name:     "Ramadan",
		IsActive: true,
		TeamID:   99999, // team not exist
	}

	tx, _ := s.db.Begin()

	err := s.userRepo.Create(context.Background(), tx, user)
	require.NoError(s.T(), tx.Rollback())
	assert.Error(s.T(), err) // FK constraint violation
}

func (s *IntegrationTestSuite) TestUserGetByID_Found() {
	team := &domain.Team{Name: "team-2"}
	tx, _ := s.db.Begin()
	s.teamRepo.Create(context.Background(), tx, team)

	user := &domain.User{ID: "user-10", Name: "Diana", IsActive: true, TeamID: team.ID}
	s.userRepo.Create(context.Background(), tx, user)
	require.NoError(s.T(), tx.Commit())

	result, err := s.userRepo.GetByID(context.Background(), "user-10")

	require.NoError(s.T(), err)
	assert.Equal(s.T(), "Diana", result.Name)
	assert.Equal(s.T(), team.ID, result.TeamID)
}

func (s *IntegrationTestSuite) TestUserGetByID_NotFound() {
	_, err := s.userRepo.GetByID(context.Background(), "nonexistent-user")
	assert.ErrorIs(s.T(), err, domain.ErrNotFound)
}

func (s *IntegrationTestSuite) TestUserUpdate_Success() {
	team := &domain.Team{Name: "team-3"}
	tx, _ := s.db.Begin()
	s.teamRepo.Create(context.Background(), tx, team)

	user := &domain.User{ID: "user-20", Name: "Eve", IsActive: true, TeamID: team.ID}
	s.userRepo.Create(context.Background(), tx, user)

	// update
	user.Name = "Eve Updated"
	user.IsActive = false

	err := s.userRepo.Update(context.Background(), tx, user)
	require.NoError(s.T(), tx.Commit())
	require.NoError(s.T(), err)

	// Check updates
	updated, _ := s.userRepo.GetByID(context.Background(), "user-20")
	assert.Equal(s.T(), "Eve Updated", updated.Name)
	assert.False(s.T(), updated.IsActive)
}

func (s *IntegrationTestSuite) TestUserGetActiveTeamMembers_FilterCorrectly() {
	team := &domain.Team{Name: "team-4"}
	tx, _ := s.db.Begin()
	s.teamRepo.Create(context.Background(), tx, team)

	// Create users
	s.userRepo.Create(context.Background(), tx, &domain.User{ID: "u1", Name: "A", IsActive: true, TeamID: team.ID})
	s.userRepo.Create(context.Background(), tx, &domain.User{ID: "u2", Name: "B", IsActive: false, TeamID: team.ID})
	s.userRepo.Create(context.Background(), tx, &domain.User{ID: "u3", Name: "C", IsActive: true, TeamID: team.ID})
	s.userRepo.Create(context.Background(), tx, &domain.User{ID: "u4", Name: "D", IsActive: true, TeamID: team.ID})

	require.NoError(s.T(), tx.Commit())

	// Get active members, excluding u1
	teamIDStr := team.ID
	members, err := s.userRepo.GetActiveTeamMembersIDs(context.Background(), teamIDStr, "u1")

	require.NoError(s.T(), err)
	assert.Len(s.T(), members, 2) // u3, u4
	assert.Contains(s.T(), members, "u3")
	assert.Contains(s.T(), members, "u4")
	assert.NotContains(s.T(), members, "u1") // excluded
	assert.NotContains(s.T(), members, "u2") // inactive
}

// ==== PullRequestRepository tests ====
func (s *IntegrationTestSuite) TestPRCreate_Success() {
	team := &domain.Team{Name: "team-5"}
	tx, _ := s.db.Begin()
	s.teamRepo.Create(context.Background(), tx, team)

	user := &domain.User{ID: "author-1", Name: "Author", IsActive: true, TeamID: team.ID}
	s.userRepo.Create(context.Background(), tx, user)

	pr := &domain.PullRequest{
		ID:       "pr-1001",
		Name:     "Feature X",
		AuthorID: "author-1",
		Status:   domain.StatusOpen,
	}

	err := s.prRepo.Create(context.Background(), tx, pr)
	require.NoError(s.T(), tx.Commit())

	require.NoError(s.T(), err)
	assert.False(s.T(), pr.CreatedAt.IsZero())
}

func (s *IntegrationTestSuite) TestPRCreate_InvalidAuthorID() {
	pr := &domain.PullRequest{
		ID:       "pr-1002",
		Name:     "Feature Y",
		AuthorID: "nonexistent-user",
		Status:   domain.StatusOpen,
	}

	tx, err := s.db.Begin()
	require.NoError(s.T(), err)

	err = s.prRepo.Create(context.Background(), tx, pr)
	assert.NoError(s.T(), tx.Rollback())
	assert.Error(s.T(), err) // FK constraint
}

func (s *IntegrationTestSuite) TestPRGetByID_Found() {
	team := &domain.Team{Name: "team-6"}
	tx, _ := s.db.Begin()
	s.teamRepo.Create(context.Background(), tx, team)

	author := &domain.User{ID: "author-2", Name: "Author2", IsActive: true, TeamID: team.ID}
	s.userRepo.Create(context.Background(), tx, author)

	pr := &domain.PullRequest{ID: "pr-2001", Name: "Fix bug", AuthorID: "author-2", Status: domain.StatusOpen}
	s.prRepo.Create(context.Background(), tx, pr)
	require.NoError(s.T(), tx.Commit())

	result, err := s.prRepo.GetByID(context.Background(), "pr-2001")

	require.NoError(s.T(), err)
	assert.Equal(s.T(), "Fix bug", result.Name)
	assert.Equal(s.T(), domain.StatusOpen, result.Status)
	assert.Nil(s.T(), result.MergedAt)
}

func (s *IntegrationTestSuite) TestPRGetByID_NotFound() {
	_, err := s.prRepo.GetByID(context.Background(), "pr-9999")
	assert.ErrorIs(s.T(), err, domain.ErrNotFound)
}

func (s *IntegrationTestSuite) TestPRUpdate_StatusToMerged() {
	team := &domain.Team{Name: "team-7"}
	tx, _ := s.db.Begin()
	s.teamRepo.Create(context.Background(), tx, team)

	author := &domain.User{ID: "author-3", Name: "Author3", IsActive: true, TeamID: team.ID}
	s.userRepo.Create(context.Background(), tx, author)

	pr := &domain.PullRequest{ID: "pr-3001", Name: "Release", AuthorID: "author-3", Status: domain.StatusOpen}
	s.prRepo.Create(context.Background(), tx, pr)
	require.NoError(s.T(), tx.Commit())

	// Update status
	pr.Status = domain.StatusMerged
	now := time.Now()
	pr.MergedAt = &now

	tx, _ = s.db.Begin()
	err := s.prRepo.Update(context.Background(), tx, pr)
	require.NoError(s.T(), err)
	require.NoError(s.T(), tx.Commit())

	// Check changes
	updated, _ := s.prRepo.GetByID(context.Background(), "pr-3001")
	assert.Equal(s.T(), domain.StatusMerged, updated.Status)
	assert.NotNil(s.T(), updated.MergedAt)
}

func (s *IntegrationTestSuite) TestPRUpdate_NotFound() {
	tx, _ := s.db.Begin()
	defer tx.Rollback()

	pr := &domain.PullRequest{ID: "pr-9999", Status: domain.StatusMerged}
	err := s.prRepo.Update(context.Background(), tx, pr)

	assert.ErrorIs(s.T(), err, domain.ErrNotFound)
}

// ==== Reviewer tests ====
func (s *IntegrationTestSuite) TestPRAddReviewer_Success() {
	team := &domain.Team{Name: "team-8"}
	tx, _ := s.db.Begin()
	s.teamRepo.Create(context.Background(), tx, team)

	author := &domain.User{ID: "author-4", Name: "Author4", IsActive: true, TeamID: team.ID}
	reviewer := &domain.User{ID: "reviewer-1", Name: "Reviewer1", IsActive: true, TeamID: team.ID}
	s.userRepo.Create(context.Background(), tx, author)
	s.userRepo.Create(context.Background(), tx, reviewer)

	pr := &domain.PullRequest{ID: "pr-4001", Name: "Feature", AuthorID: "author-4", Status: domain.StatusOpen}
	s.prRepo.Create(context.Background(), tx, pr)

	// Add reviewer
	err := s.prRepo.AddReviewer(context.Background(), tx, "pr-4001", "reviewer-1")
	require.NoError(s.T(), err)
	require.NoError(s.T(), tx.Commit())

	// Check that reviewer was added
	result, _ := s.prRepo.GetByID(context.Background(), "pr-4001")
	assert.Contains(s.T(), result.ReviewersIDs, "reviewer-1")
}

func (s *IntegrationTestSuite) TestPRAddReviewer_Duplicate_Idempotent() {
	team := &domain.Team{Name: "team-9"}
	tx, _ := s.db.Begin()
	s.teamRepo.Create(context.Background(), tx, team)

	author := &domain.User{ID: "author-5", Name: "Author5", IsActive: true, TeamID: team.ID}
	reviewer := &domain.User{ID: "reviewer-2", Name: "Reviewer2", IsActive: true, TeamID: team.ID}
	err := s.userRepo.Create(context.Background(), tx, author)
	require.NoError(s.T(), err)
	err = s.userRepo.Create(context.Background(), tx, reviewer)
	require.NoError(s.T(), err)

	pr := &domain.PullRequest{ID: "pr-5001", Name: "Feature", AuthorID: "author-5", Status: domain.StatusOpen}
	err = s.prRepo.Create(context.Background(), tx, pr)
	require.NoError(s.T(), err)

	// Add twice
	err = s.prRepo.AddReviewer(context.Background(), tx, "pr-5001", "reviewer-2")
	require.NoError(s.T(), err)
	err = s.prRepo.AddReviewer(context.Background(), tx, "pr-5001", "reviewer-2")
	require.NoError(s.T(), tx.Commit())

	// Operation must not conflict (ON CONFLICT DO NOTHING)
	assert.NoError(s.T(), err)

	// Check that reviewer was added only once
	result, err := s.prRepo.GetByID(context.Background(), "pr-5001")
	assert.NoError(s.T(), err)
	assert.Len(s.T(), result.ReviewersIDs, 1)
}

func (s *IntegrationTestSuite) TestPRRemoveReviewer_Success() {
	team := &domain.Team{Name: "team-10"}

	// Create team, two users, and PR
	tx, _ := s.db.Begin()
	s.teamRepo.Create(context.Background(), tx, team)

	author := &domain.User{ID: "author-6", Name: "Author6", IsActive: true, TeamID: team.ID}
	reviewer := &domain.User{ID: "reviewer-3", Name: "Reviewer3", IsActive: true, TeamID: team.ID}
	s.userRepo.Create(context.Background(), tx, author)
	s.userRepo.Create(context.Background(), tx, reviewer)

	pr := &domain.PullRequest{ID: "pr-6001", Name: "Feature", AuthorID: "author-6", Status: domain.StatusOpen}
	s.prRepo.Create(context.Background(), tx, pr)
	require.NoError(s.T(), tx.Commit())

	// Add reviewer
	tx, _ = s.db.Begin()
	s.prRepo.AddReviewer(context.Background(), tx, "pr-6001", "reviewer-3")
	require.NoError(s.T(), tx.Commit())

	// Delete reviewer
	tx, _ = s.db.Begin()
	err := s.prRepo.RemoveReviewer(context.Background(), tx, "pr-6001", "reviewer-3")
	require.NoError(s.T(), err)
	require.NoError(s.T(), tx.Commit())

	// Check that reviewer was deleted
	result, _ := s.prRepo.GetByID(context.Background(), "pr-6001")
	assert.Empty(s.T(), result.ReviewersIDs)
}

func (s *IntegrationTestSuite) TestPRRemoveReviewer_NotAssigned() {
	team := &domain.Team{Name: "team-11"}
	tx, _ := s.db.Begin()
	s.teamRepo.Create(context.Background(), tx, team)

	author := &domain.User{ID: "author-7", Name: "Author7", IsActive: true, TeamID: team.ID}
	s.userRepo.Create(context.Background(), tx, author)

	pr := &domain.PullRequest{ID: "pr-7001", Name: "Feature", AuthorID: "author-7", Status: domain.StatusOpen}
	s.prRepo.Create(context.Background(), tx, pr)

	// Try to delete reviewer that does not exist
	err := s.prRepo.RemoveReviewer(context.Background(), tx, "pr-7001", "nonexistent-reviewer")
	require.NoError(s.T(), tx.Rollback())

	assert.ErrorIs(s.T(), err, domain.ErrNotAssigned)
}

func (s *IntegrationTestSuite) TestPRGetPRsByReviewer_MultipleFound() {
	team := &domain.Team{Name: "team-12"}
	tx, _ := s.db.Begin()
	s.teamRepo.Create(context.Background(), tx, team)

	author := &domain.User{ID: "author-8", Name: "Author8", IsActive: true, TeamID: team.ID}
	reviewer := &domain.User{ID: "reviewer-4", Name: "Reviewer4", IsActive: true, TeamID: team.ID}
	s.userRepo.Create(context.Background(), tx, author)
	s.userRepo.Create(context.Background(), tx, reviewer)

	// Create several PRs
	pr1 := &domain.PullRequest{ID: "pr-8001", Name: "F1", AuthorID: "author-8", Status: domain.StatusOpen}
	pr2 := &domain.PullRequest{ID: "pr-8002", Name: "F2", AuthorID: "author-8", Status: domain.StatusOpen}
	s.prRepo.Create(context.Background(), tx, pr1)
	s.prRepo.Create(context.Background(), tx, pr2)

	// Assign reviewers for both PRs
	s.prRepo.AddReviewer(context.Background(), tx, "pr-8001", "reviewer-4")
	s.prRepo.AddReviewer(context.Background(), tx, "pr-8002", "reviewer-4")
	require.NoError(s.T(), tx.Commit())

	// Get PR of the reviewer
	prs, err := s.prRepo.GetPRsByReviewer(context.Background(), "reviewer-4")

	require.NoError(s.T(), err)
	assert.Len(s.T(), prs, 2)
}

func TestIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests")
	}

	suite.Run(t, new(IntegrationTestSuite))
}
