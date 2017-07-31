package main

import (
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/dataset"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/filter"
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
	fc := filter.New(cfg.FilterAPIURL)
	dc := dataset.New(cfg.DatasetAPIURL)
	filter := handlers.NewFilter(rend, fc, dc)

	r.Path("/filters/{filterID}").Methods("GET").HandlerFunc(filter.PreviewPage)
	r.Path("/filters/{filterID}/dimensions").Methods("GET").HandlerFunc(filter.FilterOverview)
	r.Path("/filters/{filterID}/dimensions/age-range").Methods("GET").HandlerFunc(filter.RangeSelector)
	r.Path("/filters/{filterID}/dimensions/age-list").Methods("GET").HandlerFunc(filter.ListSelector)
	r.Path("/filters/{filterID}/dimensions/geography").Methods("GET").HandlerFunc(filter.Geography)

	s := server.New(cfg.BindAddr, r)

	log.Debug("listening...", log.Data{
		"bind_address": cfg.BindAddr,
	})

	if err := s.ListenAndServe(); err != nil {
		log.Error(err, nil)
		return
	}
}
