package main

import (
	"learn-gin/handlers/home"
	"learn-gin/pkg/router"
	"net/http"
)

func main() {
	r := router.NewRouter()

	homeHandler := home.NewHandler()

	router.HTMLGet(r, "/", homeHandler.Index)
	router.HTMLGet(r, "/ds/{datasetId}", homeHandler.Index)

	if err := http.ListenAndServe(":8081", r.Mux()); err != nil {
		panic(err)
	}
}
