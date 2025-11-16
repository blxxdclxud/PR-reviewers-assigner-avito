package handler

import (
	"net/http"

	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/adapter/http/model"
)

func WriteErrorResponse(code model.ErrorCode) (int, model.ErrorResponse) {
	switch code {
	case model.ErrCodeTeamExists:
		return http.StatusBadRequest, model.NewErrorResponse(code, "team_name already exists")
	case model.ErrCodePRExists:
		return http.StatusConflict, model.NewErrorResponse(code, "PR id already exists")
	case model.ErrCodePRMerged:
		return http.StatusConflict, model.NewErrorResponse(code, "cannot reassign on merged PR")
	case model.ErrCodeNotAssigned:
		return http.StatusConflict, model.NewErrorResponse(code, "reviewer is not assigned to this PR")
	case model.ErrCodeNoCandidate:
		return http.StatusConflict, model.NewErrorResponse(code, "no active replacement candidate in team")
	case model.ErrCodeNotFound:
		return http.StatusNotFound, model.NewErrorResponse(code, "resource not found")
	case model.ErrCodeInvalidInput:
		return http.StatusBadRequest, model.NewErrorResponse(code, "invalid input data")
	case model.ErrCodeInternal:
		return http.StatusInternalServerError, model.NewErrorResponse(code, "internal server error")
	default:
		return http.StatusInternalServerError, model.NewErrorResponse(model.ErrCodeInternal, "internal server error")
	}
}
