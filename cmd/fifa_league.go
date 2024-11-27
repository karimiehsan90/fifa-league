package main

import (
	"flag"
	"log"
	"net/http"

	"ehskarimi.ir/fifa_league/internal/dao"
	"ehskarimi.ir/fifa_league/internal/handler"
	_ "github.com/lib/pq"
)

func main() {
	webListenAddress := flag.String("web-listen-address", ":8091", "Web listen address")
	dbHost := flag.String("db-host", "localhost", "Database Host")
	dbPort := flag.Int("db-port", 5432, "Database Port")
	dbName := flag.String("db-name", "", "Database Name")
	dbUser := flag.String("db-user", "", "Database User")
	dbPass := flag.String("db-pass", "", "Database Pass")
	rocketchatURL := flag.String("rocketchat-url", "", "Rocketchat URL")
	flag.Parse()

	matchDAO := &dao.MatchDAO{}
	err := matchDAO.MustInstantiateDB(*dbUser, *dbPass, *dbHost, *dbName, *dbPort)
	if err != nil {
		log.Fatalf("unable to instantiate database: %v", err)
	}
	err = matchDAO.CreateSchemas()
	if err != nil {
		log.Fatalf("unable to create schemas: %v", err)
	}

	uiHandler := &handler.UIHandler{
		DAO:           matchDAO,
		RocketchatURL: *rocketchatURL,
	}

	getAllMatchesHandler := &handler.GetAllMatchesHandler{
		DAO: matchDAO,
	}
	getTableHandler := &handler.GetTableHandler{
		DAO: matchDAO,
	}
	http.Handle("/", uiHandler)
	http.Handle("/api/v1/matches", getAllMatchesHandler)
	http.Handle("/api/v1/table", getTableHandler)
	panic(http.ListenAndServe(*webListenAddress, nil))
}
