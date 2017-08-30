package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/routes"
	"github.com/ONSdigital/go-ns/healthcheck"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/server"
	"github.com/gorilla/mux"
)

func main() {
	log.Namespace = "dp-frontend-filter-dataset-controller"
	cfg := config.Get()

	r := mux.NewRouter()

	rend, fc, dc, clc, hc := routes.Init(r)

	s := server.New(cfg.BindAddr, r)
	s.HandleOSSignals = false

	log.Debug("listening...", log.Data{
		"bind_address": cfg.BindAddr,
	})

	go func() {
		if err := s.ListenAndServe(); err != nil {
			log.Error(err, nil)
			return
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill)

	for {
		healthcheck.MonitorExternal(fc, dc, clc, hc, rend)

		timer := time.NewTimer(time.Second * 60)

		select {
		case <-timer.C:
			continue
		case <-stop:
			log.Info("shutting service down gracefully", nil)
			timer.Stop()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := s.Server.Shutdown(ctx); err != nil {
				log.Error(err, nil)
			}
			return
		}
	}
}
