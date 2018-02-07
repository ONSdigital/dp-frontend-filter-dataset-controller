package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/routes"
	"github.com/ONSdigital/go-ns/healthcheck"
	"github.com/ONSdigital/go-ns/identity"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/server"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
)

func main() {
	log.Namespace = "dp-frontend-filter-dataset-controller"
	cfg := config.Get()

	log.Debug("got service configuration", log.Data{"config": cfg})

	r := mux.NewRouter()

	alice := alice.New(identity.Handler(true)).Then(r)

	rend, fc, dc, clc, hc := routes.Init(r)

	s := server.New(cfg.BindAddr, alice)
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

		log.Debug("conducting service healthcheck", log.Data{
			"services": []string{
				"filter-api",
				"dataset-api",
				"code-list-api",
				"hierarchy-api",
				"renderer",
			},
		})

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
