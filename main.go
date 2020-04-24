package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

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
	"github.com/pkg/errors"
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

func main() {
	log.Namespace = "dp-frontend-filter-dataset-controller"
	ctx := context.Background()

	if err := run(ctx); err != nil {
		log.Event(ctx, "application unexpectedly failed", log.ERROR, log.Error(err))
		os.Exit(1)
	}

	os.Exit(0)
}

func run(ctx context.Context) error {
	defer func() {
		if iface := recover(); iface != nil {
			switch val := iface.(type) {
			case error:
				log.Event(ctx, "++++++ David and Nathan look here - we have caught our unexpecting panic, with error", log.ERROR, log.Error(val))
			default:
				log.Event(ctx, "++++++ David and Nathan look here - we have caught our unexpecting panic, with generic interface", log.ERROR, log.Data{"value": fmt.Sprintf("%#v", val)})
			}
		}
		log.Event(ctx, "++++++ David and Nathan look here - it is NOT panicking", log.INFO)
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)

	cfg, err := config.Get()
	if err != nil {
		log.Event(ctx, "unable to retrieve service configuration", log.FATAL, log.Error(err))
		return err
	}

	log.Event(ctx, "got service configuration", log.INFO, log.Data{"config": cfg})

	versionInfo, err := health.NewVersionInfo(
		BuildTime,
		GitCommit,
		Version,
	)
	if err != nil {
		log.Event(ctx, "failed to create service version information", log.ERROR, log.Error(err))
		return err
	}

	r := mux.NewRouter()

	clients := routes.Clients{
		Renderer:  renderer.New(cfg.RendererURL),
		Filter:    filter.New(cfg.FilterAPIURL),
		Dataset:   dataset.NewAPIClient(cfg.DatasetAPIURL),
		Hierarchy: hierarchy.New(cfg.HierarchyAPIURL),
		Search:    search.New(cfg.SearchAPIURL),
	}

	healthcheck := health.New(versionInfo, cfg.HealthCheckCriticalTimeout, cfg.HealthCheckInterval)
	clients.Healthcheck = &healthcheck

	if err = registerCheckers(ctx, clients); err != nil {
		return err
	}

	routes.Init(ctx, r, cfg, clients)

	s := server.New(cfg.BindAddr, r)
	s.HandleOSSignals = false

	s.Middleware["CollectionID"] = collectionID.CheckCookie
	s.MiddlewareOrder = append(s.MiddlewareOrder, "CollectionID")

	log.Event(ctx, "service listening...", log.Data{
		"bind_address": cfg.BindAddr,
	})

	svcErrors := make(chan error, 1)
	go func() {
		if err := s.ListenAndServe(); err != nil {
			svcErrors <- errors.Wrap(err, "failure in http listen and serve")
		}
	}()

	// Start healthcheck ticker
	healthcheck.Start(ctx)

	// Block until a fatal error occurs
	select {
	case err := <-svcErrors:
		log.Event(ctx, "service error received", log.ERROR, log.Error(err))
	case signal := <-signals:
		log.Event(ctx, "quitting after os signal received", log.INFO, log.Data{"signal": signal})
	}

	log.Event(ctx, fmt.Sprintf("shutdown with timeout: %s", cfg.GracefulShutdownTimeout), log.INFO)

	// give the app `Timeout` seconds to close gracefully before killing it.
	ctx, cancel := context.WithTimeout(context.Background(), cfg.GracefulShutdownTimeout)

	var gracefulShutdown bool

	go func() {
		defer cancel()
		hasShutdownErrs := false

		log.Event(ctx, "stop health checkers", log.INFO)
		healthcheck.Stop()

		if err := s.Shutdown(ctx); err != nil {
			log.Event(ctx, "failed to gracefully shutdown http server", log.ERROR, log.Error(err))
			hasShutdownErrs = true
		}

		if !hasShutdownErrs {
			gracefulShutdown = true
		}
	}()

	// wait for timeout or success (via cancel)
	<-ctx.Done()
	if ctx.Err() == context.DeadlineExceeded {
		log.Event(ctx, "context deadline exceeded", log.ERROR, log.Error(ctx.Err()))
		return ctx.Err()
	}

	if !gracefulShutdown {
		err = errors.New("failed to shutdown gracefully")
		log.Event(ctx, "failed to shutdown gracefully ", log.ERROR, log.Error(err))
		return err
	}

	log.Event(ctx, "graceful shutdown complete", log.INFO)
	return nil
}

func registerCheckers(ctx context.Context, clients routes.Clients) (err error) {

	hasErrors := false

	if err = clients.Healthcheck.AddCheck("frontend renderer", clients.Renderer.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "failed to add frontend renderer checker", log.ERROR, log.Error(err))
	}

	if err = clients.Healthcheck.AddCheck("filter API", clients.Filter.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "failed to add filter API checker", log.ERROR, log.Error(err))
	}

	if err = clients.Healthcheck.AddCheck("dataste API", clients.Dataset.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "failed to add dataset API checker", log.ERROR, log.Error(err))
	}

	if err = clients.Healthcheck.AddCheck("hierarchy API", clients.Hierarchy.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "failed to add hierarchy API checker", log.ERROR, log.Error(err))
	}

	if err = clients.Healthcheck.AddCheck("search API", clients.Search.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "failed to add search API checker", log.ERROR, log.Error(err))
	}

	if hasErrors {
		return errors.New("Error(s) registering checkers for healthcheck")
	}
	return nil
}
