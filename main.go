package main

import (
	"os"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/codelist"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/dataset"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/filter"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/handlers"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/renderer"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/server"
	"github.com/ONSdigital/go-ns/validator"
	"github.com/gorilla/mux"
)

func main() {
	log.Namespace = "dp-frontend-filter-dataset-controller"
	cfg := config.Get()

	r := mux.NewRouter()

	fi, err := os.Open("rules.json")
	if err != nil {
		log.ErrorC("could not open rules for validation", err, nil)
	}
	defer fi.Close()

	v, err := validator.New(fi)
	if err != nil {
		log.ErrorC("failed to crare form validator", err, nil)
	}

	rend := renderer.New()
	fc := filter.New(cfg.FilterAPIURL)
	dc := dataset.New(cfg.DatasetAPIURL)
	clc := codelist.New(cfg.CodeListAPIURL)
	filter := handlers.NewFilter(rend, fc, dc, clc, v)

	r.Path("/filters/{filterID}").Methods("GET").HandlerFunc(filter.PreviewPage)
	r.Path("/filters/{filterID}/dimensions").Methods("GET").HandlerFunc(filter.FilterOverview)
	r.Path("/filters/{filterID}/dimensions/age-range").Methods("GET").HandlerFunc(filter.RangeSelector)
	r.Path("/filters/{filterID}/dimensions/age-list").Methods("GET").HandlerFunc(filter.ListSelector)
	r.Path("/filters/{filterID}/dimensions/geography").Methods("GET").HandlerFunc(filter.Geography)

	r.Path("/filters/{filterID}/dimensions/{name}").Methods("GET").HandlerFunc(filter.RangeSelector)
	r.Path("/filters/{filterID}/dimensions/{name}/range-add").Methods("POST").HandlerFunc(filter.AddRange)
	r.Path("/filters/{filterID}/dimensions/{name}/range-delete").Methods("POST").HandlerFunc(filter.RemoveRange)

	s := server.New(cfg.BindAddr, r)

	log.Debug("listening...", log.Data{
		"bind_address": cfg.BindAddr,
	})

	if err := s.ListenAndServe(); err != nil {
		log.Error(err, nil)
		return
	}
}
