package main

import (
	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-dataset-controller/handlers"
	"github.com/ONSdigital/dp-frontend-dataset-controller/renderer"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/server"
	"github.com/gorilla/mux"
)

func main() {
	log.Namespace = "dp-frontend-dataset-controller"
	cfg := config.Get()

	r := mux.NewRouter()

	rend := renderer.New()
	cmd := handlers.NewCMD(rend)

	r.Path("/dataset/cmd").Methods("GET").HandlerFunc(cmd.Landing)
	r.Path("/dataset/cmd/middle").Methods("POST").HandlerFunc(cmd.CreateJobID)
	r.Path("/dataset/cmd/{jobID}").Methods("GET").HandlerFunc(cmd.Middle)
	r.Path("/dataset/cmd/{jobID}/finish").Methods("GET").HandlerFunc(cmd.PreviewAndDownload)

	s := server.New(cfg.BindAddr, r)

	log.Debug("listening...", log.Data{
		"bind_address": cfg.BindAddr,
	})

	if err := s.ListenAndServe(); err != nil {
		log.Error(err, nil)
		return
	}
}
