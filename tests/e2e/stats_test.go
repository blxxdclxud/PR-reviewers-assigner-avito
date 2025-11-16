//go:build e2e

package e2e

import (
	"fmt"

	"github.com/stretchr/testify/assert"
)

func (s *E2ETestSuite) TestGetStats_Success() {
	// Create test data
	teamPayload := map[string]interface{}{
		"team_name": "backend",
		"members": []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
			{"user_id": "u3", "username": "Charlie", "is_active": true},
		},
	}
	s.post("/team/add", teamPayload)

	// Create some PRs
	for i := 1; i <= 3; i++ {
		prPayload := map[string]interface{}{
			"pull_request_id":   fmt.Sprintf("pr-%d", i),
			"pull_request_name": fmt.Sprintf("Feature %d", i),
			"author_id":         "u1",
		}
		s.post("/pullRequest/create", prPayload)
	}

	// Merge one PR
	s.post("/pullRequest/merge", map[string]interface{}{
		"pull_request_id": "pr-1",
	})

	// Get stats
	resp := s.get("/stats")
	assert.Equal(s.T(), 200, resp.StatusCode)

	var stats map[string]interface{}
	s.parseJSON(resp, &stats)

	assert.Equal(s.T(), float64(1), stats["total_teams"])
	assert.Equal(s.T(), float64(3), stats["total_users"])
	assert.Equal(s.T(), float64(3), stats["total_prs"])
	assert.Equal(s.T(), float64(2), stats["open_prs"])
	assert.Equal(s.T(), float64(1), stats["merged_prs"])

	topReviewers := stats["top_reviewers"].([]interface{})
	assert.NotEmpty(s.T(), topReviewers)
}
