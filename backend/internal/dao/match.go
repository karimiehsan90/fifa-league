package dao

import (
	"database/sql"
	"fmt"
	"time"

	"ehskarimi.ir/fifa_league/internal/entity"
)

type MatchDAO struct {
	DB *sql.DB
}

func (d *MatchDAO) MustInstantiateDB(pgUser, pgPass, pgHost, pgDB string, pgPort int) error {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", pgUser, pgPass, pgHost, pgPort, pgDB)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	d.DB = db
	return nil
}

func (d *MatchDAO) CreateSchemas() error {
	matchTableSQLStatement := `
	CREATE TABLE IF NOT EXISTS match (
		match_id serial primary key,
		home varchar(63),
		home_from int references match(match_id),
		home_goals int,
		away varchar(63),
		away_from int references match(match_id),
		away_goals int,
		is_end bool,
		start_time int,
		level varchar(63)
	)
	`
	return d.CreateTable(matchTableSQLStatement)
}

func (d *MatchDAO) CreateTable(sqlStatement string) error {
	_, err := d.DB.Exec(sqlStatement)
	return err
}

func (d *MatchDAO) GetAllMatches(levelFilter string) ([]*entity.Match, error) {
	sqlStatement := `
	select match_id, home, home_from, home_goals, away, away_from, away_goals, is_end, start_time, level
	from match where level like $1 order by start_time desc
	`
	rows, err := d.DB.Query(sqlStatement, levelFilter)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	matches := make([]*entity.Match, 0)
	for rows.Next() {
		var matchID int64
		var homeTeam string
		var homeFrom int64
		var homeGoals int64
		var awayTeam string
		var awayFrom int64
		var awayGoals int64
		var isEnd bool
		var startTime int64
		var level string
		rows.Scan(&matchID, &homeTeam, &homeFrom, &homeGoals, &awayTeam, &awayFrom, &awayGoals, &isEnd, &startTime, &level)
		match := entity.Match{
			ID:        matchID,
			HomeTeam:  homeTeam,
			HomeFrom:  int(homeFrom),
			HomeGoals: int(homeGoals),
			AwayTeam:  awayTeam,
			AwayFrom:  int(awayFrom),
			AwayGoals: int(awayGoals),
			End:       isEnd,
			StartTime: time.Unix(int64(startTime), 0),
			Level:     level,
		}
		matches = append(matches, &match)
	}
	return matches, nil
}
