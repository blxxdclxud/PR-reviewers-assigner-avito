package model

import "net/http"

type ErrorCode string

const (
	ErrCodeTeamExists   ErrorCode = "TEAM_EXISTS"
	ErrCodePRExists     ErrorCode = "PR_EXISTS"
	ErrCodePRMerged     ErrorCode = "PR_MERGED"
	ErrCodeNotAssigned  ErrorCode = "NOT_ASSIGNED"
	ErrCodeNoCandidate  ErrorCode = "NO_CANDIDATE"
	ErrCodeNotFound     ErrorCode = "NOT_FOUND"
	ErrCodeInternal     ErrorCode = "INTERNAL_ERROR"
	ErrCodeInvalidInput ErrorCode = "INVALID_INPUT"
)

func WriteErrorResponse(code ErrorCode) (int, ErrorResponse) {
	switch code {
	case ErrCodeTeamExists:
		return http.StatusBadRequest, NewErrorResponse(code, "team_name already exists")
	case ErrCodePRExists:
		return http.StatusConflict, NewErrorResponse(code, "PR id already exists")
	case ErrCodePRMerged:
		return http.StatusConflict, NewErrorResponse(code, "cannot reassign on merged PR")
	case ErrCodeNotAssigned:
		return http.StatusConflict, NewErrorResponse(code, "reviewer is not assigned to this PR")
	case ErrCodeNoCandidate:
		return http.StatusConflict, NewErrorResponse(code, "no active replacement candidate in team")
	case ErrCodeNotFound:
		return http.StatusNotFound, NewErrorResponse(code, "resource not found")
	case ErrCodeInvalidInput:
		return http.StatusBadRequest, NewErrorResponse(code, "invalid input data")
	case ErrCodeInternal:
		return http.StatusInternalServerError, NewErrorResponse(code, "internal server error")
	default:
		return http.StatusInternalServerError, NewErrorResponse(ErrCodeInternal, "internal server error")
	}
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

// NewErrorResponse creates a new error response
func NewErrorResponse(code ErrorCode, message string) ErrorResponse {
	return ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	}
}
