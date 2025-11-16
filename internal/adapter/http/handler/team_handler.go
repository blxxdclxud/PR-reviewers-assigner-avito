package handler

import (
	"errors"
	"net/http"

	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/adapter/http/model"
	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/domain"
	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/usecase"
	"github.com/gin-gonic/gin"
)

type TeamHandler struct {
	teamUC *usecase.TeamUseCase
}

func NewTeamHandler(teamUC *usecase.TeamUseCase) *TeamHandler {
	return &TeamHandler{teamUC: teamUC}
}

// Add handles POST /team/add, creating a team with members.
// Response:
//
//	201 Created with the team object.
//
// Errors:
//
//	400 Bad Request (INVALID_INPUT, TEAM_EXISTS)
//	500 Internal Server Error (INTERNAL_ERROR)
func (h *TeamHandler) Add(c *gin.Context) {
	var req model.CreateTeamRequest

	// Parse json body
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(WriteErrorResponse(model.ErrCodeInvalidInput))
		return
	}

	team, err := h.teamUC.CreateTeam(c.Request.Context(), req.ToDomain())
	if err != nil {
		if errors.Is(err, domain.ErrTeamExists) {
			c.JSON(WriteErrorResponse(model.ErrCodeTeamExists))
			return
		}

		c.JSON(WriteErrorResponse(model.ErrCodeInternal))
		return
	}

	c.JSON(http.StatusCreated, gin.H{"team": model.TeamFromDomain(team)})
}

// Get handles GET /team/get, returning a team by name.
// Response:
//
//	200 OK with the team object.
//
// Errors:
//
//	400 Bad Request (INVALID_INPUT)
//	404 Not Found (NOT_FOUND)
//	500 Internal Server Error (INTERNAL_ERROR)
func (h *TeamHandler) Get(c *gin.Context) {
	teamName := c.Query("team_name")

	team, err := h.teamUC.GetTeam(c.Request.Context(), teamName)
	if err != nil {
		// Check error type
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(WriteErrorResponse(model.ErrCodeNotFound))
			return
		}

		// Any other error - internal server error
		c.JSON(WriteErrorResponse(model.ErrCodeInternal))
		return
	}

	c.JSON(http.StatusOK, model.TeamFromDomain(team))
}
