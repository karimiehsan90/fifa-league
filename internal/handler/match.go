package handler

import (
	"net/http"
	"sort"
	"text/template"
	"time"

	"ehskarimi.ir/fifa_league/internal/dao"
	"ehskarimi.ir/fifa_league/internal/entity"
	ptime "github.com/yaa110/go-persian-calendar"
)

type UIHandler struct {
	DAO           *dao.MatchDAO
	RocketchatURL string
}

func (h *UIHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	indexTpl, err := template.New("index.html").Funcs(getFuncMap()).ParseFiles("index.html")
	matches, err := h.DAO.GetAllMatches()
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
	}
	indexTpl.ExecuteTemplate(writer, "index.html", map[string]interface{}{
		"matches":       matches,
		"table":         createTable(matches),
		"rocketchatURL": h.RocketchatURL,
	})
}

func getFuncMap() template.FuncMap {
	return template.FuncMap{
		"toPersianTime": func(t time.Time) string {
			if t.Year() < 2024 {
				return "مشخص نشده"
			}
			return ptime.New(t).Format("E d MMM HH:mm")
		},
		"add": func(a, b int) int {
			return a + b
		},
		"started": func(m entity.Match) bool {
			return time.Now().After(m.StartTime)
		},
	}
}

func createTable(matches []*entity.Match) []entity.LeagueMember {
	allTeamsRanks := make(map[string]entity.LeagueMember)
	for _, match := range matches {
		homeRank, ok := allTeamsRanks[match.HomeTeam]
		if !ok {
			homeRank = entity.LeagueMember{
				Team: match.HomeTeam,
			}
		}
		awayRank, ok := allTeamsRanks[match.AwayTeam]
		if !ok {
			awayRank = entity.LeagueMember{
				Team: match.AwayTeam,
			}
		}
		if match.End {
			// matches
			homeRank.Matches += 1
			awayRank.Matches += 1
			// goals for and goals conceded
			homeRank.GoalsFor += match.HomeGoals
			homeRank.GoalsConceded += match.AwayGoals
			awayRank.GoalsFor += match.AwayGoals
			awayRank.GoalsConceded += match.HomeGoals
			// goals difference
			homeRank.GoalsDifference += (match.HomeGoals - match.AwayGoals)
			awayRank.GoalsDifference -= (match.HomeGoals - match.AwayGoals)
			// points and w/d/l
			if match.HomeGoals > match.AwayGoals {
				homeRank.Win += 1
				awayRank.Lose += 1
				homeRank.Points += 3
			} else if match.HomeGoals == match.AwayGoals {
				homeRank.Draw += 1
				awayRank.Draw += 1
				homeRank.Points += 1
				awayRank.Points += 1
			} else {
				homeRank.Lose += 1
				awayRank.Win += 1
				awayRank.Points += 3
			}
		}
		allTeamsRanks[match.HomeTeam] = homeRank
		allTeamsRanks[match.AwayTeam] = awayRank
	}
	result := make([]entity.LeagueMember, 0)
	for _, leagueMember := range allTeamsRanks {
		result = append(result, leagueMember)
	}
	return sortTable(result)
}

func sortTable(members []entity.LeagueMember) []entity.LeagueMember {
	sort.Slice(members, func(i, j int) bool {
		a := members[i]
		b := members[j]
		if a.Points != b.Points {
			return a.Points > b.Points
		}
		if a.GoalsDifference != b.GoalsDifference {
			return a.GoalsDifference > b.GoalsDifference
		}
		if a.GoalsFor != b.GoalsFor {
			return a.GoalsFor > b.GoalsFor
		}
		return a.Team < b.Team
	})
	return members
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
