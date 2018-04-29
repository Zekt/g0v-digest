package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"os/signal"
)

var config struct {
	Server  string `json:"server"`
	Port    int    `json:"port"`
	SQLHost string `json:"sqlHost"`
	SQLPort int    `json:"sqlPort"`
	SQLUser string `json:"sqlUser"`
	SQLPass string `json:"sqlPass"`
	DBName  string `json:"dbname"`
	RssUrl  string `json:"rssUrl"`
}

func main() {
	configFile, err := os.Open("config.json")
	if err != nil {
		log.Fatal("opening config file: ", err.Error())
	}

	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&config); err != nil {
		log.Fatal("parsing config file: ", err.Error())
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=\"%s\" dbname=%s sslmode=disable",
		config.SQLHost, config.SQLPort, config.SQLUser,
		config.SQLPass, config.DBName,
	)

	DB, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("connecting database: ", err.Error())
	}
	defer DB.Close()

	router := mux.NewRouter()
	srv := &http.Server{
		Addr:    config.Server + ":" + fmt.Sprint(config.Port),
		Handler: router,
	}

	mediumSub := router.PathPrefix("/medium").Methods("PUT").Subrouter()
	apiSub := router.PathPrefix("/api").Subrouter()
	RouteMedium(mediumSub)
	RouteAPI(apiSub)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal("running server: ", err.Error())
		}
	}()

	c := make(chan os.Signal, 1)
	// accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	log.Println("shutting down")
	os.Exit(0)
}
