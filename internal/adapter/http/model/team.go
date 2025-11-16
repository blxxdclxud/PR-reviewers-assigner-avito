package model

import "github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/domain"

// CreateTeamRequest represents request body for POST /team/add
type CreateTeamRequest struct {
	TeamName string       `json:"team_name" binding:"required"`
	Members  []TeamMember `json:"members" binding:"required,min=1"`
}

type TeamMember struct {
	UserID   string `json:"user_id" binding:"required"`
	Username string `json:"username" binding:"required"`
	IsActive bool   `json:"is_active"`
}

// ToDomain converts HTTP request to domain model
func (r *CreateTeamRequest) ToDomain() domain.Team {
	members := make([]domain.User, len(r.Members))
	for i, m := range r.Members {
		members[i] = domain.User{
			ID:       m.UserID,
			Name:     m.Username,
			IsActive: m.IsActive,
		}
	}

	return domain.Team{
		Name:    r.TeamName,
		Members: members,
	}
}

// TeamResponse represents response for team endpoints
type TeamResponse struct {
	TeamName string               `json:"team_name"`
	Members  []TeamMemberResponse `json:"members"`
}

type TeamMemberResponse struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

// TeamFromDomain converts domain.Team to TeamResponse
func TeamFromDomain(team *domain.Team) TeamResponse {
	members := make([]TeamMemberResponse, len(team.Members))
	for i, m := range team.Members {
		members[i] = TeamMemberResponse{
			UserID:   m.ID,
			Username: m.Name,
			IsActive: m.IsActive,
		}
	}

	return TeamResponse{
		TeamName: team.Name,
		Members:  members,
	}
}

// CreateTeamResponse represents response for POST /team/add
type CreateTeamResponse struct {
	Team TeamResponse `json:"team"`
}
