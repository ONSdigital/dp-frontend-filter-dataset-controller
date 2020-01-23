package main

import (
	"context"
	"os"
	"os/signal"
	"strconv"
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
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/server"
	"github.com/gorilla/mux"
)

// App version informaton retrieved on runtime
var (
	BuildTime, GitCommit, Version string
)

func main() {
	log.Namespace = "dp-frontend-filter-dataset-controller"

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)

	cfg, err := config.Get()
	if err != nil {
		log.Error(err, nil)
		os.Exit(1)
	}
	ctx := context.Background()

	log.InfoCtx(ctx, "got service configuration", log.Data{"config": cfg})

	var buildTime int64

	if BuildTime == "" {
		buildTime = 0
	}
	buildTime, err = strconv.ParseInt(BuildTime, 10, 64)
	if err != nil {
		log.Error(err, nil)
		os.Exit(1)
	}

	versionInfo := health.CreateVersionInfo(
		time.Unix(buildTime, 0),
		GitCommit,
		Version,
	)

	r := mux.NewRouter()

	clients := routes.Clients{
		Renderer:  renderer.New(cfg.RendererURL),
		Filter:    filter.New(cfg.FilterAPIURL),
		Dataset:   dataset.NewAPIClient(cfg.DatasetAPIURL),
		Hierarchy: hierarchy.New(cfg.HierarchyAPIURL),
		Search:    search.New(cfg.SearchAPIURL),
	}

	checkers := createCheckers(ctx, clients.Renderer, clients.Filter, clients.Dataset, clients.Hierarchy, clients.Search)

	criticalTimeout := time.Minute
	interval := 10 * time.Second
	healthcheck := health.Create(versionInfo, criticalTimeout, interval, checkers...)

	// Start healthcheck ticker
	healthcheck.Start(ctx)

	clients.Healthcheck = &healthcheck

	routes.Init(ctx, r, cfg, clients)

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

	<-signals

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	log.InfoCtx(ctx, "shutting service down gracefully", nil)
	defer cancel()

	// Stop healthcheck ticker
	healthcheck.Stop()

	if err := s.Server.Shutdown(ctx); err != nil {
		log.ErrorCtx(ctx, err, nil)
	}
}

func createCheckers(ctx context.Context, rend *renderer.Renderer, fc *filter.Client,
	dc *dataset.Client, hc *hierarchy.Client, sc *search.Client) []*health.Checker {

	rendChecker := health.Checker(func(ctx context.Context) (*health.Check, error) {
		return rend.Checker(ctx)
	})
	fcChecker := health.Checker(func(ctx context.Context) (*health.Check, error) {
		return fc.Checker(ctx)
	})
	dcChecker := health.Checker(func(ctx context.Context) (*health.Check, error) {
		return dc.Checker(ctx)
	})
	hcChecker := health.Checker(func(ctx context.Context) (*health.Check, error) {
		return hc.Checker(ctx)
	})
	scChecker := health.Checker(func(ctx context.Context) (*health.Check, error) {
		return sc.Checker(ctx)
	})

	return []*health.Checker{&rendChecker, &fcChecker, &dcChecker, &hcChecker, &scChecker}
}
