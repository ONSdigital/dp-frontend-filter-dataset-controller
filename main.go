package main

import (
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/handlers"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/renderer"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/server"
	"github.com/gorilla/mux"
)

func main() {
	log.Namespace = "dp-frontend-filter-dataset-controller"
	cfg := config.Get()

	r := mux.NewRouter()

	rend := renderer.New()
	filter := handlers.NewFilter(rend)

	r.Path("/jobs/{jobID}").Methods("GET").HandlerFunc(filter.PreviewPage)
	r.Path("/jobs/{jobID}/dimensions").Methods("GET").HandlerFunc(filter.FilterOverview)
	r.Path("/jobs/{jobID}/dimensions/age-range").Methods("GET").HandlerFunc(filter.AgeSelectorRange)
	r.Path("/jobs/{jobID}/dimensions/age-list").Methods("GET").HandlerFunc(filter.AgeSelectorList)
	r.Path("/jobs/{jobID}/dimensions/geography").Methods("GET").HandlerFunc(filter.Geography)

	s := server.New(cfg.BindAddr, r)

	log.Debug("listening...", log.Data{
		"bind_address": cfg.BindAddr,
	})

	if err := s.ListenAndServe(); err != nil {
		log.Error(err, nil)
		return
	}
}
