package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/adapter/http/model"
	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/domain"
	"github.com/gin-gonic/gin"
)

type userUseCase interface {
	SetUserIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error)
	GetAssignedPRs(ctx context.Context, userID string) ([]*domain.PullRequest, error)
}

type UserHandler struct {
	userUC userUseCase
}

func NewUserHandler(userUC userUseCase) *UserHandler {
	return &UserHandler{userUC: userUC}
}

// SetIsActive handles POST /users/setIsActive, updating a user's active status.
// Response:
//
//	200 OK with the updated user object.
//
// Errors:
//
//	400 Bad Request (INVALID_INPUT)
//	404 Not Found (NOT_FOUND)
//	500 Internal Server Error (INTERNAL_ERROR)
func (h *UserHandler) SetIsActive(c *gin.Context) {
	var req model.SetIsActiveRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(model.WriteErrorResponse(model.ErrCodeInvalidInput))
		return
	}

	user, err := h.userUC.SetUserIsActive(c.Request.Context(), req.UserID, *req.IsActive)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(model.WriteErrorResponse(model.ErrCodeNotFound))
			return
		}

		c.JSON(model.WriteErrorResponse(model.ErrCodeInternal))
		return
	}

	c.JSON(http.StatusOK, model.SetIsActiveResponse{User: model.UserFromDomain(user)})
}

// GetReview handles GET /users/getReview, returning PRs where the user is assigned as a reviewer.
// Response:
//
//	200 OK with the list of PRs (both OPEN and MERGED).
//
// Errors:
//
//	400 Bad Request (INVALID_INPUT)
//	404 Not Found (NOT_FOUND)
//	500 Internal Server Error (INTERNAL_ERROR)
func (h *UserHandler) GetReview(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(model.WriteErrorResponse(model.ErrCodeInvalidInput))
		return
	}

	prs, err := h.userUC.GetAssignedPRs(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(model.WriteErrorResponse(model.ErrCodeNotFound))
			return
		}

		c.JSON(model.WriteErrorResponse(model.ErrCodeInternal))
		return
	}

	c.JSON(http.StatusOK, model.GetAssignedPRsFromDomain(userID, prs))
}
