package handler

import (
	"errors"
	"net/http"

	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/adapter/http/model"
	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/domain"
	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/usecase"
	"github.com/gin-gonic/gin"
)

type PRHandler struct {
	prUC *usecase.PRUseCase
}

func NewPRHandler(prUC *usecase.PRUseCase) *PRHandler {
	return &PRHandler{prUC: prUC}
}

// Create handles POST /pullRequest/create, creating a PR and auto-assigning reviewers.
// Response:
//
//	201 Created with the PR object.
//
// Errors:
//
//	400 Bad Request (INVALID_INPUT)
//	404 Not Found (NOT_FOUND - author/team not found)
//	409 Conflict (PR_EXISTS)
//	500 Internal Server Error (INTERNAL_ERROR)
func (h *PRHandler) Create(c *gin.Context) {
	var req model.CreatePRRequest

	// Parse JSON body
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(model.WriteErrorResponse(model.ErrCodeInvalidInput))
		return
	}

	pr, err := h.prUC.CreatePRAndSetReviewers(c.Request.Context(), req.ToDomain())
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(model.WriteErrorResponse(model.ErrCodeNotFound))
			return
		}

		if errors.Is(err, domain.ErrPRExists) {
			c.JSON(model.WriteErrorResponse(model.ErrCodePRExists))
			return
		}

		c.JSON(model.WriteErrorResponse(model.ErrCodeInternal))
		return
	}

	c.JSON(http.StatusCreated, gin.H{"pr": model.PRFromDomain(pr)})
}

// Merge handles POST /pullRequest/merge, marking a PR as MERGED.
// This operation is idempotent - calling it multiple times has no additional effect.
// Response:
//
//	200 OK with the PR object in MERGED state.
//
// Errors:
//
//	400 Bad Request (INVALID_INPUT)
//	404 Not Found (NOT_FOUND)
//	500 Internal Server Error (INTERNAL_ERROR)
func (h *PRHandler) Merge(c *gin.Context) {
	var req model.MergePRRequest

	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(model.WriteErrorResponse(model.ErrCodeInvalidInput))
		return
	}

	pr, err := h.prUC.MergePR(c.Request.Context(), req.PullRequestID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(model.WriteErrorResponse(model.ErrCodeNotFound))
			return
		}

		c.JSON(model.WriteErrorResponse(model.ErrCodeInternal))
		return
	}

	c.JSON(http.StatusOK, gin.H{"pr": model.PRFromDomain(pr)})
}

// Reassign handles POST /pullRequest/reassign, replacing a reviewer with another team member.
// Response:
//
//	200 OK with the PR object and the new reviewer's user_id.
//
// Errors:
//
//	400 Bad Request (INVALID_INPUT)
//	404 Not Found (NOT_FOUND - PR or user not found)
//	409 Conflict (PR_MERGED, NOT_ASSIGNED, NO_CANDIDATE)
//	500 Internal Server Error (INTERNAL_ERROR)
func (h *PRHandler) Reassign(c *gin.Context) {
	var req model.ReassignReviewerRequest

	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(model.WriteErrorResponse(model.ErrCodeInvalidInput))
		return
	}

	pr, newReviewerID, err := h.prUC.ReassignReviewer(c.Request.Context(), req.PullRequestID, req.OldUserID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(model.WriteErrorResponse(model.ErrCodeNotFound))
			return
		}

		if errors.Is(err, domain.ErrPRMerged) {
			c.JSON(model.WriteErrorResponse(model.ErrCodePRMerged))
			return
		}

		if errors.Is(err, domain.ErrNotAssigned) {
			c.JSON(model.WriteErrorResponse(model.ErrCodeNotAssigned))
			return
		}

		if errors.Is(err, domain.ErrNoCandidate) {
			c.JSON(model.WriteErrorResponse(model.ErrCodeNoCandidate))
			return
		}

		c.JSON(model.WriteErrorResponse(model.ErrCodeInternal))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pr":          model.PRFromDomain(pr),
		"replaced_by": newReviewerID,
	})
}
