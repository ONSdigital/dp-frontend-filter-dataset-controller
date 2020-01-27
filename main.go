package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-api-clients-go/filter"
	"github.com/ONSdigital/dp-api-clients-go/hierarchy"
	"github.com/ONSdigital/dp-api-clients-go/renderer"
	"github.com/ONSdigital/dp-api-clients-go/search"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/config"
	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/routes"
	health "github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/go-ns/handlers/collectionID"
	"github.com/ONSdigital/go-ns/server"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
)

// App version informaton retrieved on runtime
var (
	// BuildTime represents the time in which the service was built
	BuildTime string
	// GitCommit represents the commit (SHA-1) hash of the service that is running
	GitCommit string
	// Version represents the version of the service that is running
	Version string
)

const (
	criticalTimeout = time.Minute
	interval        = 10 * time.Second
)

func main() {
	log.Namespace = "dp-frontend-filter-dataset-controller"

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)

	ctx := context.Background()

	cfg, err := config.Get()
	if err != nil {
		log.Event(ctx, "unable to retrieve service configuration", log.Error(err))
		os.Exit(1)
	}

	log.Event(ctx, "got service configuration", log.Data{"config": cfg})

	versionInfo, err := health.CreateVersionInfo(
		BuildTime,
		GitCommit,
		Version,
	)
	if err != nil {
		log.Event(ctx, "failed to create service version information", log.Error(err))
		os.Exit(1)
	}

	r := mux.NewRouter()

	clients := routes.Clients{
		Renderer:  renderer.New(cfg.RendererURL),
		Filter:    filter.New(cfg.FilterAPIURL),
		Dataset:   dataset.NewAPIClient(cfg.DatasetAPIURL),
		Hierarchy: hierarchy.New(cfg.HierarchyAPIURL),
		Search:    search.New(cfg.SearchAPIURL),
	}

	healthcheck, err := health.Create(versionInfo, criticalTimeout, interval,
		clients.Renderer.Checker, clients.Filter.Checker, clients.Dataset.Checker,
		clients.Hierarchy.Checker, clients.Search.Checker)
	if err != nil {
		log.Event(ctx, "failed to create service health checks", log.Error(err))
		os.Exit(1)
	}

	clients.Healthcheck = &healthcheck

	routes.Init(ctx, r, cfg, clients)

	// Start healthcheck ticker
	healthcheck.Start(ctx)

	s := server.New(cfg.BindAddr, r)
	s.HandleOSSignals = false

	s.Middleware["CollectionID"] = collectionID.CheckCookie
	s.MiddlewareOrder = append(s.MiddlewareOrder, "CollectionID")

	log.Event(ctx, "service listening...", log.Data{
		"bind_address": cfg.BindAddr,
	})

	go func() {
		if err := s.ListenAndServe(); err != nil {
			log.Event(ctx, "failed to start http listen and serve", log.Error(err))
			return
		}
	}()

	<-signals

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	log.Event(ctx, "shutting service down gracefully")
	defer cancel()

	// Stop healthcheck ticker
	healthcheck.Stop()

	if err := s.Server.Shutdown(ctx); err != nil {
		log.Event(ctx, "failed to shutdown http server", log.Error(err))
	}
}
