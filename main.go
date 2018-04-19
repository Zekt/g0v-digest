package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"os/signal"
)

var config struct {
	Server string `json:"server"`
	Port   int    `json:"port"`
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
