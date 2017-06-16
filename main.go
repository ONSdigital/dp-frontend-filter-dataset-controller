package main

import (
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-dataset-controller/handlers"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/server"
	"github.com/gorilla/mux"
)

func main() {
	log.Namespace = "dp-frontend-dataset-controller"
	cfg := config.Get()

	r := mux.NewRouter()

	r.Path("/dataset/cmd").Methods("GET").HandlerFunc(handlers.Landing)
	r.Path("/dataset/cmd/middle").Methods("POST").HandlerFunc(handlers.CreateJobID)
	r.Path("/dataset/cmd/{jobID}").Methods("GET").HandlerFunc(handlers.Middle)
	r.Path("/dataset/cmd/{jobID}/finish").Methods("GET").HandlerFunc(handlers.PreviewAndDownload)

	s := server.New(cfg.BindAddr, r)

	log.Debug("listening...", log.Data{
		"bind_address": cfg.BindAddr,
	})

	if err := s.ListenAndServe(); err != nil {
		log.Error(err, nil)
		return
	}
}
