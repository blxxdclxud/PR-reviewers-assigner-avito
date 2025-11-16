package model

import (
	"time"

	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/domain"
)

// CreatePRRequest represents request body for POST /pullRequest/create
type CreatePRRequest struct {
	PullRequestID   string `json:"pull_request_id" binding:"required"`
	PullRequestName string `json:"pull_request_name" binding:"required"`
	AuthorID        string `json:"author_id" binding:"required"`
}

// ToDomain converts HTTP request to domain model
func (r *CreatePRRequest) ToDomain() domain.PullRequest {
	return domain.PullRequest{
		ID:       r.PullRequestID,
		Name:     r.PullRequestName,
		AuthorID: r.AuthorID,
	}
}

// MergePRRequest represents request body for POST /pullRequest/merge
type MergePRRequest struct {
	PullRequestID string `json:"pull_request_id" binding:"required"`
}

// ReassignReviewerRequest represents request body for POST /pullRequest/reassign
type ReassignReviewerRequest struct {
	PullRequestID string `json:"pull_request_id" binding:"required"`
	OldUserID     string `json:"old_user_id" binding:"required"`
}

// PullRequestResponse represents full PR object in responses
type PullRequestResponse struct {
	PullRequestID     string   `json:"pull_request_id"`
	PullRequestName   string   `json:"pull_request_name"`
	AuthorID          string   `json:"author_id"`
	Status            string   `json:"status"`
	AssignedReviewers []string `json:"assigned_reviewers"`
	CreatedAt         *string  `json:"createdAt,omitempty"`
	MergedAt          *string  `json:"mergedAt,omitempty"`
}

// PullRequestShortResponse represents short PR object in list responses
type PullRequestShortResponse struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
	Status          string `json:"status"`
}

// PRFromDomain converts domain.PullRequest to PullRequestResponse
func PRFromDomain(pr *domain.PullRequest) PullRequestResponse {
	var createdAt, mergedAt *string

	if !pr.CreatedAt.IsZero() {
		t := pr.CreatedAt.Format(time.RFC3339)
		createdAt = &t
	}

	if pr.MergedAt != nil {
		t := pr.MergedAt.Format(time.RFC3339)
		mergedAt = &t
	}

	return PullRequestResponse{
		PullRequestID:     pr.ID,
		PullRequestName:   pr.Name,
		AuthorID:          pr.AuthorID,
		Status:            string(pr.Status),
		AssignedReviewers: pr.ReviewersIDs,
		CreatedAt:         createdAt,
		MergedAt:          mergedAt,
	}
}

// PRShortFromDomain converts domain.PullRequest to PullRequestShortResponse
func PRShortFromDomain(pr *domain.PullRequest) PullRequestShortResponse {
	return PullRequestShortResponse{
		PullRequestID:   pr.ID,
		PullRequestName: pr.Name,
		AuthorID:        pr.AuthorID,
		Status:          string(pr.Status),
	}
}

// CreatePRResponse represents response for POST /pullRequest/create
type CreatePRResponse struct {
	PR PullRequestResponse `json:"pr"`
}

// MergePRResponse represents response for POST /pullRequest/merge
type MergePRResponse struct {
	PR PullRequestResponse `json:"pr"`
}

// ReassignReviewerResponse represents response for POST /pullRequest/reassign
type ReassignReviewerResponse struct {
	PR         PullRequestResponse `json:"pr"`
	ReplacedBy string              `json:"replaced_by"`
}

// GetAssignedPRsResponse represents response for GET /users/getReview
type GetAssignedPRsResponse struct {
	UserID       string                     `json:"user_id"`
	PullRequests []PullRequestShortResponse `json:"pull_requests"`
}

// GetAssignedPRsFromDomain converts slice of PRs to GetAssignedPRsResponse
func GetAssignedPRsFromDomain(userID string, prs []*domain.PullRequest) GetAssignedPRsResponse {
	prResponses := make([]PullRequestShortResponse, len(prs))
	for i, pr := range prs {
		prResponses[i] = PRShortFromDomain(pr)
	}

	return GetAssignedPRsResponse{
		UserID:       userID,
		PullRequests: prResponses,
	}
}
