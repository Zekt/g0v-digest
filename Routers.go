package main

import (
	"github.com/gorilla/mux"
	"net/http"
)

func RouteMedium(sub *mux.Router) {
	sub.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		//TODO: fetch all new Medium posts and store to DB.
	})
}

func RouteAPI(sub *mux.Router) {
	sub.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		//TODO: return JSON
	})
}
