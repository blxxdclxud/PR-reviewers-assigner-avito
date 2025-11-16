//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	httpAdapter "github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/adapter/http"
	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/adapter/postgres"
	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/usecase"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// E2ETestSuite contains end-to-end tests for the entire application
type E2ETestSuite struct {
	suite.Suite
	db       *sql.DB
	server   *httptest.Server
	baseURL  string
	teamRepo *postgres.TeamRepository
	userRepo *postgres.UserRepository
	prRepo   *postgres.PullRequestRepository
}

// SetupSuite runs once before all tests
func (s *E2ETestSuite) SetupSuite() {
	dsn := "host=localhost port=5433 user=test_user password=test_pass dbname=test_db sslmode=disable"

	db, err := sql.Open("postgres", dsn)
	require.NoError(s.T(), err)

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	require.NoError(s.T(), err, "failed to connect to test database")

	s.db = db

	// Initialize repositories
	s.teamRepo = postgres.NewTeamRepository(db)
	s.userRepo = postgres.NewUserRepository(db)
	s.prRepo = postgres.NewPullRequestRepository(db)

	// Initialize use cases
	teamUC := usecase.NewTeamUseCase(s.teamRepo, s.userRepo, db)
	userUC := usecase.NewUserUseCase(s.userRepo, s.prRepo, s.teamRepo, db)
	prUC := usecase.NewPRUseCase(s.userRepo, s.prRepo, db)

	// Setup router and test server
	router := httpAdapter.SetupRouter(teamUC, userUC, prUC)
	s.server = httptest.NewServer(router)
	s.baseURL = s.server.URL
}

// TearDownSuite runs once after all tests
func (s *E2ETestSuite) TearDownSuite() {
	if s.server != nil {
		s.server.Close()
	}
	if s.db != nil {
		s.db.Close()
	}
}

// SetupTest runs before each test
func (s *E2ETestSuite) SetupTest() {
	s.cleanupTables()
}

// TearDownTest runs after each test
func (s *E2ETestSuite) TearDownTest() {
	s.cleanupTables()
}

func (s *E2ETestSuite) cleanupTables() {
	tables := []string{"pr_reviewers", "pull_requests", "users", "teams"}
	for _, table := range tables {
		_, err := s.db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		require.NoError(s.T(), err)
	}
}

// HTTP helpers
func (s *E2ETestSuite) post(path string, body interface{}) *http.Response {
	jsonBody, err := json.Marshal(body)
	require.NoError(s.T(), err)

	resp, err := http.Post(s.baseURL+path, "application/json", bytes.NewBuffer(jsonBody))
	require.NoError(s.T(), err)

	return resp
}

func (s *E2ETestSuite) get(path string) *http.Response {
	resp, err := http.Get(s.baseURL + path)
	require.NoError(s.T(), err)

	return resp
}

func (s *E2ETestSuite) parseJSON(resp *http.Response, to interface{}) {
	body, err := io.ReadAll(resp.Body)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	err = json.Unmarshal(body, to)
	require.NoError(s.T(), err)
}

func (s *E2ETestSuite) parseError(resp *http.Response) map[string]interface{} {
	var errResp map[string]interface{}
	s.parseJSON(resp, &errResp)
	return errResp
}

// TestE2ETestSuite runs the suite
func TestE2ETestSuite(t *testing.T) {
	suite.Run(t, new(E2ETestSuite))
}
