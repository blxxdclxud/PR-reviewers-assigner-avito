package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/adapter/http/model"
	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/domain"
	"github.com/gin-gonic/gin"
)

type teamUseCase interface {
	CreateTeam(ctx context.Context, team domain.Team) (*domain.Team, error)
	GetTeam(ctx context.Context, teamName string) (*domain.Team, error)
}

type TeamHandler struct {
	teamUC teamUseCase
}

func NewTeamHandler(teamUC teamUseCase) *TeamHandler {
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
		c.JSON(model.WriteErrorResponse(model.ErrCodeInvalidInput))
		return
	}

	team, err := h.teamUC.CreateTeam(c.Request.Context(), req.ToDomain())
	if err != nil {
		if errors.Is(err, domain.ErrTeamExists) {
			c.JSON(model.WriteErrorResponse(model.ErrCodeTeamExists))
			return
		}

		c.JSON(model.WriteErrorResponse(model.ErrCodeInternal))
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
	if teamName == "" {
		c.JSON(model.WriteErrorResponse(model.ErrCodeInvalidInput))
		return
	}

	team, err := h.teamUC.GetTeam(c.Request.Context(), teamName)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(model.WriteErrorResponse(model.ErrCodeNotFound))
			return
		}

		c.JSON(model.WriteErrorResponse(model.ErrCodeInternal))
		return
	}

	c.JSON(http.StatusOK, model.TeamFromDomain(team))
}
