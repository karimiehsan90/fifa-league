package handler

import (
	"encoding/json"
	"log"
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

type GetAllMatchesHandler struct {
	DAO *dao.MatchDAO
}

type GetTableHandler struct {
	DAO *dao.MatchDAO
}

func (h *UIHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	indexTpl, err := template.New("index.html").Funcs(getFuncMap()).ParseFiles("index.html")
	matches, err := h.DAO.GetAllMatches("%")
	levelMatches := map[string][]*entity.Match{
		"final": make([]*entity.Match, 0),
		"1/2":   make([]*entity.Match, 0),
		"1/4":   make([]*entity.Match, 0),
		"1/8":   make([]*entity.Match, 0),
		"1/16":  make([]*entity.Match, 0),
		"Group": make([]*entity.Match, 0),
	}

	for _, match := range matches {
		allMatchesOfLevel := levelMatches[match.Level]
		allMatchesOfLevel = append(allMatchesOfLevel, match)
		levelMatches[match.Level] = allMatchesOfLevel
	}

	levelMatches["1/2"] = sortMatches(levelMatches["1/2"], levelMatches["final"])
	levelMatches["1/4"] = sortMatches(levelMatches["1/4"], levelMatches["1/2"])
	levelMatches["1/8"] = sortMatches(levelMatches["1/8"], levelMatches["1/4"])
	levelMatches["1/16"] = sortMatches(levelMatches["1/16"], levelMatches["1/8"])

	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
	}
	indexTpl.ExecuteTemplate(writer, "index.html", map[string]interface{}{
		"matches":       matches,
		"table":         createTable(matches),
		"rocketchatURL": h.RocketchatURL,
		"levelMatches":  levelMatches,
	})
}

func (h *GetAllMatchesHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	matches, err := h.DAO.GetAllMatches("%")
	result := make(map[string]interface{})
	if err != nil {
		result["success"] = false
		result["error"] = err.Error()
		writer.WriteHeader(http.StatusInternalServerError)
		log.Printf("unable to get list of matches: %v\n", err)
	} else {
		result["success"] = true
		result["error"] = ""
		result["data"] = matches
	}
	sendJSONResponse(writer, result)
}

func (h *GetTableHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	matches, err := h.DAO.GetAllMatches("Group")
	result := make(map[string]interface{})
	if err != nil {
		result["success"] = false
		result["error"] = err.Error()
		writer.WriteHeader(http.StatusInternalServerError)
		log.Printf("unable to get list of matches: %v\n", err)
	} else {
		result["success"] = true
		result["error"] = ""
		result["data"] = createTable(matches)
	}
	sendJSONResponse(writer, result)
}

func sendJSONResponse(writer http.ResponseWriter, data map[string]interface{}) {
	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(data)
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
		"default": func(a, b string) string {
			if b != "" {
				return b
			}
			return a
		},
	}
}

func createTable(matches []*entity.Match) []entity.LeagueMember {
	allTeamsRanks := make(map[string]entity.LeagueMember)
	for _, match := range matches {
		if match.Level == "Group" {
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
			if match.End && match.Level == "Group" {
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
			allTeamsRanks[match.AwayTeam] = awayRank
			allTeamsRanks[match.HomeTeam] = homeRank
		}
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

func sortMatches(toBeSorted, sortFrom []*entity.Match) []*entity.Match {
	result := make([]*entity.Match, 0)
	for _, match := range sortFrom {
		homeMatch := getMatchByID(toBeSorted, match.HomeFrom)
		if homeMatch != nil {
			result = append(result, homeMatch)
		}
		awayMatch := getMatchByID(toBeSorted, match.AwayFrom)
		if awayMatch != nil {
			result = append(result, awayMatch)
		}
	}
	return result
}

func getMatchByID(matches []*entity.Match, id int) *entity.Match {
	for _, match := range matches {
		if int(match.ID) == id {
			return match
		}
	}
	return nil
}
