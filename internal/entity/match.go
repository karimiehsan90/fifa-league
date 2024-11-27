package entity

import "time"

type Match struct {
	ID        int64     `json:"id"`
	HomeTeam  string    `json:"home_team"`
	AwayTeam  string    `json:"away_team"`
	StartTime time.Time `json:"start_time"`
	End       bool      `json:"is_end"`
	HomeGoals int       `json:"home_goals"`
	AwayGoals int       `json:"away_goals"`
	Level     string    `json:"level"`
}

type LeagueMember struct {
	Team            string `json:"team"`
	Matches         int    `json:"natches"`
	Win             int    `json:"win"`
	Draw            int    `json:"draw"`
	Lose            int    `json:"lose"`
	GoalsFor        int    `json:"goals_for"`
	GoalsConceded   int    `json:"goals_conceded"`
	GoalsDifference int    `json:"goals_difference"`
	Points          int    `json:"points"`
}
