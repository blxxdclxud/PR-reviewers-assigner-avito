package domain

import "time"

type PullRequest struct {
	ID                string   `json:"pull_request_id"`
	Name              string   `json:"pull_request_name"`
	AuthorID          string   `json:"author_id"`
	Status            PRStatus `json:"status"`
	Reviewers         []User   `json:"assigned_reviewers"`
	NeedMoreReviewers bool
	CreatedAt         time.Time `json:"createdAt"`
	MergedAt          time.Time `json:"mergedAt"`
}

type PRStatus string

const StatusOpen = "OPEN"
const StatusMerged = "MERGED"
