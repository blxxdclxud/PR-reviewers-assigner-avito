package handler

import (
	"net/http"

	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/adapter/http/model"
	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/usecase"
	"github.com/gin-gonic/gin"
)

type StatsHandler struct {
	statsUC *usecase.StatsUseCase
}

func NewStatsHandler(statsUC *usecase.StatsUseCase) *StatsHandler {
	return &StatsHandler{statsUC: statsUC}
}

// GetStats handles GET /stats, returning service statistics.
// Response:
//
//	200 OK with statistics - StatsResponse model.
//
// Errors:
//
//	500 Internal Server Error (INTERNAL_ERROR)
func (h *StatsHandler) GetStats(c *gin.Context) {
	stats, err := h.statsUC.GetStats(c.Request.Context())
	if err != nil {
		c.JSON(model.WriteErrorResponse(model.ErrCodeInternal))
		return
	}

	c.JSON(http.StatusOK, model.StatsFromDomain(stats))
}
