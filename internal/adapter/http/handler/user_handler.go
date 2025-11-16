package handler

import (
	"errors"
	"net/http"

	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/adapter/http/model"
	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/domain"
	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/usecase"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userUC *usecase.UserUseCase
}

func NewUserHandler(userUC *usecase.UserUseCase) *UserHandler {
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
		c.JSON(WriteErrorResponse(model.ErrCodeInvalidInput))
		return
	}

	user, err := h.userUC.SetUserIsActive(c.Request.Context(), req.UserID, *req.IsActive)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(WriteErrorResponse(model.ErrCodeNotFound))
			return
		}

		// Any other error - internal server error
		c.JSON(WriteErrorResponse(model.ErrCodeInternal))
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

	prs, err := h.userUC.GetAssignedPRs(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(WriteErrorResponse(model.ErrCodeNotFound))
			return
		}

		// Any other error - internal server error
		c.JSON(WriteErrorResponse(model.ErrCodeInternal))
		return
	}

	c.JSON(http.StatusOK, model.GetAssignedPRsFromDomain(userID, prs))
}
