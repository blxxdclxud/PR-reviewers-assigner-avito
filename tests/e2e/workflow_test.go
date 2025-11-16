//go:build e2e

package e2e

import (
	"github.com/stretchr/testify/assert"
)

func (s *E2ETestSuite) TestCompleteWorkflow() {
	// Create team
	teamPayload := map[string]interface{}{
		"team_name": "backend",
		"members": []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
			{"user_id": "u3", "username": "Charlie", "is_active": true},
		},
	}
	resp := s.post("/team/add", teamPayload)
	assert.Equal(s.T(), 201, resp.StatusCode)

	// Create PR with auto-assigned reviewers
	prPayload := map[string]interface{}{
		"pull_request_id":   "pr-1",
		"pull_request_name": "Add feature",
		"author_id":         "u1",
	}
	resp = s.post("/pullRequest/create", prPayload)
	assert.Equal(s.T(), 201, resp.StatusCode)

	var prResult map[string]interface{}
	s.parseJSON(resp, &prResult)
	pr := prResult["pr"].(map[string]interface{})
	reviewers := pr["assigned_reviewers"].([]interface{})
	assert.Len(s.T(), reviewers, 2)

	// Deactivate one reviewer
	deactivatePayload := map[string]interface{}{
		"user_id":   reviewers[0].(string),
		"is_active": false,
	}
	resp = s.post("/users/setIsActive", deactivatePayload)
	assert.Equal(s.T(), 200, resp.StatusCode)

	// Check assigned PRs for the other reviewer
	resp = s.get("/users/getReview?user_id=" + reviewers[1].(string))
	assert.Equal(s.T(), 200, resp.StatusCode)

	var reviewResult map[string]interface{}
	s.parseJSON(resp, &reviewResult)
	prs := reviewResult["pull_requests"].([]interface{})
	assert.Len(s.T(), prs, 1)

	// Merge PR
	mergePayload := map[string]interface{}{
		"pull_request_id": "pr-1",
	}
	resp = s.post("/pullRequest/merge", mergePayload)
	assert.Equal(s.T(), 200, resp.StatusCode)

	var mergeResult map[string]interface{}
	s.parseJSON(resp, &mergeResult)
	mergedPR := mergeResult["pr"].(map[string]interface{})
	assert.Equal(s.T(), "MERGED", mergedPR["status"])
}
