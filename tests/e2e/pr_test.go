//go:build e2e

package e2e

import (
	"github.com/stretchr/testify/assert"
)

func (s *E2ETestSuite) TestPRCreate_Success_TwoReviewers() {
	// Create team with 3 members
	teamPayload := map[string]interface{}{
		"team_name": "backend",
		"members": []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
			{"user_id": "u3", "username": "Charlie", "is_active": true},
		},
	}
	s.post("/team/add", teamPayload)

	// Create PR
	prPayload := map[string]interface{}{
		"pull_request_id":   "pr-1",
		"pull_request_name": "Add feature",
		"author_id":         "u1",
	}

	resp := s.post("/pullRequest/create", prPayload)
	assert.Equal(s.T(), 201, resp.StatusCode)

	var result map[string]interface{}
	s.parseJSON(resp, &result)

	pr := result["pr"].(map[string]interface{})
	assert.Equal(s.T(), "pr-1", pr["pull_request_id"])
	assert.Equal(s.T(), "OPEN", pr["status"])

	reviewers := pr["assigned_reviewers"].([]interface{})
	assert.Len(s.T(), reviewers, 2)
	assert.NotContains(s.T(), reviewers, "u1")
}

func (s *E2ETestSuite) TestPRCreate_OneReviewer() {
	teamPayload := map[string]interface{}{
		"team_name": "backend",
		"members": []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
		},
	}
	s.post("/team/add", teamPayload)

	prPayload := map[string]interface{}{
		"pull_request_id":   "pr-1",
		"pull_request_name": "Add feature",
		"author_id":         "u1",
	}

	resp := s.post("/pullRequest/create", prPayload)
	assert.Equal(s.T(), 201, resp.StatusCode)

	var result map[string]interface{}
	s.parseJSON(resp, &result)

	pr := result["pr"].(map[string]interface{})
	reviewers := pr["assigned_reviewers"].([]interface{})
	assert.Len(s.T(), reviewers, 1)
}

func (s *E2ETestSuite) TestPRCreate_DuplicateID() {
	teamPayload := map[string]interface{}{
		"team_name": "backend",
		"members": []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
		},
	}
	s.post("/team/add", teamPayload)

	prPayload := map[string]interface{}{
		"pull_request_id":   "pr-1",
		"pull_request_name": "Add feature",
		"author_id":         "u1",
	}

	resp1 := s.post("/pullRequest/create", prPayload)
	assert.Equal(s.T(), 201, resp1.StatusCode)

	resp2 := s.post("/pullRequest/create", prPayload)
	assert.Equal(s.T(), 409, resp2.StatusCode)

	errResp := s.parseError(resp2)
	errorObj := errResp["error"].(map[string]interface{})
	assert.Equal(s.T(), "PR_EXISTS", errorObj["code"])
}

func (s *E2ETestSuite) TestPRMerge_Success() {
	teamPayload := map[string]interface{}{
		"team_name": "backend",
		"members": []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
		},
	}
	s.post("/team/add", teamPayload)

	prPayload := map[string]interface{}{
		"pull_request_id":   "pr-1",
		"pull_request_name": "Add feature",
		"author_id":         "u1",
	}
	s.post("/pullRequest/create", prPayload)

	mergePayload := map[string]interface{}{
		"pull_request_id": "pr-1",
	}

	resp := s.post("/pullRequest/merge", mergePayload)
	assert.Equal(s.T(), 200, resp.StatusCode)

	var result map[string]interface{}
	s.parseJSON(resp, &result)

	pr := result["pr"].(map[string]interface{})
	assert.Equal(s.T(), "MERGED", pr["status"])
	assert.NotNil(s.T(), pr["mergedAt"])
}

func (s *E2ETestSuite) TestPRMerge_Idempotent() {
	teamPayload := map[string]interface{}{
		"team_name": "backend",
		"members": []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
		},
	}
	s.post("/team/add", teamPayload)

	prPayload := map[string]interface{}{
		"pull_request_id":   "pr-1",
		"pull_request_name": "Add feature",
		"author_id":         "u1",
	}
	s.post("/pullRequest/create", prPayload)

	mergePayload := map[string]interface{}{
		"pull_request_id": "pr-1",
	}

	resp1 := s.post("/pullRequest/merge", mergePayload)
	assert.Equal(s.T(), 200, resp1.StatusCode)

	resp2 := s.post("/pullRequest/merge", mergePayload)
	assert.Equal(s.T(), 200, resp2.StatusCode)

	var result map[string]interface{}
	s.parseJSON(resp2, &result)

	pr := result["pr"].(map[string]interface{})
	assert.Equal(s.T(), "MERGED", pr["status"])
}

func (s *E2ETestSuite) TestPRReassign_Success() {
	teamPayload := map[string]interface{}{
		"team_name": "backend",
		"members": []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
			{"user_id": "u3", "username": "Charlie", "is_active": true},
			{"user_id": "u4", "username": "Dave", "is_active": true},
		},
	}
	s.post("/team/add", teamPayload)

	prPayload := map[string]interface{}{
		"pull_request_id":   "pr-1",
		"pull_request_name": "Add feature",
		"author_id":         "u1",
	}
	createResp := s.post("/pullRequest/create", prPayload)

	var createResult map[string]interface{}
	s.parseJSON(createResp, &createResult)

	pr := createResult["pr"].(map[string]interface{})
	reviewers := pr["assigned_reviewers"].([]interface{})
	oldReviewerID := reviewers[0].(string)

	reassignPayload := map[string]interface{}{
		"pull_request_id": "pr-1",
		"old_user_id":     oldReviewerID,
	}

	resp := s.post("/pullRequest/reassign", reassignPayload)
	assert.Equal(s.T(), 200, resp.StatusCode)

	var result map[string]interface{}
	s.parseJSON(resp, &result)

	newPR := result["pr"].(map[string]interface{})
	newReviewers := newPR["assigned_reviewers"].([]interface{})
	replacedBy := result["replaced_by"].(string)

	assert.NotContains(s.T(), newReviewers, oldReviewerID)
	assert.Contains(s.T(), newReviewers, replacedBy)
}

func (s *E2ETestSuite) TestPRReassign_AfterMerge() {
	teamPayload := map[string]interface{}{
		"team_name": "backend",
		"members": []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
			{"user_id": "u3", "username": "Charlie", "is_active": true},
		},
	}
	s.post("/team/add", teamPayload)

	prPayload := map[string]interface{}{
		"pull_request_id":   "pr-1",
		"pull_request_name": "Add feature",
		"author_id":         "u1",
	}
	s.post("/pullRequest/create", prPayload)

	mergePayload := map[string]interface{}{
		"pull_request_id": "pr-1",
	}
	s.post("/pullRequest/merge", mergePayload)

	reassignPayload := map[string]interface{}{
		"pull_request_id": "pr-1",
		"old_user_id":     "u2",
	}

	resp := s.post("/pullRequest/reassign", reassignPayload)
	assert.Equal(s.T(), 409, resp.StatusCode)

	errResp := s.parseError(resp)
	errorObj := errResp["error"].(map[string]interface{})
	assert.Equal(s.T(), "PR_MERGED", errorObj["code"])
}

func (s *E2ETestSuite) TestPRReassign_NotAssigned() {
	teamPayload := map[string]interface{}{
		"team_name": "backend",
		"members": []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
			{"user_id": "u3", "username": "Charlie", "is_active": true},
		},
	}
	s.post("/team/add", teamPayload)

	prPayload := map[string]interface{}{
		"pull_request_id":   "pr-1",
		"pull_request_name": "Add feature",
		"author_id":         "u1",
	}
	s.post("/pullRequest/create", prPayload)

	reassignPayload := map[string]interface{}{
		"pull_request_id": "pr-1",
		"old_user_id":     "u1", // it is author, not reviewer
	}

	resp := s.post("/pullRequest/reassign", reassignPayload)
	assert.Equal(s.T(), 409, resp.StatusCode)

	errResp := s.parseError(resp)
	errorObj := errResp["error"].(map[string]interface{})
	assert.Equal(s.T(), "NOT_ASSIGNED", errorObj["code"])
}
