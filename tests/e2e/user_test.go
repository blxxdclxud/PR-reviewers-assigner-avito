//go:build e2e

package e2e

import (
	"github.com/stretchr/testify/assert"
)

func (s *E2ETestSuite) TestSetIsActive_Success() {
	// Create team with users
	teamPayload := map[string]interface{}{
		"team_name": "backend",
		"members": []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
		},
	}
	s.post("/team/add", teamPayload)

	// Deactivate user
	payload := map[string]interface{}{
		"user_id":   "u1",
		"is_active": false,
	}

	resp := s.post("/users/setIsActive", payload)
	assert.Equal(s.T(), 200, resp.StatusCode)

	var result map[string]interface{}
	s.parseJSON(resp, &result)

	user := result["user"].(map[string]interface{})
	assert.Equal(s.T(), "u1", user["user_id"])
	assert.Equal(s.T(), false, user["is_active"])
}

func (s *E2ETestSuite) TestSetIsActive_UserNotFound() {
	payload := map[string]interface{}{
		"user_id":   "nonexistent",
		"is_active": false,
	}

	resp := s.post("/users/setIsActive", payload)
	assert.Equal(s.T(), 404, resp.StatusCode)

	errResp := s.parseError(resp)
	errorObj := errResp["error"].(map[string]interface{})
	assert.Equal(s.T(), "NOT_FOUND", errorObj["code"])
}

func (s *E2ETestSuite) TestGetReview_Success() {
	// Create team
	teamPayload := map[string]interface{}{
		"team_name": "backend",
		"members": []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
		},
	}
	s.post("/team/add", teamPayload)

	// Create PR (u2 will reviewer)
	prPayload := map[string]interface{}{
		"pull_request_id":   "pr-1",
		"pull_request_name": "Add feature",
		"author_id":         "u1",
	}
	s.post("/pullRequest/create", prPayload)

	// Get u2's reviews
	resp := s.get("/users/getReview?user_id=u2")
	assert.Equal(s.T(), 200, resp.StatusCode)

	var result map[string]interface{}
	s.parseJSON(resp, &result)

	assert.Equal(s.T(), "u2", result["user_id"])
	prs := result["pull_requests"].([]interface{})
	assert.NotEmpty(s.T(), prs)
}
