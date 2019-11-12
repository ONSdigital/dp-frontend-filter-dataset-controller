package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/routes"
	"github.com/ONSdigital/go-ns/handlers/collectionID"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/server"
	"github.com/gorilla/mux"
)

func main() {
	log.Namespace = "dp-frontend-filter-dataset-controller"
	cfg, err := config.Get()
	if err != nil {
		log.Error(err, nil)
		os.Exit(1)
	}
	ctx := context.Background()

	log.InfoCtx(ctx, "got service configuration", log.Data{"config": cfg})

	r := mux.NewRouter()

	routes.Init(r, cfg)

	s := server.New(cfg.BindAddr, r)
	s.HandleOSSignals = false

	s.Middleware["CollectionID"] = collectionID.CheckCookie
	s.MiddlewareOrder = append(s.MiddlewareOrder, "CollectionID")

	log.InfoCtx(ctx, "listening...", log.Data{
		"bind_address": cfg.BindAddr,
	})

	go func() {
		if err := s.ListenAndServe(); err != nil {
			log.ErrorCtx(ctx, err, nil)
			return
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill)

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	log.InfoCtx(ctx, "shutting service down gracefully", nil)
	defer cancel()
	if err := s.Server.Shutdown(ctx); err != nil {
		log.ErrorCtx(ctx, err, nil)
	}
}
