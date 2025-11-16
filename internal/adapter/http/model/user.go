package model

import "github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/domain"

// SetIsActiveRequest represents request body for POST /users/setIsActive
type SetIsActiveRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	IsActive *bool  `json:"is_active" binding:"required"`
}

// UserResponse represents user object in responses
type UserResponse struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

// UserFromDomain converts domain.User to UserResponse
// Note: teamName must be provided separately as domain.User doesn't contain it
func UserFromDomain(user *domain.User) UserResponse {
	return UserResponse{
		UserID:   user.ID,
		Username: user.Name,
		TeamName: user.TeamName,
		IsActive: user.IsActive,
	}
}

// SetIsActiveResponse represents response for POST /users/setIsActive
type SetIsActiveResponse struct {
	User UserResponse `json:"user"`
}
