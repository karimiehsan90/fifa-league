package entity

import "time"

type Match struct {
	ID               int64
	HomeTeam         string
	AwayTeam         string
	StartTime        time.Time
	End              bool
	HomeGoals        int
	AwayGoals        int
	Level            string
}

type LeagueMember struct {
	Team string
	Matches int
	Win int
	Draw int
	Lose int
	GoalsFor int
	GoalsConceded int
	GoalsDifference int
	Points int
}
