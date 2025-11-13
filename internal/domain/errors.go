package domain

import "errors"

var (
	ErrTeamExists  = errors.New("team already exists")
	ErrPRExists    = errors.New("pull request already exists")
	ErrPRMerged    = errors.New("cannot modify merged pull request")
	ErrNotAssigned = errors.New("reviewer not assigned")
	ErrNoCandidate = errors.New("no candidate available")
	ErrNotFound    = errors.New("resource not found")
)
