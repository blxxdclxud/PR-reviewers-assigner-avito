//go:build e2e

package e2e

import (
	"github.com/stretchr/testify/assert"
)

func (s *E2ETestSuite) TestTeamAdd_Success() {
	payload := map[string]interface{}{
		"team_name": "backend",
		"members": []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
		},
	}

	resp := s.post("/team/add", payload)
	assert.Equal(s.T(), 201, resp.StatusCode)

	var result map[string]interface{}
	s.parseJSON(resp, &result)

	team := result["team"].(map[string]interface{})
	assert.Equal(s.T(), "backend", team["team_name"])
	assert.Len(s.T(), team["members"], 2)
}

func (s *E2ETestSuite) TestTeamAdd_DuplicateName() {
	payload := map[string]interface{}{
		"team_name": "backend",
		"members": []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
		},
	}

	// First request - success
	resp1 := s.post("/team/add", payload)
	assert.Equal(s.T(), 201, resp1.StatusCode)

	// Second request - conflict
	resp2 := s.post("/team/add", payload)
	assert.Equal(s.T(), 400, resp2.StatusCode)

	errResp := s.parseError(resp2)
	errorObj := errResp["error"].(map[string]interface{})
	assert.Equal(s.T(), "TEAM_EXISTS", errorObj["code"])
}

func (s *E2ETestSuite) TestTeamGet_Success() {
	// Create team first
	payload := map[string]interface{}{
		"team_name": "backend",
		"members": []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
		},
	}
	s.post("/team/add", payload)

	// Get team
	resp := s.get("/team/get?team_name=backend")
	assert.Equal(s.T(), 200, resp.StatusCode)

	var team map[string]interface{}
	s.parseJSON(resp, &team)
	assert.Equal(s.T(), "backend", team["team_name"])
}

func (s *E2ETestSuite) TestTeamGet_NotFound() {
	resp := s.get("/team/get?team_name=nonexistent")
	assert.Equal(s.T(), 404, resp.StatusCode)

	errResp := s.parseError(resp)
	errorObj := errResp["error"].(map[string]interface{})
	assert.Equal(s.T(), "NOT_FOUND", errorObj["code"])
}
