package domain

type UserReviewStats struct {
	UserID      string
	Username    string
	ReviewCount int
}

type Stats struct {
	TotalTeams int
	TotalUsers int
	TotalPRs   int
	OpenPRs    int
	MergedPRs  int
	Reviewers  []UserReviewStats
}
