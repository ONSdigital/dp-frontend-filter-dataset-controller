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
	cmd := handlers.NewCMD(rend)

	r.Path("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}").Methods("GET").HandlerFunc(cmd.Landing)
	r.Path("/datasets/{datasetID}/editions/{editionID}/versions/{versionID}/filter").Methods("POST").HandlerFunc(cmd.CreateJobID)
	r.Path("/jobs/{jobID}").Methods("GET").HandlerFunc(cmd.Middle)
	r.Path("/jobs/{jobID}/dimensions").Methods("GET").HandlerFunc(cmd.PreviewAndDownload)

	s := server.New(cfg.BindAddr, r)

	log.Debug("listening...", log.Data{
		"bind_address": cfg.BindAddr,
	})

	if err := s.ListenAndServe(); err != nil {
		log.Error(err, nil)
		return
	}
}
