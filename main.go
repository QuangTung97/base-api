package main

import (
	"learn-gin/handlers/home"
	"learn-gin/pkg/router"
	"learn-gin/pkg/urls"
	"net/http"
)

var homePath = urls.NewEmpty("/")

type datasetParams struct {
	DatasetID int64 `json:"dataset_id"`
}

var datasetPath = urls.New[datasetParams]("/ds/{dataset_id}")

func main() {
	r := router.NewRouter()

	homeHandler := home.NewHandler()

	router.HTMLGet(r, homePath, homeHandler.Index)
	router.HTMLGet(r, datasetPath, homeHandler.GetDataset)

	if err := http.ListenAndServe(":8081", r.Mux()); err != nil {
		panic(err)
	}
}
