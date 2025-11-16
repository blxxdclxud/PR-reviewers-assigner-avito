package usecase

import (
	"context"
	"database/sql"
	"errors"

	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/domain"
	"github.com/blxxdclxud/PR-reviewers-assigner-avito/internal/repository"
)

type TeamUseCase struct {
	teamRepo repository.TeamRepository
	userRepo repository.UserRepository
	db       *sql.DB
}

func NewTeamUseCase(
	teamRepo repository.TeamRepository,
	userRepo repository.UserRepository,
	db *sql.DB) *TeamUseCase {
	return &TeamUseCase{
		teamRepo: teamRepo,
		userRepo: userRepo,
		db:       db,
	}
}

// CreateTeam creates a new team with the specified name and members.
// For each member: if user exists, updates their data; if user doesn't exist, creates a new user.
//
// Returns:
//   - *domain.Team: created team with assigned ID and list of members
//   - error: domain.ErrTeamExists if team name already exists, or any database error
func (u *TeamUseCase) CreateTeam(ctx context.Context, team domain.Team) (*domain.Team, error) {
	// Start transaction
	tx, err := u.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() //nolint:errcheck

	// Create team itself, created team ID is assigned to team.ID inside
	err = u.teamRepo.Create(ctx, tx, &team)
	if err != nil {
		if errors.Is(err, domain.ErrTeamExists) {
			return nil, err
		}
		return nil, err
	}

	// Add members to the team (updating existing ones and creating new ones)
	err = u.addTeamMembers(ctx, tx, &team)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &team, nil
}

// addTeamMembers creates or updates users and assigns them to the specified team.
// This is a helper method.
//
// Returns:
//   - error: any database error during user creation or update
func (u *TeamUseCase) addTeamMembers(ctx context.Context, tx *sql.Tx, team *domain.Team) error {
	// Iterate through each member and add/update it
	for i := range team.Members {
		member := &team.Members[i]
		// Update teamID field for each user
		member.TeamID = team.ID

		// Try to get user if it exists
		_, err := u.userRepo.GetByID(ctx, member.ID)
		if err != nil {
			// If user not created yet - create it
			if errors.Is(err, domain.ErrNotFound) {
				err = u.userRepo.Create(ctx, tx, member)
				if err != nil {
					return err
				}
				continue
			}
			return err
		}

		// User exists - update
		err = u.userRepo.Update(ctx, tx, member)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetTeam retrieves a team by name with all its members.
//
// Returns:
//   - *domain.Team: team object with members
//   - error: domain.ErrNotFound if team doesn't exist, or any database error
func (u *TeamUseCase) GetTeam(ctx context.Context, teamName string) (*domain.Team, error) {
	team, err := u.teamRepo.GetByName(ctx, teamName)
	if err != nil {
		return nil, err
	}

	return team, nil
}
